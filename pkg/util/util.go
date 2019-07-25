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

func KubeConfig() *rest.Config {
	return config.GetConfigOrDie()
}

func KubeClient() client.Client {
	config := KubeConfig()
	mapper := meta.NewLazyRESTMapperLoader(func() (meta.RESTMapper, error) {
		dc := discovery.NewDiscoveryClientForConfigOrDie(config)
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
	Fatal(err)
	return c
}

func KubeObject(text string) runtime.Object {
	// Decode text (yaml/json) to kube api object
	deserializer := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()
	obj, group, err := deserializer.Decode([]byte(text), nil, nil)
	// obj, group, err := scheme.Codecs.UniversalDecoder().Decode([]byte(text), nil, nil)
	Fatal(err)
	// not sure if really needed, but set it anyway
	obj.GetObjectKind().SetGroupVersionKind(*group)
	return obj
}

func KubeApply(c client.Client, obj runtime.Object) bool {
	metaObj, _ := meta.Accessor(obj)
	objKey := client.ObjectKey{Namespace: metaObj.GetNamespace(), Name: metaObj.GetName()}
	gvk := obj.GetObjectKind().GroupVersionKind()
	clone := obj.DeepCopyObject()
	err := c.Get(ctx, objKey, clone)
	if err == nil {
		err = c.Update(ctx, obj)
		if err == nil {
			logrus.Printf("✅ %s %s Updated.\n", gvk.Kind, metaObj.GetName())
			return false
		}
	}
	if errors.IsNotFound(err) {
		err = c.Create(ctx, obj)
		if err == nil {
			logrus.Printf("✅ %s %s Created.\n", gvk.Kind, metaObj.GetName())
			return true
		}
	}
	Fatal(err)
	return false
}

func KubeCreateSkipExisting(c client.Client, obj runtime.Object) bool {
	metaObj, _ := meta.Accessor(obj)
	objKey := client.ObjectKey{Namespace: metaObj.GetNamespace(), Name: metaObj.GetName()}
	gvk := obj.GetObjectKind().GroupVersionKind()
	clone := obj.DeepCopyObject()
	err := c.Get(ctx, objKey, clone)
	if err == nil {
		logrus.Printf("✅ %s %s Already exists.\n", gvk.Kind, metaObj.GetName())
		return false
	}
	if errors.IsNotFound(err) {
		err = c.Create(ctx, obj)
		if err == nil {
			logrus.Printf("✅ %s %s Created.\n", gvk.Kind, metaObj.GetName())
			return true
		}
	}
	Fatal(err)
	return false
}

func KubeDelete(c client.Client, obj runtime.Object) bool {
	metaObj, _ := meta.Accessor(obj)
	// objKey := client.ObjectKey{Namespace: metaObj.GetNamespace(), Name: metaObj.GetName()}
	gvk := obj.GetObjectKind().GroupVersionKind()
	err := c.Delete(ctx, obj)
	if err == nil {
		logrus.Printf("✅ %s %s Deleted.\n", gvk.Kind, metaObj.GetName())
		return true
	}
	if errors.IsNotFound(err) {
		logrus.Printf("❌ %s %s Not found.\n", gvk.Kind, metaObj.GetName())
		return false
	}
	Fatal(err)
	return false
}

func KubeCheck(c client.Client, obj runtime.Object) bool {
	metaObj, _ := meta.Accessor(obj)
	objKey := client.ObjectKey{Namespace: metaObj.GetNamespace(), Name: metaObj.GetName()}
	gvk := obj.GetObjectKind().GroupVersionKind()
	err := c.Get(ctx, objKey, obj)
	if err == nil {
		logrus.Printf("✅ %s %s Exists.\n", gvk.Kind, metaObj.GetName())
		return true
	}
	if errors.IsNotFound(err) {
		logrus.Printf("❌ %s %s Not found.\n", gvk.Kind, metaObj.GetName())
		return false
	}
	Fatal(err)
	return false
}

func Fatal(err error) {
	if err != nil {
		logrus.Println("☠️  Fatal Error", err)
		panic(err)
	}
}
