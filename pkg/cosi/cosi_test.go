package cosi

import (
	"context"
	"fmt"
	"os"

	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	namespacestore "github.com/noobaa/noobaa-operator/v5/pkg/namespacestore"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cosi "sigs.k8s.io/container-object-storage-interface-spec"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("COSI driver/provisioner tests", func() {
	os.Setenv("TEST_ENV", "true")
	options.Namespace = "test"

	defaultBackingStoreName := "noobaa-default-backing-store"
	cosiNSResourceName := "cosi-nsr"
	firstBucket := "first.bucket"
	cosiRegBucketName := "cosi-regular-bucket"
	cosiRegBucketDeleteName := "cosi-regular-bucket-to-delete"
	cosiRegReplicationBucketName := "cosi-regular-bucket-replication"
	cosiRegQuotaBucketName := "cosi-regular-bucket-quota"
	cosiRegTargetBucketName := "cosi-regular-bucket-target-to-ns"
	cosiNsBucketName := "cosi-namespace-bucket"
	cosiAccountName := "cosi-account"

	log := util.Logger()
	nbClient := system.GetNBClient()
	sysClient, err := system.Connect(true)
	if err != nil {
		log.Infof("COSI bucket creation tests, can't create sysclient: err= %q", err)
	}
	var nsResource *nbv1.NamespaceStore
	p := &Provisioner{
		Logger:    log,
		Namespace: options.Namespace,
	}

	AfterSuite(func() {
		bucketsToDelete := []string{cosiRegBucketName, cosiNsBucketName, cosiRegTargetBucketName, cosiRegReplicationBucketName, cosiRegQuotaBucketName}
		cleanupKubeResources(nsResource)
		cleanupBuckets(nbClient, bucketsToDelete)
		cleanupAccounts(nbClient, cosiAccountName)
	})

	Context("Create/Delete bucket using COSI driver ", func() {

		It("create COSI regular bucket - default bucketclass", func() {
			// call create bucket of cosi driver
			driverCreateBucketParams := getCreateBucketParams(cosiRegBucketName, &defaultBackingStoreName, nil, nil, false)
			resp, _ := p.DriverCreateBucket(context.TODO(), driverCreateBucketParams)
			Expect(resp.BucketId).To(Equal(cosiRegBucketName))

			// call read bucket and compare
			readBucketResp, err := nbClient.ReadBucketAPI(nb.ReadBucketParams{Name: cosiRegBucketName})
			Expect(err).To(BeNil())
			Expect(readBucketResp.Name).To(Equal(cosiRegBucketName))
			Expect(readBucketResp.BucketType).To(Equal("REGULAR"))
		})

		It("create COSI regular bucket - default bucketclass & replication", func() {
			// call create bucket of cosi driver
			driverCreateBucketParams := getCreateBucketParams(cosiRegReplicationBucketName, &defaultBackingStoreName, nil, &firstBucket, false)
			resp, _ := p.DriverCreateBucket(context.TODO(), driverCreateBucketParams)
			Expect(resp.BucketId).To(Equal(cosiRegReplicationBucketName))

			// call read bucket and compare
			readBucketResp, err := nbClient.ReadBucketAPI(nb.ReadBucketParams{Name: cosiRegReplicationBucketName})
			Expect(err).To(BeNil())
			Expect(readBucketResp.Name).To(Equal(cosiRegReplicationBucketName))
			Expect(readBucketResp.BucketType).To(Equal("REGULAR"))

			replicationResp, err := nbClient.GetBucketReplicationAPI(nb.ReadBucketParams{Name: cosiRegReplicationBucketName})
			Expect(err).To(BeNil())
			var expectedReplication = map[string]interface{}{"filter": map[string]interface{}{"prefix": "a"}, "rule_id": "rule-1", "destination_bucket": "first.bucket"}
			Expect(replicationResp.Rules[0]).To(Equal(expectedReplication))
		})

		It("create COSI regular bucket - default bucketclass & quota", func() {
			// call create bucket of cosi driver
			driverCreateBucketParams := getCreateBucketParams(cosiRegQuotaBucketName, &defaultBackingStoreName, nil, nil, true)
			resp, _ := p.DriverCreateBucket(context.TODO(), driverCreateBucketParams)
			Expect(resp.BucketId).To(Equal(cosiRegQuotaBucketName))

			// call read bucket and compare
			readBucketResp, err := nbClient.ReadBucketAPI(nb.ReadBucketParams{Name: cosiRegQuotaBucketName})
			Expect(err).To(BeNil())
			Expect(readBucketResp.Name).To(Equal(cosiRegQuotaBucketName))
			Expect(readBucketResp.BucketType).To(Equal("REGULAR"))
			Expect(readBucketResp.Quota.Size.Value).To(Equal(float64(2)))
			Expect(readBucketResp.Quota.Quantity.Value).To(Equal(5))
		})

		It("delete COSI regular bucket - default bucket", func() {

			// call create bucket of cosi driver
			driverCreateBucketParams := getCreateBucketParams(cosiRegBucketDeleteName, &defaultBackingStoreName, nil, nil, false)
			resp, err := p.DriverCreateBucket(context.TODO(), driverCreateBucketParams)
			Expect(resp.BucketId).To(Equal(cosiRegBucketDeleteName))
			Expect(err).To(BeNil())

			// call read bucket
			readBucketResp, err := nbClient.ReadBucketAPI(nb.ReadBucketParams{Name: cosiRegBucketDeleteName})
			Expect(err).To(BeNil())
			Expect(readBucketResp.Name).To(Equal(cosiRegBucketDeleteName))

			// call delete bucket
			_, err = p.DriverDeleteBucket(context.TODO(), &cosi.DriverDeleteBucketRequest{BucketId: cosiRegBucketDeleteName})
			Expect(err).To(BeNil())

			// check bucket was deleted
			_, err = nbClient.ReadBucketAPI(nb.ReadBucketParams{Name: cosiRegBucketDeleteName})
			Expect(err).NotTo(BeNil())
			rpcErr, isRPCErr := err.(*nb.RPCError)
			Expect(isRPCErr).To(BeTrue())
			Expect(rpcErr.RPCCode).To(Equal("NO_SUCH_BUCKET"))
		})
	})

	Context("Create/Delete namespace bucket using COSI driver ", func() {

		BeforeEach(func() {
			// call create bucket of cosi driver
			driverCreateTargetBucketParams := getCreateBucketParams(cosiRegTargetBucketName, &defaultBackingStoreName, nil, nil, false)
			createBucketResp, _ := p.DriverCreateBucket(context.TODO(), driverCreateTargetBucketParams)
			Expect(createBucketResp.BucketId).To(Equal(cosiRegTargetBucketName))

			// create namespace resource on top of first.bucket
			nsResource = getNamespaceResourceObj(cosiNSResourceName, cosiRegTargetBucketName, sysClient)
			Expect(util.KubeCreateFailExisting(nsResource)).To(BeTrue())
			namespacestore.WaitReady(nsResource)
		}, 30)

		It("create COSI bucket - namespace bucket", func() {
			// call create bucket of cosi driver
			driverCreateBucketParams := getCreateBucketParams(cosiNsBucketName, nil, &cosiNSResourceName, nil, false)
			resp, err := p.DriverCreateBucket(context.TODO(), driverCreateBucketParams)
			Expect(err).To(BeNil())
			Expect(resp.BucketId).To(Equal(cosiNsBucketName))

			// call read bucket and compare
			readBucketResp, err := nbClient.ReadBucketAPI(nb.ReadBucketParams{Name: cosiNsBucketName})
			Expect(err).To(BeNil())
			expectedBucketInfo := getExpectedBucketInfo(cosiNSResourceName, cosiNsBucketName)
			Expect(readBucketResp.Name).To(Equal(expectedBucketInfo.Name))
			Expect(readBucketResp.Namespace).To(Equal(expectedBucketInfo.Namespace))
			Expect(readBucketResp.BucketType).To(Equal("NAMESPACE"))
		})
	})

	Context("Grant/Revoke bucket access using COSI driver ", func() {
		It("Grant COSI bucket access", func() {

			// call grant access of cosi driver
			cosiBucketAccessReq := getGrantAccessParams(firstBucket, cosiAccountName)
			resp, _ := p.DriverGrantBucketAccess(context.TODO(), cosiBucketAccessReq)
			Expect(resp.AccountId).To(Equal(cosiAccountName))

			// call read account and compare
			readAccountResp, err := nbClient.ReadAccountAPI(nb.ReadAccountParams{Email: cosiAccountName})
			Expect(err).To(BeNil())
			Expect(readAccountResp.Name).To(Equal(cosiAccountName))
			Expect(readAccountResp.Email).To(Equal(cosiAccountName))
		})

		It("Revoke COSI bucket access", func() {
			cosiAccountToDeleteName := "cosi-account-to-delete"

			// call grant access of cosi driver
			cosiGrantAccessReq := getGrantAccessParams(firstBucket, cosiAccountToDeleteName)
			grantResp, err := p.DriverGrantBucketAccess(context.TODO(), cosiGrantAccessReq)
			Expect(grantResp.AccountId).To(Equal(cosiAccountToDeleteName))
			Expect(err).To(BeNil())

			// call revoke access of cosi driver
			cosiRevokeAccessReq := getRevokeAccessParams(firstBucket, cosiAccountToDeleteName)
			_, err = p.DriverRevokeBucketAccess(context.TODO(), cosiRevokeAccessReq)
			Expect(err).To(BeNil())

			// call read account and expect an error
			_, err = nbClient.ReadAccountAPI(nb.ReadAccountParams{Email: cosiAccountToDeleteName})
			Expect(err).NotTo(BeNil())
			rpcErr, isRPCErr := err.(*nb.RPCError)
			Expect(isRPCErr).To(BeTrue())
			Expect(rpcErr.RPCCode).To(Equal("NO_SUCH_ACCOUNT"))
		})
	})
})

func getCreateBucketParams(cosiBucketName string, resourceName *string, nsResourceName *string, replicationDstBucket *string, addQuota bool) *cosi.DriverCreateBucketRequest {
	var placementPolicy string
	var namespacePolicy string
	var replicationPolicy string
	var quota string

	if resourceName != nil {
		placementPolicy = fmt.Sprintf("{\"tiers\":[{\"backingStores\":[\"%s\"]}]}", *resourceName)
	}
	if nsResourceName != nil {
		namespacePolicy = fmt.Sprintf("{\"type\": \"Single\", \"single\":{\"resource\":\"%s\"}}", *nsResourceName)
	}
	if replicationDstBucket != nil {
		replicationPolicy = fmt.Sprintf("\"{\\\"rules\\\":[{\\\"rule_id\\\":\\\"rule-1\\\",\\\"destination_bucket\\\":\\\"%s\\\",\\\"filter\\\":{\\\"prefix\\\":\\\"a\\\"}}]}\"", *replicationDstBucket)
	}
	if addQuota == true {
		quota = "{ \"maxObjects\": \"5\", \"maxSize\": \"2Gi\"}"
	}

	return &cosi.DriverCreateBucketRequest{
		Name: cosiBucketName,
		Parameters: map[string]string{
			"placementPolicy":   placementPolicy,
			"namespacePolicy":   namespacePolicy,
			"replicationPolicy": replicationPolicy,
			"quota":             quota,
		},
	}
}

func getGrantAccessParams(bucketToAccess string, AccountName string) *cosi.DriverGrantBucketAccessRequest {
	return &cosi.DriverGrantBucketAccessRequest{
		BucketId:           bucketToAccess,
		Name:               AccountName,
		AuthenticationType: cosi.AuthenticationType_Key,
	}
}

func getRevokeAccessParams(bucketToAccess string, AccountName string) *cosi.DriverRevokeBucketAccessRequest {
	return &cosi.DriverRevokeBucketAccessRequest{
		BucketId:  bucketToAccess,
		AccountId: AccountName,
	}
}

func getNamespaceResourceObj(cosiNSResourceName string, TargetBucket string, sysClient *system.Client) *nbv1.NamespaceStore {
	return &nbv1.NamespaceStore{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cosiNSResourceName,
			Namespace: options.Namespace,
		},
		Spec: nbv1.NamespaceStoreSpec{
			Type: nbv1.NSStoreTypeS3Compatible,
			S3Compatible: &nbv1.S3CompatibleSpec{
				TargetBucket: TargetBucket,
				Secret: corev1.SecretReference{
					Name:      sysClient.SecretAdmin.Name,
					Namespace: sysClient.SecretAdmin.Namespace,
				},
				Endpoint: sysClient.S3URL.String(),
			},
		},
	}
}

func getExpectedBucketInfo(cosiNSResourceName string, cosiNsBucketName string) *nb.BucketInfo {
	var readResources []nb.NamespaceResourceFullConfig
	expectedBucketInfo := &nb.BucketInfo{
		Name: cosiNsBucketName,
		Namespace: &nb.NamespaceBucketInfo{
			WriteResource: nb.NamespaceResourceFullConfig{
				Resource: cosiNSResourceName,
			},
			ReadResources: append(readResources, nb.NamespaceResourceFullConfig{
				Resource: cosiNSResourceName,
			}),
		},
	}
	return expectedBucketInfo
}

func cleanupBuckets(nbClient nb.Client, bucketNames []string) {
	for _, bucketName := range bucketNames {
		err := nbClient.DeleteBucketAPI(nb.DeleteBucketParams{
			Name: bucketName,
		})
		Expect(err).To(BeNil())
	}
}

func cleanupAccounts(nbClient nb.Client, accountEmail string) {
	err := nbClient.DeleteAccountAPI(nb.DeleteAccountParams{
		Email: accountEmail,
	})
	Expect(err).To(BeNil())
}

func cleanupKubeResources(obj client.Object) {
	util.KubeCheck(obj)
	emptyFinalizers := []string{}
	obj.SetFinalizers(emptyFinalizers)
	util.KubeUpdate(obj)
	Expect(util.KubeDelete(obj)).To(BeTrue())
}
