package obc

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/pkg/nb"
	"github.com/noobaa/noobaa-operator/pkg/options"
	"github.com/noobaa/noobaa-operator/pkg/system"
	"github.com/noobaa/noobaa-operator/pkg/util"

	"github.com/kube-object-storage/lib-bucket-provisioner/pkg/provisioner"
	obAPI "github.com/kube-object-storage/lib-bucket-provisioner/pkg/provisioner/api"
	obErrors "github.com/kube-object-storage/lib-bucket-provisioner/pkg/provisioner/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	obStateAccountNameKey = "UserName"
	allNamespaces         = ""
)

var (
	namespace       = os.Getenv("WATCH_NAMESPACE")
	provisionerName = "noobaa.io/" + namespace + ".bucket"
	log             = util.Logger().WithField("mod", "OBC")
)

// RunProvisioner will run OBC provisioner
func RunProvisioner(client client.Client, scheme *runtime.Scheme, recorder record.EventRecorder) error {
	log.Info("OBC Provisioner - start..")

	config, clientset, err := createConfigAndClient()
	if err != nil {
		return err
	}

	p := Provisioner{
		clientset: clientset,
		client:    client,
		scheme:    scheme,
		recorder:  recorder,
	}

	// Create and run the s3 provisioner controller.
	// It implements the Provisioner interface expected by the bucket
	// provisioning lib.
	prov, err := provisioner.NewProvisioner(config, provisionerName, p, allNamespaces)
	if err != nil {
		log.Error(err, "failed to create noobaa provisioner")
		return err
	}

	log.Info("running noobaa provisioner ", provisionerName)
	stopChan := make(chan struct{})
	go func() {
		prov.Run(stopChan)
	}()

	return nil
}

// Provisioner implements lib-bucket-provisioner callbacks
type Provisioner struct {
	clientset *kubernetes.Clientset
	client    client.Client
	scheme    *runtime.Scheme
	recorder  record.EventRecorder
	nbClient  nb.Client

	// request info
	bucketName string

	// noobaa system info
	isSSL  bool
	s3Host string
	s3Port int

	// account details
	accountUserName string
	accessKey       string
	secretKey       string
}

// Provision implements lib-bucket-provisioner callback to create a new bucket
func (p Provisioner) Provision(options *obAPI.BucketOptions) (*nbv1.ObjectBucket, error) {
	p.bucketName = options.BucketName
	log.Infof("Provision: got request to provision bucket %q", p.bucketName)

	err := p.initNoobaaInfo()
	if err != nil {
		log.Info("GetNBClient returned error ", err)
		return nil, err
	}

	// TODO: we need to better handle the case that a bucket was created, but Provision failed
	// right now we will fail on create bucket when Provision is called the second time
	err = p.createBucket()
	if err != nil {
		return nil, err
	}

	// create account and give permissions for bucket
	p.createAccountForBucket()
	if err != nil {
		return nil, err
	}

	return p.getObjectBucket(), nil
}

// Delete implements lib-bucket-provisioner callback to delete a bucket
func (p Provisioner) Delete(ob *nbv1.ObjectBucket) error {
	p.bucketName = ob.Spec.Endpoint.BucketName
	p.accountUserName = ob.Spec.AdditionalState[obStateAccountNameKey]
	log.Infof("Delete: got request to delete bucket %q and account %q", p.bucketName, p.accountUserName)

	err := p.initNoobaaInfo()

	// TODO: delete the bucket and all its object

	err = p.deleteAccount()
	if err != nil {
		return err
	}

	return nil
}

// Grant implements lib-bucket-provisioner callback to use an existing bucket
func (p Provisioner) Grant(options *obAPI.BucketOptions) (*nbv1.ObjectBucket, error) {
	p.bucketName = options.BucketName
	log.Infof("Grant: got request to grant access to bucket %q", p.bucketName)

	err := p.initNoobaaInfo()
	if err != nil {
		log.Info("GetNBClient returned error ", err)
		return nil, err
	}

	// create account and give permissions for bucket
	p.createAccountForBucket()
	if err != nil {
		return nil, err
	}

	return p.getObjectBucket(), nil
}

// Revoke implements lib-bucket-provisioner callback to stop using an existing bucket
func (p Provisioner) Revoke(ob *nbv1.ObjectBucket) error {
	p.bucketName = ob.Spec.Endpoint.BucketName
	p.accountUserName = ob.Spec.AdditionalState[obStateAccountNameKey]
	log.Infof("Revoke: got request to revoke access to bucket %q for account %q", p.bucketName, p.accountUserName)

	err := p.initNoobaaInfo()

	err = p.deleteAccount()
	if err != nil {
		return err
	}

	return nil
}

func (p *Provisioner) createBucket() error {
	_, err := p.nbClient.CreateBucketAPI(nb.CreateBucketParams{Name: p.bucketName})
	if err != nil {
		if nbErr, ok := err.(*nb.RPCError); ok {
			if nbErr.RPCCode == "BUCKET_ALREADY_EXISTS" {
				msg := fmt.Sprintf("Bucket %q already exists", p.bucketName)
				log.Error(msg)
				return obErrors.NewBucketExistsError(msg)
			}
		}
		return fmt.Errorf("Failed to create bucket %q with error: %v", p.bucketName, err)
	}

	log.Infof("✅ Successfully created bucket %q", p.bucketName)
	return nil
}

func (p *Provisioner) createAccountForBucket() error {
	timeSuffix := time.Now().Unix()
	name := fmt.Sprintf("%s.account-%v@noobaa.io", p.bucketName, timeSuffix)
	accountInfo, err := p.nbClient.CreateAccountAPI(nb.CreateAccountParams{
		Name:              name,
		Email:             name,
		HasLogin:          false,
		S3Access:          true,
		AllowBucketCreate: false,
		AllowedBuckets: nb.AccountAllowedBuckets{
			FullPermission: false,
			PermissionList: []string{p.bucketName},
		},
	})
	if err != nil {
		return err
	}

	if len(accountInfo.AccessKeys) == 0 {
		return fmt.Errorf("Create account did not return access keys")
	}

	p.accountUserName = name
	p.accessKey = accountInfo.AccessKeys[0].AccessKey
	p.secretKey = accountInfo.AccessKeys[0].SecretKey

	log.Infof("✅ Successfully created account %q with access to bucket %q", name, p.bucketName)
	return nil
}

func (p *Provisioner) deleteAccount() error {

	log.Infof("deleting account %q", p.accountUserName)
	_, err := p.nbClient.DeleteAccountAPI(nb.DeleteAccountParams{Email: p.accountUserName})
	if err != nil {
		return fmt.Errorf("failed to delete account %q. got error: %v", p.accountUserName, err)
	}

	log.Infof("✅ Successfully deleted account %q", p.accountUserName)
	return nil
}

func (p *Provisioner) initNoobaaInfo() error {

	// TODO how to get the system name of the provisioner?
	r := system.NewReconciler(
		types.NamespacedName{Namespace: namespace, Name: options.SystemName},
		p.client, p.scheme, nil,
	)
	r.Load()

	mgmtStatus := &r.NooBaa.Status.Services.ServiceMgmt
	s3Status := &r.NooBaa.Status.Services.ServiceS3
	token := r.SecretOp.StringData["auth_token"]

	if len(mgmtStatus.NodePorts) == 0 {
		err := fmt.Errorf("❌ Mgmt service not ready")
		log.Error(err)
		return err
	}
	if len(s3Status.NodePorts) == 0 {
		err := fmt.Errorf("❌ S3 service not ready")
		log.Error(err)
		return err
	}
	if token == "" {
		err := fmt.Errorf("❌ Auth token not ready")
		log.Error(err)
		return err
	}

	s3URL, err := url.Parse(s3Status.NodePorts[0])
	if err != nil {
		return fmt.Errorf("failed to parse s3 endpoint %q. got error: %v", s3Status.NodePorts[0], err)
	}

	nodePort := mgmtStatus.NodePorts[0]
	nodeIP := nodePort[strings.Index(nodePort, "://")+3 : strings.LastIndex(nodePort, ":")]
	nbClient := nb.NewClient(&nb.APIRouterNodePort{
		ServiceMgmt: r.ServiceMgmt,
		NodeIP:      nodeIP,
	})

	nbClient.SetAuthToken(token)
	p.nbClient = nbClient
	p.s3Host = s3URL.Hostname()
	p.s3Port, err = strconv.Atoi(s3URL.Port())
	p.isSSL = strings.HasPrefix(s3Status.NodePorts[0], "https")
	if err != nil {
		return fmt.Errorf("failed to parse s3 port %q. got error: %v", s3URL.Port(), err)
	}

	return nil
}

func (p *Provisioner) getObjectBucket() *nbv1.ObjectBucket {
	conn := &nbv1.Connection{
		Endpoint: &nbv1.Endpoint{
			BucketHost: p.s3Host,
			BucketPort: p.s3Port,
			BucketName: p.bucketName,
			SSL:        p.isSSL,
		},
		Authentication: &nbv1.Authentication{
			AccessKeys: &nbv1.AccessKeys{
				AccessKeyID:     p.accessKey,
				SecretAccessKey: p.secretKey,
			},
		},
		// store the user information so we can remove the user once the OB is deleted
		AdditionalState: map[string]string{
			obStateAccountNameKey: p.accountUserName,
		},
	}

	return &nbv1.ObjectBucket{
		Spec: nbv1.ObjectBucketSpec{
			Connection: conn,
		},
	}
}

// create k8s config and client for the runtime-controller.
// Note: panics on errors.
func createConfigAndClient() (*restclient.Config, *kubernetes.Clientset, error) {
	config, err := config.GetConfig()
	if err != nil {
		log.Error(err, "Failed to create config")
		return nil, nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error(err, "Failed to create client")
		return nil, nil, err
	}
	return config, clientset, nil
}
