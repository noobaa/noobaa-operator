package system

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/noobaa/noobaa-operator/build/_output/bundle"
	nbv1 "github.com/noobaa/noobaa-operator/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/pkg/nb"
	"github.com/noobaa/noobaa-operator/pkg/options"
	"github.com/noobaa/noobaa-operator/pkg/util"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	dockerref "github.com/docker/distribution/reference"
	semver "github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	// ContainerImageConstraint is the instantiated semver contraints used for image verification
	ContainerImageConstraint, _ = semver.NewConstraint(options.ContainerImageConstraintSemver)

	// NooBaaType is and empty noobaa struct used for passing the object type
	NooBaaType = &nbv1.NooBaa{}
)

// Reconciler is the context for loading or reconciling a noobaa system
type Reconciler struct {
	Request  types.NamespacedName
	Client   client.Client
	Scheme   *runtime.Scheme
	Ctx      context.Context
	Logger   *logrus.Entry
	Recorder record.EventRecorder
	NBClient nb.Client

	NooBaa       *nbv1.NooBaa
	CoreApp      *appsv1.StatefulSet
	ServiceMgmt  *corev1.Service
	ServiceS3    *corev1.Service
	SecretServer *corev1.Secret
	SecretOp     *corev1.Secret
	SecretAdmin  *corev1.Secret
}

// NewReconciler initializes a reconciler to be used for loading or reconciling a noobaa system
func NewReconciler(
	req types.NamespacedName,
	client client.Client,
	scheme *runtime.Scheme,
	recorder record.EventRecorder,
) *Reconciler {

	r := &Reconciler{
		Request:      req,
		Client:       client,
		Scheme:       scheme,
		Recorder:     recorder,
		Ctx:          context.TODO(),
		Logger:       logrus.WithFields(logrus.Fields{"ns": req.Namespace, "sys": req.Name}),
		NooBaa:       util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_noobaa_cr_yaml).(*nbv1.NooBaa),
		CoreApp:      util.KubeObject(bundle.File_deploy_internal_statefulset_core_yaml).(*appsv1.StatefulSet),
		ServiceMgmt:  util.KubeObject(bundle.File_deploy_internal_service_mgmt_yaml).(*corev1.Service),
		ServiceS3:    util.KubeObject(bundle.File_deploy_internal_service_s3_yaml).(*corev1.Service),
		SecretServer: util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret),
		SecretOp:     util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret),
		SecretAdmin:  util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret),
	}
	util.SecretResetStringDataFromData(r.SecretServer)
	util.SecretResetStringDataFromData(r.SecretOp)
	util.SecretResetStringDataFromData(r.SecretAdmin)

	// Set Namespace
	r.NooBaa.Namespace = r.Request.Namespace
	r.CoreApp.Namespace = r.Request.Namespace
	r.ServiceMgmt.Namespace = r.Request.Namespace
	r.ServiceS3.Namespace = r.Request.Namespace
	r.SecretServer.Namespace = r.Request.Namespace
	r.SecretOp.Namespace = r.Request.Namespace
	r.SecretAdmin.Namespace = r.Request.Namespace

	// Set Names
	r.NooBaa.Name = r.Request.Name
	r.CoreApp.Name = r.Request.Name + "-core"
	r.ServiceMgmt.Name = r.Request.Name + "-mgmt"
	r.ServiceS3.Name = "s3" // TODO: handle collision in namespace
	r.SecretServer.Name = r.Request.Name + "-server"
	r.SecretOp.Name = r.Request.Name + "-operator"
	r.SecretAdmin.Name = r.Request.Name + "-admin"

	return r
}

// Load reads the state of the kubernetes objects of the system
func (r *Reconciler) Load() {
	util.KubeCheck(r.NooBaa)
	util.KubeCheck(r.CoreApp)
	util.KubeCheck(r.ServiceMgmt)
	util.KubeCheck(r.ServiceS3)
	util.KubeCheck(r.SecretServer)
	util.KubeCheck(r.SecretOp)
	util.KubeCheck(r.SecretAdmin)
	util.SecretResetStringDataFromData(r.SecretServer)
	util.SecretResetStringDataFromData(r.SecretOp)
	util.SecretResetStringDataFromData(r.SecretAdmin)
}

// Reconcile reads that state of the cluster for a System object,
// and makes changes based on the state read and what is in the System.Spec.
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *Reconciler) Reconcile() (reconcile.Result, error) {

	log := r.Logger.WithField("func", "Reconcile")
	log.Infof("Start ...")

	util.KubeCheck(r.NooBaa)
	if r.NooBaa.UID == "" {
		log.Infof("NooBaa not found or already deleted. Skip reconcile.")
		return reconcile.Result{}, nil
	}

	err := util.CombineErrors(
		r.RunReconcile(),
		r.UpdateStatus(),
	)
	if util.IsPersistentError(err) {
		log.Errorf("❌ Persistent Error: %s", err)
		return reconcile.Result{}, nil
	}
	if err != nil {
		log.Warnf("⏳ Temporary Error: %s", err)
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}
	log.Infof("✅ Done")
	return reconcile.Result{}, nil
}

// UpdateStatus updates the system status in kubernetes from the memory
func (r *Reconciler) UpdateStatus() error {
	log := r.Logger.WithField("func", "UpdateStatus")
	log.Infof("Updating noobaa status")
	r.NooBaa.Status.ObservedGeneration = r.NooBaa.Generation
	return r.Client.Status().Update(r.Ctx, r.NooBaa)
}

// RunReconcile runs the reconcile flow and populates System.Status.
func (r *Reconciler) RunReconcile() error {

	r.SetPhase(nbv1.SystemPhaseVerifying)

	if err := r.CheckSpecImage(); err != nil {
		return err
	}

	r.SetPhase(nbv1.SystemPhaseCreating)

	if err := r.ReconcileSecretServer(); err != nil {
		r.setErrorCondition(err)
		return err
	}
	if err := r.ReconcileObject(r.CoreApp, r.SetDesiredCoreApp); err != nil {
		r.setErrorCondition(err)
		return err
	}
	if err := r.ReconcileObject(r.ServiceMgmt, r.SetDesiredServiceMgmt); err != nil {
		r.setErrorCondition(err)
		return err
	}
	if err := r.ReconcileObject(r.ServiceS3, r.SetDesiredServiceS3); err != nil {
		r.setErrorCondition(err)
		return err
	}

	r.SetPhase(nbv1.SystemPhaseConnecting)

	if err := r.Connect(); err != nil {
		r.setErrorCondition(err)
		return err
	}

	r.SetPhase(nbv1.SystemPhaseConfiguring)

	if err := r.ReconcileSecretOp(); err != nil {
		r.setErrorCondition(err)
		return err
	}

	if err := r.ReconcileSecretAdmin(); err != nil {
		r.setErrorCondition(err)
		return err
	}

	r.SetPhase(nbv1.SystemPhaseReady)

	return r.Complete()
}

// ReconcileSecretServer creates a secret needed for the server pod
func (r *Reconciler) ReconcileSecretServer() error {
	util.KubeCheck(r.SecretServer)
	util.SecretResetStringDataFromData(r.SecretServer)

	if r.SecretServer.StringData["jwt"] == "" {
		r.SecretServer.StringData["jwt"] = util.RandomBase64(16)
	}
	if r.SecretServer.StringData["server_secret"] == "" {
		r.SecretServer.StringData["server_secret"] = util.RandomHex(4)
	}
	r.Own(r.SecretServer)
	util.KubeCreateSkipExisting(r.SecretServer)
	return nil
}

// SetDesiredCoreApp updates the CoreApp as desired for reconciling
func (r *Reconciler) SetDesiredCoreApp() {
	r.CoreApp.Spec.Template.Labels["noobaa-core"] = r.Request.Name
	r.CoreApp.Spec.Template.Labels["noobaa-mgmt"] = r.Request.Name
	r.CoreApp.Spec.Template.Labels["noobaa-s3"] = r.Request.Name
	r.CoreApp.Spec.Selector.MatchLabels["noobaa-core"] = r.Request.Name
	r.CoreApp.Spec.ServiceName = r.ServiceMgmt.Name

	podSpec := &r.CoreApp.Spec.Template.Spec
	podSpec.ServiceAccountName = "noobaa-operator" // TODO do we use the same SA?
	for i := range podSpec.InitContainers {
		c := &podSpec.InitContainers[i]
		if c.Image == "NOOBAA_IMAGE" {
			c.Image = r.NooBaa.Status.ActualImage
		}
	}
	for i := range podSpec.Containers {
		c := &podSpec.Containers[i]
		if c.Image == "NOOBAA_IMAGE" {
			c.Image = r.NooBaa.Status.ActualImage
			for j := range c.Env {
				if c.Env[j].Value == "NOOBAA_IMAGE" {
					c.Env[j].Value = r.NooBaa.Status.ActualImage
				}
			}
		} else if c.Image == "MONGO_IMAGE" {
			if r.NooBaa.Spec.MongoImage == nil {
				c.Image = options.MongoImage
			} else {
				c.Image = *r.NooBaa.Spec.MongoImage
			}
		}
	}
	if r.NooBaa.Spec.ImagePullSecret == nil {
		podSpec.ImagePullSecrets =
			[]corev1.LocalObjectReference{}
	} else {
		podSpec.ImagePullSecrets =
			[]corev1.LocalObjectReference{*r.NooBaa.Spec.ImagePullSecret}
	}
	for i := range r.CoreApp.Spec.VolumeClaimTemplates {
		pvc := &r.CoreApp.Spec.VolumeClaimTemplates[i]
		pvc.Spec.StorageClassName = r.NooBaa.Spec.StorageClassName

		// TODO we want to own the PVC's by NooBaa system but get errors on openshift:
		//   Warning  FailedCreate  56s  statefulset-controller
		//   create Pod noobaa-core-0 in StatefulSet noobaa-core failed error:
		//   Failed to create PVC mongo-datadir-noobaa-core-0:
		//   persistentvolumeclaims "mongo-datadir-noobaa-core-0" is forbidden:
		//   cannot set blockOwnerDeletion if an ownerReference refers to a resource
		//   you can't set finalizers on: , <nil>, ...

		// r.Own(pvc)
	}
}

// SetDesiredServiceMgmt updates the ServiceMgmt as desired for reconciling
func (r *Reconciler) SetDesiredServiceMgmt() {
	r.ServiceMgmt.Spec.Selector["noobaa-mgmt"] = r.Request.Name
}

// SetDesiredServiceS3 updates the ServiceS3 as desired for reconciling
func (r *Reconciler) SetDesiredServiceS3() {
	r.ServiceS3.Spec.Selector["noobaa-s3"] = r.Request.Name
}

// CheckSpecImage checks the System.Spec.Image property,
// and sets System.Status.ActualImage
func (r *Reconciler) CheckSpecImage() error {

	log := r.Logger.WithField("func", "CheckSpecImage")

	specImage := options.ContainerImage
	if r.NooBaa.Spec.Image != nil {
		specImage = *r.NooBaa.Spec.Image
	}

	// Parse the image spec as a docker image url
	imageRef, err := dockerref.Parse(specImage)

	// If the image cannot be parsed log the incident and mark as persistent error
	// since we don't need to retry until the spec is updated.
	if err != nil {
		log.Errorf("Invalid image %s: %s", specImage, err)
		if r.Recorder != nil {
			r.Recorder.Eventf(r.NooBaa, corev1.EventTypeWarning,
				"BadImage", `Invalid image requested %q`, specImage)
		}
		r.SetPhase(nbv1.SystemPhaseRejected)
		return util.NewPersistentError(err)
	}

	// Get the image name and tag
	imageName := ""
	imageTag := ""
	switch image := imageRef.(type) {
	case dockerref.NamedTagged:
		log.Infof("Parsed image (NamedTagged) %v", image)
		imageName = image.Name()
		imageTag = image.Tag()
	case dockerref.Tagged:
		log.Infof("Parsed image (Tagged) %v", image)
		imageTag = image.Tag()
	case dockerref.Named:
		log.Infof("Parsed image (Named) %v", image)
		imageName = image.Name()
	default:
		log.Infof("Parsed image (unstructured) %v", image)
	}

	if imageName == options.ContainerImageName {
		version, err := semver.NewVersion(imageTag)
		if err == nil {
			log.Infof("Parsed version %q from image tag %q", version.String(), imageTag)
			if !ContainerImageConstraint.Check(version) {
				log.Errorf("Unsupported image version %q for contraints %q",
					imageRef.String(), ContainerImageConstraint.String())
				if r.Recorder != nil {
					r.Recorder.Eventf(r.NooBaa, corev1.EventTypeWarning,
						"BadImage", `Unsupported image version requested %q not matching constraints %q`,
						imageRef, ContainerImageConstraint)
				}
				r.SetPhase(nbv1.SystemPhaseRejected)
				return util.NewPersistentError(fmt.Errorf(`Unsupported image version "%+v"`, imageRef))
			}
		} else {
			log.Infof("Using custom image %q contraints %q", imageRef.String(), ContainerImageConstraint.String())
			if r.Recorder != nil {
				r.Recorder.Eventf(r.NooBaa, corev1.EventTypeNormal,
					"CustomImage", `Custom image version requested %q, I hope you know what you're doing ...`, imageRef)
			}
		}
	} else {
		log.Infof("Using custom image name %q the default is %q", imageRef.String(), options.ContainerImageName)
		if r.Recorder != nil {
			r.Recorder.Eventf(r.NooBaa, corev1.EventTypeNormal,
				"CustomImage", `Custom image requested %q, I hope you know what you're doing ...`, imageRef)
		}
	}

	// Set ActualImage to be updated in the noobaa status
	r.NooBaa.Status.ActualImage = specImage
	return nil
}

// CheckServiceStatus populates the status of a service by detecting all of its addresses
func (r *Reconciler) CheckServiceStatus(srv *corev1.Service, status *nbv1.ServiceStatus, portName string) {

	log := r.Logger.WithField("func", "CheckServiceStatus").WithField("service", srv.Name)
	*status = nbv1.ServiceStatus{}
	servicePort := nb.FindPortByName(srv, portName)
	proto := "http"
	if strings.HasSuffix(portName, "https") {
		proto = "https"
	}

	// Node IP:Port
	// Pod IP:Port
	pods := corev1.PodList{}
	podsListOptions := &client.ListOptions{
		Namespace:     r.Request.Namespace,
		LabelSelector: labels.SelectorFromSet(srv.Spec.Selector),
	}
	err := r.Client.List(r.Ctx, podsListOptions, &pods)
	if err == nil {
		for _, pod := range pods.Items {
			if pod.Status.Phase == corev1.PodRunning {
				if pod.Status.HostIP != "" {
					status.NodePorts = append(
						status.NodePorts,
						fmt.Sprintf("%s://%s:%d", proto, pod.Status.HostIP, servicePort.NodePort),
					)
				}
				if pod.Status.PodIP != "" {
					status.PodPorts = append(
						status.PodPorts,
						fmt.Sprintf("%s://%s:%s", proto, pod.Status.PodIP, servicePort.TargetPort.String()),
					)
				}
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

	log.Infof("Collected addresses: %+v", status)
}

// Connect initializes the noobaa client for making calls to the server.
func (r *Reconciler) Connect() error {

	r.CheckServiceStatus(r.ServiceMgmt, &r.NooBaa.Status.Services.ServiceMgmt, "mgmt-https")
	r.CheckServiceStatus(r.ServiceS3, &r.NooBaa.Status.Services.ServiceS3, "s3-https")

	if len(r.NooBaa.Status.Services.ServiceMgmt.NodePorts) == 0 {
		return fmt.Errorf("core pod port not ready yet")
	}

	nodePort := r.NooBaa.Status.Services.ServiceMgmt.NodePorts[0]
	nodeIP := nodePort[strings.Index(nodePort, "://")+3 : strings.LastIndex(nodePort, ":")]

	r.NBClient = nb.NewClient(&nb.APIRouterNodePort{
		ServiceMgmt: r.ServiceMgmt,
		NodeIP:      nodeIP,
	})

	r.NBClient.SetAuthToken(r.SecretOp.StringData["auth_token"])

	// Check that the server is indeed serving the API already
	// we use the read_auth call here because it's an API that always answers
	// even when auth_token is empty.
	_, err := r.NBClient.ReadAuthAPI()
	return err

	// if len(r.NooBaa.Status.Services.ServiceMgmt.PodPorts) != 0 {
	// 	podPort := r.NooBaa.Status.Services.ServiceMgmt.PodPorts[0]
	// 	podIP := podPort[strings.Index(podPort, "://")+3 : strings.LastIndex(podPort, ":")]
	// 	r.NBClient = nb.NewClient(&nb.APIRouterPodPort{
	// 		ServiceMgmt: r.ServiceMgmt,
	// 		PodIP:       podIP,
	// 	})
	// 	r.NBClient.SetAuthToken(r.SecretOp.StringData["auth_token"])
	// 	return nil
	// }

}

// ReconcileSecretOp creates a new system in the noobaa server if not created yet.
func (r *Reconciler) ReconcileSecretOp() error {

	// log := r.Logger.WithName("ReconcileSecretOp")
	util.KubeCheck(r.SecretOp)
	util.SecretResetStringDataFromData(r.SecretOp)

	if r.SecretOp.StringData["auth_token"] != "" {
		return nil
	}

	if r.SecretOp.StringData["email"] == "" {
		r.SecretOp.StringData["email"] = options.AdminAccountEmail
	}

	if r.SecretOp.StringData["password"] == "" {
		r.SecretOp.StringData["password"] = util.RandomBase64(16)
		r.Own(r.SecretOp)
		err := r.Client.Create(r.Ctx, r.SecretOp)
		if err != nil {
			return err
		}
	}

	res, err := r.NBClient.CreateAuthAPI(nb.CreateAuthParams{
		System:   r.Request.Name,
		Role:     "admin",
		Email:    r.SecretOp.StringData["email"],
		Password: r.SecretOp.StringData["password"],
	})
	if err == nil {
		// TODO this recovery flow does not allow us to get OperatorToken like CreateSystem
		r.SecretOp.StringData["auth_token"] = res.Token
	} else {
		res, err := r.NBClient.CreateSystemAPI(nb.CreateSystemParams{
			Name:     r.Request.Name,
			Email:    r.SecretOp.StringData["email"],
			Password: r.SecretOp.StringData["password"],
		})
		if err != nil {
			return err
		}
		// TODO use res.OperatorToken after https://github.com/noobaa/noobaa-core/issues/5635
		r.SecretOp.StringData["auth_token"] = res.Token
	}
	r.NBClient.SetAuthToken(r.SecretOp.StringData["auth_token"])
	return r.Client.Update(r.Ctx, r.SecretOp)
}

// ReconcileSecretAdmin creates the admin secret
func (r *Reconciler) ReconcileSecretAdmin() error {

	log := r.Logger.WithField("func", "ReconcileSecretAdmin")

	util.KubeCheck(r.SecretAdmin)
	util.SecretResetStringDataFromData(r.SecretAdmin)

	ns := r.Request.Namespace
	name := r.Request.Name
	secretAdminName := name + "-admin"

	r.SecretAdmin = &corev1.Secret{}
	err := r.GetObject(secretAdminName, r.SecretAdmin)
	if err == nil {
		return nil
	}
	if !errors.IsNotFound(err) {
		log.Errorf("Failed getting admin secret: %v", err)
		return err
	}

	r.SecretAdmin = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      secretAdminName,
			Labels:    map[string]string{"app": "noobaa"},
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"system":   name,
			"email":    options.AdminAccountEmail,
			"password": string(r.SecretOp.Data["password"]),
		},
	}

	log.Infof("listing accounts")
	res, err := r.NBClient.ListAccountsAPI()
	if err != nil {
		return err
	}
	for _, account := range res.Accounts {
		if account.Email == options.AdminAccountEmail {
			if len(account.AccessKeys) > 0 {
				r.SecretAdmin.StringData["AWS_ACCESS_KEY_ID"] = account.AccessKeys[0].AccessKey
				r.SecretAdmin.StringData["AWS_SECRET_ACCESS_KEY"] = account.AccessKeys[0].SecretKey
			}
		}
	}

	r.Own(r.SecretAdmin)
	return r.Client.Create(r.Ctx, r.SecretAdmin)
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

func (r *Reconciler) setErrorCondition(err error) {
	reason := "ReconcileFailed"
	message := fmt.Sprintf("Error while reconciling: %v", err)
	currentTime := metav1.NewTime(time.Now())
	conditionsv1.SetStatusCondition(&r.NooBaa.Status.Conditions, conditionsv1.Condition{
		//LastHeartbeatTime should be set by the custom-resource-status just like lastTransitionTime
		// Setting it here temporarity
		LastHeartbeatTime: currentTime,
		Type:   conditionsv1.ConditionAvailable,
		Status: corev1.ConditionUnknown,
		Reason: reason,
		Message: message,
	})
	conditionsv1.SetStatusCondition(&r.NooBaa.Status.Conditions, conditionsv1.Condition{
		LastHeartbeatTime: currentTime,
		Type:   conditionsv1.ConditionProgressing,
		Status: corev1.ConditionFalse,
		Reason: reason,
		Message: message,
	})
	conditionsv1.SetStatusCondition(&r.NooBaa.Status.Conditions, conditionsv1.Condition{
		LastHeartbeatTime: currentTime,
		Type:   conditionsv1.ConditionDegraded,
		Status: corev1.ConditionTrue,
		Reason: reason,
		Message: message,
	})
	conditionsv1.SetStatusCondition(&r.NooBaa.Status.Conditions, conditionsv1.Condition{
		LastHeartbeatTime: currentTime,
		Type:   conditionsv1.ConditionUpgradeable,
		Status: corev1.ConditionUnknown,
		Reason: reason,
		Message: message,
	})
}

func (r *Reconciler) setAvailableCondition(reason string, message string) {
	currentTime := metav1.NewTime(time.Now())
	conditionsv1.SetStatusCondition(&r.NooBaa.Status.Conditions, conditionsv1.Condition{
		LastHeartbeatTime: currentTime,
		Type:   conditionsv1.ConditionAvailable,
		Status: corev1.ConditionTrue,
		Reason: reason,
		Message: message,
	})
	conditionsv1.SetStatusCondition(&r.NooBaa.Status.Conditions, conditionsv1.Condition{
		LastHeartbeatTime: currentTime,
		Type:   conditionsv1.ConditionProgressing,
		Status: corev1.ConditionFalse,
		Reason: reason,
		Message: message,
	})
	conditionsv1.SetStatusCondition(&r.NooBaa.Status.Conditions, conditionsv1.Condition{
		LastHeartbeatTime: currentTime,
		Type:   conditionsv1.ConditionDegraded,
		Status: corev1.ConditionFalse,
		Reason: reason,
		Message: message,
	})
	conditionsv1.SetStatusCondition(&r.NooBaa.Status.Conditions, conditionsv1.Condition{
		LastHeartbeatTime: currentTime,
		Type:   conditionsv1.ConditionUpgradeable,
		Status: corev1.ConditionTrue,
		Reason: reason,
		Message: message,
	})
}

func (r *Reconciler) setProgressingCondition(reason string, message string) {
	currentTime := metav1.NewTime(time.Now())
	conditionsv1.SetStatusCondition(&r.NooBaa.Status.Conditions, conditionsv1.Condition{
		LastHeartbeatTime: currentTime,
		Type:   conditionsv1.ConditionAvailable,
		Status: corev1.ConditionFalse,
		Reason: reason,
		Message: message,
	})
	conditionsv1.SetStatusCondition(&r.NooBaa.Status.Conditions, conditionsv1.Condition{
		LastHeartbeatTime: currentTime,
		Type:   conditionsv1.ConditionProgressing,
		Status: corev1.ConditionTrue,
		Reason: reason,
		Message: message,
	})
	conditionsv1.SetStatusCondition(&r.NooBaa.Status.Conditions, conditionsv1.Condition{
		LastHeartbeatTime: currentTime,
		Type:   conditionsv1.ConditionDegraded,
		Status: corev1.ConditionFalse,
		Reason: reason,
		Message: message,
	})
	conditionsv1.SetStatusCondition(&r.NooBaa.Status.Conditions, conditionsv1.Condition{
		LastHeartbeatTime: currentTime,
		Type:   conditionsv1.ConditionUpgradeable,
		Status: corev1.ConditionFalse,
		Reason: reason,
		Message: message,
	})
}

// SetPhase updates the status phase and conditions
func (r *Reconciler) SetPhase(phase nbv1.SystemPhase) {
	r.Logger.Infof("SetPhase: %s", phase)
	r.NooBaa.Status.Phase = phase
	reason := fmt.Sprintf("%v", phase)
	message := fmt.Sprintf("%v", phase)
	switch phase {
		case nbv1.SystemPhaseVerifying:
			reason = "ReconcileInit"
			message = "Initializing noobaa cluster"
			r.setAvailableCondition(reason, message)
		case nbv1.SystemPhaseCreating:
			r.setProgressingCondition(reason, message)
		case nbv1.SystemPhaseConnecting:
			r.setProgressingCondition(reason, message)
		case nbv1.SystemPhaseConfiguring:
			r.setProgressingCondition(reason, message)
		case nbv1.SystemPhaseReady:
			reason = "Reconcilecompleted"
			message = "ReconcileCompleted"
			r.setAvailableCondition(reason, message)
		default:
	}
}

// Complete populates the noobaa status at the end of reconcile.
func (r *Reconciler) Complete() error {

	var readmeBuffer bytes.Buffer
	err := readmeTemplate.Execute(&readmeBuffer, r)
	if err != nil {
		return err
	}
	r.NooBaa.Status.Readme = readmeBuffer.String()
	r.NooBaa.Status.Accounts.Admin.SecretRef.Name = r.SecretAdmin.Name
	r.NooBaa.Status.Accounts.Admin.SecretRef.Namespace = r.SecretAdmin.Namespace
	return nil
}

// Own sets the object owner references to the noobaa system
func (r *Reconciler) Own(obj metav1.Object) {
	util.Panic(controllerutil.SetControllerReference(r.NooBaa, obj, r.Scheme))
}

// GetObject gets an object by name from the request namespace.
func (r *Reconciler) GetObject(name string, obj runtime.Object) error {
	return r.Client.Get(r.Ctx, client.ObjectKey{Namespace: r.Request.Namespace, Name: name}, obj)
}

// ReconcileObject is a generic call to reconcile a kubernetes object
// desiredFunc can be passed to modify the object before create/update.
// Currently we ignore enforcing a desired state, but it might be needed on upgrades.
func (r *Reconciler) ReconcileObject(obj runtime.Object, desiredFunc func()) error {

	kind := obj.GetObjectKind().GroupVersionKind().Kind
	objMeta, _ := meta.Accessor(obj)
	log := r.Logger.
		WithField("func", "ReconcileObject").
		WithField("kind", kind).
		WithField("name", objMeta.GetName())

	r.Own(objMeta)

	op, err := controllerutil.CreateOrUpdate(
		r.Ctx, r.Client, obj.(runtime.Object),
		func(obj runtime.Object) error {
			if desiredFunc != nil {
				desiredFunc()
			}
			return nil
		},
	)
	if err != nil {
		log.Errorf("ReconcileObject Failed: %v", err)
		return err
	}

	log.Infof("Done. op=%s", op)
	return nil
}
