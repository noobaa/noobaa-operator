package system

import (
	"fmt"
	"os"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	operv2 "github.com/operator-framework/api/pkg/operators/v2"
	"github.com/operator-framework/operator-lib/conditions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

const (
	// operatorConditionNameEnv is set by OLM on every deployment of the CSV.
	// It holds the name of the OperatorCondition CR of the running CSV.
	operatorConditionNameEnv = "OPERATOR_CONDITION_NAME"

	upgradeableReasonNoSystem            = "NoSystem"
	upgradeableReasonDbReady             = "DBReady"
	upgradeableReasonDbExternal          = "DBExternal"
	notUpgradeableReasonDBCreating       = "DBCreating"
	notUpgradeableReasonOperatorStarting = "OperatorStarting"
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
func (r *Reconciler) ReconcileOLMUpgradeable() error {
	if os.Getenv(operatorConditionNameEnv) == "" {
		// not deployed by OLM - there is no OperatorCondition to update
		return nil
	}

	status, reason, message := r.getOLMUpgradeableStatus()
	if err := setOLMUpgradeableCondition(status, reason, message); err != nil {
		r.Logger.Warnf("failed to set OLM Upgradeable=%s (%s): %v", status, reason, err)
		return err
	}
	return nil
}

// getOLMUpgradeableStatus returns the desired Upgradeable status, reason and message
func (r *Reconciler) getOLMUpgradeableStatus() (metav1.ConditionStatus, string, string) {
	if r.NooBaa.UID == "" || r.NooBaa.DeletionTimestamp != nil {
		return metav1.ConditionTrue, upgradeableReasonNoSystem, "no NooBaa system to protect"
	}

	if r.NooBaa.Spec.ExternalPgSecret != nil {
		// external DB - no migration is managed by the operator
		return metav1.ConditionTrue, upgradeableReasonDbExternal, "the operator does not manage a CNPG cluster"
	}

	if dbStatus := r.NooBaa.Status.DBStatus; dbStatus != nil {
		if (dbStatus.DBClusterStatus == nbv1.DBClusterStatusCreating) {
			return metav1.ConditionFalse, notUpgradeableReasonDBCreating,
				fmt.Sprintf("the DB cluster status is %s", dbStatus.DBClusterStatus)
		}
	}

	return metav1.ConditionTrue, upgradeableReasonDbReady, "DB is ready"
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

func InitOLMUpgradeable() error {
	if os.Getenv(operatorConditionNameEnv) == "" {
		return nil
	}
	sys := &nbv1.NooBaa{
		ObjectMeta: metav1.ObjectMeta {
			Name:      options.SystemName,
			Namespace: options.Namespace,
		},
	}
	var status metav1.ConditionStatus
	var reason string
	var message string
	if !util.KubeCheckQuiet(sys) || sys.GetDeletionTimestamp() != nil {
		status = metav1.ConditionTrue
		reason = upgradeableReasonNoSystem
		message = "no NooBaa system to protect"
	} else {
		status = metav1.ConditionFalse
		reason = notUpgradeableReasonOperatorStarting
		message = "System is starting up"
	}
	util.Logger().Infof("initial OLM Upgradeable=%s (%s): %s", status, reason, message)
	return setOLMUpgradeableCondition(status, reason, message)
}
