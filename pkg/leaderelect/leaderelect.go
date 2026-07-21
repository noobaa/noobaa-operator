// Package leaderelect implements a PID-1 leader-election exec wrapper.
//
// Exposed as a Hidden cobra subcommand of the noobaa-operator binary and
// copied into the core pod from the operator image:
//
//	noobaa-operator leader-elect -- <command> [args...]
//
// Timings and lease name come from environment variables (see NOOBAA_CORE_*).
// The command is Hidden so it does not appear in public CLI help. The wrapper
// acquires a namespaced Lease, spawns the given command in its own process
// group while holding leadership, reaps orphaned children (PID 1), and on
// SIGTERM/SIGINT or leadership loss stops the child group, cancels the elector
// (stopping renew), then releases the Lease once le.Run has returned.
package leaderelect

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/spf13/cobra"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

const (
	exitOK     = 0
	exitConfig = 1
	exitLost   = 2

	envLeaseName     = "NOOBAA_CORE_LEASE_NAME"
	envLeaseDuration = "NOOBAA_CORE_LEASE_DURATION"
	envRenewDeadline = "NOOBAA_CORE_RENEW_DEADLINE"
	envRetryPeriod   = "NOOBAA_CORE_RETRY_PERIOD"
	envShutdownGrace = "NOOBAA_CORE_SHUTDOWN_GRACE"
	envLostGrace     = "NOOBAA_CORE_LOST_GRACE"
	envPodNamespace  = "POD_NAMESPACE"
	saNamespaceFile  = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

	defaultLeaseDuration = 20 * time.Second
	defaultRenewDeadline = 10 * time.Second
	defaultRetryPeriod   = 3 * time.Second
	defaultShutdownGrace = 25 * time.Second
	defaultLostGrace     = 8 * time.Second
	// childSigkillWait is how long stopChild waits after SIGKILL for the child process group to exit before returning.
	childSigkillWait = 5 * time.Second
	// leaseReleaseMargin is extra time after child teardown for the manual lease release Update.
	leaseReleaseMargin = 10 * time.Second
)

// RecommendedTerminationGracePeriod is the pod terminationGracePeriodSeconds
// that leaves room for ShutdownGrace + SIGKILL wait + manual lease release.
func RecommendedTerminationGracePeriod(shutdownGrace time.Duration) int64 {
	if shutdownGrace <= 0 {
		shutdownGrace = defaultShutdownGrace
	}
	return int64((shutdownGrace + childSigkillWait + leaseReleaseMargin) / time.Second)
}

// Config holds leader-elect wrapper settings.
type Config struct {
	LeaseName     string
	LeaseDuration time.Duration
	RenewDeadline time.Duration
	RetryPeriod   time.Duration
	ShutdownGrace time.Duration
	LostGrace     time.Duration
	Command       []string
}

// Cmd returns the leader-elect CLI command (Hidden — for in-pod use).
// Configuration is read from environment variables; args are only the wrapped command.
func Cmd() *cobra.Command {
	return &cobra.Command{
		Hidden:                true,
		Use:                   "leader-elect -- <command> [args...]",
		Short:                 "Acquire a Lease then exec a command (PID 1 leader-elect wrapper)",
		DisableFlagsInUseLine: true,
		DisableFlagParsing:    true,
		SilenceUsage:          true,
		Run: func(_ *cobra.Command, args []string) {
			cfg := defaultConfig()
			cfg.Command = commandFromArgs(args)
			if err := cfg.Validate(); err != nil {
				fmt.Fprintf(os.Stderr, "leader-elect: %v\n", err)
				os.Exit(exitConfig)
			}
			os.Exit(Run(cfg))
		},
	}
}

func defaultConfig() *Config {
	return &Config{
		LeaseName:     os.Getenv(envLeaseName),
		LeaseDuration: durationFromEnv(envLeaseDuration, defaultLeaseDuration),
		RenewDeadline: durationFromEnv(envRenewDeadline, defaultRenewDeadline),
		RetryPeriod:   durationFromEnv(envRetryPeriod, defaultRetryPeriod),
		ShutdownGrace: durationFromEnv(envShutdownGrace, defaultShutdownGrace),
		LostGrace:     durationFromEnv(envLostGrace, defaultLostGrace),
	}
}

// durationFromEnv parses a duration env var; returns fallback if unset or invalid.
func durationFromEnv(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}

// commandFromArgs returns the wrapped command, stripping a leading "--" if present.
func commandFromArgs(args []string) []string {
	if len(args) == 0 {
		return nil
	}
	if args[0] == "--" {
		return append([]string(nil), args[1:]...)
	}
	return append([]string(nil), args...)
}

// ParseArgs builds Config from the environment and the wrapped command args.
// It does not run the election or spawn processes.
func ParseArgs(args []string) (*Config, error) {
	cfg := defaultConfig()
	cfg.Command = commandFromArgs(args)
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate checks required fields and client-go timing invariants.
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("nil config")
	}
	if c.LeaseName == "" {
		return fmt.Errorf("$%s is required", envLeaseName)
	}
	if len(c.Command) == 0 {
		return fmt.Errorf("command is required after --")
	}
	maxLost := c.LeaseDuration - c.RenewDeadline
	if maxLost <= 0 {
		return fmt.Errorf("%s (%v) must be greater than %s (%v)",
			envLeaseDuration, c.LeaseDuration, envRenewDeadline, c.RenewDeadline)
	}
	if c.LostGrace >= maxLost {
		return fmt.Errorf("%s (%v) must be < %s - %s (%v)",
			envLostGrace, c.LostGrace, envLeaseDuration, envRenewDeadline, maxLost)
	}
	// Match NewLeaderElector: renewDeadline must be > retryPeriod*JitterFactor (1.2).
	if c.RetryPeriod <= 0 {
		return fmt.Errorf("%s (%v) must be greater than zero", envRetryPeriod, c.RetryPeriod)
	}
	minRenew := time.Duration(leaderelection.JitterFactor * float64(c.RetryPeriod))
	if c.RenewDeadline <= minRenew {
		return fmt.Errorf("%s (%v) must be greater than %s*%g (%v)",
			envRenewDeadline, c.RenewDeadline, envRetryPeriod, leaderelection.JitterFactor, minRenew)
	}
	return nil
}

// Run acquires the Lease, runs the command while leading, and returns a process exit code.
func Run(cfg *Config) int {
	log := util.Logger()

	namespace := resolveNamespace()
	identity := leaderIdentity()

	clientset, err := kubernetes.NewForConfig(util.KubeConfig())
	if err != nil {
		log.Errorf("leader-elect: create clientset: %v", err)
		return exitConfig
	}

	lock, err := resourcelock.New(
		resourcelock.LeasesResourceLock,
		namespace,
		cfg.LeaseName,
		clientset.CoreV1(),
		clientset.CoordinationV1(),
		resourcelock.ResourceLockConfig{Identity: identity},
	)
	if err != nil {
		log.Errorf("leader-elect: create lock: %v", err)
		return exitConfig
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := &runner{
		cfg:      cfg,
		cancel:   cancel,
		log:      log,
		lock:     lock,
		identity: identity,
	}

	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(sigCh)

	go func() {
		sig := <-sigCh
		r.log.Infof("leader-elect: received %v, stopping child then cancelling election", sig)
		r.stopChild(cfg.ShutdownGrace)
		r.forceExit(exitOK)
		cancel()
	}()

	// closed when OnStartedLeading returns, or immediately if we never led.
	leadingDone := make(chan struct{})
	closeLeadingDone := sync.OnceFunc(func() { close(leadingDone) })
	var led bool
	var ledMu sync.Mutex

	le, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:          lock,
		LeaseDuration: cfg.LeaseDuration,
		RenewDeadline: cfg.RenewDeadline,
		RetryPeriod:   cfg.RetryPeriod,
		// Do not release on cancel: renew cancel can race with stopChild and
		// hand the Lease to a standby while core is still running.
		// We release manually after le.Run returns and OnStartedLeading finishes.
		ReleaseOnCancel: false,
		Name:            cfg.LeaseName,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(leadCtx context.Context) {
				ledMu.Lock()
				led = true
				ledMu.Unlock()
				defer closeLeadingDone()
				r.onStartedLeading(leadCtx)
			},
			OnStoppedLeading: func() {
				r.log.Infof("leader-elect: stopped leading (identity %s)", identity)
				ledMu.Lock()
				didLead := led
				ledMu.Unlock()
				if !didLead {
					closeLeadingDone()
				}
			},
		},
	})
	if err != nil {
		log.Errorf("leader-elect: create elector: %v", err)
		return exitConfig
	}

	log.Infof("leader-elect: waiting for lease %s/%s as %s", namespace, cfg.LeaseName, identity)
	le.Run(ctx)
	// Renew has stopped. Wait for leading teardown (esp. loss-path stopChild) before release.
	<-leadingDone
	r.releaseLease()

	return r.getExit()
}

func resolveNamespace() string {
	if ns := os.Getenv(envPodNamespace); ns != "" {
		return ns
	}
	data, err := os.ReadFile(saNamespaceFile)
	if err == nil {
		if ns := strings.TrimSpace(string(data)); ns != "" {
			return ns
		}
	}
	return options.Namespace
}

func leaderIdentity() string {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "unknown"
	}
	return hostname + "_" + string(uuid.NewUUID())
}

type logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type runner struct {
	cfg      *Config
	cancel   context.CancelFunc
	log      logger
	lock     resourcelock.Interface
	identity string

	mu            sync.Mutex
	childPID      int
	childExited   chan struct{} // closed once the main child is reaped
	childCode     int
	exitCode      int
	exitSet       bool
	termRequested bool
	releaseOnce   sync.Once
}

func (r *runner) setExit(code int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.exitSet {
		r.exitCode = code
		r.exitSet = true
	}
}

func (r *runner) forceExit(code int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.exitCode = code
	r.exitSet = true
}

func (r *runner) getExit() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.exitSet {
		// Cancelled before leading (e.g. SIGTERM while waiting for lease).
		return exitOK
	}
	return r.exitCode
}

func (r *runner) onStartedLeading(leadCtx context.Context) {
	r.log.Infof("leader-elect: acquired lease, starting %v", r.cfg.Command)

	cmd := exec.Command(r.cfg.Command[0], r.cfg.Command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		r.log.Errorf("leader-elect: spawn: %v", err)
		r.setExit(exitConfig)
		r.cancel()
		return
	}

	childPID := cmd.Process.Pid
	childDone := make(chan struct{})

	r.mu.Lock()
	r.childPID = childPID
	r.childExited = childDone
	// SIGTERM may have arrived before Start() while childPID was 0; stopChild
	// returned without signaling. Re-check under the same lock as the publish.
	alreadyTerm := r.termRequested
	r.mu.Unlock()

	// Single Wait4(-1) loop: observes the main child and reaps orphans (PID 1).
	go r.waitAndReap(childPID, childDone)

	if alreadyTerm {
		r.stopChild(r.cfg.ShutdownGrace)
		return
	}

	select {
	case <-childDone:
		r.mu.Lock()
		termRequested := r.termRequested
		r.mu.Unlock()
		if termRequested {
			// SIGTERM path owns the wrapper exit code; lease release is after le.Run.
			return
		}
		code := r.getChildCode()
		r.log.Infof("leader-elect: child exited with code %d", code)
		r.setExit(code)
		r.cancel()
	case <-leadCtx.Done():
		r.mu.Lock()
		termRequested := r.termRequested
		r.mu.Unlock()
		if termRequested {
			// SIGTERM path owns teardown; wait for child then return.
			select {
			case <-childDone:
			case <-time.After(r.cfg.ShutdownGrace + childSigkillWait):
			}
			return
		}
		r.log.Errorf("leader-elect: leadership lost, stopping child")
		r.stopChild(r.cfg.LostGrace)
		r.forceExit(exitLost)
		select {
		case <-childDone:
		case <-time.After(r.cfg.LostGrace + childSigkillWait):
		}
	}
}

// releaseLease clears our HolderIdentity after renew has stopped and the child
// is down so a standby can acquire immediately. Crashes still rely on Lease TTL.
// Workaround until client-go waits for OnStartedLeading before ReleaseOnCancel:
// https://github.com/kubernetes/kubernetes/pull/139991
func (r *runner) releaseLease() {
	r.releaseOnce.Do(func() {
		if r.lock == nil {
			return
		}
		deadline := r.cfg.RenewDeadline
		if deadline <= 0 {
			deadline = defaultRenewDeadline
		}
		ctx, cancel := context.WithTimeout(context.Background(), deadline)
		defer cancel()

		old, _, err := r.lock.Get(ctx)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				r.log.Errorf("leader-elect: get lease for release: %v", err)
			}
			return
		}
		if old.HolderIdentity != r.identity {
			return
		}

		now := metav1.NewTime(time.Now())
		rec := resourcelock.LeaderElectionRecord{
			LeaderTransitions:    old.LeaderTransitions,
			LeaseDurationSeconds: 1,
			RenewTime:            now,
			AcquireTime:          now,
		}
		if err := r.lock.Update(ctx, rec); err != nil {
			r.log.Errorf("leader-elect: release lease: %v", err)
			return
		}
		r.log.Infof("leader-elect: released lease")
	})
}

func (r *runner) getChildCode() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.childCode
}

// waitAndReap blocks on Wait4(-1), records the main child's exit code once, and
// continues reaping any orphaned grandchildren until no children remain.
func (r *runner) waitAndReap(childPID int, done chan struct{}) {
	childReported := false
	for {
		var status syscall.WaitStatus
		wpid, err := syscall.Wait4(-1, &status, 0, nil)
		if err == syscall.EINTR {
			continue
		}
		if err != nil {
			// ECHILD: no children left.
			if !childReported {
				r.mu.Lock()
				r.childCode = exitConfig
				r.mu.Unlock()
				close(done)
			}
			return
		}
		if wpid == childPID && !childReported {
			childReported = true
			r.mu.Lock()
			r.childCode = waitStatusExitCode(status)
			r.mu.Unlock()
			close(done)
		}
	}
}

func waitStatusExitCode(status syscall.WaitStatus) int {
	if status.Exited() {
		return status.ExitStatus()
	}
	if status.Signaled() {
		return 128 + int(status.Signal())
	}
	return exitConfig
}

func (r *runner) stopChild(grace time.Duration) {
	r.mu.Lock()
	r.termRequested = true
	pid := r.childPID
	done := r.childExited
	r.mu.Unlock()

	if pid <= 0 {
		return
	}

	// Signal the whole process group (Setpgid makes pgid == child pid).
	if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil && err != syscall.ESRCH {
		r.log.Errorf("leader-elect: kill pgid %d SIGTERM: %v", pid, err)
	}

	if done == nil {
		return
	}

	timer := time.NewTimer(grace)
	defer timer.Stop()
	select {
	case <-done:
		return
	case <-timer.C:
		r.log.Infof("leader-elect: grace %v elapsed, SIGKILL process group %d", grace, pid)
		if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil && err != syscall.ESRCH {
			r.log.Errorf("leader-elect: kill pgid %d SIGKILL: %v", pid, err)
		}
		select {
		case <-done:
		case <-time.After(childSigkillWait):
		}
	}
}
