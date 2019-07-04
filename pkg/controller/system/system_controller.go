package system

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"text/template"
	"time"

	"k8s.io/apimachinery/pkg/labels"

	dockerref "github.com/docker/distribution/reference"
	"github.com/go-logr/logr"
	semver "github.com/hashicorp/go-version"
	nbv1 "github.com/noobaa/noobaa-operator/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/pkg/nb"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	// ContainerImageOrg is the org of the default image url
	ContainerImageOrg = "noobaa"
	// ContainerImageRepo is the repo of the default image url
	ContainerImageRepo = "noobaa-core"
	// ContainerImageTag is the tag of the default image url
	ContainerImageTag = "4.0"
	// ContainerImageConstraintSemver is the contraints of supported image versions
	ContainerImageConstraintSemver = ">=4, <5"
	// ContainerImageName is the default image name without the tag/version
	ContainerImageName = ContainerImageOrg + "/" + ContainerImageRepo
	// ContainerImage is the full default image url
	ContainerImage = ContainerImageName + ":" + ContainerImageTag

	// AdminAccountEmail is the default email used for admin account
	AdminAccountEmail = "admin@noobaa.io"
)

var (
	// ContainerImageConstraint is the instantiated semver contraints used for image verification
	ContainerImageConstraint, _ = semver.NewConstraint(ContainerImageConstraintSemver)
	// SystemType is and empty system struct used for passing the object type
	SystemType = &nbv1.System{}

	logger = logf.Log.WithName("noobaa").WithName("system")
)

// Add creates a System Controller and adds it to the Manager.
// The Manager will set fields on the Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {

	// Create a new controller
	c, err := controller.New("system-controller", mgr, controller.Options{
		MaxConcurrentReconciles: 1,
		Reconciler: reconcile.Func(
			func(req reconcile.Request) (reconcile.Result, error) {
				return NewSystem(mgr, req).Reconcile()
			}),
	})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource System
	err = c.Watch(&source.Kind{Type: SystemType}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resources and requeue the owner System
	ownerHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    SystemType,
	}
	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, ownerHandler)
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, ownerHandler)
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, ownerHandler)
	if err != nil {
		return err
	}

	return nil
}

// System is the context for reconciling a noobaa system.
// It is created per every reconcile call and keeps state between calls of the reconcile flow.
type System struct {
	Mgr      manager.Manager
	Request  reconcile.Request
	Ctx      context.Context
	Logger   logr.Logger
	Recorder record.EventRecorder

	System      *nbv1.System
	CoreApp     *appsv1.StatefulSet
	ServiceMgmt *corev1.Service
	ServiceS3   *corev1.Service
	SecretOp    *corev1.Secret
	SecretAdmin *corev1.Secret
	NBClient    nb.Client
}

// NewSystem initializes a new system reconciler
func NewSystem(mgr manager.Manager, req reconcile.Request) *System {
	return &System{
		Mgr:      mgr,
		Request:  req,
		Ctx:      context.TODO(),
		Logger:   logger.WithValues("system", req.Namespace+"/"+req.Name),
		Recorder: mgr.GetRecorder("noobaa-operator"),
		System:   &nbv1.System{},
	}
}

// Reconcile reads that state of the cluster for a System object,
// and makes changes based on the state read and what is in the System.Spec.
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (s *System) Reconcile() (reconcile.Result, error) {

	log := s.Logger.WithName("Reconcile")
	log.Info("Start ...")

	err := s.Mgr.GetClient().Get(s.Ctx, s.Request.NamespacedName, s.System)
	if errors.IsNotFound(err) {
		log.Info("Ignoring request on deleted system")
		return reconcile.Result{}, nil
	}
	if err != nil {
		log.Error(err, "Failed getting system")
		return reconcile.Result{}, err
	}

	err = CombineErrors(
		s.ReconcileSystem(),
		s.UpdateSystemStatus(),
	)
	if err == nil {
		log.Info("✅ Done")
		return reconcile.Result{}, nil
	}
	if !IsPersistentError(err) {
		log.Error(err, "⏳ Temporary Error")
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}
	log.Error(err, "❌ Persistent Error")
	return reconcile.Result{}, nil
}

// ReconcileSystem runs the reconcile flow and populates System.Status.
func (s *System) ReconcileSystem() error {

	s.System.Status.Phase = nbv1.SystemPhaseVerifying

	if err := s.VerifySpecImage(); err != nil {
		return err
	}

	s.System.Status.Phase = nbv1.SystemPhaseCreating

	if err := CombineErrors(
		s.ReconcileServiceMgmt(),
		s.ReconcileServiceS3(),
		s.ReconcileCoreApp(),
		s.UpdateServiceStatus(s.ServiceMgmt, &s.System.Status.Services.ServiceMgmt, "mgmt-https"),
		s.UpdateServiceStatus(s.ServiceS3, &s.System.Status.Services.ServiceS3, "s3-https"),
	); err != nil {
		return err
	}

	s.System.Status.Phase = nbv1.SystemPhaseConfiguring

	if err := s.SetupNooBaaClient(); err != nil {
		return err
	}

	if err := s.SetupNooBaaSystem(); err != nil {
		return err
	}

	if err := s.SetupAdminAccount(); err != nil {
		return err
	}

	s.System.Status.Phase = nbv1.SystemPhaseReady

	return s.Complete()
}

// UpdateSystemStatus will update the System.Status field to match the observed status
func (s *System) UpdateSystemStatus() error {
	log := s.Logger.WithName("UpdateSystemStatus")
	log.Info("Updating system status")
	s.System.Status.ObservedGeneration = s.System.Generation
	return s.Mgr.GetClient().Status().Update(s.Ctx, s.System)
}

// VerifySpecImage checks the System.Spec.Image property,
// and sets System.Status.ActualImage
func (s *System) VerifySpecImage() error {

	log := s.Logger.WithName("VerifySpecImage")

	specImage := ContainerImage
	if s.System.Spec.Image != "" {
		specImage = s.System.Spec.Image
	}

	// Parse the image spec as a docker image url
	imageRef, err := dockerref.Parse(specImage)

	// If the image cannot be parsed log the incident and mark as persistent error
	// since we don't need to retry until the spec is updated.
	if err != nil {
		log.Error(err, "Invalid image", "image", specImage)
		s.Recorder.Eventf(s.System, corev1.EventTypeWarning,
			"BadImage", `Invalid image requested "%s"`, specImage)
		s.System.Status.Phase = nbv1.SystemPhaseRejected
		return NewPersistentError(err)
	}

	// Get the image name and tag
	imageName := ""
	imageTag := ""
	switch image := imageRef.(type) {
	case dockerref.NamedTagged:
		log.Info("Parsed image (NamedTagged)", "image", image)
		imageName = image.Name()
		imageTag = image.Tag()
	case dockerref.Tagged:
		log.Info("Parsed image (Tagged)", "image", image)
		imageTag = image.Tag()
	case dockerref.Named:
		log.Info("Parsed image (Named)", "image", image)
		imageName = image.Name()
	default:
		log.Info("Parsed image (unstructured)", "image", image)
	}

	if imageName == ContainerImageName {
		version, err := semver.NewVersion(imageTag)
		if err == nil {
			log.Info("Parsed version from image tag", "tag", imageTag, "version", version)
			if !ContainerImageConstraint.Check(version) {
				log.Error(nil, "Unsupported image version",
					"image", imageRef, "contraints", ContainerImageConstraint)
				s.Recorder.Eventf(s.System, corev1.EventTypeWarning,
					"BadImage", `Unsupported image version requested "%s" not matching constraints "%s"`,
					imageRef, ContainerImageConstraint)
				s.System.Status.Phase = nbv1.SystemPhaseRejected
				return NewPersistentError(fmt.Errorf(`Unsupported image version "%+v"`, imageRef))
			}
		} else {
			log.Info("Using custom image version", "image", imageRef, "contraints", ContainerImageConstraint)
			s.Recorder.Eventf(s.System, corev1.EventTypeNormal,
				"CustomImage", `Custom image version requested "%s", I hope you know what you're doing ...`, imageRef)
		}
	} else {
		log.Info("Using custom image name", "image", imageRef, "default", ContainerImageName)
		s.Recorder.Eventf(s.System, corev1.EventTypeNormal,
			"CustomImage", `Custom image requested "%s", I hope you know what you're doing ...`, imageRef)
	}

	// Set ActualImage to be updated in the system status
	s.System.Status.ActualImage = specImage
	return nil
}

// ReconcileObject is a generic call to reconcile a kubernetes object
// desiredFunc can be passed to modify the object before create/update.
// Currently we ignore enforcing a desired state, but it might be needed on upgrades.
func (s *System) ReconcileObject(obj metav1.Object, desiredFunc func()) error {

	log := s.Logger.WithName("ReconcileObject")

	s.Own(obj)

	op, err := controllerutil.CreateOrUpdate(
		s.Ctx, s.Mgr.GetClient(), obj.(runtime.Object),
		func(obj runtime.Object) error {
			if desiredFunc != nil {
				desiredFunc()
			}
			return nil
		},
	)
	if err != nil {
		log.Error(err, "Failed")
		return err
	}

	log.Info("Done", "op", op)
	return nil
}

// ReconcileServiceMgmt create or update the mgmt service.
func (s *System) ReconcileServiceMgmt() error {

	s.ServiceMgmt = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.Request.Namespace,
			Name:      s.Request.Name + "-mgmt",
			Labels:    map[string]string{"app": "noobaa"},
			Annotations: map[string]string{
				"prometheus.io/scrape": "true",
				"prometheus.io/scheme": "http",
				"prometheus.io/port":   "8080",
			},
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeLoadBalancer,
			Selector: map[string]string{"noobaa-mgmt": s.Request.Name},
			Ports: []corev1.ServicePort{
				{Port: 8080, Name: "mgmt"},
				{Port: 8443, Name: "mgmt-https"},
				{Port: 8444, Name: "md-https"},
				{Port: 8445, Name: "bg-https"},
				{Port: 8446, Name: "hosted-agents-https"},
				{Port: 80, TargetPort: intstr.FromInt(6001), Name: "s3"},
				{Port: 443, TargetPort: intstr.FromInt(6443), Name: "s3-https"},
			},
		},
	}
	return s.ReconcileObject(s.ServiceMgmt, nil)
}

// ReconcileServiceS3 create or update the s3 service.
func (s *System) ReconcileServiceS3() error {

	s.ServiceS3 = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.Request.Namespace,
			Name:      "s3", // TODO: handle collision in namespace
			Labels:    map[string]string{"app": "noobaa"},
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeLoadBalancer,
			Selector: map[string]string{"noobaa-s3": s.Request.Name},
			Ports: []corev1.ServicePort{
				{Port: 80, TargetPort: intstr.FromInt(6001), Name: "s3"},
				{Port: 443, TargetPort: intstr.FromInt(6443), Name: "s3-https"},
			},
		},
	}
	return s.ReconcileObject(s.ServiceS3, nil)
}

// ReconcileCoreApp create or update the core statefulset.
func (s *System) ReconcileCoreApp() error {

	ns := s.Request.Namespace
	name := s.Request.Name
	coreAppName := name + "-core"
	serviceAccountName := "noobaa-operator" // TODO do we use the same SA?
	image := s.System.Status.ActualImage
	replicas := int32(1)

	s.CoreApp = &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      coreAppName,
			Labels:    map[string]string{"app": "noobaa"},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"noobaa-core": name},
			},
			ServiceName: s.ServiceMgmt.Name,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: ns,
					Name:      coreAppName,
					Labels: map[string]string{
						"app":         "noobaa",
						"noobaa-core": name,
						"noobaa-mgmt": name,
						"noobaa-s3":   name,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: serviceAccountName,
					Containers: []corev1.Container{{
						Name:            coreAppName,
						Image:           image,
						ImagePullPolicy: corev1.PullIfNotPresent,
						VolumeMounts: []corev1.VolumeMount{
							{MountPath: "/data", Name: "datadir"},
							{MountPath: "/log", Name: "logdir"},
						},
						Env: []corev1.EnvVar{
							{Name: "CONTAINER_PLATFORM", Value: "KUBERNETES"},
						},
						Ports: []corev1.ContainerPort{
							{ContainerPort: 80},
							{ContainerPort: 443},
							{ContainerPort: 8080},
							{ContainerPort: 8443},
							{ContainerPort: 8444},
							{ContainerPort: 8445},
							{ContainerPort: 8446},
							{ContainerPort: 60100},
						},
						// # https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#container-probes
						// # ready when s3 port is open
						ReadinessProbe: &corev1.Probe{
							TimeoutSeconds: 5,
							Handler: corev1.Handler{
								TCPSocket: &corev1.TCPSocketAction{
									Port: intstr.FromInt(6001),
								},
							},
						},
						// # https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("500m"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("4"),
								corev1.ResourceMemory: resource.MustParse("8Gi"),
							},
						},
					}},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: ns,
						Name:      "datadir",
						Labels:    map[string]string{"app": "noobaa"},
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("50Gi"),
							},
						},
					},
				}, {
					ObjectMeta: metav1.ObjectMeta{
						Namespace: ns,
						Name:      "logdir",
						Labels:    map[string]string{"app": "noobaa"},
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("10Gi"),
							},
						},
					},
				},
			},
		},
	}
	return s.ReconcileObject(s.CoreApp, func() {
		s.CoreApp.Spec.Template.Spec.Containers[0].Image = image
	})
}

// UpdateServiceStatus populates the status of a service by detecting all of its addresses
func (s *System) UpdateServiceStatus(srv *corev1.Service, status *nbv1.SystemServiceStatus, portName string) error {

	log := s.Logger.WithName("UpdateServiceStatus").WithValues("service", srv.Name)
	*status = nbv1.SystemServiceStatus{}
	servicePort := nb.FindPortByName(srv, portName)
	proto := "http"
	if strings.HasSuffix(portName, "https") {
		proto = "https"
	}

	// Node IP:Port

	nodes := corev1.NodeList{}
	nodesListOptions := &client.ListOptions{}
	err := s.Mgr.GetClient().List(s.Ctx, nodesListOptions, &nodes)
	if err == nil {
		for _, node := range nodes.Items {
			for _, addr := range node.Status.Addresses {
				switch addr.Type {
				case corev1.NodeHostName:
					break // currently ignoring
				case corev1.NodeExternalIP:
					fallthrough
				case corev1.NodeExternalDNS:
					fallthrough
				case corev1.NodeInternalIP:
					fallthrough
				case corev1.NodeInternalDNS:
					// log.Info("Adding NodePorts", "addr", addr)
					status.NodePorts = append(
						status.NodePorts,
						fmt.Sprintf("%s://%s:%d", proto, addr.Address, servicePort.NodePort),
					)
				}
			}
		}
	}

	// Pod IP:Port

	pods := corev1.PodList{}
	podsListOptions := &client.ListOptions{
		Namespace:     s.Request.Namespace,
		LabelSelector: labels.SelectorFromSet(srv.Spec.Selector),
	}
	err = s.Mgr.GetClient().List(s.Ctx, podsListOptions, &pods)
	if err == nil {
		for _, pod := range pods.Items {
			if pod.Status.PodIP != "" && pod.Status.Phase == corev1.PodRunning {
				status.PodPorts = append(
					status.PodPorts,
					fmt.Sprintf("%s://%s:%s", proto, pod.Status.PodIP, servicePort.TargetPort.String()),
				)
			}
		}
	}

	// Cluster IP:Port (of the service)

	if srv.Spec.ClusterIP != "" {
		status.InternalIP = append(
			status.InternalIP,
			fmt.Sprintf("%s://%s:%d", proto, srv.Spec.ClusterIP, servicePort.Port),
		)
		status.InternalDNS = append(
			status.InternalDNS,
			fmt.Sprintf("%s://%s.%s:%d", proto, srv.Name, srv.Namespace, servicePort.Port),
		)
	}

	// LoadBalancer IP:Port (of the service)

	if srv.Status.LoadBalancer.Ingress != nil {
		for _, lb := range srv.Status.LoadBalancer.Ingress {
			if lb.IP != "" {
				status.ExternalIP = append(
					status.ExternalIP,
					fmt.Sprintf("%s://%s:%d", proto, lb.IP, servicePort.Port),
				)
			}
			if lb.Hostname != "" {
				status.ExternalDNS = append(
					status.ExternalDNS,
					fmt.Sprintf("%s://%s:%d", proto, lb.Hostname, servicePort.Port),
				)
			}
		}
	}

	// External IP:Port (of the service)

	if srv.Spec.ExternalIPs != nil {
		for _, ip := range srv.Spec.ExternalIPs {
			status.ExternalIP = append(
				status.ExternalIP,
				fmt.Sprintf("%s://%s:%d", proto, ip, servicePort.Port),
			)
		}
	}

	log.Info("Collected addresses", "status", status)
	return nil
}

// SetupNooBaaClient initializes the noobaa client for making calls to the server.
func (s *System) SetupNooBaaClient() error {

	// log := s.Logger.WithName("SetupNooBaaClient")

	if len(s.System.Status.Services.ServiceMgmt.PodPorts) == 0 ||
		len(s.System.Status.Services.ServiceMgmt.NodePorts) == 0 {
		return fmt.Errorf("core pod not ready yet")
	}

	if true {
		nodePort := s.System.Status.Services.ServiceMgmt.NodePorts[0]
		nodeIP := nodePort[strings.Index(nodePort, "://")+3 : strings.LastIndex(nodePort, ":")]
		s.NBClient = nb.NewClient(&nb.APIRouterNodePort{
			ServiceMgmt: s.ServiceMgmt,
			NodeIP:      nodeIP,
		})
	} else {
		podPort := s.System.Status.Services.ServiceMgmt.PodPorts[0]
		podIP := podPort[strings.Index(podPort, "://")+3 : strings.LastIndex(podPort, ":")]
		s.NBClient = nb.NewClient(&nb.APIRouterPodPort{
			ServiceMgmt: s.ServiceMgmt,
			PodIP:       podIP,
		})
	}

	return nil
}

// SetupNooBaaSystem creates a new system in the noobaa server if not created yet.
func (s *System) SetupNooBaaSystem() error {

	log := s.Logger.WithName("SetupNooBaaSystem")
	ns := s.Request.Namespace
	name := s.Request.Name
	secretOpName := name + "-operator"

	s.SecretOp = &corev1.Secret{}
	err := s.GetObject(secretOpName, s.SecretOp)
	if err == nil {
		s.NBClient.SetAuthToken(string(s.SecretOp.Data["auth_token"]))
		return nil
	}
	if !errors.IsNotFound(err) {
		log.Error(err, "Failed getting operator secret")
		return err
	}

	randomBytes := make([]byte, 16)
	_, err = rand.Read(randomBytes)
	fatal(err)
	randomPassword := base64.StdEncoding.EncodeToString(randomBytes)

	res, err := s.NBClient.CreateSystemAPI(nb.CreateSystemParams{
		Name:     name,
		Email:    AdminAccountEmail,
		Password: randomPassword,
	})

	if err != nil {
		return err
	}
	s.NBClient.SetAuthToken(res.Token)

	s.SecretOp = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      secretOpName,
			Labels:    map[string]string{"app": "noobaa"},
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"system":     name,
			"email":      AdminAccountEmail,
			"password":   randomPassword,
			"auth_token": res.Token,
		},
	}

	s.Own(s.SecretOp)
	return s.Mgr.GetClient().Create(s.Ctx, s.SecretOp)
}

// SetupAdminAccount creates the admin secret
func (s *System) SetupAdminAccount() error {

	log := s.Logger.WithName("SetupAdminAccount")
	ns := s.Request.Namespace
	name := s.Request.Name
	secretAdminName := name + "-admin"

	s.SecretAdmin = &corev1.Secret{}
	err := s.GetObject(secretAdminName, s.SecretAdmin)
	if err == nil {
		return nil
	}
	if !errors.IsNotFound(err) {
		log.Error(err, "Failed getting admin secret")
		return err
	}

	s.SecretAdmin = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      secretAdminName,
			Labels:    map[string]string{"app": "noobaa"},
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"system":   name,
			"email":    AdminAccountEmail,
			"password": string(s.SecretOp.Data["password"]),
		},
	}

	log.Info("listing accounts")
	res, err := s.NBClient.ListAccountsAPI()
	if err != nil {
		return err
	}
	for _, account := range res.Accounts {
		if account.Email == AdminAccountEmail {
			if len(account.AccessKeys) > 0 {
				s.SecretAdmin.StringData["AWS_ACCESS_KEY_ID"] = account.AccessKeys[0].AccessKey
				s.SecretAdmin.StringData["AWS_SECRET_ACCESS_KEY"] = account.AccessKeys[0].SecretKey
			}
		}
	}

	s.Own(s.SecretAdmin)
	return s.Mgr.GetClient().Create(s.Ctx, s.SecretAdmin)
}

var readmeTemplate = template.Must(template.New("NooBaaSystem.Status.Readme").Parse(`

	Welcome to NooBaa!
	-----------------

	Lets get started:

	1. Connect to Management console:

		Read your mgmt console login information (email & password) from secret: "{{.SecretAdmin.Name}}".

			kubectl get secret {{.SecretAdmin.Name}} -n {{.SecretAdmin.Namespace}} -o json | jq '.data|map_values(@base64d)'

		Open the management console service - take External IP/DNS or Node Port or use port forwarding:

			kubectl port-forward -n {{.ServiceMgmt.Namespace}} service/{{.ServiceMgmt.Name}} 11443:8443 &
			open https://localhost:11443

	2. Test S3 client:

		kubectl port-forward -n {{.ServiceS3.Namespace}} service/{{.ServiceS3.Name}} 10443:443 &
		NOOBAA_ACCESS_KEY=$(kubectl get secret {{.SecretAdmin.Name}} -n {{.SecretAdmin.Namespace}} -o json | jq -r '.data.AWS_ACCESS_KEY_ID|@base64d')
		NOOBAA_SECRET_KEY=$(kubectl get secret {{.SecretAdmin.Name}} -n {{.SecretAdmin.Namespace}} -o json | jq -r '.data.AWS_SECRET_ACCESS_KEY|@base64d')
		alias s3='AWS_ACCESS_KEY_ID=$NOOBAA_ACCESS_KEY AWS_SECRET_ACCESS_KEY=$NOOBAA_SECRET_KEY aws --endpoint https://localhost:10443 --no-verify-ssl s3'
		s3 ls

`))

// Complete populates the system status at the end of reconcile.
func (s *System) Complete() error {

	var readmeBuffer bytes.Buffer
	err := readmeTemplate.Execute(&readmeBuffer, s)
	if err != nil {
		return err
	}
	s.System.Status.Readme = readmeBuffer.String()
	s.System.Status.Accounts.Admin.SecretRef.Name = s.SecretAdmin.Name
	s.System.Status.Accounts.Admin.SecretRef.Namespace = s.SecretAdmin.Namespace
	return nil
}

// GetObject gets an object by name from the request namespace.
func (s *System) GetObject(name string, obj runtime.Object) error {
	return s.Mgr.GetClient().Get(s.Ctx, client.ObjectKey{Namespace: s.Request.Namespace, Name: name}, obj)
}

// Own marks the object as owned by the system controller,
// so that it will be garbage collected once the system is deleted.
func (s *System) Own(obj metav1.Object) {
	fatal(controllerutil.SetControllerReference(s.System, obj, s.Mgr.GetScheme()))
}

func fatal(err error) {
	if err != nil {
		logger.Error(err, "PANIC")
		panic(err)
	}
}

// PersistentError is an error type that tells the reconcile to avoid requeueing.
type PersistentError struct {
	E error
}

// Error function makes PersistentError implement error interface
func (e *PersistentError) Error() string { return e.E.Error() }

// assert implement error interface
var _ error = &PersistentError{}

// NewPersistentError returns a new persistent error.
func NewPersistentError(err error) *PersistentError {
	if err == nil {
		panic("NewPersistentError expects non nil error")
	}
	return &PersistentError{E: err}
}

// IsPersistentError checks if the provided error is persistent.
func IsPersistentError(err error) bool {
	_, persistent := err.(*PersistentError)
	return persistent
}

// CombineErrors takes a list of errors and combines them to one.
// Generally it will return the first non-nil error,
// but if a persistent error is found it will be returned
// instead of non-persistent errors.
func CombineErrors(errs ...error) error {
	combined := error(nil)
	for _, err := range errs {
		if err == nil {
			continue
		}
		if combined == nil {
			combined = err
			continue
		}
		if IsPersistentError(err) && !IsPersistentError(combined) {
			combined = err
		}
	}
	return combined
}
