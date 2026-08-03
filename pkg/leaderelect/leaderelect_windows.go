//go:build windows

package leaderelect

import (
	"fmt"
	"os/exec"
	"time"
)

func checkPlatform() error {
	return fmt.Errorf("not supported on Windows (Linux PID-1 wrapper only)")
}

// configureChildProc is a no-op on Windows (no process groups / Setpgid).
func configureChildProc(cmd *exec.Cmd) {}

// waitAndReap cannot use Wait4 on Windows. Leader-elect is a Linux PID-1
// wrapper; this stub only exists so the operator CLI still cross-compiles.
func (r *runner) waitAndReap(childPID int, done chan struct{}) {
	r.mu.Lock()
	r.childCode = exitConfig
	r.mu.Unlock()
	close(done)
}

// stopChild cannot signal process groups on Windows.
func (r *runner) stopChild(grace time.Duration) {
	r.mu.Lock()
	r.termRequested = true
	r.mu.Unlock()
}
