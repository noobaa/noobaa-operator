module github.com/noobaa/noobaa-operator/v2

go 1.13

require (
	cloud.google.com/go/storage v1.3.0
	github.com/Azure/azure-sdk-for-go v39.2.0+incompatible
	github.com/Azure/azure-storage-blob-go v0.8.0
	github.com/Azure/go-autorest/autorest v0.9.4
	github.com/Azure/go-autorest/autorest/adal v0.8.1
	github.com/Azure/go-autorest/autorest/to v0.3.1-0.20191028180845-3492b2aff503
	github.com/asaskevich/govalidator v0.0.0-20200108200545-475eaeb16496
	github.com/aws/aws-sdk-go v1.25.48
	github.com/blang/semver v3.5.1+incompatible
	github.com/coreos/go-semver v0.3.0
	github.com/coreos/prometheus-operator v0.38.0
	github.com/docker/distribution v2.7.1+incompatible
	github.com/emicklei/go-restful v2.9.6+incompatible // indirect
	github.com/go-openapi/spec v0.19.4
	github.com/kube-object-storage/lib-bucket-provisioner v0.0.0-20200107223247-51020689f1fb
	github.com/marstr/randname v0.0.0-20200428202425-99aca53a2176
	github.com/openshift/api v3.9.1-0.20190924102528-32369d4db2ad+incompatible
	github.com/openshift/cloud-credential-operator v0.0.0-20190614194054-1ccced634f6c
	github.com/openshift/custom-resource-status v0.0.0-20190801200128-4c95b3a336cd
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-20200321030439-57b580e57e88 // 0.11.0
	github.com/operator-framework/operator-sdk v0.17.0
	github.com/rook/rook v1.1.2
	github.com/sirupsen/logrus v1.5.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975
	k8s.io/api v0.17.4
	k8s.io/apiextensions-apiserver v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/cli-runtime v0.17.4
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.17.4
	k8s.io/gengo v0.0.0-20191010091904-7fa3014cb28f
	k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a
	k8s.io/kubectl v0.17.4
	nhooyr.io/websocket v1.7.4
	sigs.k8s.io/controller-runtime v0.5.2
	sigs.k8s.io/yaml v1.1.0
)

// Pinned to kubernetes-1.17.4
replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/api => k8s.io/api v0.17.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.4
	k8s.io/apiserver => k8s.io/apiserver v0.17.4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.4
	k8s.io/client-go => k8s.io/client-go v0.17.4 // Required by prometheus-operator
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.17.4
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.17.4
	k8s.io/code-generator => k8s.io/code-generator v0.17.4
	k8s.io/component-base => k8s.io/component-base v0.17.4
	k8s.io/cri-api => k8s.io/cri-api v0.17.4
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.17.4
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.17.4
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.17.4
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.17.4
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.17.4
	k8s.io/kubectl => k8s.io/kubectl v0.17.4
	k8s.io/kubelet => k8s.io/kubelet v0.17.4
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.17.4
	k8s.io/metrics => k8s.io/metrics v0.17.4
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.17.4
)
