package validations

import (
	"context"
	"fmt"
	"testing"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestValidateBackingStoreIgnoresStalePVPoolPriorityClass(t *testing.T) {
	restore := usePriorityClassClient(t, nil, false)
	defer restore()

	bs := nbv1.BackingStore{
		ObjectMeta: metav1.ObjectMeta{Name: "aws-bs"},
		Spec: nbv1.BackingStoreSpec{
			Type: nbv1.StoreTypeAWSS3,
			AWSS3: &nbv1.AWSS3Spec{
				TargetBucket: "bucket",
				Secret:       corev1.SecretReference{Name: "secret", Namespace: "ns"},
				Region:       "us-east-1",
			},
			PVPool: &nbv1.PVPoolSpec{
				NumVolumes:        1,
				PriorityClassName: "missing-priority",
			},
		},
	}
	if err := ValidateBackingStore(bs); err != nil {
		t.Fatalf("stale pvPool priority class on aws-s3 should be ignored, got: %v", err)
	}
	if err := ValidatePvpoolPriorityClass(bs); err != nil {
		t.Fatalf("non pv-pool priority class check should be a no-op, got: %v", err)
	}
}

func TestValidatePvpoolPriorityClass(t *testing.T) {
	pc := &schedulingv1.PriorityClass{ObjectMeta: metav1.ObjectMeta{Name: "high-priority"}, Value: 1000}
	restore := usePriorityClassClient(t, []client.Object{pc}, false)
	defer restore()

	existing := pvPoolBackingStore("high-priority", 1)
	if err := ValidatePvpoolPriorityClass(existing); err != nil {
		t.Fatalf("existing priority class should be accepted, got: %v", err)
	}

	missing := pvPoolBackingStore("missing-priority", 1)
	err := ValidateBackingStore(missing)
	if !util.IsValidationError(err) || !util.IsPriorityClassNotFound(err) {
		t.Fatalf("missing priority class should be invalid configuration, got: %v", err)
	}
	if util.IsPriorityClassLookupError(err) {
		t.Fatalf("missing priority class must not be reported as a lookup failure: %v", err)
	}
}

func TestValidatePvpoolPriorityClassLookupError(t *testing.T) {
	restore := usePriorityClassClient(t, nil, true)
	defer restore()

	err := ValidatePvpoolPriorityClass(pvPoolBackingStore("high-priority", 1))
	if !util.IsPriorityClassLookupError(err) {
		t.Fatalf("expected lookup failure, got: %v", err)
	}
	if util.IsValidationError(err) || util.IsPriorityClassNotFound(err) {
		t.Fatalf("lookup failure must stay retryable, got: %v", err)
	}
}

func TestValidateBackingStoreUpdateSkipsUnchangedPriorityClass(t *testing.T) {
	restore := usePriorityClassClient(t, nil, false)
	defer restore()

	bs := pvPoolBackingStore("missing-priority", 1)
	bs.Finalizers = []string{"noobaa.io/finalizer"}
	updated := bs.DeepCopy()
	updated.Finalizers = nil

	if err := ValidateBackingStoreUpdate(*updated, bs); err != nil {
		t.Fatalf("finalizer removal should be allowed when the priority class is already gone, got: %v", err)
	}

	changed := bs.DeepCopy()
	changed.Spec.PVPool.NumVolumes = 2
	err := ValidateBackingStoreUpdate(*changed, bs)
	if !util.IsValidationError(err) || !util.IsPriorityClassNotFound(err) {
		t.Fatalf("spec change should still reject a missing priority class, got: %v", err)
	}
}

func TestValidateNoobaaPriorityClassUpdate(t *testing.T) {
	pc := &schedulingv1.PriorityClass{ObjectMeta: metav1.ObjectMeta{Name: "db-priority"}, Value: 10}
	restore := usePriorityClassClient(t, []client.Object{pc}, false)
	defer restore()

	nb := nbv1.NooBaa{Spec: nbv1.NooBaaSpec{CorePriorityClassName: "missing-priority"}}
	nb.Finalizers = []string{"noobaa.io/graceful_finalizer"}
	updated := nb.DeepCopy()
	updated.Finalizers = nil
	if err := ValidateNoobaaUpdate(*updated, nb); err != nil {
		t.Fatalf("finalizer removal should be allowed when the priority class is already gone, got: %v", err)
	}

	changed := nb.DeepCopy()
	changed.Spec.DBPriorityClassName = "db-priority"
	if err := ValidateNoobaaUpdate(*changed, nb); err == nil || !util.IsPriorityClassNotFound(err) {
		t.Fatalf("spec change should still reject a missing priority class, got: %v", err)
	}

	fixed := changed.DeepCopy()
	fixed.Spec.CorePriorityClassName = ""
	if err := ValidateNoobaaUpdate(*fixed, nb); err != nil {
		t.Fatalf("clearing a missing priority class and setting an existing one should be allowed, got: %v", err)
	}
}

func TestValidatePriorityClassesLookupError(t *testing.T) {
	restore := usePriorityClassClient(t, nil, true)
	defer restore()

	err := ValidatePriorityClasses(nbv1.NooBaa{Spec: nbv1.NooBaaSpec{EndpointPriorityClassName: "endpoint-priority"}})
	if !util.IsPriorityClassLookupError(err) || util.IsValidationError(err) {
		t.Fatalf("lookup failure must stay retryable, got: %v", err)
	}
}

func pvPoolBackingStore(priorityClassName string, numVolumes int) nbv1.BackingStore {
	return nbv1.BackingStore{
		ObjectMeta: metav1.ObjectMeta{Name: "pv-pool"},
		Spec: nbv1.BackingStoreSpec{
			Type: nbv1.StoreTypePVPool,
			PVPool: &nbv1.PVPoolSpec{
				NumVolumes:        numVolumes,
				PriorityClassName: priorityClassName,
			},
		},
	}
}

func usePriorityClassClient(t *testing.T, objects []client.Object, failGet bool) func() {
	t.Helper()
	builder := fake.NewClientBuilder().WithScheme(scheme.Scheme)
	if len(objects) > 0 {
		builder = builder.WithObjects(objects...)
	}
	if failGet {
		builder = builder.WithInterceptorFuncs(interceptor.Funcs{
			Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
				return apierrors.NewInternalError(fmt.Errorf("apiserver unavailable"))
			},
		})
	}
	prev := util.SetKubeClientForTest(builder.Build())
	return func() { util.SetKubeClientForTest(prev) }
}
