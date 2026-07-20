// Package leaderelect implements a PID-1 leader-election exec wrapper.
//
// Exposed as a Hidden cobra subcommand of the noobaa-operator binary and
// copied into the core pod from the operator image:
//
//	noobaa-operator leader-elect [flags] -- <command> [args...]
//
// The command is Hidden so it does not appear in public CLI help. The wrapper
// acquires a namespaced Lease, spawns the given command in its own process
// group while holding leadership, reaps orphaned children (PID 1), and on
// SIGTERM/SIGINT or leadership loss stops the child group before releasing
// the Lease (ReleaseOnCancel).
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

	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

const (
	exitOK     = 0
	exitConfig = 1
	exitLost   = 2

	envLeaseName    = "NOOBAA_CORE_LEASE_NAME"
	envPodNamespace = "POD_NAMESPACE"
	saNamespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

	defaultLeaseDuration = 20 * time.Second
	defaultRenewDeadline = 10 * time.Second
	defaultRetryPeriod   = 3 * time.Second
	defaultShutdownGrace = 25 * time.Second
	defaultLostGrace     = 8 * time.Second
)

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
func Cmd() *cobra.Command {
	cfg := defaultConfig()
	cmd := &cobra.Command{
		Hidden:                true,
		Use:                   "leader-elect -- <command> [args...]",
		Short:                 "Acquire a Lease then exec a command (PID 1 leader-elect wrapper)",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Args:                  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cfg.Command = args
			if err := cfg.Validate(); err != nil {
				fmt.Fprintf(os.Stderr, "leader-elect: %v\n", err)
				os.Exit(exitConfig)
			}
			os.Exit(Run(cfg))
		},
	}
	bindFlags(cmd, cfg)
	return cmd
}

func defaultConfig() *Config {
	return &Config{
		LeaseName:     os.Getenv(envLeaseName),
		LeaseDuration: defaultLeaseDuration,
		RenewDeadline: defaultRenewDeadline,
		RetryPeriod:   defaultRetryPeriod,
		ShutdownGrace: defaultShutdownGrace,
		LostGrace:     defaultLostGrace,
	}
}

func bindFlags(cmd *cobra.Command, cfg *Config) {
	cmd.Flags().StringVar(&cfg.LeaseName, "lease-name", cfg.LeaseName, "Lease object name (default: $NOOBAA_CORE_LEASE_NAME)")
	cmd.Flags().DurationVar(&cfg.LeaseDuration, "lease-duration", cfg.LeaseDuration, "Lease duration")
	cmd.Flags().DurationVar(&cfg.RenewDeadline, "renew-deadline", cfg.RenewDeadline, "Leadership renew deadline")
	cmd.Flags().DurationVar(&cfg.RetryPeriod, "retry-period", cfg.RetryPeriod, "Leadership retry period")
	cmd.Flags().DurationVar(&cfg.ShutdownGrace, "shutdown-grace", cfg.ShutdownGrace, "Time to wait for child exit after SIGTERM/SIGINT before SIGKILL")
	cmd.Flags().DurationVar(&cfg.LostGrace, "lost-grace", cfg.LostGrace, "Time to wait for child exit after leadership loss before SIGKILL")
}

// ParseArgs parses leader-elect flags and the wrapped command (args after flags/--).
// It does not run the election or spawn processes.
func ParseArgs(args []string) (*Config, error) {
	cfg := defaultConfig()
	cmd := &cobra.Command{
		Use:           "leader-elect",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ArbitraryArgs,
		Run: func(_ *cobra.Command, a []string) {
			cfg.Command = a
		},
	}
	bindFlags(cmd, cfg)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate checks required fields and the lost-grace invariant.
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("nil config")
	}
	if c.LeaseName == "" {
		return fmt.Errorf("--lease-name is required (or set $%s)", envLeaseName)
	}
	if len(c.Command) == 0 {
		return fmt.Errorf("command is required after --")
	}
	maxLost := c.LeaseDuration - c.RenewDeadline
	if maxLost <= 0 {
		return fmt.Errorf("lease-duration (%v) must be greater than renew-deadline (%v)", c.LeaseDuration, c.RenewDeadline)
	}
	if c.LostGrace >= maxLost {
		return fmt.Errorf("--lost-grace (%v) must be < lease-duration - renew-deadline (%v)", c.LostGrace, maxLost)
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
		cfg:    cfg,
		cancel: cancel,
		log:    log,
	}

	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(sigCh)

	go func() {
		sig := <-sigCh
		r.log.Infof("leader-elect: received %v, stopping child then releasing lease", sig)
		r.stopChild(cfg.ShutdownGrace)
		r.forceExit(exitOK)
		cancel()
	}()

	le, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:            lock,
		LeaseDuration:   cfg.LeaseDuration,
		RenewDeadline:   cfg.RenewDeadline,
		RetryPeriod:     cfg.RetryPeriod,
		ReleaseOnCancel: true,
		Name:            cfg.LeaseName,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(leadCtx context.Context) {
				r.onStartedLeading(leadCtx)
			},
			OnStoppedLeading: func() {
				r.log.Infof("leader-elect: stopped leading (identity %s)", identity)
			},
		},
	})
	if err != nil {
		log.Errorf("leader-elect: create elector: %v", err)
		return exitConfig
	}

	log.Infof("leader-elect: waiting for lease %s/%s as %s", namespace, cfg.LeaseName, identity)
	le.Run(ctx)

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
	cfg    *Config
	cancel context.CancelFunc
	log    logger

	mu            sync.Mutex
	childPID      int
	childExited   chan struct{} // closed once the main child is reaped
	childCode     int
	exitCode      int
	exitSet       bool
	termRequested bool
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
			// SIGTERM/lost path owns the wrapper exit code.
			return
		}
		code := r.getChildCode()
		r.log.Infof("leader-elect: child exited with code %d", code)
		r.setExit(code)
		r.cancel() // release lease after child is down
	case <-leadCtx.Done():
		r.mu.Lock()
		termRequested := r.termRequested
		r.mu.Unlock()
		if termRequested {
			// SIGTERM path owns teardown; wait for child then return.
			select {
			case <-childDone:
			case <-time.After(r.cfg.ShutdownGrace + 5*time.Second):
			}
			return
		}
		r.log.Errorf("leader-elect: leadership lost, stopping child")
		r.stopChild(r.cfg.LostGrace)
		r.forceExit(exitLost)
		select {
		case <-childDone:
		case <-time.After(r.cfg.LostGrace + 5*time.Second):
		}
	}
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
		case <-time.After(5 * time.Second):
		}
	}
}
