//go:build unix

package leaderelect

import (
	"os/exec"
	"syscall"
	"time"
)

func checkPlatform() error { return nil }

func configureChildProc(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
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
