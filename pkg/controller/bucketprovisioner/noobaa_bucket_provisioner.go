package bucketprovisioner

import ( // "flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	obAPI "github.com/kube-object-storage/lib-bucket-provisioner/pkg/apis/objectbucket.io/v1alpha1"
	libbkt "github.com/kube-object-storage/lib-bucket-provisioner/pkg/provisioner"
	apibkt "github.com/kube-object-storage/lib-bucket-provisioner/pkg/provisioner/api"
	obError "github.com/kube-object-storage/lib-bucket-provisioner/pkg/provisioner/api/errors"
	"github.com/noobaa/noobaa-operator/pkg/nb"
	"github.com/noobaa/noobaa-operator/pkg/system"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	obStateAccountNameKey = "UserName"
)

var (
	namespace       = os.Getenv("WATCH_NAMESPACE")
	logger          = logrus.WithFields(logrus.Fields{"mod": "bucket-provisioner"})
	provisionerName = "noobaa.io/" + namespace + ".bucket"
)

type noobaaBucketProvisioner struct {
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

func (p *noobaaBucketProvisioner) getObjectBucket() *obAPI.ObjectBucket {
	conn := &obAPI.Connection{
		Endpoint: &obAPI.Endpoint{
			BucketHost: p.s3Host,
			BucketPort: p.s3Port,
			BucketName: p.bucketName,
			SSL:        p.isSSL,
		},
		Authentication: &obAPI.Authentication{
			AccessKeys: &obAPI.AccessKeys{
				AccessKeyID:     p.accessKey,
				SecretAccessKey: p.secretKey,
			},
		},
		// store the user information so we can remove the user once the OB is deleted
		AdditionalState: map[string]string{
			obStateAccountNameKey: p.accountUserName,
		},
	}

	return &obAPI.ObjectBucket{
		Spec: obAPI.ObjectBucketSpec{
			Connection: conn,
		},
	}

	// return &obAPI.ObjectBucket{}
}

func (p *noobaaBucketProvisioner) createBucket() error {
	_, err := p.nbClient.CreateBucketAPI(nb.CreateBucketParams{Name: p.bucketName})
	if err != nil {
		if nbErr, ok := err.(*nb.RPCError); ok {
			if nbErr.RPCCode == "BUCKET_ALREADY_EXISTS" {
				msg := fmt.Sprintf("Bucket %q already exists", p.bucketName)
				logger.Error(msg)
				return obError.NewBucketExistsError(msg)
			}
		}
		return fmt.Errorf("Failed to create bucket %q with error: %v", p.bucketName, err)
	}
	logger.Infof("Successfully created bucket %q", p.bucketName)
	return nil
}

func (p *noobaaBucketProvisioner) createAccountForBucket() error {
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
	logger.Infof("Successfully created account %q with access to bucket %q", name, p.bucketName)
	return nil
}

func (p *noobaaBucketProvisioner) deleteAccount() error {

	logger.Infof("deleting account %q", p.accountUserName)
	_, err := p.nbClient.DeleteAccountAPI(nb.DeleteAccountParams{Email: p.accountUserName})
	if err != nil {
		return fmt.Errorf("failed to delete account %q. got error: %v", p.accountUserName, err)
	}
	return nil
}

func (p noobaaBucketProvisioner) Provision(options *apibkt.BucketOptions) (*obAPI.ObjectBucket, error) {
	p.bucketName = options.BucketName
	logger.Infof("Provision: got request to provision bucket %q", p.bucketName)

	err := p.initNoobaaInfo()
	if err != nil {
		logger.Info("GetNBClient returned error ", err)
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

func (p noobaaBucketProvisioner) Delete(ob *obAPI.ObjectBucket) error {
	p.bucketName = ob.Spec.Endpoint.BucketName
	p.accountUserName = ob.Spec.AdditionalState[obStateAccountNameKey]
	logger.Infof("Delete: got request to delete bucket %q and account %q", p.bucketName, p.accountUserName)

	err := p.initNoobaaInfo()

	// TODO: delete the bucket and all its object

	err = p.deleteAccount()
	if err != nil {
		return err
	}

	return nil
}

func (p noobaaBucketProvisioner) Grant(options *apibkt.BucketOptions) (*obAPI.ObjectBucket, error) {
	p.bucketName = options.BucketName
	logger.Infof("Grant: got request to grant access to bucket %q", p.bucketName)

	err := p.initNoobaaInfo()
	if err != nil {
		logger.Info("GetNBClient returned error ", err)
		return nil, err
	}

	// create account and give permissions for bucket
	p.createAccountForBucket()
	if err != nil {
		return nil, err
	}

	return p.getObjectBucket(), nil
}

func (p noobaaBucketProvisioner) Revoke(ob *obAPI.ObjectBucket) error {
	p.bucketName = ob.Spec.Endpoint.BucketName
	p.accountUserName = ob.Spec.AdditionalState[obStateAccountNameKey]
	logger.Infof("Revoke: got request to revoke access to bucket %q for account %q", p.bucketName, p.accountUserName)

	err := p.initNoobaaInfo()

	err = p.deleteAccount()
	if err != nil {
		return err
	}

	return nil
}

// create k8s config and client for the runtime-controller.
// Note: panics on errors.
func createConfigAndClient() (*restclient.Config, *kubernetes.Clientset, error) {
	config, err := config.GetConfig()
	if err != nil {
		logger.Error(err, "Failed to create config")
		return nil, nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Error(err, "Failed to create client")
		return nil, nil, err
	}
	return config, clientset, nil
}

func (p *noobaaBucketProvisioner) initNoobaaInfo() error {
	s := system.New(types.NamespacedName{Namespace: namespace, Name: "noobaa"}, p.client, p.scheme, nil)
	s.Load()

	mgmtStatus := s.NooBaa.Status.Services.ServiceMgmt
	if len(mgmtStatus.NodePorts) == 0 {
		err := fmt.Errorf("❌ Mgmt service not ready")
		logger.Error(err)
		return err
	}
	if s.SecretOp.StringData["auth_token"] == "" {
		err := fmt.Errorf("❌ Auth token not ready")
		logger.Error(err)
		return err
	}

	s3Status := s.NooBaa.Status.Services.ServiceS3
	if len(s3Status.NodePorts) == 0 {
		err := fmt.Errorf("❌ S3 service not ready")
		logger.Error(err)
		return err
	}

	s3URL, err := url.Parse(s3Status.NodePorts[0])
	if err != nil {
		return fmt.Errorf("failed to parse s3 endpoint %q. got error: %v", s3Status.NodePorts[0], err)
	}

	nodePort := mgmtStatus.NodePorts[0]
	nodeIP := nodePort[strings.Index(nodePort, "://")+3 : strings.LastIndex(nodePort, ":")]
	nbClient := nb.NewClient(&nb.APIRouterNodePort{
		ServiceMgmt: s.ServiceMgmt,
		NodeIP:      nodeIP,
	})

	nbClient.SetAuthToken(s.SecretOp.StringData["auth_token"])
	p.nbClient = nbClient
	p.s3Host = s3URL.Hostname()
	p.s3Port, err = strconv.Atoi(s3URL.Port())
	p.isSSL = strings.HasPrefix(s3Status.NodePorts[0], "https")
	if err != nil {
		return fmt.Errorf("failed to parse s3 port %q. got error: %v", s3URL.Port(), err)
	}

	return nil
}

// RunNoobaaProvisioner will run Noobaa OBC provisioner
func RunNoobaaProvisioner(client client.Client, scheme *runtime.Scheme, recorder record.EventRecorder) error {
	logger.Info("NooBaa Provisioner - start..")

	config, clientset, err := createConfigAndClient()
	if err != nil {
		return err
	}

	s3Prov := noobaaBucketProvisioner{
		clientset: clientset,
		client:    client,
		scheme:    scheme,
		recorder:  recorder,
	}

	// Create and run the s3 provisioner controller.
	// It implements the Provisioner interface expected by the bucket
	// provisioning lib.
	const allNamespaces = ""
	S3Provisioner, err := libbkt.NewProvisioner(config, provisionerName, s3Prov, allNamespaces)
	if err != nil {
		logger.Error(err, "failed to create noobaa provisioner")
		return err
	}
	logger.Info("running noobaa provisioner ", provisionerName)
	go func() {
		S3Provisioner.Run(make(chan struct{}))
	}()
	return nil
}
