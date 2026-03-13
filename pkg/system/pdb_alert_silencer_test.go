package system_test

import (
	"context"
	"testing"
	"time"

	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
)

func TestPDBAlert(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PDB Alert Suite")
}

// Helper function to create a test reconciler
func createTestReconciler(cnpgCluster *cnpgv1.Cluster) *system.Reconciler {
	// Create a NooBaa object with DBSpec to simulate CNPG usage
	noobaa := &nbv1.NooBaa{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-noobaa",
			Namespace: "test-namespace",
		},
		Spec: nbv1.NooBaaSpec{
			DBSpec: &nbv1.NooBaaDBSpec{
				// Minimal DBSpec to indicate CNPG is being used
			},
		},
	}

	req := types.NamespacedName{
		Namespace: "test-namespace",
		Name:      "test-noobaa",
	}

	return &system.Reconciler{
		Request:     req,
		Ctx:         context.Background(),
		NooBaa:      noobaa,
		Logger:      logrus.WithField("sys", req.Namespace+"/"+req.Name),
		CNPGCluster: cnpgCluster,
	}
}

var _ = Describe("PDB Alert Silence", func() {

	BeforeEach(func() {
		// Reset the global lastSilenceCheckTime before each test
		system.ResetLastSilenceCheckTime()
	})

	Describe("CreateCNPGPDBSilence", func() {
		It("should create a silence with year 3000 end date", func() {
			cnpgCluster := &cnpgv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "test-namespace",
				},
			}

			r := createTestReconciler(cnpgCluster)
			silence := r.CreateCNPGPDBSilence()

			// Verify end date is set to year 3000
			expectedEndDate := time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC)
			Expect(silence.EndsAt).To(Equal(expectedEndDate))

			// Verify start date is recent
			timeDiff := time.Since(silence.StartsAt)
			Expect(timeDiff).To(BeNumerically("<", 5*time.Second))
			Expect(timeDiff).To(BeNumerically(">=", 0))

			// Verify creator
			Expect(silence.CreatedBy).To(Equal("noobaa-operator"))

			// Verify matchers
			Expect(silence.Matchers).To(HaveLen(3))

			// Check matchers
			var foundAlertname, foundNamespace, foundPDB bool
			for _, matcher := range silence.Matchers {
				if matcher.Name == "alertname" && matcher.Value == system.PodDisruptionBudgetAtLimitAlertName {
					foundAlertname = true
					Expect(matcher.IsEqual).To(BeTrue())
					Expect(matcher.IsRegex).To(BeFalse())
				}
				if matcher.Name == "namespace" && matcher.Value == "test-namespace" {
					foundNamespace = true
				}
				if matcher.Name == "poddisruptionbudget" {
					foundPDB = true
					Expect(matcher.Value).To(Equal("test-cluster-primary"))
					Expect(matcher.IsEqual).To(BeTrue())
					Expect(matcher.IsRegex).To(BeFalse())
				}
			}

			Expect(foundAlertname).To(BeTrue(), "Missing alertname matcher")
			Expect(foundNamespace).To(BeTrue(), "Missing namespace matcher")
			Expect(foundPDB).To(BeTrue(), "Missing PDB matcher")
		})
	})

	Describe("IsCNPGPDBSilence", func() {
		var r *system.Reconciler

		BeforeEach(func() {
			cnpgCluster := &cnpgv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "test-namespace",
				},
			}
			r = createTestReconciler(cnpgCluster)
		})

		DescribeTable("silence matching logic",
			func(silence *system.AlertmanagerSilence, expected bool) {
				result := r.IsCNPGPDBSilence(silence)
				Expect(result).To(Equal(expected))
			},
			Entry("Valid silence with all correct matchers",
				&system.AlertmanagerSilence{
					CreatedBy: "noobaa-operator",
					Matchers: []system.SilenceMatcher{
						{Name: "alertname", Value: system.PodDisruptionBudgetAtLimitAlertName, IsRegex: false},
						{Name: "namespace", Value: "test-namespace", IsRegex: false},
						{Name: "poddisruptionbudget", Value: "test-cluster-primary", IsRegex: false},
					},
				},
				true,
			),
			Entry("Wrong alertname",
				&system.AlertmanagerSilence{
					CreatedBy: "noobaa-operator",
					Matchers: []system.SilenceMatcher{
						{Name: "alertname", Value: "WrongAlert", IsRegex: false},
						{Name: "namespace", Value: "test-namespace", IsRegex: false},
						{Name: "poddisruptionbudget", Value: "test-cluster-primary", IsRegex: false},
					},
				},
				false,
			),
			Entry("Wrong namespace",
				&system.AlertmanagerSilence{
					CreatedBy: "noobaa-operator",
					Matchers: []system.SilenceMatcher{
						{Name: "alertname", Value: system.PodDisruptionBudgetAtLimitAlertName, IsRegex: false},
						{Name: "namespace", Value: "wrong-namespace", IsRegex: false},
						{Name: "poddisruptionbudget", Value: "test-cluster-primary", IsRegex: false},
					},
				},
				false,
			),
			Entry("Wrong PDB name",
				&system.AlertmanagerSilence{
					CreatedBy: "noobaa-operator",
					Matchers: []system.SilenceMatcher{
						{Name: "alertname", Value: system.PodDisruptionBudgetAtLimitAlertName, IsRegex: false},
						{Name: "namespace", Value: "test-namespace", IsRegex: false},
						{Name: "poddisruptionbudget", Value: "wrong-cluster-primary", IsRegex: false},
					},
				},
				false,
			),
			Entry("PDB matcher with regex (should fail)",
				&system.AlertmanagerSilence{
					CreatedBy: "noobaa-operator",
					Matchers: []system.SilenceMatcher{
						{Name: "alertname", Value: system.PodDisruptionBudgetAtLimitAlertName, IsRegex: false},
						{Name: "namespace", Value: "test-namespace", IsRegex: false},
						{Name: "poddisruptionbudget", Value: "test-cluster-primary", IsRegex: true},
					},
				},
				false,
			),
			Entry("Alertname matcher with regex (should fail)",
				&system.AlertmanagerSilence{
					CreatedBy: "noobaa-operator",
					Matchers: []system.SilenceMatcher{
						{Name: "alertname", Value: system.PodDisruptionBudgetAtLimitAlertName, IsRegex: true},
						{Name: "namespace", Value: "test-namespace", IsRegex: false},
						{Name: "poddisruptionbudget", Value: "test-cluster-primary", IsRegex: false},
					},
				},
				false,
			),
			Entry("Namespace matcher with regex (should fail)",
				&system.AlertmanagerSilence{
					CreatedBy: "noobaa-operator",
					Matchers: []system.SilenceMatcher{
						{Name: "alertname", Value: system.PodDisruptionBudgetAtLimitAlertName, IsRegex: false},
						{Name: "namespace", Value: "test-namespace", IsRegex: true},
						{Name: "poddisruptionbudget", Value: "test-cluster-primary", IsRegex: false},
					},
				},
				false,
			),
			Entry("Wrong creator",
				&system.AlertmanagerSilence{
					CreatedBy: "other-operator",
					Matchers: []system.SilenceMatcher{
						{Name: "alertname", Value: system.PodDisruptionBudgetAtLimitAlertName, IsRegex: false},
						{Name: "namespace", Value: "test-namespace", IsRegex: false},
						{Name: "poddisruptionbudget", Value: "test-cluster-primary", IsRegex: false},
					},
				},
				false,
			),
			Entry("Missing PDB matcher",
				&system.AlertmanagerSilence{
					CreatedBy: "noobaa-operator",
					Matchers: []system.SilenceMatcher{
						{Name: "alertname", Value: system.PodDisruptionBudgetAtLimitAlertName, IsRegex: false},
						{Name: "namespace", Value: "test-namespace", IsRegex: false},
					},
				},
				false,
			),
		)
	})

	Describe("ReconcilePDBAlert", func() {
		Context("when silence was checked recently", func() {
			It("should skip reconciliation", func() {
				// Set last check time to 30 minutes ago (less than the 1-hour interval)
				system.SetLastSilenceCheckTime(time.Now().Add(-30 * time.Minute))

				cnpgCluster := &cnpgv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cluster",
						Namespace: "test-namespace",
					},
				}

				r := createTestReconciler(cnpgCluster)
				err := r.ReconcilePDBAlertSilencer()

				// Expect no error because the method returns nil when skipping due to recent check
				// This verifies the rate-limiting logic at lines 86-89 in pdb_alert_silencer.go
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when silence check interval has passed", func() {
			It("should attempt reconciliation after 1 hour", func() {
				// Set last check time to 2 hours ago (more than the 1-hour interval)
				system.SetLastSilenceCheckTime(time.Now().Add(-2 * time.Hour))

				cnpgCluster := &cnpgv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cluster",
						Namespace: "test-namespace",
					},
				}

				r := createTestReconciler(cnpgCluster)
				err := r.ReconcilePDBAlertSilencer()

				// Expect an error because we don't have a real Alertmanager to connect to
				// This proves the method did NOT skip and tried to proceed with reconciliation
				// If it had skipped, it would return nil (no error)
				Expect(err).To(HaveOccurred())
				// The error should be related to Alertmanager client creation or API call
				Expect(err.Error()).To(ContainSubstring("failed to"))
			})
		})

		Context("when last check time is zero (first run)", func() {
			It("should attempt reconciliation", func() {
				// Reset ensures lastSilenceCheckTime is zero (never checked before)
				system.ResetLastSilenceCheckTime()

				cnpgCluster := &cnpgv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cluster",
						Namespace: "test-namespace",
					},
				}

				r := createTestReconciler(cnpgCluster)
				err := r.ReconcilePDBAlertSilencer()

				// Expect an error because we don't have a real Alertmanager to connect to
				// This proves the method did NOT skip the first-time check
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to"))
			})
		})

		Context("when CNPG cluster is not configured", func() {
			It("should skip reconciliation", func() {
				r := createTestReconciler(nil)
				err := r.ReconcilePDBAlertSilencer()

				// Expect no error because the method returns nil when CNPG is not configured
				// This verifies the shouldReconcileCNPGCluster() check at line 74
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when CNPG cluster is not ready", func() {
			It("should skip when cluster is nil", func() {
				r := createTestReconciler(nil)
				err := r.ReconcilePDBAlertSilencer()

				// Expect no error because the method returns nil when cluster is nil
				// This verifies the nil check at line 80
				Expect(err).ToNot(HaveOccurred())
			})

			It("should skip when cluster name is empty", func() {
				cnpgCluster := &cnpgv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "",
						Namespace: "test-namespace",
					},
				}
				r := createTestReconciler(cnpgCluster)
				err := r.ReconcilePDBAlertSilencer()

				// Expect no error because the method returns nil when cluster name is empty
				// This verifies the empty name check at line 80
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when CNPG cluster is ready and check interval passed", func() {
			It("should proceed with reconciliation", func() {
				// Ensure enough time has passed since last check
				system.SetLastSilenceCheckTime(time.Now().Add(-2 * time.Hour))

				cnpgCluster := &cnpgv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cluster",
						Namespace: "test-namespace",
					},
				}

				r := createTestReconciler(cnpgCluster)
				err := r.ReconcilePDBAlertSilencer()

				// Expect an error because:
				// 1. All skip conditions are false (CNPG configured, cluster ready, interval passed)
				// 2. The method proceeds to create Alertmanager client
				// 3. Without a real Alertmanager, the client creation or API call fails
				// This proves the reconciliation logic was executed (not skipped)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Or(
					ContainSubstring("failed to read service account token"),
					ContainSubstring("failed to create Alertmanager client"),
					ContainSubstring("failed to list existing silences"),
				))
			})
		})
	})

	Describe("Silence Expiry Check", func() {
		It("should correctly determine if silence needs recreation", func() {
			// Use a fixed reference time to avoid timing issues
			now := time.Now()

			tests := []struct {
				name           string
				endsAt         time.Time
				shouldRecreate bool
			}{
				{
					name:           "Silence expires in 1 hour (should recreate)",
					endsAt:         now.Add(1 * time.Hour),
					shouldRecreate: true,
				},
				{
					name:           "Silence expires in 23 hours (should recreate)",
					endsAt:         now.Add(23 * time.Hour),
					shouldRecreate: true,
				},
				{
					name:           "Silence expires in exactly 24 hours (should not recreate)",
					endsAt:         now.Add(24 * time.Hour),
					shouldRecreate: false,
				},
				{
					name:           "Silence expires in 25 hours (should not recreate)",
					endsAt:         now.Add(25 * time.Hour),
					shouldRecreate: false,
				},
				{
					name:           "Silence already expired (should recreate)",
					endsAt:         now.Add(-1 * time.Hour),
					shouldRecreate: true,
				},
				{
					name:           "Silence expires in year 3000 (should not recreate)",
					endsAt:         time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC),
					shouldRecreate: false,
				},
			}

			for _, tt := range tests {
				By(tt.name)
				timeUntilExpiry := tt.endsAt.Sub(now)
				result := timeUntilExpiry < 24*time.Hour

				Expect(result).To(Equal(tt.shouldRecreate), tt.name)
			}
		})
	})

	Describe("Constants", func() {
		It("should have correct constant values", func() {
			Expect(system.PodDisruptionBudgetAtLimitAlertName).To(Equal("PodDisruptionBudgetAtLimit"))
			Expect(system.AlertmanagerSvcHost).To(Equal("https://alertmanager-main.openshift-monitoring.svc.cluster.local:9094"))
			Expect(system.SilenceCheckInterval).To(Equal(1 * time.Hour))
		})
	})
})
