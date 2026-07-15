package system

import (
	"strings"
	"testing"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newTestUpgradeReconciler(objs ...client.Object) *Reconciler {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	_ = batchv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = nbv1.SchemeBuilder.AddToScheme(scheme)

	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
	r := NewReconciler(types.NamespacedName{Namespace: "test-ns", Name: "noobaa"}, cl, scheme, nil)
	r.NooBaa.Status.ActualImage = "noobaa/noobaa-core:new"
	return r
}

func setCoreImage(r *Reconciler, image string) {
	r.CoreApp.ObjectMeta = metav1.ObjectMeta{
		Name:      r.CoreApp.Name,
		Namespace: r.CoreApp.Namespace,
		UID:       "core-uid",
	}
	r.CoreApp.Spec.Template.Spec.Containers = []corev1.Container{{
		Name:  "core",
		Image: image,
	}}
}

func TestShouldBlockWorkloadRestore(t *testing.T) {
	tests := []struct {
		name     string
		phase    nbv1.UpgradePhase
		oldImage string
		want     bool
	}{
		{name: "no upgrade needed", phase: nbv1.UpgradePhaseNone, oldImage: "noobaa/noobaa-core:new", want: false},
		{name: "upgrading", phase: nbv1.UpgradePhaseUpgrade, oldImage: "noobaa/noobaa-core:old", want: true},
		{name: "failed", phase: nbv1.UpgradePhaseFailed, oldImage: "noobaa/noobaa-core:old", want: true},
		{name: "done upgrade", phase: nbv1.UpgradePhaseFinished, oldImage: "noobaa/noobaa-core:old", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestUpgradeReconciler()
			setCoreImage(r, tt.oldImage)
			r.NooBaa.Status.UpgradePhase = tt.phase
			if got := r.shouldBlockWorkloadRestore(); got != tt.want {
				t.Fatalf("shouldBlockWorkloadRestore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultUpgradeJobCommandAndEnv(t *testing.T) {
	const upgradeScriptsDir = "/root/node_modules/noobaa-core/src/upgrade/upgrade_scripts"

	r := newTestUpgradeReconciler()
	defaultContainer := r.DefaultUpgradeJob.Containers[0]
	if len(defaultContainer.Command) == 0 {
		t.Fatal("expected upgrade command in default upgrade job template")
	}
	cmd := strings.Join(defaultContainer.Command, " ")
	if !strings.Contains(cmd, "upgrade_manager.js") {
		t.Fatalf("upgrade command missing upgrade_manager.js: %q", cmd)
	}
	if !strings.Contains(cmd, "--upgrade_scripts_dir "+upgradeScriptsDir) {
		t.Fatalf("upgrade command missing scripts dir: %q", cmd)
	}

	var upgradeScriptsDirEnv string
	for _, env := range defaultContainer.Env {
		if env.Name == "UPGRADE_SCRIPTS_DIR" {
			upgradeScriptsDirEnv = env.Value
			break
		}
	}
	if upgradeScriptsDirEnv != upgradeScriptsDir {
		t.Fatalf("UPGRADE_SCRIPTS_DIR = %q, want %q", upgradeScriptsDirEnv, upgradeScriptsDir)
	}

	if upgradeJobBackoffLimit != 4 {
		t.Fatalf("upgradeJobBackoffLimit = %d, want 4", upgradeJobBackoffLimit)
	}
}

func TestSetDesiredUpgradeJobRestoresTemplateCommand(t *testing.T) {
	r := newTestUpgradeReconciler()
	r.SecretRootMasterKey = "test-root-key"
	r.UpgradeJob.Spec.Template.Spec.Containers[0].Command = nil

	if err := r.setDesiredUpgradeJob(); err != nil {
		t.Fatalf("setDesiredUpgradeJob() error: %v", err)
	}

	cmd := strings.Join(r.UpgradeJob.Spec.Template.Spec.Containers[0].Command, " ")
	if !strings.Contains(cmd, "upgrade_manager.js") {
		t.Fatalf("setDesiredUpgradeJob did not restore upgrade command: %q", cmd)
	}
}

func TestShouldBlockWorkloadRestoreBlocksCoreReplicaRestore(t *testing.T) {
	r := newTestUpgradeReconciler()
	setCoreImage(r, "noobaa/noobaa-core:old")
	r.NooBaa.Status.UpgradePhase = nbv1.UpgradePhaseUpgrade
	if !r.shouldBlockWorkloadRestore() {
		t.Fatal("expected workload restore to be blocked during upgrade")
	}
}

func TestIsCoreUpgradeFinishedForActualImage(t *testing.T) {
	tests := []struct {
		name           string
		phase          nbv1.UpgradePhase
		actualImage    string
		completedImage string
		want           bool
	}{
		{
			name:           "finished for current image",
			phase:          nbv1.UpgradePhaseFinished,
			actualImage:    "noobaa/noobaa-core:new",
			completedImage: "noobaa/noobaa-core:new",
			want:           true,
		},
		{
			name:           "finished for previous image",
			phase:          nbv1.UpgradePhaseFinished,
			actualImage:    "noobaa/noobaa-core:newer",
			completedImage: "noobaa/noobaa-core:new",
			want:           false,
		},
		{
			name:           "not finished phase",
			phase:          nbv1.UpgradePhaseUpgrade,
			actualImage:    "noobaa/noobaa-core:new",
			completedImage: "noobaa/noobaa-core:new",
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestUpgradeReconciler()
			r.NooBaa.Status.UpgradePhase = tt.phase
			r.NooBaa.Status.ActualImage = tt.actualImage
			r.NooBaa.Status.LastSuccessfulUpgradeImage = tt.completedImage
			if got := r.isCoreUpgradeFinishedForActualImage(); got != tt.want {
				t.Fatalf("isCoreUpgradeFinishedForActualImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDeployedCoreImage(t *testing.T) {
	r := newTestUpgradeReconciler()
	setCoreImage(r, "noobaa/noobaa-core:deployed")
	if got := r.getDeployedCoreImage(); got != "noobaa/noobaa-core:deployed" {
		t.Fatalf("getDeployedCoreImage() = %q", got)
	}
}

func TestIsUpgradeJobFailed(t *testing.T) {
	tests := []struct {
		name   string
		failed int32
		want   bool
	}{
		{name: "below backoffLimit", failed: upgradeJobBackoffLimit - 1, want: false},
		{name: "equal to backoffLimit", failed: upgradeJobBackoffLimit, want: false},
		{name: "above backoffLimit", failed: upgradeJobBackoffLimit + 1, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestUpgradeReconciler()
			r.UpgradeJob.Status.Failed = tt.failed
			if got := r.isUpgradeJobFailed(); got != tt.want {
				t.Fatalf("isUpgradeJobFailed() with Failed=%d = %v, want %v", tt.failed, got, tt.want)
			}
		})
	}
}

func TestEnsureUpgradeJobRecreateDeletesMismatchedImageAndRequeues(t *testing.T) {
	r := newTestUpgradeReconciler()
	r.UpgradeJob.Spec.Template.Spec.Containers = []corev1.Container{{
		Name:  "noobaa-core-upgrade",
		Image: "noobaa/noobaa-core:old",
	}}
	r.NooBaa.Status.ActualImage = "noobaa/noobaa-core:new"
	if err := r.Client.Create(r.Ctx, r.UpgradeJob.DeepCopy()); err != nil {
		t.Fatalf("create stale upgrade job: %v", err)
	}

	deleted, err := r.ensureUpgradeJobRecreate()
	if err != nil {
		t.Fatalf("ensureUpgradeJobRecreate() error: %v", err)
	}
	if !deleted {
		t.Fatal("expected ensureUpgradeJobRecreate() to delete mismatched job and request requeue")
	}

	err = r.Client.Get(r.Ctx, client.ObjectKeyFromObject(r.UpgradeJob), &batchv1.Job{})
	if !k8serrors.IsNotFound(err) {
		t.Fatalf("expected upgrade job to be deleted from the client, get err = %v", err)
	}
}
