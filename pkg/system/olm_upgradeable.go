package system

import (
	"fmt"
	"os"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	operv2 "github.com/operator-framework/api/pkg/operators/v2"
	"github.com/operator-framework/operator-lib/conditions"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

const (
	// operatorConditionNameEnv is set by OLM on every deployment of the CSV.
	// It holds the name of the OperatorCondition CR of the running CSV.
	operatorConditionNameEnv = "OPERATOR_CONDITION_NAME"

	upgradeableReasonNoSystem          = "NoSystem"
	upgradeableReasonNoMigration       = "NoDBMigrationInProgress"
	notUpgradeableReasonDBMigration    = "DBMigrationInProgress"
	notUpgradeableReasonStandaloneDBUp = "StandaloneDBNotRetired"
)

// ReconcileOLMUpgradeable reports to OLM whether it is safe to replace the running operator CSV with a newer one.
// It sets the Upgradeable condition in spec.conditions of the CSV's OperatorCondition CR.
// While Upgradeable=False, OLM keeps a newer CSV in Pending phase. odf-operator also copies a False value
// into its own OperatorCondition, which holds ODF minor upgrades.
//
// The condition is False only while the DB migration from the standalone DB to the CNPG cluster is not completed.
// Upgrading the operator in the middle of the migration can leave noobaa-core on the standalone DB
// while the CNPG cluster holds a stale copy of the data (see DFBUGS-10942).
//
// Do not block on a generic "system not ready": a bug that keeps the system not ready is fixed by upgrading.
// A cluster admin can always force an upgrade with spec.overrides in the OperatorCondition.
func (r *Reconciler) ReconcileOLMUpgradeable() {
	if os.Getenv(operatorConditionNameEnv) == "" {
		// not deployed by OLM - there is no OperatorCondition to update
		return
	}

	status, reason, message := r.getOLMUpgradeableStatus()
	if err := setOLMUpgradeableCondition(status, reason, message); err != nil {
		r.Logger.Warnf("failed to set OLM Upgradeable=%s (%s): %v", status, reason, err)
	}
}

// getOLMUpgradeableStatus returns the desired Upgradeable status, reason and message
func (r *Reconciler) getOLMUpgradeableStatus() (metav1.ConditionStatus, string, string) {
	if r.NooBaa.UID == "" || r.NooBaa.DeletionTimestamp != nil {
		return metav1.ConditionTrue, upgradeableReasonNoSystem, "no NooBaa system to protect"
	}

	if !r.shouldReconcileCNPGCluster() {
		// standalone or external DB - no migration is managed by the operator
		return metav1.ConditionTrue, upgradeableReasonNoMigration, "the operator does not manage a CNPG cluster"
	}

	if dbStatus := r.NooBaa.Status.DBStatus; dbStatus != nil {
		switch dbStatus.DBClusterStatus {
		case nbv1.DBClusterStatusCreating, nbv1.DBClusterStatusImporting:
			return metav1.ConditionFalse, notUpgradeableReasonDBMigration,
				fmt.Sprintf("the DB cluster status is %s", dbStatus.DBClusterStatus)
		}
	}

	// When the CNPG cluster becomes ready, ReconcileCNPGCluster scales the standalone DB statefulset to 0 replicas.
	// A standalone DB with replicas > 0 means the migration was not completed, and noobaa-core may still use it.
	standaloneDB := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NooBaaPostgresDB.Name,
			Namespace: r.NooBaaPostgresDB.Namespace,
		},
	}
	if util.KubeCheckQuiet(standaloneDB) && (standaloneDB.Spec.Replicas == nil || *standaloneDB.Spec.Replicas > 0) {
		return metav1.ConditionFalse, notUpgradeableReasonStandaloneDBUp,
			fmt.Sprintf("the standalone DB statefulset %s is not scaled down - the migration to the CNPG cluster is not completed",
				standaloneDB.Name)
	}

	return metav1.ConditionTrue, upgradeableReasonNoMigration, "no DB migration in progress"
}

// setOLMUpgradeableCondition sets the Upgradeable condition in the OperatorCondition of the running CSV.
// It skips the update when the condition is unchanged, to avoid bumping the OperatorCondition generation on every reconcile.
func setOLMUpgradeableCondition(status metav1.ConditionStatus, reason string, message string) error {
	cond, err := conditions.InClusterFactory{Client: util.KubeClient()}.NewCondition(operv2.ConditionType(operv2.Upgradeable))
	if err != nil {
		return err
	}

	// Get returns an error also when the condition type is not set yet. In that case we set it below.
	if current, err := cond.Get(util.Context()); err == nil &&
		current.Status == status && current.Reason == reason && current.Message == message {
		return nil
	}

	// Set does a Get + Update of the whole OperatorCondition, which can conflict with OLM writing it
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return cond.Set(util.Context(), status, conditions.WithReason(reason), conditions.WithMessage(message))
	})
}
