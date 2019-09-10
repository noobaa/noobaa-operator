module github.com/noobaa/noobaa-operator

go 1.12

// Force Kubernetes => 1.13
replace (
	k8s.io/api => k8s.io/api v0.0.0-20190817221950-ebce17126a01
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190817224053-878b1211abf1
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190817221809-bf4de9df677c
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190817222206-ee6c071a42cf
)

// Operator Framework => 0.10
replace (
	github.com/coreos/prometheus-operator => github.com/coreos/prometheus-operator v0.29.0
	github.com/operator-framework/operator-lifecycle-manager => github.com/operator-framework/operator-lifecycle-manager v0.0.0-20190522155654-7b2b397ad0cc
	github.com/operator-framework/operator-sdk => github.com/operator-framework/operator-sdk v0.10.0
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.1.12
)

require (
	github.com/aws/aws-sdk-go v1.23.8
	github.com/blang/semver v3.5.1+incompatible
	github.com/coreos/prometheus-operator v0.31.1
	github.com/docker/distribution v2.7.1+incompatible
	github.com/go-openapi/spec v0.19.2
	github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/hashicorp/go-version v1.2.0
	github.com/kube-object-storage/lib-bucket-provisioner v0.0.0-20190924175516-f3ba69cc601e
	github.com/openshift/cloud-credential-operator v0.0.0-20190614194054-1ccced634f6c
	github.com/openshift/custom-resource-status v0.0.0-20190801200128-4c95b3a336cd
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-20190128024246-5eb7ae5bdb7a
	github.com/operator-framework/operator-sdk v0.8.2-0.20190522220659-031d71ef8154
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	golang.org/x/crypto v0.0.0-20190829043050-9756ffdc2472
	golang.org/x/net v0.0.0-20190827160401-ba9fcec4b297 // indirect
	golang.org/x/sys v0.0.0-20190904005037-43c01164e931 // indirect
	golang.org/x/tools v0.0.0-20190903163617-be0da057c5e3 // indirect
	k8s.io/api v0.0.0-20190820101039-d651a1528133
	k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery v0.0.0-20190826114657-e31a5531b558
	k8s.io/cli-runtime v0.0.0-20190823123533-5ef25e8d2ab0
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/code-generator v0.0.0-20190826114438-f795916aae3f
	k8s.io/gengo v0.0.0-20190822140433-26a664648505
	k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf
	k8s.io/kubectl v0.0.0-20190826124545-8fb713b895ce
	sigs.k8s.io/controller-runtime v0.1.10
	sigs.k8s.io/controller-tools v0.0.0-20190411181648-9d55346c2bde
	sigs.k8s.io/yaml v1.1.0
)
