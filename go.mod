module github.com/noobaa/noobaa-operator/v2

go 1.12

require (
	github.com/asaskevich/govalidator v0.0.0-20180720115003-f9ffefc3facf
	github.com/aws/aws-sdk-go v1.23.8
	github.com/blang/semver v3.5.1+incompatible
	github.com/coreos/go-semver v0.2.0
	github.com/coreos/prometheus-operator v0.29.0
	github.com/docker/distribution v2.7.1+incompatible
	github.com/go-openapi/spec v0.19.2
	github.com/go-openapi/validate v0.18.0 // indirect
	github.com/gobuffalo/flect v0.1.6 // indirect
	github.com/kube-object-storage/lib-bucket-provisioner v0.0.0-20191209225510-8423df408133
	github.com/openshift/api v3.9.1-0.20190424152011-77b8897ec79a+incompatible
	github.com/openshift/cloud-credential-operator v0.0.0-20190614194054-1ccced634f6c
	github.com/openshift/custom-resource-status v0.0.0-20190801200128-4c95b3a336cd
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-20190605231540-b8a4faf68e36 // 0.11.0
	github.com/operator-framework/operator-sdk v0.10.0
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/rook/rook v1.1.2
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	go.uber.org/multierr v1.2.0 // indirect
	golang.org/x/crypto v0.0.0-20191119213627-4f8c1d86b1ba
	golang.org/x/tools v0.0.0-20190903163617-be0da057c5e3 // indirect
	k8s.io/api v0.0.0-20191005115622-2e41325d9e4b
	k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery v0.0.0-20191005115455-e71eb83a557c
	k8s.io/cli-runtime v0.0.0-20191005121332-4d28aef60981
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/code-generator v0.0.0-20191003035328-700b1226c0bd
	k8s.io/gengo v0.0.0-20190822140433-26a664648505
	k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a
	k8s.io/kubectl v0.0.0-20191005122440-987b623dc1f7
	nhooyr.io/websocket v1.5.0
	sigs.k8s.io/controller-runtime v0.1.12
	sigs.k8s.io/controller-tools v0.1.12
	sigs.k8s.io/testing_frameworks v0.1.2 // indirect
	sigs.k8s.io/yaml v1.1.0
)

// Pinned to kubernetes-1.13.4
// This is because operator-sdk v0.10.0 still requires kubernetes-1.13.4 and controller-runtime v0.1.10
// once we can bump the operator-sdk version we can also bump k8s.io/* and controller-runtime
replace (
	github.com/operator-framework/operator-sdk => github.com/operator-framework/operator-sdk v0.10.0
	k8s.io/api => k8s.io/api v0.0.0-20190222213804-5cb15d344471 // kubernetes-1.13.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190228180357-d002e88f6236 // kubernetes-1.13.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190221213512-86fb29eff628 // kubernetes-1.13.4
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190228174230-b40b2a5939e4 // kubernetes-1.13.4
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.1.12
	sigs.k8s.io/controller-tools => sigs.k8s.io/controller-tools v0.1.12
)
