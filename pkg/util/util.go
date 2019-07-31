package util

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"strings"

	"github.com/noobaa/noobaa-operator/pkg/apis"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var ctx = context.TODO()
var lazyConfig *rest.Config
var lazyRest *rest.RESTClient
var lazyClient client.Client

func init() {
	Panic(apiextv1beta1.AddToScheme(scheme.Scheme))
	Panic(apis.AddToScheme(scheme.Scheme))
}

// KubeConfig loads kubernetes client config from default locations (flags, user dir, etc)
func KubeConfig() *rest.Config {
	if lazyConfig == nil {
		var err error
		lazyConfig, err = config.GetConfig()
		Panic(err)
	}
	return lazyConfig
}

// KubeRest returns a configured kubernetes REST client
func KubeRest() *rest.RESTClient {
	if lazyRest == nil {
		var err error
		config := KubeConfig()
		lazyRest, err = rest.RESTClientFor(config)
		Panic(err)
	}
	return lazyRest
}

// KubeClient resturns a controller-runtime client
// We use a lazy mapper and a specialized implementation of fast mapper
// in order to avoid lags when running a CLI client to a far away cluster.
func KubeClient() client.Client {
	if lazyClient == nil {
		config := KubeConfig()
		mapper := meta.NewLazyRESTMapperLoader(func() (meta.RESTMapper, error) {
			dc, err := discovery.NewDiscoveryClientForConfig(config)
			Panic(err)
			return NewFastRESTMapper(dc, func(g *metav1.APIGroup) bool {
				if g.Name == "" ||
					g.Name == "apps" ||
					g.Name == "noobaa.io" ||
					g.Name == "operator.openshift.io" ||
					g.Name == "cloudcredential.openshift.io" ||
					strings.HasSuffix(g.Name, ".k8s.io") {
					return true
				}
				return false
			}), nil
		})
		var err error
		lazyClient, err = client.New(config, client.Options{Mapper: mapper, Scheme: scheme.Scheme})
		Panic(err)
	}
	return lazyClient
}

// KubeObject loads a text yaml/json to a kubernets object.
func KubeObject(text string) runtime.Object {
	// Decode text (yaml/json) to kube api object
	deserializer := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()
	obj, group, err := deserializer.Decode([]byte(text), nil, nil)
	// obj, group, err := scheme.Codecs.UniversalDecoder().Decode([]byte(text), nil, nil)
	Panic(err)
	// not sure if really needed, but set it anyway
	obj.GetObjectKind().SetGroupVersionKind(*group)
	return obj
}

// KubeApply will check if the object exists and will create/update accordingly
// and report the object status.
func KubeApply(obj runtime.Object) bool {
	klient := KubeClient()
	objKey, _ := client.ObjectKeyFromObject(obj)
	gvk := obj.GetObjectKind().GroupVersionKind()
	clone := obj.DeepCopyObject()
	err := klient.Get(ctx, objKey, clone)
	if err == nil {
		err = klient.Update(ctx, obj)
		if err == nil {
			logrus.Printf("✅ Updated: %s \"%s\"\n", gvk.Kind, objKey.Name)
			return false
		}
	}
	if errors.IsNotFound(err) {
		err = klient.Create(ctx, obj)
		if err == nil {
			logrus.Printf("✅ Created: %s \"%s\"\n", gvk.Kind, objKey.Name)
			return true
		}
	}
	if errors.IsConflict(err) {
		logrus.Printf("❌ Conflict: %s \"%s\": %s\n", gvk.Kind, objKey.Name, err)
		return false
	}
	Panic(err)
	return false
}

// KubeCreateSkipExisting will check if the object exists and will create/skip accordingly
// and report the object status.
func KubeCreateSkipExisting(obj runtime.Object) bool {
	klient := KubeClient()
	objKey, _ := client.ObjectKeyFromObject(obj)
	gvk := obj.GetObjectKind().GroupVersionKind()
	clone := obj.DeepCopyObject()
	err := klient.Get(ctx, objKey, clone)
	if err == nil {
		logrus.Printf("✅ Already Exists: %s \"%s\"\n", gvk.Kind, objKey.Name)
		return false
	}
	if meta.IsNoMatchError(err) {
		logrus.Printf("❌ CRD Missing: %s \"%s\"\n", gvk.Kind, objKey.Name)
		return false
	}
	if errors.IsNotFound(err) {
		err = klient.Create(ctx, obj)
		if err == nil {
			logrus.Printf("✅ Created: %s \"%s\"\n", gvk.Kind, objKey.Name)
			return true
		}
		if errors.IsNotFound(err) {
			logrus.Printf("❌ Namespace Missing: %s \"%s\": kubectl create ns %s\n",
				gvk.Kind, objKey.Name, objKey.Namespace)
			return false
		}
	}
	if errors.IsConflict(err) {
		logrus.Printf("❌ Conflict: %s \"%s\": %s\n", gvk.Kind, objKey.Name, err)
		return false
	}
	if errors.IsForbidden(err) {
		logrus.Printf("❌ Forbidden: %s \"%s\": %s\n", gvk.Kind, objKey.Name, err)
		return false
	}
	Panic(err)
	return false
}

// KubeDelete deletes an object and reports the object status.
func KubeDelete(obj runtime.Object) bool {
	klient := KubeClient()
	objKey, _ := client.ObjectKeyFromObject(obj)
	gvk := obj.GetObjectKind().GroupVersionKind()
	err := klient.Delete(ctx, obj)
	if err == nil {
		logrus.Printf("🗑️  Deleted: %s \"%s\"\n", gvk.Kind, objKey.Name)
		return true
	}
	if errors.IsConflict(err) {
		logrus.Printf("❌ Conflict: %s \"%s\": %s\n", gvk.Kind, objKey.Name, err)
		return false
	}
	if meta.IsNoMatchError(err) || errors.IsNotFound(err) {
		logrus.Printf("❌ Not Found: %s \"%s\"\n", gvk.Kind, objKey.Name)
		return false
	}
	Panic(err)
	return false
}

// KubeCheck checks if the object exists and reports the object status.
func KubeCheck(obj runtime.Object) bool {
	klient := KubeClient()
	objKey, _ := client.ObjectKeyFromObject(obj)
	gvk := obj.GetObjectKind().GroupVersionKind()
	err := klient.Get(ctx, objKey, obj)
	if err == nil {
		logrus.Printf("✅ Exists: %s \"%s\"\n", gvk.Kind, objKey.Name)
		return true
	}
	if meta.IsNoMatchError(err) {
		logrus.Printf("❌ CRD Missing: %s \"%s\"\n", gvk.Kind, objKey.Name)
		return false
	}
	if errors.IsNotFound(err) {
		logrus.Printf("❌ Not Found: %s \"%s\"\n", gvk.Kind, objKey.Name)
		return false
	}
	if errors.IsConflict(err) {
		logrus.Printf("❌ Conflict: %s \"%s\": %s\n", gvk.Kind, objKey.Name, err)
		return false
	}
	Panic(err)
	return false
}

// Panic is conviniently calling panic only if err is not nil
func Panic(err error) {
	if err != nil {
		reason := errors.ReasonForError(err)
		logrus.Panicf("☠️  Panic Attack: [%s] %s", reason, err)
	}
}

// InitLogger initializes the logrus logger with defaults
func InitLogger() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		// FullTimestamp: true,
	})
}

// Logger returns a default logger
func Logger() *logrus.Entry {
	return logrus.WithContext(ctx)
}

// Context returns a default Context
func Context() context.Context {
	return ctx
}

func CurrentNamespace() string {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	ns, _, err := kubeConfig.Namespace()
	Panic(err)
	return ns
}

// PersistentError is an error type that tells the reconcile to avoid requeueing.
type PersistentError struct {
	E error
}

// Error function makes PersistentError implement error interface
func (e *PersistentError) Error() string { return e.E.Error() }

// assert implement error interface
var _ error = &PersistentError{}

// NewPersistentError returns a new persistent error.
func NewPersistentError(err error) *PersistentError {
	if err == nil {
		panic("NewPersistentError expects non nil error")
	}
	return &PersistentError{E: err}
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
func SecretResetStringDataFromData(secret *corev1.Secret) {
	secret.StringData = map[string]string{}
	for key, val := range secret.Data {
		secret.StringData[key] = string(val)
	}
	secret.Data = map[string][]byte{}
}

func RandomBase64(numBytes int) string {
	randomBytes := make([]byte, numBytes)
	_, err := rand.Read(randomBytes)
	Panic(err)
	return base64.StdEncoding.EncodeToString(randomBytes)
}

func RandomHex(numBytes int) string {
	randomBytes := make([]byte, numBytes)
	_, err := rand.Read(randomBytes)
	Panic(err)
	return hex.EncodeToString(randomBytes)
}
