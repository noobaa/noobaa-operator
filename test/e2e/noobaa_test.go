package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	apis "github.com/noobaa/noobaa-operator/v2/pkg/apis"
	nbv1 "github.com/noobaa/noobaa-operator/v2/pkg/apis/noobaa/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	retryInterval        = time.Second * 3
	cleanupRetryInterval = time.Second * 3
	timeout              = time.Minute * 5
	cleanupTimeout       = time.Minute * 1
)

// NOTICE - This e2e test is currently not included in CI builds
// it was running locally ok, but failed on TravisCI builds,
// so for now I just disabled it.

func TestNooBaa(t *testing.T) {
	list := &nbv1.NooBaaList{}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, list)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	// run subtests
	t.Run("operator", func(t *testing.T) {
		t.Run("first", testOperator)
		t.Run("second", testOperator)
	})
}

func testOperator(t *testing.T) {
	t.Parallel()
	ctx := framework.NewContext(t)
	defer ctx.Cleanup()

	err := ctx.InitializeClusterResources(&framework.CleanupOptions{
		TestContext:   ctx,
		Timeout:       cleanupTimeout,
		RetryInterval: cleanupRetryInterval,
	})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}

	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}

	// get global framework variables
	f := framework.Global

	// wait for noobaa-operator to be ready
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "noobaa-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	err = testInstall(t, f, ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func testInstall(t *testing.T, f *framework.Framework, ctx *framework.Context) error {

	name := "noobaa"
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}
	namespacedName := types.NamespacedName{Namespace: namespace, Name: name}

	// create noobaa custom resource
	noobaa := &nbv1.NooBaa{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: nbv1.NooBaaSpec{},
	}

	// use Context's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(
		context.TODO(),
		noobaa,
		&framework.CleanupOptions{
			TestContext:   ctx,
			Timeout:       cleanupTimeout,
			RetryInterval: cleanupRetryInterval,
		},
	)
	if err != nil {
		return err
	}

	err = waitReadyPhase(t, f, ctx, namespacedName)
	if err != nil {
		return err
	}

	t.Logf("Done.\n")
	return nil

}

// wait for noobaa to reach ready phase
func waitReadyPhase(t *testing.T, f *framework.Framework, ctx *framework.Context, namespacedName types.NamespacedName) error {
	return wait.Poll(retryInterval, timeout, func() (bool, error) {
		currentNooBaa := nbv1.NooBaa{}
		err := f.Client.Get(context.TODO(), namespacedName, &currentNooBaa)
		if err != nil {
			if !errors.IsNotFound(err) {
				return false, err
			}
			t.Logf("Waiting for availability of %s deployment\n", namespacedName.String())
			return false, nil
		}
		if currentNooBaa.Status.Phase == nbv1.SystemPhaseRejected {
			return false, fmt.Errorf("SystemPhaseRejected")
		}
		if currentNooBaa.Status.Phase != nbv1.SystemPhaseReady {
			t.Logf("Waiting for noobaa ready phase, still %s\n", currentNooBaa.Status.Phase)
			return false, nil
		}
		t.Logf("Noobaa is ready.\n")
		return true, nil
	})
}
