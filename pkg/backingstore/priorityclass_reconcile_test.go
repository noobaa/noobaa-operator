package backingstore

import (
	"fmt"
	"testing"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
)

func TestReconcileBackingStoreValidationError(t *testing.T) {
	bs := &nbv1.BackingStore{}
	bs.Name = "pool"
	bs.Spec.Type = nbv1.StoreTypePVPool
	bs.Spec.PVPool = &nbv1.PVPoolSpec{PriorityClassName: "missing-priority"}

	notFound := util.ValidationError{
		Msg: `Invalid pvPool.priorityClassName "missing-priority"`,
		Err: &util.PriorityClassNotFoundError{Name: "missing-priority"},
	}
	err := reconcileBackingStoreValidationError(bs, notFound)
	perr, ok := err.(*util.PersistentError)
	if !ok {
		t.Fatalf("missing priority class should be persistent, got: %T %v", err, err)
	}
	if perr.Reason != "InvalidPriorityClass" {
		t.Fatalf("expected InvalidPriorityClass, got %q (%s)", perr.Reason, perr.Message)
	}

	lookup := &util.PriorityClassLookupError{Name: "missing-priority", Err: fmt.Errorf("apiserver unavailable")}
	err = reconcileBackingStoreValidationError(bs, lookup)
	if util.IsPersistentError(err) || !util.IsPriorityClassLookupError(err) {
		t.Fatalf("lookup failure should stay retryable, got: %v", err)
	}

	other := util.ValidationError{Msg: "Invalid volume size, minimum volume size is 16Gi"}
	err = reconcileBackingStoreValidationError(bs, other)
	perr, ok = err.(*util.PersistentError)
	if !ok || perr.Reason != "BackingStoreValidationError" {
		t.Fatalf("other validation errors should keep BackingStoreValidationError, got: %v", err)
	}
}
