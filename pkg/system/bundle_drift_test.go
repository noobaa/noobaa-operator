package system

import (
	"context"
	"reflect"
	"testing"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// These tests check that objects created by an older operator version are updated
// to the current bundle content on reconcile (DFBUGS-11487).

const bundleTestNamespace = "test-ns"

func newBundleTestReconciler(t *testing.T, objs ...client.Object) *Reconciler {
	t.Helper()
	nb := &nbv1.NooBaa{ObjectMeta: metav1.ObjectMeta{Name: "noobaa", Namespace: bundleTestNamespace, UID: "noobaa-uid"}}
	return &Reconciler{
		Request: types.NamespacedName{Name: nb.Name, Namespace: nb.Namespace},
		Client:  fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(objs...).Build(),
		Scheme:  scheme.Scheme,
		Ctx:     context.TODO(),
		Logger:  logrus.WithField("test", t.Name()),
		NooBaa:  nb,
	}
}

// reconcileTwice runs ReconcileObject twice, each time on a fresh bundle object,
// like two reconciles of the operator. It returns the live object after the first reconcile,
// and fails if the second reconcile updated the object again.
func reconcileTwice[T client.Object](t *testing.T, r *Reconciler, fresh func() T, setter func(T) func() error) T {
	t.Helper()
	obj := fresh()
	if err := r.ReconcileObject(obj, setter(obj)); err != nil {
		t.Fatalf("first reconcile failed: %v", err)
	}
	live := fresh()
	if err := r.Client.Get(r.Ctx, client.ObjectKeyFromObject(live), live); err != nil {
		t.Fatalf("get after first reconcile failed: %v", err)
	}

	obj = fresh()
	if err := r.ReconcileObject(obj, setter(obj)); err != nil {
		t.Fatalf("second reconcile failed: %v", err)
	}
	again := fresh()
	if err := r.Client.Get(r.Ctx, client.ObjectKeyFromObject(again), again); err != nil {
		t.Fatalf("get after second reconcile failed: %v", err)
	}
	if again.GetResourceVersion() != live.GetResourceVersion() {
		t.Errorf("second reconcile updated the object again: resourceVersion %s -> %s",
			live.GetResourceVersion(), again.GetResourceVersion())
	}
	return live
}

func bundleNetworkPolicy(file, name string) func() *networkingv1.NetworkPolicy {
	return func() *networkingv1.NetworkPolicy {
		np := util.KubeObject(file).(*networkingv1.NetworkPolicy)
		np.Namespace = bundleTestNamespace
		np.Name = name
		return np
	}
}

func TestBundleDriftNetworkPolicyEndpoint(t *testing.T) {
	fresh := bundleNetworkPolicy(bundle.File_deploy_internal_networkpolicy_endpoint_yaml, "noobaa-endpoint")
	// live policy from before the N2N port range was added
	old := fresh()
	ports := old.Spec.Ingress[0].Ports
	old.Spec.Ingress[0].Ports = ports[:len(ports)-1]

	r := newBundleTestReconciler(t, old)
	live := reconcileTwice(t, r, fresh, func(np *networkingv1.NetworkPolicy) func() error {
		r.NetworkPolicyEndpoint = np
		return r.SetDesiredNetworkPolicyEndpoint
	})

	want := fresh()
	want.Spec.PodSelector.MatchLabels["noobaa-s3"] = r.Request.Name
	if !reflect.DeepEqual(live.Spec, want.Spec) {
		t.Errorf("spec was not updated from the bundle:\n got: %+v\nwant: %+v", live.Spec, want.Spec)
	}
}

func TestBundleDriftNetworkPolicyPVPool(t *testing.T) {
	fresh := bundleNetworkPolicy(bundle.File_deploy_internal_networkpolicy_pvpool_yaml, "noobaa-pvpool")
	old := fresh()
	endPort := int32(60500)
	old.Spec.Ingress[0].Ports[0].EndPort = &endPort

	r := newBundleTestReconciler(t, old)
	live := reconcileTwice(t, r, fresh, func(np *networkingv1.NetworkPolicy) func() error {
		r.NetworkPolicyPVPool = np
		return r.SetDesiredNetworkPolicyPVPool
	})
	if !reflect.DeepEqual(live.Spec, fresh().Spec) {
		t.Errorf("spec was not updated from the bundle: %+v", live.Spec)
	}
}

func bundleServiceS3() *corev1.Service {
	svc := util.KubeObject(bundle.File_deploy_internal_service_s3_yaml).(*corev1.Service)
	svc.Namespace = bundleTestNamespace
	return svc
}

func TestBundleDriftServiceS3(t *testing.T) {
	// live service from an older bundle, as the API server stores it: allocated clusterIP
	// and nodePorts, defaulted protocol and targetPort, and the old plain-HTTP metrics port
	old := bundleServiceS3()
	old.Spec.ClusterIP = "172.30.1.1"
	old.Spec.Selector["noobaa-s3"] = "noobaa"
	old.Spec.Ports = []corev1.ServicePort{
		{Name: "s3", Port: 80, TargetPort: intstr.FromInt32(6001), Protocol: corev1.ProtocolTCP, NodePort: 30001},
		{Name: "s3-https", Port: 443, TargetPort: intstr.FromInt32(6443), Protocol: corev1.ProtocolTCP, NodePort: 30002},
		{Name: "md-https", Port: 8444, TargetPort: intstr.FromInt32(8444), Protocol: corev1.ProtocolTCP, NodePort: 30003},
		{Name: "metrics", Port: 7004, TargetPort: intstr.FromInt32(7004), Protocol: corev1.ProtocolTCP, NodePort: 30004},
	}

	r := newBundleTestReconciler(t, old)
	live := reconcileTwice(t, r, bundleServiceS3, func(svc *corev1.Service) func() error {
		r.ServiceS3 = svc
		return r.SetDesiredServiceS3
	})

	wantNodePorts := map[string]int32{"s3": 30001, "s3-https": 30002, "md-https": 30003, "metrics-https": 0}
	if len(live.Spec.Ports) != len(wantNodePorts) {
		t.Fatalf("ports = %+v, want the %d bundle ports", live.Spec.Ports, len(wantNodePorts))
	}
	for _, p := range live.Spec.Ports {
		want, ok := wantNodePorts[p.Name]
		if !ok {
			t.Errorf("unexpected port %q (%d) was not removed", p.Name, p.Port)
			continue
		}
		if p.NodePort != want {
			t.Errorf("port %q nodePort = %d, want %d", p.Name, p.NodePort, want)
		}
	}
	if live.Spec.ClusterIP != "172.30.1.1" {
		t.Errorf("clusterIP = %q, want it kept", live.Spec.ClusterIP)
	}
}

func bundleServiceSts() *corev1.Service {
	svc := util.KubeObject(bundle.File_deploy_internal_service_sts_yaml).(*corev1.Service)
	svc.Namespace = bundleTestNamespace
	return svc
}

func TestBundleDriftServiceSts(t *testing.T) {
	const signedByAnnotation = "service.beta.openshift.io/serving-cert-signed-by"
	// live service from before the serving cert annotations were added,
	// with an annotation set by another controller
	old := bundleServiceSts()
	old.Annotations = map[string]string{signedByAnnotation: "openshift-service-serving-signer"}

	r := newBundleTestReconciler(t, old)
	live := reconcileTwice(t, r, bundleServiceSts, func(svc *corev1.Service) func() error {
		r.ServiceSts = svc
		return r.SetDesiredServiceSts
	})

	if live.Annotations["service.beta.openshift.io/serving-cert-secret-name"] != "noobaa-sts-serving-cert" {
		t.Errorf("bundle serving-cert annotation was not added: %v", live.Annotations)
	}
	if live.Annotations[signedByAnnotation] == "" {
		t.Errorf("annotation %q set by another controller was removed", signedByAnnotation)
	}
}

func bundlePrometheusRule() *monitoringv1.PrometheusRule {
	rule := util.KubeObject(bundle.File_deploy_internal_prometheus_rules_yaml).(*monitoringv1.PrometheusRule)
	rule.Namespace = bundleTestNamespace
	rule.Name = "noobaa-prometheus-rules"
	return rule
}

func TestBundleDriftPrometheusRule(t *testing.T) {
	// live rule from an older bundle, with a missing alert and a changed one
	old := bundlePrometheusRule()
	group := &old.Spec.Groups[len(old.Spec.Groups)-1]
	group.Rules = group.Rules[:len(group.Rules)-1]
	old.Spec.Groups[0].Rules[0].Expr = intstr.FromString("vector(1)")

	r := newBundleTestReconciler(t, old)
	live := reconcileTwice(t, r, bundlePrometheusRule, func(rule *monitoringv1.PrometheusRule) func() error {
		r.PrometheusRule = rule
		return r.SetDesiredPrometheusRule
	})
	if !reflect.DeepEqual(live.Spec, bundlePrometheusRule().Spec) {
		t.Errorf("rules were not updated from the bundle")
	}
}
