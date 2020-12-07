module github.com/noobaa/noobaa-operator/v2

go 1.15

require (
	cloud.google.com/go/storage v1.3.0
	github.com/Azure/azure-sdk-for-go v43.0.0+incompatible
	github.com/Azure/azure-storage-blob-go v0.8.0
	github.com/Azure/go-autorest/autorest v0.11.11
	github.com/Azure/go-autorest/autorest/adal v0.9.5
	github.com/Azure/go-autorest/autorest/to v0.3.1-0.20191028180845-3492b2aff503
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535
	github.com/aws/aws-sdk-go v1.28.2
	github.com/blang/semver v3.5.1+incompatible
	github.com/coreos/go-semver v0.3.0
	github.com/coreos/prometheus-operator v0.38.1-0.20200424145508-7e176fda06cc
	github.com/docker/distribution v2.7.1+incompatible
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32 // indirect
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/go-openapi/spec v0.19.5
	github.com/hashicorp/vault/api v1.0.4
	github.com/kube-object-storage/lib-bucket-provisioner v0.0.0-20200610144127-e2eec875d6d1
	github.com/marstr/randname v0.0.0-20200428202425-99aca53a2176
	github.com/openshift/api v3.9.1-0.20190924102528-32369d4db2ad+incompatible
	github.com/openshift/cloud-credential-operator v0.0.0-20190614194054-1ccced634f6c
	github.com/openshift/custom-resource-status v0.0.0-20190801200128-4c95b3a336cd
	github.com/operator-framework/api v0.3.22
	github.com/operator-framework/operator-lib v0.2.0
	github.com/rook/rook v1.4.6
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	go.uber.org/zap v1.14.1 // indirect
	golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/tools v0.0.0-20200915031644-64986481280e // indirect
	google.golang.org/api v0.15.1
	k8s.io/api v0.19.3
	k8s.io/apiextensions-apiserver v0.19.3
	k8s.io/apimachinery v0.19.3
	k8s.io/cli-runtime v0.19.3
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.19.3
	k8s.io/gengo v0.0.0-20200428234225-8167cfdcfc14
	k8s.io/kube-openapi v0.0.0-20200805222855-6aeccd4b50c6
	k8s.io/kubectl v0.18.8
	nhooyr.io/websocket v1.7.4
	sigs.k8s.io/controller-runtime v0.6.3
	sigs.k8s.io/yaml v1.2.0
)

// Pinned to kubernetes-1.19.3
replace (
	github.com/kube-object-storage/lib-bucket-provisioner => github.com/kube-object-storage/lib-bucket-provisioner v0.0.0-20201113145926-53eac50605b4
	github.com/moby/term => github.com/moby/term v0.0.0-20201110203204-bea5bbe245bf // indirect
	k8s.io/api => k8s.io/api v0.19.3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.19.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.3
	k8s.io/apiserver => k8s.io/apiserver v0.19.3
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.19.3
	k8s.io/client-go => k8s.io/client-go v0.19.3 // Required by prometheus-operator
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.19.3
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.19.3
	k8s.io/code-generator => k8s.io/code-generator v0.19.3
	k8s.io/component-base => k8s.io/component-base v0.19.3
	k8s.io/cri-api => k8s.io/cri-api v0.19.3
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.19.3
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.19.3
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.19.3
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.19.3
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.19.3
	k8s.io/kubectl => k8s.io/kubectl v0.19.3
	k8s.io/kubelet => k8s.io/kubelet v0.19.3
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.19.3
	k8s.io/metrics => k8s.io/metrics v0.19.3
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.19.3
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.6.2
)
