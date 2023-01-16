package util

import (
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var _ = ginkgo.Describe("Predicates Suite", func() {
	ginkgo.Context("IgnoreIfNotInNamespace", func() {
		ginkgo.It("should return true if the object is in the same namespace", func() {
			ns := "test"
			obj := &nbv1.NooBaa{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: ns,
				},
			}

			funcs := IgnoreIfNotInNamespace(ns)

			gomega.Expect(funcs.CreateFunc(event.CreateEvent{Object: obj})).To(gomega.BeTrue())
			gomega.Expect(funcs.DeleteFunc(event.DeleteEvent{Object: obj})).To(gomega.BeTrue())
			gomega.Expect(funcs.UpdateFunc(event.UpdateEvent{ObjectNew: obj})).To(gomega.BeTrue())
			gomega.Expect(funcs.GenericFunc(event.GenericEvent{Object: obj})).To(gomega.BeTrue())
		})

		ginkgo.It("should return false if the object is in different namespace", func() {
			ns := "test"
			obj := &nbv1.NooBaa{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "random",
				},
			}

			funcs := IgnoreIfNotInNamespace(ns)

			gomega.Expect(funcs.CreateFunc(event.CreateEvent{Object: obj})).To(gomega.BeFalse())
			gomega.Expect(funcs.DeleteFunc(event.DeleteEvent{Object: obj})).To(gomega.BeFalse())
			gomega.Expect(funcs.UpdateFunc(event.UpdateEvent{ObjectNew: obj})).To(gomega.BeFalse())
			gomega.Expect(funcs.GenericFunc(event.GenericEvent{Object: obj})).To(gomega.BeFalse())
		})
	})
})
