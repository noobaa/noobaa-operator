package util

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	semver "github.com/coreos/go-semver/semver"
	kedav1alpha1 "github.com/kedacore/keda/v2/apis/keda/v1alpha1"
	obv1 "github.com/kube-object-storage/lib-bucket-provisioner/pkg/apis/objectbucket.io/v1alpha1"
	nbapis "github.com/noobaa/noobaa-operator/v5/pkg/apis"
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	configv1 "github.com/openshift/api/config/v1"
	routev1 "github.com/openshift/api/route/v1"
	secv1 "github.com/openshift/api/security/v1"
	cloudcredsv1 "github.com/openshift/cloud-credential-operator/pkg/apis/cloudcredential/v1"
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	operv1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	cephv1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	apiregistration "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	"k8s.io/utils/ptr"
	cosiv1 "sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	oAuthWellKnownEndpoint = "https://openshift.default.svc/.well-known/oauth-authorization-server"
	ibmRegion              = "ibm-cloud.kubernetes.io/region"

	gigabyte             = 1024 * 1024 * 1024
	petabyte             = gigabyte * 1024 * 1024
	obcMaxSizeUpperLimit = petabyte * 1023

	topologyConstraintsEnabledKubeVersion = "1.26.0"
	trueStr                               = "true"
)

// OAuth2Endpoints holds OAuth2 endpoints information.
type OAuth2Endpoints struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
}

// ValidationError is a custom error if the validation failed
type ValidationError struct {
	Msg string
}

// AccessKeyRegexp validates access keys, which are 20 characters long and may include alphanumeric characters
var AccessKeyRegexp, _ = regexp.Compile(`^[a-zA-Z0-9]{20}$`)

// SecretKeyRegexp validates secret keys, which are 40 characters long and may include alphanumeric characters '+' and '/'
var SecretKeyRegexp, _ = regexp.Compile(`^[a-zA-Z0-9+/]{40}$`)

// IsValidationError check if err is of type ValidationError
func IsValidationError(err error) bool {
	_, ok := err.(ValidationError)
	return ok
}

// Error returns the ValidationError message
func (e ValidationError) Error() string {
	return e.Msg
}

var (
	ctx        = context.TODO()
	log        = logrus.WithContext(ctx)
	lazyConfig *rest.Config
	lazyClient client.Client

	// InsecureHTTPTransport is a global insecure http transport
	InsecureHTTPTransport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// GlobalCARefreshingTransport is a global secure http transport
	GlobalCARefreshingTransport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}

	// MapStorTypeToMandatoryProperties holds a map of store type -> credentials mandatory properties
	// note that this map holds the mandatory properties for both backingstores and namespacestores
	MapStorTypeToMandatoryProperties = map[string][]string{
		"aws-s3":               {"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"},         // backingstores and namespacestores
		"s3-compatible":        {"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"},         // backingstores and namespacestores
		"ibm-cos":              {"IBM_COS_ACCESS_KEY_ID", "IBM_COS_SECRET_ACCESS_KEY"}, // backingstores and namespacestores
		"google-cloud-storage": {"GoogleServiceAccountPrivateKeyJson"},                 // backingstores and namespacestores
		"azure-blob":           {"AccountName", "AccountKey"},                          // backingstores and namespacestores
		"pv-pool":              {},                                                     // backingstores
		"nsfs":                 {},                                                     // namespacestores
	}
)

// AddToRootCAs adds a local cert file to Our GlobalCARefreshingTransport
func AddToRootCAs(localCertFile string) error {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	var certFiles = []string{
		"/etc/ocp-injected-ca-bundle.crt",
		localCertFile,
	}

	for _, certFile := range certFiles {
		// Read in the cert file
		certs, err := os.ReadFile(certFile)
		if err != nil {
			return err
		}

		// Append our cert to the system pool
		if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
			log.Errorf("Failed to append %q to RootCAs", certFile)
			return fmt.Errorf("failed to append %q to RootCAs", certFile)
		}

		// Trust the augmented cert pool in our client
		log.Infof("Successfuly appended %q to RootCAs", certFile)
	}
	GlobalCARefreshingTransport.TLSClientConfig.RootCAs = rootCAs
	return nil
}

func init() {
	Panic(apiextv1.AddToScheme(scheme.Scheme))
	Panic(nbapis.AddToScheme(scheme.Scheme))
	Panic(obv1.AddToScheme(scheme.Scheme))
	Panic(cosiv1.AddToScheme(scheme.Scheme))
	Panic(monitoringv1.AddToScheme(scheme.Scheme))
	Panic(cloudcredsv1.AddToScheme(scheme.Scheme))
	Panic(operv1.AddToScheme(scheme.Scheme))
	Panic(cephv1.AddToScheme(scheme.Scheme))
	Panic(routev1.AddToScheme(scheme.Scheme))
	Panic(secv1.AddToScheme(scheme.Scheme))
	Panic(autoscalingv1.AddToScheme(scheme.Scheme))
	Panic(kedav1alpha1.AddToScheme(scheme.Scheme))
	Panic(apiregistration.AddToScheme(scheme.Scheme))
	Panic(configv1.AddToScheme(scheme.Scheme))
}

// KubeConfig loads kubernetes client config from default locations (flags, user dir, etc)
func KubeConfig() *rest.Config {
	if lazyConfig == nil {
		var err error
		lazyConfig, err = config.GetConfig()
		if err != nil {
			log.Fatalf("KubeConfig: %v", err)
		}
	}
	return lazyConfig
}

// GetKubeVersion will fetch the kubernates minor version
func GetKubeVersion() (*semver.Version, error) {
	var err error
	var discClient *discovery.DiscoveryClient
	var kubeVersionInfo *version.Info
	if discClient, err = discovery.NewDiscoveryClientForConfig(KubeConfig()); err != nil {
		return nil, err
	}
	if kubeVersionInfo, err = discClient.ServerVersion(); err != nil {
		return nil, err
	}
	kubeVersion := fmt.Sprintf("%s.%s.0", kubeVersionInfo.Major, kubeVersionInfo.Minor)
	version, err := semver.NewVersion(kubeVersion)
	return version, err
}

// MapperProvider creates RESTMapper
func MapperProvider(config *rest.Config, httpClient *http.Client) (meta.RESTMapper, error) {
	return meta.NewLazyRESTMapperLoader(func() (meta.RESTMapper, error) {
		dc, err := discovery.NewDiscoveryClientForConfig(config)
		if err != nil {
			return nil, err
		}
		return NewFastRESTMapper(dc, func(g *metav1.APIGroup) bool {
			if g.Name == "" ||
				g.Name == "apps" ||
				g.Name == "noobaa.io" ||
				g.Name == "objectbucket.io" ||
				g.Name == "operator.openshift.io" ||
				g.Name == "route.openshift.io" ||
				g.Name == "cloudcredential.openshift.io" ||
				g.Name == "security.openshift.io" ||
				g.Name == "monitoring.coreos.com" ||
				g.Name == "ceph.rook.io" ||
				g.Name == "autoscaling" ||
				g.Name == "batch" ||
				g.Name == "keda.sh" ||
				g.Name == "config.openshift.io" ||
				strings.HasSuffix(g.Name, ".k8s.io") {
				return true
			}
			return false
		}), nil
	}), nil
}

// KubeClient resturns a controller-runtime client
// We use a lazy mapper and a specialized implementation of fast mapper
// in order to avoid lags when running a CLI client to a far away cluster.
func KubeClient() client.Client {
	if lazyClient == nil {
		config := KubeConfig()
		mapper, _ := MapperProvider(config, nil)
		var err error
		lazyClient, err = client.New(config, client.Options{Mapper: mapper, Scheme: scheme.Scheme})
		if err != nil {
			log.Fatalf("KubeClient: %v", err)
		}
	}
	return lazyClient
}

// KubeObject loads a text yaml/json to a kubernetes object.
func KubeObject(text string) runtime.Object {
	// Decode text (yaml/json) to kube api object
	deserializer := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()
	obj, group, err := deserializer.Decode([]byte(text), nil, nil)
	// obj, group, err := scheme.Codecs.UniversalDecoder().Decode([]byte(text), nil, nil)
	Panic(err)
	// not sure if really needed, but set it anyway
	obj.GetObjectKind().SetGroupVersionKind(*group)
	SecretResetStringDataFromData(obj)
	return obj
}

// KubeApply will check if the object exists and will create/update accordingly
// and report the object status.
func KubeApply(obj client.Object) bool {
	klient := KubeClient()
	objKey := ObjectKey(obj)
	gvk := obj.GetObjectKind().GroupVersionKind()
	clone := obj.DeepCopyObject().(client.Object)
	err := klient.Get(ctx, objKey, clone)
	if err == nil {
		obj.SetResourceVersion(clone.GetResourceVersion())
		err = klient.Update(ctx, obj)
		if err == nil {
			log.Printf("✅ Updated: %s %q\n", gvk.Kind, objKey.Name)
			return false
		}
	}
	if meta.IsNoMatchError(err) || runtime.IsNotRegisteredError(err) {
		log.Printf("❌ CRD Missing: %s %q\n", gvk.Kind, objKey.Name)
		return false
	}
	if errors.IsNotFound(err) {
		err = klient.Create(ctx, obj)
		if err == nil {
			log.Printf("✅ Created: %s %q\n", gvk.Kind, objKey.Name)
			return true
		}
	}
	if errors.IsConflict(err) {
		log.Printf("❌ Conflict: %s %q: %s\n", gvk.Kind, objKey.Name, err)
		return false
	}
	statusErr, ok := err.(*errors.StatusError)
	if ok {
		log.Printf("❌ Status Error: %s %q: %s\n", gvk.Kind, objKey.Name, statusErr.ErrStatus.Message)
		return false
	}
	Panic(err)
	return false
}

// kubeCreateSkipOrFailExisting create k8s object,
// return true on success
// if the object exists return skipOrFail parameter
// return false on failure
func kubeCreateSkipOrFailExisting(obj client.Object, skipOrFail bool) bool {
	klient := KubeClient()
	objKey := ObjectKey(obj)
	gvk := obj.GetObjectKind().GroupVersionKind()
	clone := obj.DeepCopyObject().(client.Object)
	err := klient.Get(ctx, objKey, clone)
	if err == nil {
		log.Printf("✅ Already Exists: %s %q\n", gvk.Kind, objKey.Name)
		return skipOrFail
	}
	if meta.IsNoMatchError(err) || runtime.IsNotRegisteredError(err) {
		log.Printf("❌ CRD Missing: %s %q\n", gvk.Kind, objKey.Name)
		return false
	}
	if errors.IsNotFound(err) {
		err = klient.Create(ctx, obj)
		if err == nil {
			log.Printf("✅ Created: %s %q\n", gvk.Kind, objKey.Name)
			return true
		}
		if errors.IsNotFound(err) {
			log.Printf("❌ Namespace Missing: %s %q: kubectl create ns %s\n",
				gvk.Kind, objKey.Name, objKey.Namespace)
			return false
		}
	}
	if errors.IsConflict(err) {
		log.Printf("❌ Conflict: %s %q: %s\n", gvk.Kind, objKey.Name, err)
		return false
	}
	if errors.IsForbidden(err) {
		log.Printf("❌ Forbidden: %s %q: %s\n", gvk.Kind, objKey.Name, err)
		return false
	}
	if errors.IsInvalid(err) {
		log.Printf("❌ Invalid: %s %q: %s\n", gvk.Kind, objKey.Name, err)
		return false
	}
	statusErr, ok := err.(*errors.StatusError)
	if ok {
		log.Printf("❌ Status Error: %s %q: %s\n", gvk.Kind, objKey.Name, statusErr.ErrStatus.Message)
		return false
	}
	Panic(err)
	return false
}

// KubeCreateFailExisting will check if the object exists and will create/skip accordingly
// and report the object status.
func KubeCreateFailExisting(obj client.Object) bool {
	return kubeCreateSkipOrFailExisting(obj, false)
}

// KubeCreateSkipExisting will try to create an object
// returns true of the object exist or was created
// returns false otherwise
func KubeCreateSkipExisting(obj client.Object) bool {
	return kubeCreateSkipOrFailExisting(obj, true)
}

// KubeCreateOptional will check if the object exists and will create/skip accordingly
// It detects the situation of a missing CRD and reports it as an optional feature.
func KubeCreateOptional(obj client.Object) bool {
	klient := KubeClient()
	objKey := ObjectKey(obj)
	gvk := obj.GetObjectKind().GroupVersionKind()
	clone := obj.DeepCopyObject().(client.Object)
	err := klient.Get(ctx, objKey, clone)
	if err == nil {
		log.Printf("✅ Already Exists: %s %q\n", gvk.Kind, objKey.Name)
		return false
	}
	if meta.IsNoMatchError(err) || runtime.IsNotRegisteredError(err) {
		log.Printf("⬛ (Optional) CRD Unavailable: %s %q\n", gvk.Kind, objKey.Name)
		return false
	}
	if errors.IsNotFound(err) {
		err = klient.Create(ctx, obj)
		if err == nil {
			log.Printf("✅ Created: %s %q\n", gvk.Kind, objKey.Name)
			return true
		}
		if errors.IsNotFound(err) {
			log.Printf("❌ Namespace Missing: %s %q: kubectl create ns %s\n",
				gvk.Kind, objKey.Name, objKey.Namespace)
			return false
		}
	}
	if errors.IsConflict(err) {
		log.Printf("❌ Conflict: %s %q: %s\n", gvk.Kind, objKey.Name, err)
		return false
	}
	if errors.IsForbidden(err) {
		log.Printf("❌ Forbidden: %s %q: %s\n", gvk.Kind, objKey.Name, err)
		return false
	}
	if errors.IsInvalid(err) {
		log.Printf("❌ Invalid: %s %q: %s\n", gvk.Kind, objKey.Name, err)
		return false
	}
	Panic(err)
	return false
}

// KubeDelete deletes an object and reports the object status.
func KubeDelete(obj client.Object, opts ...client.DeleteOption) bool {
	klient := KubeClient()
	objKey := ObjectKey(obj)
	gvk := obj.GetObjectKind().GroupVersionKind()
	deleted := false
	conflicted := false

	err := klient.Delete(ctx, obj, opts...)
	if err == nil {
		deleted = true
		log.Printf("🗑️  Deleting: %s %q\n", gvk.Kind, objKey.Name)
	} else if errors.IsConflict(err) {
		conflicted = true
		log.Printf("🗑️  Conflict (OK): %s %q: %s\n", gvk.Kind, objKey.Name, err)
	} else if errors.IsNotFound(err) {
		return true
	}
	statusErr, ok := err.(*errors.StatusError)
	if ok {
		log.Printf("❌ Status Error: %s %q: %s\n", gvk.Kind, objKey.Name, statusErr.ErrStatus.Message)
		return false
	}

	time.Sleep(10 * time.Millisecond)

	err = wait.PollUntilContextCancel(ctx, time.Second, true, func(ctx context.Context) (bool, error) {
		err := klient.Delete(ctx, obj, opts...)
		if err == nil {
			if !deleted {
				deleted = true
				log.Printf("🗑️  Deleting: %s %q\n", gvk.Kind, objKey.Name)
			}
			return false, nil
		}
		if errors.IsConflict(err) {
			if !conflicted {
				conflicted = true
				log.Printf("🗑️  Conflict (OK): %s %q: %s\n", gvk.Kind, objKey.Name, err)
			}
			return false, nil
		}
		if meta.IsNoMatchError(err) || runtime.IsNotRegisteredError(err) {
			log.Printf("🗑️  CRD Missing (OK): %s %q\n", gvk.Kind, objKey.Name)
			return true, nil
		}
		if errors.IsNotFound(err) {
			log.Printf("🗑️  Deleted : %s %q\n", gvk.Kind, objKey.Name)
			return true, nil
		}
		return false, err
	})
	Panic(err)
	return deleted
}

// KubeDeleteNoPolling deletes an object without waiting for acknowledgement the object got deleted
func KubeDeleteNoPolling(obj client.Object, opts ...client.DeleteOption) bool {
	klient := KubeClient()
	objKey := ObjectKey(obj)
	gvk := obj.GetObjectKind().GroupVersionKind()
	deleted := false

	err := klient.Delete(ctx, obj, opts...)
	if err == nil {
		deleted = true
		log.Printf("🗑️  Deleting: %s %q\n", gvk.Kind, objKey.Name)
	} else if errors.IsConflict(err) {
		log.Printf("🗑️  Conflict (OK): %s %q: %s\n", gvk.Kind, objKey.Name, err)
	} else if errors.IsNotFound(err) {
		return true
	}

	Panic(err)
	return (deleted)
}

// KubeDeleteAllOf deletes an list of objects and reports the status.
func KubeDeleteAllOf(obj client.Object, opts ...client.DeleteAllOfOption) bool {
	klient := KubeClient()
	gvk := obj.GetObjectKind().GroupVersionKind()
	deleted := false

	err := klient.DeleteAllOf(ctx, obj, opts...)
	if err == nil {
		deleted = true
		log.Printf("🗑️  Deleting All of type: %s\n", gvk.Kind)
	} else if errors.IsNotFound(err) {
		return true
	}
	Panic(err)
	return deleted
}

// KubeUpdate updates an object and reports the object status.
func KubeUpdate(obj client.Object) bool {
	klient := KubeClient()
	objKey := ObjectKey(obj)
	gvk := obj.GetObjectKind().GroupVersionKind()
	err := klient.Update(ctx, obj)
	if err == nil {
		log.Printf("✅ Updated: %s %q\n", gvk.Kind, objKey.Name)
		return true
	}
	if meta.IsNoMatchError(err) || runtime.IsNotRegisteredError(err) {
		log.Printf("❌ CRD Missing: %s %q\n", gvk.Kind, objKey.Name)
		return false
	}
	if errors.IsConflict(err) {
		log.Printf("❌ Conflict: %s %q: %s\n", gvk.Kind, objKey.Name, err)
		return false
	}
	if errors.IsNotFound(err) {
		log.Printf("❌ Not Found: %s %q\n", gvk.Kind, objKey.Name)
		return false
	}
	statusErr, ok := err.(*errors.StatusError)
	if ok {
		log.Printf("❌ Status Error: %s %q: %s\n", gvk.Kind, objKey.Name, statusErr.ErrStatus.Message)
		return false
	}
	Panic(err)
	return false
}

// KubeCheck checks if the object exists and reports the object status.
func KubeCheck(obj client.Object) bool {
	name, kind, err := KubeGet(obj)
	if err == nil {
		SecretResetStringDataFromData(obj)
		log.Printf("✅ Exists: %s %q\n", kind, name)
		return true
	}
	if meta.IsNoMatchError(err) || runtime.IsNotRegisteredError(err) {
		log.Printf("❌ CRD Missing: %s %q\n", kind, name)
		return false
	}
	if errors.IsNotFound(err) {
		log.Printf("❌ Not Found: %s %q\n", kind, name)
		return false
	}
	if errors.IsConflict(err) {
		log.Printf("❌ Conflict: %s %q: %s\n", kind, name, err)
		return false
	}
	Panic(err)
	return false
}

// KubeCheckOptional checks if the object exists and reports the object status.
// It detects the situation of a missing CRD and reports it as an optional feature.
func KubeCheckOptional(obj client.Object) bool {
	name, kind, err := KubeGet(obj)
	if err == nil {
		log.Printf("✅ (Optional) Exists: %s %q\n", kind, name)
		return true
	}
	if meta.IsNoMatchError(err) || runtime.IsNotRegisteredError(err) {
		log.Printf("⬛ (Optional) CRD Unavailable: %s %q\n", kind, name)
		return false
	}
	if errors.IsNotFound(err) {
		log.Printf("⬛ (Optional) Not Found: %s %q\n", kind, name)
		return false
	}
	if errors.IsConflict(err) {
		log.Printf("❌ (Optional) Conflict: %s %q: %s\n", kind, name, err)
		return false
	}
	Panic(err)
	return false
}

// KubeCheckQuiet checks if the object exists fills the given object if found.
// returns true if the object was found. It does not print any status
func KubeCheckQuiet(obj client.Object) bool {
	_, _, err := KubeGet(obj)
	if err == nil {
		return true
	}
	if meta.IsNoMatchError(err) ||
		runtime.IsNotRegisteredError(err) ||
		errors.IsNotFound(err) ||
		errors.IsConflict(err) {
		return false
	}
	Panic(err)
	return false
}

// KubeGet gets a client.Object, fills the given object and returns the name and kind
// returns error on failure
func KubeGet(obj client.Object) (name string, kind string, err error) {
	klient := KubeClient()
	objKey := ObjectKey(obj)
	name = objKey.Name
	gvk := obj.GetObjectKind().GroupVersionKind()
	kind = gvk.Kind
	err = klient.Get(ctx, objKey, obj)
	return name, kind, err
}

// KubeList returns a list of objects.
func KubeList(list client.ObjectList, options ...client.ListOption) bool {
	klient := KubeClient()
	gvk := list.GetObjectKind().GroupVersionKind()
	err := klient.List(ctx, list, options...)
	if err == nil {
		return true
	}
	if meta.IsNoMatchError(err) || runtime.IsNotRegisteredError(err) {
		log.Printf("❌ CRD Missing: %s: %s\n", gvk.Kind, err)
		return false
	}
	if errors.IsNotFound(err) {
		log.Printf("❌ Not Found: %s\n", gvk.Kind)
		return false
	}
	if errors.IsConflict(err) {
		log.Printf("❌ Conflict: %s: %s\n", gvk.Kind, err)
		return false
	}
	Panic(err)
	return false
}

// RemoveFinalizer modifies the object and removes the finalizer
func RemoveFinalizer(obj metav1.Object, finalizer string) bool {
	finalizers := obj.GetFinalizers()
	if finalizers == nil {
		return false
	}
	// see https://github.com/golang/go/wiki/SliceTricks#filter-in-place
	n := 0
	found := false
	for _, f := range finalizers {
		if f == finalizer {
			found = true
		} else {
			finalizers[n] = f
			n++
		}
	}
	finalizers = finalizers[:n]
	obj.SetFinalizers(finalizers)
	return found
}

// AddFinalizer adds the finalizer to the object if it doesn't contains it already
func AddFinalizer(obj metav1.Object, finalizer string) bool {
	if !Contains(obj.GetFinalizers(), finalizer) {
		finalizers := append(obj.GetFinalizers(), finalizer)
		obj.SetFinalizers(finalizers)
		return true
	}
	return false
}

// GetPodStatusLine returns a one liner status for a pod
func GetPodStatusLine(pod *corev1.Pod) string {
	s := fmt.Sprintf("Phase=%q. ", pod.Status.Phase)
	for i := range pod.Status.Conditions {
		c := &pod.Status.Conditions[i]
		if c.Status != corev1.ConditionTrue {
			s += fmt.Sprintf("%s (%s). ", c.Reason, c.Message)
		}
	}
	return s
}

// GetPodLogs info
func GetPodLogs(pod corev1.Pod) (map[string]io.ReadCloser, error) {
	allContainers := append(pod.Spec.InitContainers, pod.Spec.Containers...)
	config := KubeConfig()
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf(`Could not create %s, reason: %s\n`, pod.Name, err)
		return nil, err
	}

	containerMap := make(map[string]io.ReadCloser)
	for _, container := range allContainers {
		getPodLogOpts := func(pod *corev1.Pod, logOpts *corev1.PodLogOptions, nameSuffix *string) {
			req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, logOpts)

			podLogs, err := req.Stream(ctx)
			if err != nil {
				// do not warn about lack of additional previous logs, i.e. when nameSuffix is set
				if nameSuffix == nil {
					log.Printf(`Could not read logs %s container %s, reason: %s`, pod.Name, container.Name, err)
				}
				return
			}
			mapKey := container.Name
			if nameSuffix != nil {
				mapKey += "-" + *nameSuffix
			}
			containerMap[mapKey] = podLogs
		}

		// retrieve logs from a current instantiation of pod containers
		podLogOpts := corev1.PodLogOptions{Container: container.Name}
		getPodLogOpts(&pod, &podLogOpts, nil)

		// retrieve logs from a previous instantiation of pod containers
		prevPodLogOpts := corev1.PodLogOptions{Container: container.Name, Previous: true}
		previousSuffix := "previous"
		getPodLogOpts(&pod, &prevPodLogOpts, &previousSuffix)
	}
	return containerMap, nil
}

// SaveStreamToFile info
func SaveStreamToFile(body io.ReadCloser, path string) error {
	if body == nil {
		return nil
	}
	defer body.Close()
	f, err := os.Create(path)
	if err != nil {
		log.Errorf(`Could not save stream to file: %s, reason: %s\n`, path, err)
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, body); err != nil {
		log.Errorf(`Could not write to file: %s, reason: %s\n`, path, err)
		return err
	}
	return nil
}

// SaveCRsToFile info
func SaveCRsToFile(crs runtime.Object, path string) error {
	if crs == nil {
		return nil
	}

	f, err := os.Create(path)
	if err != nil {
		log.Printf(`Could not create file %s, reason: %s\n`, path, err)
		return err
	}

	defer f.Close()
	p := printers.YAMLPrinter{}
	err = p.PrintObj(crs, f)
	if err != nil {
		log.Printf(`Could not write yaml to file %s, reason: %s\n`, path, err)
		return err
	}

	return nil
}

// GetContainerStatusLine returns a one liner status for a container
func GetContainerStatusLine(cont *corev1.ContainerStatus) string {
	s := ""
	if cont.RestartCount > 0 {
		s += fmt.Sprintf("RestartCount=%d. ", cont.RestartCount)
	}
	if cont.State.Waiting != nil {
		s += fmt.Sprintf("%s (%s). ", cont.State.Waiting.Reason, cont.State.Waiting.Message)
	}
	if cont.State.Terminated != nil {
		s += fmt.Sprintf("%s (%s). ", cont.State.Terminated.Reason, cont.State.Terminated.Message)
	}
	if s == "" {
		s = "starting..."
	}
	return s
}

// Panic is conveniently calling panic only if err is not nil
func Panic(err error) {
	if err != nil {
		reason := errors.ReasonForError(err)
		log.Panicf("☠️  Panic Attack: [%s] %s", reason, err)
	}
}

// LogError prints the error to the log and continue
func LogError(err error) {
	if err != nil {
		log.Errorf("Error discovered: %s", err)
	}
}

// IgnoreError do nothing if err is not nil
func IgnoreError(err error) {
}

// InitLogger initializes the logrus logger with defaults
func InitLogger(lvl logrus.Level) {
	logrus.SetLevel(lvl)
	logrus.SetFormatter(&logrus.TextFormatter{
		// FullTimestamp: true,
	})
}

// Logger returns a default logger
func Logger() *logrus.Entry {
	return log
}

// Context returns a default Context
func Context() context.Context {
	return ctx
}

// CurrentNamespace reads the current namespace from the kube config
func CurrentNamespace() string {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	ns, _, _ := kubeConfig.Namespace()
	// ignoring errors and just return empty string if config is missing or invalid
	return ns
}

// PersistentError is an error type that tells the reconcile to avoid requeueing.
type PersistentError struct {
	Reason  string
	Message string
}

// Error function makes PersistentError implement error interface
func (e *PersistentError) Error() string { return e.Message }

// assert implement error interface
var _ error = &PersistentError{}

// NewPersistentError returns a new persistent error.
func NewPersistentError(reason string, message string) *PersistentError {
	return &PersistentError{Reason: reason, Message: message}
}

// IsPersistentError checks if the provided error is persistent.
func IsPersistentError(err error) bool {
	_, isPersistent := err.(*PersistentError)
	return isPersistent
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

// SecretResetStringDataFromData reads the secret data into string data
// to streamline the paths that use the secret values as strings.
func SecretResetStringDataFromData(obj runtime.Object) {
	secret, isSecret := obj.(*corev1.Secret)
	if !isSecret {
		return
	}
	secret.StringData = map[string]string{}
	for key, val := range secret.Data {
		secret.StringData[key] = string(val)
	}
	secret.Data = map[string][]byte{}
}

// ObjectKey returns the objects key (namespace + name)
func ObjectKey(obj client.Object) client.ObjectKey {
	objKey := client.ObjectKeyFromObject(obj)
	return objKey
}

// RandomBase64 creates a random buffer with numBytes and returns it encoded in base64
// Returned string length is 4*math.Ceil(numBytes/3)
func RandomBase64(numBytes int) string {
	randomBytes := make([]byte, numBytes)
	_, err := rand.Read(randomBytes)
	Panic(err)
	return base64.StdEncoding.EncodeToString(randomBytes)
}

// RandomHex creates a random buffer with numBytes and returns it encoded in hex
// Returned string length is 2*numBytes
func RandomHex(numBytes int) string {
	randomBytes := make([]byte, numBytes)
	_, err := rand.Read(randomBytes)
	Panic(err)
	return hex.EncodeToString(randomBytes)
}

// SetAvailableCondition updates the status conditions to available state
func SetAvailableCondition(conditions *[]conditionsv1.Condition, reason string, message string) {
	currentTime := metav1.NewTime(time.Now())
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		LastHeartbeatTime: currentTime,
		Type:              conditionsv1.ConditionAvailable,
		Status:            corev1.ConditionTrue,
		Reason:            reason,
		Message:           message,
	})
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		LastHeartbeatTime: currentTime,
		Type:              conditionsv1.ConditionProgressing,
		Status:            corev1.ConditionFalse,
		Reason:            reason,
		Message:           message,
	})
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		LastHeartbeatTime: currentTime,
		Type:              conditionsv1.ConditionDegraded,
		Status:            corev1.ConditionFalse,
		Reason:            reason,
		Message:           message,
	})
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		LastHeartbeatTime: currentTime,
		Type:              conditionsv1.ConditionUpgradeable,
		Status:            corev1.ConditionTrue,
		Reason:            reason,
		Message:           message,
	})
}

// SetProgressingCondition updates the status conditions to in-progress state
func SetProgressingCondition(conditions *[]conditionsv1.Condition, reason string, message string) {
	currentTime := metav1.NewTime(time.Now())
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		LastHeartbeatTime: currentTime,
		Type:              conditionsv1.ConditionAvailable,
		Status:            corev1.ConditionFalse,
		Reason:            reason,
		Message:           message,
	})
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		LastHeartbeatTime: currentTime,
		Type:              conditionsv1.ConditionProgressing,
		Status:            corev1.ConditionTrue,
		Reason:            reason,
		Message:           message,
	})
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		LastHeartbeatTime: currentTime,
		Type:              conditionsv1.ConditionDegraded,
		Status:            corev1.ConditionFalse,
		Reason:            reason,
		Message:           message,
	})
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		LastHeartbeatTime: currentTime,
		Type:              conditionsv1.ConditionUpgradeable,
		Status:            corev1.ConditionFalse,
		Reason:            reason,
		Message:           message,
	})
}

// SetErrorCondition updates the status conditions to error state
func SetErrorCondition(conditions *[]conditionsv1.Condition, reason string, message string) {
	currentTime := metav1.NewTime(time.Now())
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		//LastHeartbeatTime should be set by the custom-resource-status just like lastTransitionTime
		// Setting it here temporarity
		LastHeartbeatTime: currentTime,
		Type:              conditionsv1.ConditionAvailable,
		Status:            corev1.ConditionUnknown,
		Reason:            reason,
		Message:           message,
	})
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		LastHeartbeatTime: currentTime,
		Type:              conditionsv1.ConditionProgressing,
		Status:            corev1.ConditionFalse,
		Reason:            reason,
		Message:           message,
	})
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		LastHeartbeatTime: currentTime,
		Type:              conditionsv1.ConditionDegraded,
		Status:            corev1.ConditionTrue,
		Reason:            reason,
		Message:           message,
	})
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		LastHeartbeatTime: currentTime,
		Type:              conditionsv1.ConditionUpgradeable,
		Status:            corev1.ConditionUnknown,
		Reason:            reason,
		Message:           message,
	})
}

// IsAWSPlatform returns true if this cluster is running on AWS
func IsAWSPlatform() bool {
	nodesList := &corev1.NodeList{}
	if ok := KubeList(nodesList); !ok || len(nodesList.Items) == 0 {
		Panic(fmt.Errorf("failed to list kubernetes nodes"))
	}
	isAWS := strings.HasPrefix(nodesList.Items[0].Spec.ProviderID, "aws")
	return isAWS
}

// IsFusionHCIWithScale checks if the noobaa is deployed on HCI platform and
// using Spectrum Scale storage.
func IsFusionHCIWithScale() bool {
	sc := &storagev1.StorageClass{
		TypeMeta:   metav1.TypeMeta{Kind: "StorageClass"},
		ObjectMeta: metav1.ObjectMeta{Name: "ibm-spectrum-scale-csi-storageclass-version2"},
	}
	return KubeCheck(sc)
}

// IsSTSClusterBS returns true if it is running on an STS cluster
func IsSTSClusterBS(bs *nbv1.BackingStore) bool {
	if bs.Spec.Type == nbv1.StoreTypeAWSS3 {
		return bs.Spec.AWSS3.AWSSTSRoleARN != nil
	}
	return false
}

// IsSTSClusterNS returns true if it is running on an STS cluster
func IsSTSClusterNS(ns *nbv1.NamespaceStore) bool {
	if ns.Spec.Type == nbv1.NSStoreTypeAWSS3 {
		return ns.Spec.AWSS3.AWSSTSRoleARN != nil
	}
	return false
}

// IsAzurePlatformNonGovernment returns true if this cluster is running on Azure and also not on azure government\DOD cloud
func IsAzurePlatformNonGovernment() bool {
	nodesList := &corev1.NodeList{}
	if ok := KubeList(nodesList); !ok || len(nodesList.Items) == 0 {
		Panic(fmt.Errorf("failed to list kubernetes nodes"))
	}
	const regionLabel string = "topology.kubernetes.io/region"
	node := nodesList.Items[0]
	isAzure := strings.HasPrefix(node.Spec.ProviderID, "azure")
	if isAzure {
		nodeLabels := node.GetLabels()
		region, ok := nodeLabels[regionLabel]
		if !ok {
			log.Warnf("did not find the expected label %q on node %q to determine azure region", regionLabel, node.Name)
		} else if strings.HasPrefix(region, "usgov") || strings.HasPrefix(region, "usdod") {
			log.Infof("identified the region [%q] as an Azure gov/DOD region", region)
			return false
		}
	}
	return isAzure
}

// IsGCPPlatform returns true if this cluster is running on GCP
func IsGCPPlatform() bool {
	nodesList := &corev1.NodeList{}
	if ok := KubeList(nodesList); !ok || len(nodesList.Items) == 0 {
		Panic(fmt.Errorf("failed to list kubernetes nodes"))
	}
	isGCP := strings.HasPrefix(nodesList.Items[0].Spec.ProviderID, "gce")
	return isGCP
}

// IsIBMPlatform returns true if this cluster is running on IBM Cloud
func IsIBMPlatform() bool {
	nodesList := &corev1.NodeList{}
	if ok := KubeList(nodesList); !ok || len(nodesList.Items) == 0 {
		Panic(fmt.Errorf("failed to list kubernetes nodes"))
	}
	isIBM := strings.HasPrefix(nodesList.Items[0].Spec.ProviderID, "ibm")
	if isIBM {
		// In case of Satellite cluster is deployed in user provided infrastructure
		if strings.Contains(nodesList.Items[0].Spec.ProviderID, "/sat-") {
			isIBM = false
		}
	}
	return isIBM
}

// GetIBMRegion returns the cluster's region in IBM Cloud
func GetIBMRegion() (string, error) {
	nodesList := &corev1.NodeList{}
	if ok := KubeList(nodesList, client.HasLabels{ibmRegion}); !ok {
		return "", fmt.Errorf("failed to list Kubernetes nodes with the IBM region label")
	}
	if len(nodesList.Items) == 0 {
		return "", fmt.Errorf("no Kubernetes nodes with the IBM region label found")
	}
	labels := nodesList.Items[0].GetLabels()
	region := labels[ibmRegion]
	return region, nil
}

// GetAWSRegion determines the AWS region from cluster infrastructure or node name
func GetAWSRegion() (string, error) {
	// Determine the AWS region based on cluster infrastructure or node name
	// If infrastructure details are unavailable, the node's name is parsed to extract the region
	// Refer to the following for more details:
	// https://docs.aws.amazon.com/en_pv/vpc/latest/userguide/vpc-dns.html#vpc-dns-hostnames
	// The list of regions can be found here:
	// https://docs.aws.amazon.com/general/latest/gr/rande.html
	var mapValidAWSRegions = map[string]string{
		"compute-1":      "us-east-1",
		"ec2":            "us-east-1",
		"us-east-1":      "us-east-1",
		"us-east-2":      "us-east-2",
		"us-west-1":      "us-west-1",
		"us-west-2":      "us-west-2",
		"ca-central-1":   "ca-central-1",
		"ca-west-1":      "ca-west-1",
		"eu-central-1":   "eu-central-1",
		"eu-central-2":   "eu-central-2",
		"eu-west-1":      "eu-west-1",
		"eu-west-2":      "eu-west-2",
		"eu-west-3":      "eu-west-3",
		"eu-north-1":     "eu-north-1",
		"eu-south-1":     "eu-south-1",
		"eu-south-2":     "eu-south-2",
		"ap-east-1":      "ap-east-1",
		"ap-northeast-1": "ap-northeast-1",
		"ap-northeast-2": "ap-northeast-2",
		"ap-northeast-3": "ap-northeast-3",
		"ap-southeast-1": "ap-southeast-1",
		"ap-southeast-2": "ap-southeast-2",
		"ap-southeast-3": "ap-southeast-3",
		"ap-southeast-4": "ap-southeast-4",
		"ap-southeast-5": "ap-southeast-5",
		"ap-south-1":     "ap-south-1",
		"ap-south-2":     "ap-south-2",
		"me-south-1":     "me-south-1",
		"me-central-1":   "me-central-1",
		"sa-east-1":      "sa-east-1",
		"us-gov-west-1":  "us-gov-west-1",
		"us-gov-east-1":  "us-gov-east-1",
		"af-south-1":     "af-south-1",
		"il-central-1":   "il-central-1",
	}
	var awsRegion string
	infrastructure := &configv1.Infrastructure{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
	}
	if _, _, err := KubeGet(infrastructure); err != nil {
		log.Infof("Failed to fetch cluster infrastructure details: %v", err)
	}
	if infrastructure.Status.PlatformStatus != nil && infrastructure.Status.PlatformStatus.AWS != nil {
		awsRegion = mapValidAWSRegions[infrastructure.Status.PlatformStatus.AWS.Region]
	}

	// Parsing the aws region from node name if not fetched from cluster
	log.Warn("Falling back to parsing the node name as infrastructure details are unavailable or incomplete")
	if awsRegion == "" {
		nodesList := &corev1.NodeList{}
		if ok := KubeList(nodesList); !ok || len(nodesList.Items) == 0 {
			log.Infof("Failed to list kubernetes nodes")
		}
		nameSplit := strings.Split(nodesList.Items[0].Name, ".")
		if len(nameSplit) < 2 {
			log.Infof("Unexpected node name format: %q", nodesList.Items[0].Name)
		}
		awsRegion = mapValidAWSRegions[nameSplit[1]]
	}

	// returning error if not fetched from either cluster or node name
	if awsRegion == "" {
		return "", fmt.Errorf("Failed to determine the AWS Region.")
	}
	return awsRegion, nil
}

// IsValidS3BucketName checks the name according to
// https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-s3-bucket-naming-requirements.html
func IsValidS3BucketName(name string) bool {
	validBucketNameRegex := regexp.MustCompile(`^(([a-z0-9]|[a-z0-9][a-z0-9-]*[a-z0-9])\.)*([a-z0-9]|[a-z0-9][a-z0-9-]*[a-z0-9])$`)
	return validBucketNameRegex.MatchString(name)
}

// GetFlagIntOrPrompt returns flag value but if empty will prompt to read from stdin
func GetFlagIntOrPrompt(cmd *cobra.Command, flag string) int {
	val, _ := cmd.Flags().GetInt(flag)
	if val != -1 {
		return val
	}
	fmt.Printf("Enter %s: ", flag)
	_, err := fmt.Scan(&val)

	if err != nil {
		if strings.Contains(err.Error(), "expected integer") {
			log.Fatalf(`❌ The flag %s must be an integer`, flag)
		}
	}

	Panic(err)
	if val == -1 {
		log.Fatalf(`❌ Missing %s %s`, flag, cmd.UsageString())
	}
	return val
}

// GetFlagStringOrPrompt returns flag value but if empty will prompt to read from stdin
func GetFlagStringOrPrompt(cmd *cobra.Command, flag string) string {
	str, _ := cmd.Flags().GetString(flag)
	if str != "" {
		return str
	}
	fmt.Printf("Enter %s: ", flag)
	_, err := fmt.Scan(&str)
	Panic(err)
	if str == "" {
		log.Fatalf(`❌ Missing %s %s`, flag, cmd.UsageString())
	}
	return str
}

// GetFlagStringOrPromptPassword is like GetFlagStringOrPrompt
// but does not show the input characters on the terminal
// to avoid leaking secret data in shell history
func GetFlagStringOrPromptPassword(cmd *cobra.Command, flag string) string {
	str, _ := cmd.Flags().GetString(flag)
	if str != "" {
		return str
	}
	fmt.Printf("Enter %s: ", flag)
	bytes, err := term.ReadPassword(0)
	Panic(err)
	str = string(bytes)
	fmt.Printf("[got %d characters]\n", len(str))
	return str
}

// GetBoolFlagPtr returns a pointer to the boolean flag value if set, or nil if not set
func GetBoolFlagPtr(cmd *cobra.Command, flag string) (*bool, error) {
	if cmd.Flags().Changed(flag) {
		flagVal, err := cmd.Flags().GetBool(flag)
		if err != nil {
			return nil, err
		}
		return &flagVal, nil
	}

	return nil, nil
}

// PrintThisNoteWhenFinishedApplyingAndStartWaitLoop is a common log task
func PrintThisNoteWhenFinishedApplyingAndStartWaitLoop() {
	log := Logger()
	log.Printf("NOTE:")
	log.Printf("  - This command has finished applying changes to the cluster.")
	log.Printf("  - From now on, it only loops and reads the status, to monitor the operator work.")
	log.Printf("  - You may Ctrl-C at any time to stop the loop and watch it manually.")
}

// DiscoverOAuthEndpoints uses a well known url to get info on the cluster oauth2 endpoints
func DiscoverOAuthEndpoints() (*OAuth2Endpoints, error) {
	client := http.Client{
		Timeout:   120 * time.Second,
		Transport: InsecureHTTPTransport,
	}

	res, err := client.Get(oAuthWellKnownEndpoint)
	defer func() {
		if res != nil && res.Body != nil {
			res.Body.Close()
		}
	}()
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	endpoints := OAuth2Endpoints{}
	err = json.Unmarshal(body, &endpoints)
	if err != nil {
		return nil, err
	}

	return &endpoints, nil
}

// IsStringGraphicOrSpacesCharsOnly returns true only if all the chars are graphic or spaces
func IsStringGraphicOrSpacesCharsOnly(s string) bool {
	for _, c := range s {
		if !unicode.IsGraphic(c) && !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}

// VerifyCredsInSecret throws fatal error when a given secret doesn't contain the mandatory properties
func VerifyCredsInSecret(secretName string, namespace string, mandatoryProperties []string) {
	secret := KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret)
	secret.Name = secretName
	secret.Namespace = namespace
	if !KubeCheck(secret) {
		log.Fatalf("secret %q does not exist", secretName)
	}
	for _, p := range mandatoryProperties {
		cred, ok := secret.StringData[p]
		if cred == "" || !ok {
			log.Fatalf("❌ secret %q does not contain property %q", secret.Name, p)
		}
	}
}

// Tar takes a source and variable writers and walks 'source' writing each file
// found to the tar writer; the purpose for accepting multiple writers is to allow
// for multiple outputs (for example a file, or md5 hash)
func Tar(src string, writers ...io.Writer) error {
	// ensure the src actually exists before trying to tar it
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("Unable to tar files - %v", err.Error())
	}

	mw := io.MultiWriter(writers...)

	gzw := gzip.NewWriter(mw)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// walk path
	return filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {

		// return on any error
		if err != nil {
			return err
		}

		// return on non-regular files (thanks to [kumo](https://medium.com/@komuw/just-like-you-did-fbdd7df829d3) for this suggested update)
		if !fi.Mode().IsRegular() {
			return nil
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		// update the name to correctly reflect the desired destination when untaring
		header.Name = strings.TrimPrefix(strings.Replace(file, src, "", -1), string(filepath.Separator))

		// write the header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// open files for taring
		f, err := os.Open(file)
		if err != nil {
			return err
		}

		defer f.Close()

		// copy file data into tar writer
		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		return nil
	})
}

// WriteYamlFile writes a yaml file from the given objects
func WriteYamlFile(name string, obj runtime.Object, moreObjects ...runtime.Object) error {
	p := printers.YAMLPrinter{}

	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()

	err = p.PrintObj(obj, file)
	if err != nil {
		return err
	}

	for i := range moreObjects {
		err = p.PrintObj(moreObjects[i], file)
		if err != nil {
			return err
		}
	}

	return nil
}

// Contains is a generic function
// that receives a slice and an element from comparable type (sould be from the same type)
// and returns true if the slice contains the element, otherwise false
func Contains[T comparable](arr []T, item T) bool {
	return ContainsAny(
		arr,
		item,
		func(a, b T) bool {
			return a == b
		},
	)
}

// ContainsAny is a generic function
// that receives a slice, an element and a function to compare
// and returns true if the function executed on the item and the slice return true
func ContainsAny[T any](arr []T, item T, eq func(T, T) bool) bool {
	for _, element := range arr {
		if eq(element, item) {
			return true
		}
	}

	return false
}

// GetEnvVariable is looking for env variable called name in env and return a pointer to the variable
func GetEnvVariable(env *[]corev1.EnvVar, name string) *corev1.EnvVar {
	for i := range *env {
		e := &(*env)[i]
		if e.Name == name {
			return e
		}
	}
	return nil
}

// GetAnnotationValue searches for an annotation within a map of strings and returns if it exists and what its value
func GetAnnotationValue(annotations map[string]string, name string) (string, bool) {
	if annotations != nil {
		val, exists := annotations[name]
		return val, exists
	}
	return "", false
}

// IsRemoteClientNoobaa checks for the existance and value of the remote-client-noobaa annotation
// within an annotation map, if the annotation doesnt exist it's the same as if its value is false.
func IsRemoteClientNoobaa(annotations map[string]string) bool {
	annotationValue, exists := GetAnnotationValue(annotations, "remote-client-noobaa")
	annotationBoolVal := false
	if exists {
		annotationBoolVal = strings.ToLower(annotationValue) == trueStr
	}
	return annotationBoolVal
}

// ReflectEnvVariable will add, update or remove an env variable base on the existence and value of an
// env variable with the same name on the container running this function.
func ReflectEnvVariable(env *[]corev1.EnvVar, name string) {
	if val, ok := os.LookupEnv(name); ok {
		envVar := GetEnvVariable(env, name)
		if envVar != nil {
			envVar.Value = val
		} else {
			*env = append(*env, corev1.EnvVar{
				Name:  name,
				Value: val,
			})
		}
	} else {
		for i, envVar := range *env {
			if envVar.Name == name {
				*env = append((*env)[:i], (*env)[i+1:]...)
				return
			}
		}
	}
}

// MergeVolumeList takes two Volume arrays and merge them into the first
func MergeVolumeList(existing, template *[]corev1.Volume) {
	existingElements := make(map[string]bool)

	for _, item := range *existing {
		existingElements[item.Name] = true
	}

	for _, item := range *template {
		if !existingElements[item.Name] {
			*existing = append(*existing, item)
		}
	}
}

// MergeVolumeMountList takes two VolumeMount arrays and merge them into the first
func MergeVolumeMountList(existing, template *[]corev1.VolumeMount) {
	existingElements := make(map[string]bool)

	for _, item := range *existing {
		existingElements[item.Name] = true
	}

	for _, item := range *template {
		if !existingElements[item.Name] {
			*existing = append(*existing, item)
		}
	}
}

// GetCmDataHash calculates a Hash string representing an array of key value strings
func GetCmDataHash(input map[string]string) string {
	b := new(bytes.Buffer)

	keys := make([]string, 0, len(input))
	for k := range input {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		// Convert each key/value pair in the map to a string
		fmt.Fprintf(b, "%s=\"%s\"\n", k, input[k])
	}

	sha256Bytes := sha256.Sum256(b.Bytes())
	sha256Hex := hex.EncodeToString(sha256Bytes[:])
	return sha256Hex
}

// MergeEnvArrays takes two Env variables arrays and merge them into the first
func MergeEnvArrays(envA, envB *[]corev1.EnvVar) {
	existingEnvs := make(map[string]bool)

	for _, item := range *envA {
		existingEnvs[item.Name] = true
	}

	for _, item := range *envB {
		if !existingEnvs[item.Name] {
			*envA = append(*envA, item)
		}
	}
}

// HumanizeDuration humanizes time.Duration output to a meaningful value - will show days/years
func HumanizeDuration(duration time.Duration) string {
	const (
		oneDay  = time.Minute * 60 * 24
		oneYear = 365 * oneDay
	)
	if duration < oneDay {
		return duration.String()
	}

	var builder strings.Builder

	if duration >= oneYear {
		years := duration / oneYear
		fmt.Fprintf(&builder, "%dy", years)
		duration -= years * oneYear
	}

	days := duration / oneDay
	duration -= days * oneDay
	fmt.Fprintf(&builder, "%dd%s", days, duration)

	return builder.String()
}

// IsStringArrayUnorderedEqual checks if two string arrays has the same members
func IsStringArrayUnorderedEqual(stringsArrayA, stringsArrayB []string) bool {
	if len(stringsArrayA) != len(stringsArrayB) {
		return false
	}
	existingStrings := make(map[string]bool)
	for _, item := range stringsArrayA {
		existingStrings[item] = true
	}
	for _, item := range stringsArrayB {
		if !existingStrings[item] {
			return false
		}
	}
	return true
}

// EnsureCommonMetaFields ensures that the resource has all mandatory meta fields
func EnsureCommonMetaFields(object metav1.Object, finalizer string) bool {
	updated := false

	labels := object.GetLabels()
	if labels == nil {
		object.SetLabels(map[string]string{
			"app": "noobaa",
		})
		updated = true

	} else if labels["app"] != "noobaa" {
		labels["app"] = "noobaa"
		object.SetLabels(labels)
		updated = true

	}
	if AddFinalizer(object, finalizer) {
		updated = true

	}
	return updated
}

// GetWatchNamespace returns the namespace the operator should be watching for changes
// this was implemented in operator-sdk v0.17 and removed. copied from here:
// https://github.com/operator-framework/operator-sdk/blob/53b00d125fb12515cd74fb169149913b401c8995/pkg/k8sutil/k8sutil.go#L45
func GetWatchNamespace() (string, error) {
	const WatchNamespaceEnvVar = "WATCH_NAMESPACE"
	ns, found := os.LookupEnv(WatchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", WatchNamespaceEnvVar)
	}

	return ns, nil
}

// DeleteStorageClass deletes storage class
func DeleteStorageClass(sc *storagev1.StorageClass) error {
	log.Infof("storageclass %v found, deleting..", sc.Name)
	if !KubeDelete(sc) {
		return fmt.Errorf("storageclass %q failed to delete", sc)
	}
	log.Infof("deleted storageclass %v successfully", sc.Name)
	return nil
}

// LoadConfigurationJSON loads the bucket replication from a json file
func LoadConfigurationJSON(configurationJSONPath string) (string, error) {

	logrus.Infof("loading JSON configuration file %v", configurationJSONPath)
	bytes, err := os.ReadFile(configurationJSONPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %q: %v", configurationJSONPath, err)
	}
	var configurationJSON interface{}
	err = json.Unmarshal(bytes, &configurationJSON)
	if err != nil {
		return "", fmt.Errorf("failed to parse JSON file %q: %v", configurationJSONPath, err)
	}

	logrus.Infof("✅ Successfully loaded JSON configuration %v", string(bytes))

	return string(bytes), nil
}

// NoobaaStatus returns true if NooBaa condition type and status matches
// returns false otherwise
func NoobaaStatus(nb *nbv1.NooBaa, t conditionsv1.ConditionType, status corev1.ConditionStatus) bool {
	for _, cond := range nb.Status.Conditions {
		log.Printf("condition type %v status %v", cond.Type, cond.Status)
		if cond.Type == t && cond.Status == status {
			return true
		}
	}
	return false
}

// ValidateQuotaConfig maxSize and maxObjects value of obc or bucketclass
func ValidateQuotaConfig(name string, maxSize string, maxObjects string) error {

	// Positive number
	if maxObjects != "" {
		obcMaxObjectsInt, err := strconv.ParseInt(maxObjects, 10, 32)
		if err != nil {
			return ValidationError{
				Msg: fmt.Sprintf("ob %q validation error: failed to parse maxObjects %v, %v", name, maxObjects, err),
			}
		}
		if obcMaxObjectsInt < 0 {
			return ValidationError{
				Msg: fmt.Sprintf("ob %q validation error: invalid maxObjects value. O or any positive number ", name),
			}
		}
	}

	//Valid range 0, 1G - 1023P
	if maxSize != "" {
		quantity, err := resource.ParseQuantity(maxSize)
		if err != nil {
			log.Errorf("failed to parse quantity: %v", maxSize)
			return err
		}
		obcMaxSizeValue := quantity.Value()
		if obcMaxSizeValue != 0 && (obcMaxSizeValue < gigabyte || obcMaxSizeValue > obcMaxSizeUpperLimit) {
			return ValidationError{
				Msg: fmt.Sprintf("ob %q validation error: invalid obcMaxSizeValue value: min 1Gi, max 1023Pi, 0 to remove quota", name),
			}
		}
	}

	return nil
}

//
// Test shared utilities
//

// NooBaaCondStatus waits for requested NooBaa CR KMS condition status
// returns false if timeout
func NooBaaCondStatus(noobaa *nbv1.NooBaa, s corev1.ConditionStatus) bool {
	return NooBaaCondition(noobaa, nbv1.ConditionTypeKMSStatus, s)
}

// NooBaaCondition waits for requested NooBaa CR KMS condition type & status
// returns false if timeout
func NooBaaCondition(noobaa *nbv1.NooBaa, t conditionsv1.ConditionType, s corev1.ConditionStatus) bool {
	found := false

	timeout := 120 // seconds
	for i := 0; i < timeout; i++ {
		_, _, err := KubeGet(noobaa)
		Panic(err)

		if NoobaaStatus(noobaa, t, s) {
			found = true
			break
		}
		time.Sleep(time.Second)
	}

	return found
}

// GetAvailabeKubeCli will check which k8s cli command is availabe in the system: oc or kubectl
// returns one of: "oc" or "kubectl"
func GetAvailabeKubeCli() string {
	kubeCommand := "kubectl"
	cmd := exec.Command(kubeCommand)
	err := cmd.Run()
	if err == nil {
		log.Printf("✅ kubectl exists - will use it for diagnostics\n")
	} else {
		log.Printf("❌ Could not find kubectl, will try to use oc instead, error: %s\n", err)
		kubeCommand = "oc"
		cmd = exec.Command(kubeCommand)
		err = cmd.Run()
		if err == nil {
			log.Printf("✅ oc exists - will use it for diagnostics\n")
		} else {
			log.Fatalf("❌ Could not find both kubectl and oc, will stop running diagnostics, error: %s", err)
		}
	}
	return kubeCommand
}

// GetEndpointByBackingStoreType returns the endpoint url the of the backing store if it is relevant to the type
func GetEndpointByBackingStoreType(bs *nbv1.BackingStore) (string, error) {
	var endpoint string
	switch bs.Spec.Type {
	case nbv1.StoreTypeAWSS3:
		awsS3 := bs.Spec.AWSS3
		u := url.URL{
			Scheme: "https",
			Host:   "s3.amazonaws.com",
		}
		if awsS3.SSLDisabled {
			u.Scheme = "http"
		}
		if awsS3.Region != "" {
			u.Host = fmt.Sprintf("s3.%s.amazonaws.com", awsS3.Region)
		}
		endpoint = u.String()
	case nbv1.StoreTypeS3Compatible:
		endpoint = bs.Spec.S3Compatible.Endpoint
	case nbv1.StoreTypeIBMCos:
		endpoint = bs.Spec.IBMCos.Endpoint
	case nbv1.StoreTypeAzureBlob:
		endpoint = "https://blob.core.windows.net"
	case nbv1.StoreTypeGoogleCloudStorage:
		endpoint = "https://www.googleapis.com"
	case nbv1.StoreTypePVPool:
		return endpoint, fmt.Errorf("%q type does not have endpoint parameter %q", bs.Spec.Type, bs.Name)
	default:
		return endpoint, fmt.Errorf("failed to get endpoint url from backingstore %q", bs.Name)
	}
	return endpoint, nil
}

// GetEndpointByNamespaceStoreType returns the endpoint url the of the backing store if it is relevant to the type
func GetEndpointByNamespaceStoreType(ns *nbv1.NamespaceStore) (string, error) {
	var endpoint string
	switch ns.Spec.Type {
	case nbv1.NSStoreTypeAWSS3:
		awsS3 := ns.Spec.AWSS3
		u := url.URL{
			Scheme: "https",
			Host:   "s3.amazonaws.com",
		}
		if awsS3.SSLDisabled {
			u.Scheme = "http"
		}
		if awsS3.Region != "" {
			u.Host = fmt.Sprintf("s3.%s.amazonaws.com", awsS3.Region)
		}
		endpoint = u.String()
	case nbv1.NSStoreTypeS3Compatible:
		endpoint = ns.Spec.S3Compatible.Endpoint
	case nbv1.NSStoreTypeIBMCos:
		endpoint = ns.Spec.IBMCos.Endpoint
	case nbv1.NSStoreTypeAzureBlob:
		endpoint = "https://blob.core.windows.net"
	case nbv1.NSStoreTypeGoogleCloudStorage:
		endpoint = "https://www.googleapis.com"
	case nbv1.NSStoreTypeNSFS:
		return endpoint, fmt.Errorf("%q type does not have endpoint parameter %q", ns.Spec.Type, ns.Name)
	default:
		return endpoint, fmt.Errorf("failed to get endpoint url from backingstore %q", ns.Name)
	}
	return endpoint, nil
}

// GetBackingStoreSecretByType returns the secret reference of the backing store if it is relevant to the type
func GetBackingStoreSecretByType(bs *nbv1.BackingStore) (*corev1.SecretReference, error) {
	var secretRef corev1.SecretReference
	switch bs.Spec.Type {
	case nbv1.StoreTypeAWSS3:
		secretRef = bs.Spec.AWSS3.Secret
	case nbv1.StoreTypeS3Compatible:
		secretRef = bs.Spec.S3Compatible.Secret
	case nbv1.StoreTypeIBMCos:
		secretRef = bs.Spec.IBMCos.Secret
	case nbv1.StoreTypeAzureBlob:
		secretRef = bs.Spec.AzureBlob.Secret
	case nbv1.StoreTypeGoogleCloudStorage:
		secretRef = bs.Spec.GoogleCloudStorage.Secret
	case nbv1.StoreTypePVPool:
		secretRef = bs.Spec.PVPool.Secret
	default:
		return nil, fmt.Errorf("failed to get secret reference from backingstore %q", bs.Name)
	}
	return &secretRef, nil
}

// GetBackingStoreSecret returns the secret and adding the namespace if it is missing
func GetBackingStoreSecret(bs *nbv1.BackingStore) (*corev1.SecretReference, error) {
	secretRef, err := GetBackingStoreSecretByType(bs)
	if err != nil {
		return nil, err
	}
	if secretRef.Namespace == "" {
		secretRef.Namespace = bs.Namespace
	}
	return secretRef, nil
}

// SetBackingStoreSecretRef setting a backingstore secret reference to the provided one
func SetBackingStoreSecretRef(bs *nbv1.BackingStore, ref *corev1.SecretReference) error {
	switch bs.Spec.Type {
	case nbv1.StoreTypeAWSS3:
		bs.Spec.AWSS3.Secret = *ref
		return nil
	case nbv1.StoreTypeS3Compatible:
		bs.Spec.S3Compatible.Secret = *ref
		return nil
	case nbv1.StoreTypeIBMCos:
		bs.Spec.IBMCos.Secret = *ref
		return nil
	case nbv1.StoreTypeAzureBlob:
		bs.Spec.AzureBlob.Secret = *ref
		return nil
	case nbv1.StoreTypeGoogleCloudStorage:
		bs.Spec.GoogleCloudStorage.Secret = *ref
		return nil
	case nbv1.StoreTypePVPool:
		bs.Spec.PVPool.Secret = *ref
		return nil
	default:
		return fmt.Errorf("failed to set backingstore %q secret reference", bs.Name)
	}
}

// GetBackingStoreTargetBucket returns the target bucket of the backing store if it is relevant to the type
func GetBackingStoreTargetBucket(bs *nbv1.BackingStore) (string, error) {
	switch bs.Spec.Type {
	case nbv1.StoreTypeAWSS3:
		return bs.Spec.AWSS3.TargetBucket, nil
	case nbv1.StoreTypeS3Compatible:
		return bs.Spec.S3Compatible.TargetBucket, nil
	case nbv1.StoreTypeIBMCos:
		return bs.Spec.IBMCos.TargetBucket, nil
	case nbv1.StoreTypeAzureBlob:
		return bs.Spec.AzureBlob.TargetBlobContainer, nil
	case nbv1.StoreTypeGoogleCloudStorage:
		return bs.Spec.GoogleCloudStorage.TargetBucket, nil
	case nbv1.StoreTypePVPool:
		return "", nil
	default:
		return "", fmt.Errorf("failed to get backingstore %q target bucket", bs.Name)
	}
}

// GetNamespaceStoreSecretByType returns the secret reference of the namespace store if it is relevant to the type
func GetNamespaceStoreSecretByType(ns *nbv1.NamespaceStore) (*corev1.SecretReference, error) {
	var secretRef corev1.SecretReference
	switch ns.Spec.Type {
	case nbv1.NSStoreTypeAWSS3:
		secretRef = ns.Spec.AWSS3.Secret
	case nbv1.NSStoreTypeS3Compatible:
		secretRef = ns.Spec.S3Compatible.Secret
	case nbv1.NSStoreTypeIBMCos:
		secretRef = ns.Spec.IBMCos.Secret
	case nbv1.NSStoreTypeAzureBlob:
		secretRef = ns.Spec.AzureBlob.Secret
	case nbv1.NSStoreTypeGoogleCloudStorage:
		secretRef = ns.Spec.GoogleCloudStorage.Secret
	case nbv1.NSStoreTypeNSFS:
		return nil, nil
	default:
		return nil, fmt.Errorf("failed to get namespacestore %q secret", ns.Name)
	}

	return &secretRef, nil
}

// GetNamespaceStoreSecret returns the secret and adding the namespace if it is missing
func GetNamespaceStoreSecret(ns *nbv1.NamespaceStore) (*corev1.SecretReference, error) {
	secretRef, err := GetNamespaceStoreSecretByType(ns)
	if err != nil {
		return nil, err
	}
	if secretRef != nil && secretRef.Namespace == "" {
		secretRef.Namespace = ns.Namespace
	}
	return secretRef, nil
}

// SetNamespaceStoreSecretRef setting a namespacestore secret reference to the provided one
func SetNamespaceStoreSecretRef(ns *nbv1.NamespaceStore, ref *corev1.SecretReference) error {
	switch ns.Spec.Type {
	case nbv1.NSStoreTypeAWSS3:
		ns.Spec.AWSS3.Secret = *ref
		return nil
	case nbv1.NSStoreTypeS3Compatible:
		ns.Spec.S3Compatible.Secret = *ref
		return nil
	case nbv1.NSStoreTypeIBMCos:
		ns.Spec.IBMCos.Secret = *ref
		return nil
	case nbv1.NSStoreTypeAzureBlob:
		ns.Spec.AzureBlob.Secret = *ref
		return nil
	case nbv1.NSStoreTypeGoogleCloudStorage:
		ns.Spec.GoogleCloudStorage.Secret = *ref
		return nil
	case nbv1.NSStoreTypeNSFS:
		return nil
	default:
		return fmt.Errorf("failed to set namespacestore %q secret reference", ns.Name)
	}
}

// GetNamespaceStoreTargetBucket returns the target bucket of the namespace store if it is relevant to the type
func GetNamespaceStoreTargetBucket(ns *nbv1.NamespaceStore) (string, error) {
	switch ns.Spec.Type {
	case nbv1.NSStoreTypeAWSS3:
		return ns.Spec.AWSS3.TargetBucket, nil
	case nbv1.NSStoreTypeS3Compatible:
		return ns.Spec.S3Compatible.TargetBucket, nil
	case nbv1.NSStoreTypeIBMCos:
		return ns.Spec.IBMCos.TargetBucket, nil
	case nbv1.NSStoreTypeAzureBlob:
		return ns.Spec.AzureBlob.TargetBlobContainer, nil
	case nbv1.NSStoreTypeGoogleCloudStorage:
		return ns.Spec.GoogleCloudStorage.TargetBucket, nil
	case nbv1.NSStoreTypeNSFS:
		return "", nil
	default:
		return "", fmt.Errorf("failed to ger namespacestore %q target bucket", ns.Name)
	}
}

// GetSecretFromSecretReference search and retruns a secret obj from a provided secret reference
func GetSecretFromSecretReference(secretRef *corev1.SecretReference) (*corev1.Secret, error) {
	if secretRef == nil || secretRef.Name == "" {
		return nil, nil
	}

	o := KubeObject(bundle.File_deploy_internal_secret_empty_yaml)
	secret := o.(*corev1.Secret)

	secret.Name = secretRef.Name
	secret.Namespace = secretRef.Namespace

	if !KubeCheck(secret) {
		return nil, fmt.Errorf(`❌ Could not get Secret %q in namespace %q`, secret.Name, secret.Namespace)
	}

	return secret, nil
}

// CheckForIdenticalSecretsCreds search and returns a secret name with identical credentials in the provided secret
// the credentials to compare stored in mandatoryProp
func CheckForIdenticalSecretsCreds(secret *corev1.Secret, storeTypeStr string) *corev1.Secret {
	mandatoryProp, ok := MapStorTypeToMandatoryProperties[storeTypeStr]
	if !ok {
		log.Errorf("❌  failed to map store type %q to mandatory properties", storeTypeStr)
	}
	if secret == nil || len(mandatoryProp) == 0 {
		return nil
	}
	nsList := &nbv1.NamespaceStoreList{
		TypeMeta: metav1.TypeMeta{Kind: "NamespaceStoreList"},
	}
	bsList := &nbv1.BackingStoreList{
		TypeMeta: metav1.TypeMeta{Kind: "BackingStoreList"},
	}
	KubeList(bsList, &client.ListOptions{Namespace: secret.Namespace})
	KubeList(nsList, &client.ListOptions{Namespace: secret.Namespace})

	for _, bs := range bsList.Items {
		if bs.Spec.Type != nbv1.StoreTypePVPool {
			secretRef, err := GetBackingStoreSecret(&bs)
			if err != nil {
				log.Errorf("%s", err)
			}
			if secretRef != nil {
				usedSecret, err := GetSecretFromSecretReference(secretRef)
				if err != nil {
					log.Errorf("%s", err)
				}
				if usedSecret != nil && usedSecret.Name != secret.Name && string(bs.Spec.Type) == storeTypeStr {
					found := true
					for _, key := range mandatoryProp {
						found = found && MapAlternateKeysValue(usedSecret.StringData, key) == secret.StringData[key]
					}
					if found {
						return usedSecret
					}
				}
			}
		}
	}

	for _, ns := range nsList.Items {
		if ns.Spec.Type != nbv1.NSStoreTypeNSFS {
			secretRef, err := GetNamespaceStoreSecret(&ns)
			if err != nil {
				log.Errorf("%s", err)
			}
			if secretRef != nil {
				usedSecret, err := GetSecretFromSecretReference(secretRef)
				if err != nil {
					log.Errorf("%s", err)
				}
				if usedSecret != nil && usedSecret.Name != secret.Name && string(ns.Spec.Type) == storeTypeStr {
					found := true
					for _, key := range mandatoryProp {
						found = found && MapAlternateKeysValue(usedSecret.StringData, key) == secret.StringData[key]
					}
					if found {
						return usedSecret
					}
				}
			}
		}
	}
	return nil
}

// SetOwnerReference setting a owner reference of owner to dependent metadata with the field of blockOwnerDeletion: true
// controllerutil.SetOwnerReference is doing the same thing but without blockOwnerDeletion: true
// If a reference to the same object already exists, it'll return an AlreadyOwnedError. see:
// https://github.com/kubernetes-sigs/controller-runtime/blob/master/pkg/controller/controllerutil/controllerutil.go#L93
func SetOwnerReference(owner, dependent metav1.Object, scheme *runtime.Scheme) error {
	// Validate the owner.
	ro, ok := owner.(runtime.Object)
	if !ok {
		return fmt.Errorf("%T is not a runtime.Object, cannot call SetControllerReference", owner)
	}
	if err := validateOwner(owner, dependent); err != nil {
		return err
	}
	// Create a new ref.
	gvk, err := apiutil.GVKForObject(ro, scheme)
	if err != nil {
		return err
	}
	ref := metav1.OwnerReference{
		APIVersion:         gvk.GroupVersion().String(),
		Kind:               gvk.Kind,
		Name:               owner.GetName(),
		UID:                owner.GetUID(),
		BlockOwnerDeletion: ptr.To[bool](true),
	}

	owners := dependent.GetOwnerReferences()
	if idx := indexOwnerRef(owners, ref); idx == -1 {
		owners = append(owners, ref)
	} else {
		return &controllerutil.AlreadyOwnedError{
			Object: dependent,
			Owner:  ref,
		}
	}
	dependent.SetOwnerReferences(owners)
	return nil
}

func validateOwner(owner, object metav1.Object) error {
	ownerNs := owner.GetNamespace()
	if ownerNs != "" {
		objNs := object.GetNamespace()
		if objNs == "" {
			return fmt.Errorf("cluster-scoped resource must not have a namespace-scoped owner, owner's namespace %s", ownerNs)
		}
		if ownerNs != objNs {
			return fmt.Errorf("cross-namespace owner references are disallowed, owner's namespace %s, object's namespace %s", owner.GetNamespace(), object.GetNamespace())
		}
	}
	return nil
}

// indexOwnerRef returns the index of the owner reference in the slice if found, or -1.
func indexOwnerRef(ownerReferences []metav1.OwnerReference, ref metav1.OwnerReference) int {
	for index, r := range ownerReferences {
		if referSameObject(r, ref) {
			return index
		}
	}
	return -1
}

// Returns true if a and b point to the same object.
func referSameObject(a, b metav1.OwnerReference) bool {
	aGV, err := schema.ParseGroupVersion(a.APIVersion)
	if err != nil {
		return false
	}

	bGV, err := schema.ParseGroupVersion(b.APIVersion)
	if err != nil {
		return false
	}

	return aGV.Group == bGV.Group && a.Kind == b.Kind && a.Name == b.Name
}

// IsOwnedByNoobaa receives an array of owner references and returns true if one of them is of a noobaa resource
func IsOwnedByNoobaa(ownerReferences []metav1.OwnerReference) bool {
	for _, ownerRef := range ownerReferences {
		if strings.Contains(ownerRef.APIVersion, "noobaa") {
			return true
		}
	}
	return false
}

// PrettyPrint the string array in multiple lines, if length greater than 1
func PrettyPrint(key string, strArray []string) {
	if len(strArray) > 1 {
		fmt.Printf("%s : [ %s,\n", key, strArray[0])
		for indx, i := range strArray[1:] {
			if indx != len(strArray)-2 {
				fmt.Printf("\t\t%s,\n", i)
			} else {
				fmt.Printf("\t\t%s ]\n", i)
			}
		}
	} else {
		fmt.Printf("%s : %s\n", key, strArray)
	}
}

// MapAlternateKeysValue scans the map, returning the alternative key name's value if present
// the canonical key name value otherwise
// used to map values from secrets using alternative names
func MapAlternateKeysValue(stringData map[string]string, key string) string {
	alternativeNames := map[string][]string{
		"AWS_ACCESS_KEY_ID":     {"aws_access_key_id", "AccessKey"},
		"AWS_SECRET_ACCESS_KEY": {"aws_secret_access_key", "SecretKey"},
	}

	if stringData[key] == "" {
		for _, altKey := range alternativeNames[key] {
			if stringData[altKey] != "" {
				return stringData[altKey]
			}
		}
	}

	return stringData[key]
}

// FilterSlice takes in a slice and a filter function which
// must return false for the all the elements that need to be
// renoved from the slice
func FilterSlice[V any](slice []V, f func(V) bool) []V {
	var r []V
	for _, v := range slice {
		if f(v) {
			r = append(r, v)
		}
	}
	return r
}

// IsTestEnv checks for TEST_ENV env var existance and equality
// to true and returns true or false accordingly
func IsTestEnv() bool {
	testEnv, ok := os.LookupEnv("TEST_ENV")
	if ok && testEnv == "true" {
		return true
	}
	return false
}

// IsDevEnv checks for DEV_ENV env var existance and equality
// to true and returns true or false accordingly
func IsDevEnv() bool {
	devEnv, ok := os.LookupEnv("DEV_ENV")
	if ok && devEnv == "true" {
		return true
	}
	return false
}

// HasNodeInclusionPolicyInPodTopologySpread checks if the cluster supports the spread topology policy
func HasNodeInclusionPolicyInPodTopologySpread() bool {
	kubeVersion, err := GetKubeVersion()
	if err != nil {
		fmt.Printf("❌ Failed to get kube version %s", err)
		return false
	}
	enabledKubeVersion, err := semver.NewVersion(topologyConstraintsEnabledKubeVersion)
	if err != nil {
		Panic(err)
		return false
	}
	if kubeVersion.LessThan(*enabledKubeVersion) {
		return false
	}
	return true
}
