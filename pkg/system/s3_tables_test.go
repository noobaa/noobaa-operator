package system

import (
	"testing"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TestIsS3TablesEnabled checks that the Developer Preview is gated solely on the
// noobaa.io/enable_s3_tables_dev_preview annotation being set to "true".
func TestIsS3TablesEnabled(t *testing.T) {
	cases := []struct {
		name        string
		annotations map[string]string
		want        bool
	}{
		{"no annotations", nil, false},
		{"unrelated annotation", map[string]string{"noobaa.io/other": "true"}, false},
		{"set to true", map[string]string{nbv1.EnableS3TablesDevPreview: "true"}, true},
		{"set to TRUE", map[string]string{nbv1.EnableS3TablesDevPreview: "TRUE"}, true},
		{"set to false", map[string]string{nbv1.EnableS3TablesDevPreview: "false"}, false},
		{"set to empty", map[string]string{nbv1.EnableS3TablesDevPreview: ""}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &Reconciler{NooBaa: &nbv1.NooBaa{ObjectMeta: metav1.ObjectMeta{Annotations: tc.annotations}}}
			if got := r.IsS3TablesEnabled(); got != tc.want {
				t.Errorf("IsS3TablesEnabled() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestSetDesiredS3TablesEnv checks that CONFIG_JS_S3_TABLES_ENABLED is added when the
// feature is enabled, updated in place, and removed when the feature is disabled - so
// core and endpoints never disagree on the flag.
func TestSetDesiredS3TablesEnv(t *testing.T) {
	const envName = "CONFIG_JS_S3_TABLES_ENABLED"

	reconciler := func(enabled bool) *Reconciler {
		annotations := map[string]string{}
		if enabled {
			annotations[nbv1.EnableS3TablesDevPreview] = "true"
		}
		return &Reconciler{NooBaa: &nbv1.NooBaa{ObjectMeta: metav1.ObjectMeta{Annotations: annotations}}}
	}

	getEnv := func(c *corev1.Container) *corev1.EnvVar {
		return util.GetEnvVariable(&c.Env, envName)
	}

	t.Run("adds when enabled", func(t *testing.T) {
		c := &corev1.Container{Env: []corev1.EnvVar{{Name: "OTHER", Value: "x"}}}
		reconciler(true).setDesiredS3TablesEnv(c)
		ev := getEnv(c)
		if ev == nil || ev.Value != "true" {
			t.Fatalf("expected %s=true, got %+v", envName, ev)
		}
	})

	t.Run("updates existing stale value", func(t *testing.T) {
		c := &corev1.Container{Env: []corev1.EnvVar{{Name: envName, Value: "false"}}}
		reconciler(true).setDesiredS3TablesEnv(c)
		if ev := getEnv(c); ev == nil || ev.Value != "true" {
			t.Fatalf("expected %s updated to true, got %+v", envName, ev)
		}
		// no duplicate entry
		count := 0
		for _, e := range c.Env {
			if e.Name == envName {
				count++
			}
		}
		if count != 1 {
			t.Fatalf("expected a single %s entry, got %d", envName, count)
		}
	})

	t.Run("removes when disabled", func(t *testing.T) {
		c := &corev1.Container{Env: []corev1.EnvVar{{Name: "OTHER", Value: "x"}, {Name: envName, Value: "true"}}}
		reconciler(false).setDesiredS3TablesEnv(c)
		if ev := getEnv(c); ev != nil {
			t.Fatalf("expected %s removed, got %+v", envName, ev)
		}
		if len(c.Env) != 1 || c.Env[0].Name != "OTHER" {
			t.Fatalf("unrelated env vars must be preserved, got %+v", c.Env)
		}
	})

	t.Run("no-op when disabled and absent", func(t *testing.T) {
		c := &corev1.Container{Env: []corev1.EnvVar{{Name: "OTHER", Value: "x"}}}
		reconciler(false).setDesiredS3TablesEnv(c)
		if len(c.Env) != 1 {
			t.Fatalf("expected env unchanged, got %+v", c.Env)
		}
	})
}

// TestSetDesiredServiceTables checks the Service type follows DisableLoadBalancerService
// and that, unlike the vector service, no loadBalancerSourceRanges are set.
func TestSetDesiredServiceTables(t *testing.T) {
	t.Run("load balancer by default", func(t *testing.T) {
		r := &Reconciler{
			NooBaa:        &nbv1.NooBaa{},
			ServiceTables: util.KubeObject(bundle.File_deploy_internal_service_tables_yaml).(*corev1.Service),
		}
		r.Request.Name = "noobaa"
		if err := r.SetDesiredServiceTables(); err != nil {
			t.Fatal(err)
		}
		if r.ServiceTables.Spec.Type != corev1.ServiceTypeLoadBalancer {
			t.Errorf("expected LoadBalancer, got %s", r.ServiceTables.Spec.Type)
		}
		if r.ServiceTables.Spec.Selector["noobaa-s3"] != "noobaa" {
			t.Errorf("expected selector noobaa-s3=noobaa, got %v", r.ServiceTables.Spec.Selector)
		}
		if len(r.ServiceTables.Spec.LoadBalancerSourceRanges) != 0 {
			t.Errorf("expected no loadBalancerSourceRanges, got %v", r.ServiceTables.Spec.LoadBalancerSourceRanges)
		}
	})

	t.Run("cluster ip when load balancer disabled", func(t *testing.T) {
		r := &Reconciler{
			NooBaa:        &nbv1.NooBaa{Spec: nbv1.NooBaaSpec{DisableLoadBalancerService: true}},
			ServiceTables: util.KubeObject(bundle.File_deploy_internal_service_tables_yaml).(*corev1.Service),
		}
		r.Request.Name = "noobaa"
		if err := r.SetDesiredServiceTables(); err != nil {
			t.Fatal(err)
		}
		if r.ServiceTables.Spec.Type != corev1.ServiceTypeClusterIP {
			t.Errorf("expected ClusterIP, got %s", r.ServiceTables.Spec.Type)
		}
	})
}

// TestSetDesiredRouteTables checks the re-encrypt route honours DenyHTTP.
func TestSetDesiredRouteTables(t *testing.T) {
	newR := func(denyHTTP bool) *Reconciler {
		return &Reconciler{
			NooBaa:      &nbv1.NooBaa{Spec: nbv1.NooBaaSpec{DenyHTTP: denyHTTP}},
			RouteTables: util.KubeObject(bundle.File_deploy_internal_route_tables_yaml).(*routev1.Route),
		}
	}

	t.Run("allows insecure by default", func(t *testing.T) {
		r := newR(false)
		if err := r.SetDesiredRouteTables(); err != nil {
			t.Fatal(err)
		}
		if r.RouteTables.Spec.TLS.InsecureEdgeTerminationPolicy != routev1.InsecureEdgeTerminationPolicyAllow {
			t.Errorf("expected Allow, got %s", r.RouteTables.Spec.TLS.InsecureEdgeTerminationPolicy)
		}
		if r.RouteTables.Spec.TLS.Termination != routev1.TLSTerminationReencrypt {
			t.Errorf("expected reencrypt termination, got %s", r.RouteTables.Spec.TLS.Termination)
		}
	})

	t.Run("denies insecure when DenyHTTP", func(t *testing.T) {
		r := newR(true)
		if err := r.SetDesiredRouteTables(); err != nil {
			t.Fatal(err)
		}
		if r.RouteTables.Spec.TLS.InsecureEdgeTerminationPolicy != routev1.InsecureEdgeTerminationPolicyNone {
			t.Errorf("expected None, got %s", r.RouteTables.Spec.TLS.InsecureEdgeTerminationPolicy)
		}
	})
}

// TestReconcileS3TablesServiceCreatedWhenEnabled verifies the tables Service is created
// through the reconcile client while the annotation is set.
func TestReconcileS3TablesServiceCreatedWhenEnabled(t *testing.T) {
	r := newBundleTestReconciler(t)
	r.NooBaa.Annotations = map[string]string{nbv1.EnableS3TablesDevPreview: "true"}
	// DisableRoutes so the reconcile skips the Route (route CRD not needed for this check)
	r.NooBaa.Spec.DisableRoutes = true

	r.ServiceTables = util.KubeObject(bundle.File_deploy_internal_service_tables_yaml).(*corev1.Service)
	r.ServiceTables.Namespace = bundleTestNamespace
	r.ServiceTables.Name = "tables"
	r.RouteTables = util.KubeObject(bundle.File_deploy_internal_route_tables_yaml).(*routev1.Route)
	r.RouteTables.Namespace = bundleTestNamespace
	r.RouteTables.Name = "tables"

	if err := r.ReconcileS3TablesServiceAndRoute(); err != nil {
		t.Fatalf("reconcile failed: %v", err)
	}

	live := &corev1.Service{}
	key := client.ObjectKey{Namespace: bundleTestNamespace, Name: "tables"}
	if err := r.Client.Get(r.Ctx, key, live); err != nil {
		t.Fatalf("expected tables Service to be created, get failed: %v", err)
	}
	if live.Spec.Selector["noobaa-s3"] != r.Request.Name {
		t.Errorf("expected selector noobaa-s3=%s, got %v", r.Request.Name, live.Spec.Selector)
	}
}
