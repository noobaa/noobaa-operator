package system

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
)

const (
	// PodDisruptionBudgetAtLimitAlertName is the name of the alert to silence
	PodDisruptionBudgetAtLimitAlertName = "PodDisruptionBudgetAtLimit"
	// AlertmanagerSvcHost is the hardcoded Alertmanager service URL
	AlertmanagerSvcHost = "https://alertmanager-main.openshift-monitoring.svc.cluster.local:9094"
	// ServiceAccountTokenPath is the path to the service account token
	ServiceAccountTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	// ServiceAccountCACertPath is the path to the service account CA certificate
	ServiceAccountCACertPath = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	// SilenceCheckInterval is how often to check if silence exists (1 hour)
	SilenceCheckInterval = 1 * time.Hour
)

var (
	// lastSilenceCheckTime tracks when we last checked/created the silence
	lastSilenceCheckTime time.Time
)

// ResetLastSilenceCheckTime resets the last silence check time (for testing)
func ResetLastSilenceCheckTime() {
	lastSilenceCheckTime = time.Time{}
}

// SetLastSilenceCheckTime sets the last silence check time (for testing)
func SetLastSilenceCheckTime(t time.Time) {
	lastSilenceCheckTime = t
}

// AlertmanagerSilence represents an Alertmanager silence
type AlertmanagerSilence struct {
	ID        string           `json:"id,omitempty"`
	Status    *SilenceStatus   `json:"status,omitempty"`
	Matchers  []SilenceMatcher `json:"matchers"`
	StartsAt  time.Time        `json:"startsAt"`
	EndsAt    time.Time        `json:"endsAt"`
	CreatedBy string           `json:"createdBy"`
	Comment   string           `json:"comment"`
}

// SilenceStatus represents the status of a silence
type SilenceStatus struct {
	State string `json:"state"`
}

// SilenceMatcher represents a matcher for a silence
type SilenceMatcher struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	IsRegex bool   `json:"isRegex"`
	IsEqual bool   `json:"isEqual"`
}

// ReconcilePDBAlertSilencer creates or updates an Alertmanager silence for the PodDisruptionBudgetAtLimit alert
// for the CNPG cluster PDB. This is necessary because CNPG creates a PDB with minAvailable=1 for the
// primary instance, which always has disruptionsAllowed=0, causing the alert to fire continuously.
func (r *Reconciler) ReconcilePDBAlertSilencer() error {
	if !r.shouldReconcileCNPGCluster() {
		// If not using CNPG, no need to create silence
		return nil
	}

	// Get the CNPG cluster name to create a specific silence
	if r.CNPGCluster == nil || r.CNPGCluster.Name == "" {
		r.cnpgLog("CNPG cluster not yet created, skipping PDB alert silence")
		return nil
	}

	// Check if we need to query the API based on the last check time
	if !lastSilenceCheckTime.IsZero() && time.Since(lastSilenceCheckTime) < SilenceCheckInterval {
		return nil
	}

	r.cnpgLog("reconciling PDB alert silence for CNPG cluster %s", r.CNPGCluster.Name)

	// Create Alertmanager API client using configured or default URL
	alertmanagerHost := r.NooBaa.Annotations[nbv1.AlertmanagerHostOverride]
	if alertmanagerHost == "" {
		alertmanagerHost = AlertmanagerSvcHost
	}
	amClient, err := r.newAlertmanagerClient(alertmanagerHost)
	if err != nil {
		r.cnpgLogError("failed to create Alertmanager client: %v", err)
		return err
	}

	// Check if silence already exists
	existingSilences, err := amClient.listSilences()
	if err != nil {
		// if the alertsmanager host is not found (DNSError), assume we are not running in openshift and skip
		// will retry in the next silence check interval (1h)
		var dnsError *net.DNSError
		if errors.As(err, &dnsError) {
			r.cnpgLog("alertsmanager host is not found (DNSError), assuming we are not running in openshift and skipping")
			SetLastSilenceCheckTime(time.Now())
			return nil
		}
		r.cnpgLogError("failed to list existing silences: %v", err)
		return err
	}

	// Look for our silence
	var existingSilence *AlertmanagerSilence
	for i := range existingSilences {
		silence := &existingSilences[i]
		if r.IsCNPGPDBSilence(silence) {
			existingSilence = silence
			break
		}
	}

	// Create the silence specification
	silence := r.CreateCNPGPDBSilence()

	if existingSilence != nil {
		// Check if silence is expired or about to expire in the next 24 hours
		timeUntilExpiry := time.Until(existingSilence.EndsAt)
		if timeUntilExpiry < 24*time.Hour {
			r.cnpgLog("PDB alert silence is expired or expiring soon (ID: %s, expires: %s), recreating",
				existingSilence.ID, existingSilence.EndsAt.Format(time.RFC3339))
			silence.ID = existingSilence.ID
			if err := amClient.createOrUpdateSilence(silence); err != nil {
				r.cnpgLogError("failed to recreate silence: %v", err)
				return err
			}
			r.cnpgLog("successfully recreated PDB alert silence (indefinite)")
		} else {
			// Silence exists and is valid, update last check time and skip
			r.cnpgLog("PDB alert silence already exists and is valid (ID: %s, expires: %s)",
				existingSilence.ID, existingSilence.EndsAt.Format(time.RFC3339))
		}
		SetLastSilenceCheckTime(time.Now())
		return nil
	}

	// Create new silence
	r.cnpgLog("creating new PDB alert silence")
	if err := amClient.createOrUpdateSilence(silence); err != nil {
		r.cnpgLogError("failed to create silence: %v", err)
		return err
	}
	r.cnpgLog("successfully created PDB alert silence (indefinite)")

	// Update last check time
	SetLastSilenceCheckTime(time.Now())

	return nil
}

// CreateCNPGPDBSilence creates a silence specification for CNPG PDB alerts
func (r *Reconciler) CreateCNPGPDBSilence() *AlertmanagerSilence {
	now := time.Now()
	// Set expiration to year 3000 to silence for a very long time
	maxTime := time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC)

	silenceComment := fmt.Sprintf(
		"The PodDisruptionBudgetAtLimit alert for the NooBaa CNPG cluster %s is silenced by noobaa-operator. "+
			"The alert is expected and does not indicate a problem. "+
			"The PDB %s has disruptionsAllowed=0 by design to prevent the primary instance from being deleted during node drains. "+
			"This is safe because CNPG automatically performs a graceful switchover to a secondary instance when a node is drained. "+
			"Once the instance in the node is downgraded and is no longer the primary, draining can resume. ",
		r.CNPGCluster.Name, r.CNPGCluster.Name+"-primary")

	return &AlertmanagerSilence{
		Matchers: []SilenceMatcher{
			{
				Name:    "alertname",
				Value:   PodDisruptionBudgetAtLimitAlertName,
				IsRegex: false,
				IsEqual: true,
			},
			{
				Name:    "namespace",
				Value:   r.Request.Namespace,
				IsRegex: false,
				IsEqual: true,
			},
			{
				Name:    "poddisruptionbudget",
				Value:   fmt.Sprintf("%s-primary", r.CNPGCluster.Name),
				IsRegex: false,
				IsEqual: true,
			},
		},
		StartsAt:  now,
		EndsAt:    maxTime,
		CreatedBy: "noobaa-operator",
		Comment:   silenceComment,
	}
}

// IsCNPGPDBSilence checks if a silence is for our CNPG PDB
func (r *Reconciler) IsCNPGPDBSilence(silence *AlertmanagerSilence) bool {
	if silence.CreatedBy != "noobaa-operator" {
		return false
	}

	expectedPDBName := fmt.Sprintf("%s-primary", r.CNPGCluster.Name)

	hasAlertname := false
	hasNamespace := false
	hasPDB := false

	for _, matcher := range silence.Matchers {
		if matcher.Name == "alertname" && matcher.Value == PodDisruptionBudgetAtLimitAlertName && !matcher.IsRegex {
			hasAlertname = true
		}
		if matcher.Name == "namespace" && matcher.Value == r.Request.Namespace && !matcher.IsRegex {
			hasNamespace = true
		}
		if matcher.Name == "poddisruptionbudget" && matcher.Value == expectedPDBName && !matcher.IsRegex {
			hasPDB = true
		}
	}

	return hasAlertname && hasNamespace && hasPDB
}

// AlertmanagerClient is a client for the Alertmanager API
type AlertmanagerClient struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

// newAlertmanagerClient creates a new Alertmanager API client
func (r *Reconciler) newAlertmanagerClient(baseURL string) (*AlertmanagerClient, error) {
	// Read service account token
	tokenBytes, err := os.ReadFile(ServiceAccountTokenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read service account token: %w", err)
	}
	token := string(tokenBytes)

	// Use the global CA refreshing transport which includes system certs
	// and any injected CA bundles (e.g., from OpenShift service CA)
	httpClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: util.GlobalCARefreshingTransport,
	}

	return &AlertmanagerClient{
		httpClient: httpClient,
		baseURL:    baseURL,
		token:      token,
	}, nil
}

// listSilences lists all silences from Alertmanager
func (c *AlertmanagerClient) listSilences() ([]AlertmanagerSilence, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v2/silences", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			util.Logger().Errorf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var silences []AlertmanagerSilence
	if err := json.NewDecoder(resp.Body).Decode(&silences); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return silences, nil
}

// createOrUpdateSilence creates or updates a silence in Alertmanager
func (c *AlertmanagerClient) createOrUpdateSilence(silence *AlertmanagerSilence) error {
	body, err := json.Marshal(silence)
	if err != nil {
		return fmt.Errorf("failed to marshal silence: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/v2/silences", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			util.Logger().Errorf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
