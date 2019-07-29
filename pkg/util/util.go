package util

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var ctx = context.TODO()

// KubeConfig loads kubernetes client config from default locations (flags, user dir, etc)
func KubeConfig() *rest.Config {
	config, err := config.GetConfig()
	Panic(err)
	return config
}

// KubeRest returns a configured kubernetes REST client
func KubeRest() *rest.RESTClient {
	config := KubeConfig()
	restClient, err := rest.RESTClientFor(config)
	Panic(err)
	return restClient
}

// KubeClient resturns a controller-runtime client
// We use a lazy mapper and a specialized implementation of fast mapper
// in order to avoid lags when running a CLI client to a far away cluster.
func KubeClient() client.Client {
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
	c, err := client.New(config, client.Options{Mapper: mapper, Scheme: scheme.Scheme})
	Panic(err)
	return c
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
func KubeApply(c client.Client, obj runtime.Object) bool {
	objKey, _ := client.ObjectKeyFromObject(obj)
	gvk := obj.GetObjectKind().GroupVersionKind()
	clone := obj.DeepCopyObject()
	err := c.Get(ctx, objKey, clone)
	if err == nil {
		err = c.Update(ctx, obj)
		if err == nil {
			logrus.Printf("✅ Updated: %s \"%s\"\n", gvk.Kind, objKey.Name)
			return false
		}
	}
	if errors.IsNotFound(err) {
		err = c.Create(ctx, obj)
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
func KubeCreateSkipExisting(c client.Client, obj runtime.Object) bool {
	objKey, _ := client.ObjectKeyFromObject(obj)
	gvk := obj.GetObjectKind().GroupVersionKind()
	clone := obj.DeepCopyObject()
	err := c.Get(ctx, objKey, clone)
	if err == nil {
		logrus.Printf("✅ Already Exists: %s \"%s\"\n", gvk.Kind, objKey.Name)
		return false
	}
	if meta.IsNoMatchError(err) {
		logrus.Printf("❌ CRD Missing: %s \"%s\"\n", gvk.Kind, objKey.Name)
		return false
	}
	if errors.IsNotFound(err) {
		err = c.Create(ctx, obj)
		if err == nil {
			logrus.Printf("✅ Created: %s \"%s\"\n", gvk.Kind, objKey.Name)
			return true
		}
		if errors.IsNotFound(err) {
			logrus.Printf("❌ Namespace Missing: %s \"%s\": kubectl create ns %s\n",
				gvk.Kind, objKey.Name, objKey.Namespace)
			return true
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
func KubeDelete(c client.Client, obj runtime.Object) bool {
	objKey, _ := client.ObjectKeyFromObject(obj)
	gvk := obj.GetObjectKind().GroupVersionKind()
	err := c.Delete(ctx, obj)
	if err == nil {
		logrus.Printf("❌ Deleted: %s \"%s\"\n", gvk.Kind, objKey.Name)
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
func KubeCheck(c client.Client, obj runtime.Object) bool {
	objKey, _ := client.ObjectKeyFromObject(obj)
	gvk := obj.GetObjectKind().GroupVersionKind()
	err := c.Get(ctx, objKey, obj)
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
		FullTimestamp: true,
	})
}
