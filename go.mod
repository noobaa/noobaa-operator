module github.com/noobaa/noobaa-operator/v5

go 1.24.7

replace (
	// TODO: remove this replace once https://github.com/libopenstorage/secrets/pull/83 is merged
	github.com/libopenstorage/secrets => github.com/rook/secrets v0.0.0-20240315053144-3195f6906937
	github.com/portworx/sched-ops => github.com/portworx/sched-ops v0.20.4-openstorage-rc3
)

require (
	cloud.google.com/go/storage v1.57.2
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.20.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.13.1
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage v1.8.1
	github.com/Azure/azure-storage-blob-go v0.15.0
	github.com/IBM/keyprotect-go-client v0.15.1
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2
	github.com/aws/aws-sdk-go v1.48.12
	github.com/blang/semver/v4 v4.0.0
	github.com/cloudnative-pg/cloudnative-pg v1.27.0
	github.com/coreos/go-semver v0.3.1
	github.com/distribution/reference v0.6.0
	github.com/gemalto/kmip-go v0.0.10
	github.com/go-test/deep v1.1.1
	github.com/google/uuid v1.6.0
	github.com/kedacore/keda/v2 v2.18.1
	github.com/kube-object-storage/lib-bucket-provisioner v0.0.0-20221122204822-d1a8c34382f1
	github.com/kubernetes-csi/external-snapshotter/client/v8 v8.4.0
	github.com/libopenstorage/secrets v0.0.0-20240416031220-a17cf7f72c6c
	github.com/marstr/randname v0.0.0-20200428202425-99aca53a2176
	github.com/onsi/ginkgo/v2 v2.27.2
	github.com/onsi/gomega v1.38.2
	github.com/openshift/api v0.0.0-20251120040117-916c7003ed78
	github.com/openshift/cloud-credential-operator v0.0.0-20251215212231-8f1a631d825f
	github.com/openshift/custom-resource-status v1.1.3-0.20220503160415-f2fdb4999d87
	github.com/operator-framework/api v0.36.0
	github.com/operator-framework/operator-lib v0.19.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.86.2
	github.com/robfig/cron/v3 v3.0.1
	github.com/rook/rook/pkg/apis v0.0.0-20251119183400-8782e9183d04
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.10
	github.com/spf13/viper v1.21.0
	golang.org/x/term v0.37.0
	google.golang.org/api v0.256.0
	google.golang.org/grpc v1.77.0
	k8s.io/api v0.34.2
	k8s.io/apiextensions-apiserver v0.34.2
	k8s.io/apimachinery v0.34.2
	k8s.io/cli-runtime v0.34.2
	k8s.io/client-go v0.34.2
	k8s.io/code-generator v0.34.2
	k8s.io/gengo v0.0.0-20230829151522-9cce18d56c01
	k8s.io/kube-aggregator v0.34.2
	k8s.io/kube-openapi v0.0.0-20251125145642-4e65d59e963e
	k8s.io/kubectl v0.34.2
	k8s.io/utils v0.0.0-20251002143259-bc988d571ff4
	nhooyr.io/websocket v1.8.7
	sigs.k8s.io/container-object-storage-interface-api v0.1.0
	sigs.k8s.io/container-object-storage-interface-provisioner-sidecar v0.1.0
	sigs.k8s.io/container-object-storage-interface-spec v0.1.0
	sigs.k8s.io/controller-runtime v0.22.4
	sigs.k8s.io/yaml v1.6.0
)

require (
	cel.dev/expr v0.25.1 // indirect
	cloud.google.com/go/auth v0.17.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/monitoring v1.24.3 // indirect
	github.com/Azure/azure-pipeline-go v0.2.3 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.11.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets v0.12.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/keyvault/internal v0.7.1 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.6.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.30.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.54.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.54.0 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cloudnative-pg/barman-cloud v0.3.3 // indirect
	github.com/cloudnative-pg/cnpg-i v0.3.0 // indirect
	github.com/cloudnative-pg/machinery v0.3.1 // indirect
	github.com/cncf/xds/go v0.0.0-20251110193048-8bfbf64dc13e // indirect
	github.com/containernetworking/cni v1.2.3 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.36.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/expr-lang/expr v1.17.7 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/go-jose/go-jose/v4 v4.1.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-logr/zapr v1.3.0 // indirect
	github.com/go-openapi/swag/cmdutils v0.25.4 // indirect
	github.com/go-openapi/swag/conv v0.25.4 // indirect
	github.com/go-openapi/swag/fileutils v0.25.4 // indirect
	github.com/go-openapi/swag/jsonname v0.25.4 // indirect
	github.com/go-openapi/swag/jsonutils v0.25.4 // indirect
	github.com/go-openapi/swag/loading v0.25.4 // indirect
	github.com/go-openapi/swag/mangling v0.25.4 // indirect
	github.com/go-openapi/swag/netutils v0.25.4 // indirect
	github.com/go-openapi/swag/stringutils v0.25.4 // indirect
	github.com/go-openapi/swag/typeutils v0.25.4 // indirect
	github.com/go-openapi/swag/yamlutils v0.25.4 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.0 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/pprof v0.0.0-20251114195745-4902fdda35c8 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/hashicorp/vault/api/auth/kubernetes v0.8.0 // indirect
	github.com/k8snetworkplumbingwg/network-attachment-definition-client v1.7.7 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/mattn/go-ieproxy v0.0.1 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/spiffe/go-spiffe/v2 v2.6.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.38.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.63.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.63.0 // indirect
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/sync v0.18.0 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
	k8s.io/gengo/v2 v2.0.0-20250604051438-85fd79dbfd9f // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.1 // indirect
)

require (
	cloud.google.com/go v0.123.0 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	cloud.google.com/go/iam v1.5.3 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/ansel1/merry v1.8.0 // indirect
	github.com/ansel1/merry/v2 v2.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bytedance/sonic v1.10.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/emicklei/go-restful/v3 v3.13.0 // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/gemalto/flume v0.13.1 // indirect
	github.com/gin-gonic/gin v1.9.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-openapi/jsonpointer v0.22.3 // indirect
	github.com/go-openapi/jsonreference v0.21.3 // indirect
	github.com/go-openapi/swag v0.25.4 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/gnostic-models v0.7.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.7 // indirect
	github.com/googleapis/gax-go/v2 v2.15.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.8 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.2.0 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.7 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-7 // indirect
	github.com/hashicorp/vault/api v1.22.0 // indirect
	github.com/hashicorp/vault/api/auth/approle v0.8.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.1 // indirect
	github.com/lib/pq v1.10.9
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/marstr/collection v1.0.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/moby/spdystream v0.5.0 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/portworx/sched-ops v1.20.4-rc1.0.20220208024433-611d861089d4 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.4 // indirect
	github.com/prometheus/procfs v0.19.2 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	golang.org/x/arch v0.4.0 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/mod v0.30.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/oauth2 v0.33.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	golang.org/x/tools v0.39.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.5.0 // indirect
	google.golang.org/genproto v0.0.0-20251124214823-79d6a2a48846 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251124214823-79d6a2a48846 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251124214823-79d6a2a48846 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	knative.dev/pkg v0.0.0-20250326102644-9f3e60a9244c // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
)

// see https://github.com/rook/rook/blob/master/go.mod#L42-L43
//replace github.com/portworx/sched-ops => github.com/portworx/sched-ops v0.20.4-openstorage-rc3

// we need to get rid of "github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage" in azure utils.go
replace github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v62.0.0+incompatible

// TODO fix the KNS
// replace github.com/libopenstorage/secrets => github.com/libopenstorage/secrets v0.0.0-20230117230814-885ae38d82f8

// https://github.com/rook/rook/blob/master/go.mod#L47-L49
exclude (
	// This tag doesn't exist, but is imported by github.com/portworx/sched-ops.
	github.com/kubernetes-incubator/external-storage v0.20.4-openstorage-rc2
	// Exclude pre-go-mod kubernetes tags, because they are older
	// than v0.x releases but are picked when updating dependencies.
	k8s.io/client-go v1.4.0
	k8s.io/client-go v1.5.0
	k8s.io/client-go v1.5.1
	k8s.io/client-go v1.5.2
	k8s.io/client-go v2.0.0-alpha.1+incompatible
	k8s.io/client-go v2.0.0+incompatible
	k8s.io/client-go v3.0.0-beta.0+incompatible
	k8s.io/client-go v3.0.0+incompatible
	k8s.io/client-go v4.0.0-beta.0+incompatible
	k8s.io/client-go v4.0.0+incompatible
	k8s.io/client-go v5.0.0+incompatible
	k8s.io/client-go v5.0.1+incompatible
	k8s.io/client-go v6.0.0+incompatible
	k8s.io/client-go v7.0.0+incompatible
	k8s.io/client-go v8.0.0+incompatible
	k8s.io/client-go v9.0.0-invalid+incompatible
	k8s.io/client-go v9.0.0+incompatible
	k8s.io/client-go v10.0.0+incompatible
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/client-go v12.0.0+incompatible
)

replace github.com/kube-object-storage/lib-bucket-provisioner => github.com/kube-object-storage/lib-bucket-provisioner v0.0.0-20220811192603-abc54caa34ac
