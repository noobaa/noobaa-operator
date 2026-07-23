package leaderelect

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParseArgs(t *testing.T) {
	// Clear env so default lease-name does not leak across cases.
	t.Setenv(envLeaseName, "")

	tests := []struct {
		name         string
		args         []string
		wantLease    string
		wantCmd      []string
		wantLost     time.Duration
		wantShutdown time.Duration
		wantErr      string
	}{
		{
			name:         "double-dash splits command",
			args:         []string{"--lease-name=foo", "--", "/usr/local/bin/node", "core_init.js"},
			wantLease:    "foo",
			wantCmd:      []string{"/usr/local/bin/node", "core_init.js"},
			wantLost:     defaultLostGrace,
			wantShutdown: defaultShutdownGrace,
		},
		{
			name:      "command without double-dash",
			args:      []string{"--lease-name=bar", "sleep", "100"},
			wantLease: "bar",
			wantCmd:   []string{"sleep", "100"},
		},
		{
			name:         "custom grace flags",
			args:         []string{"--lease-name=x", "--lost-grace=5s", "--shutdown-grace=15s", "--", "echo", "hi"},
			wantLease:    "x",
			wantCmd:      []string{"echo", "hi"},
			wantLost:     5 * time.Second,
			wantShutdown: 15 * time.Second,
		},
		{
			name:      "lease-duration and renew-deadline",
			args:      []string{"--lease-name=x", "--lease-duration=30s", "--renew-deadline=12s", "--retry-period=4s", "--", "true"},
			wantLease: "x",
			wantCmd:   []string{"true"},
		},
		{
			name:    "missing lease-name",
			args:    []string{"--", "sleep", "1"},
			wantErr: "--lease-name is required",
		},
		{
			name:    "missing command",
			args:    []string{"--lease-name=foo"},
			wantErr: "command is required after --",
		},
		{
			name:    "lost-grace equal to lease-duration minus renew-deadline",
			args:    []string{"--lease-name=foo", "--lease-duration=20s", "--renew-deadline=10s", "--lost-grace=10s", "--", "sleep", "1"},
			wantErr: "--lost-grace",
		},
		{
			name:    "lost-grace greater than lease-duration minus renew-deadline",
			args:    []string{"--lease-name=foo", "--lease-duration=20s", "--renew-deadline=10s", "--lost-grace=15s", "--", "sleep", "1"},
			wantErr: "--lost-grace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := ParseArgs(tt.args)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.LeaseName != tt.wantLease {
				t.Errorf("LeaseName = %q, want %q", cfg.LeaseName, tt.wantLease)
			}
			if !reflect.DeepEqual(cfg.Command, tt.wantCmd) {
				t.Errorf("Command = %#v, want %#v", cfg.Command, tt.wantCmd)
			}
			if tt.wantLost != 0 && cfg.LostGrace != tt.wantLost {
				t.Errorf("LostGrace = %v, want %v", cfg.LostGrace, tt.wantLost)
			}
			if tt.wantShutdown != 0 && cfg.ShutdownGrace != tt.wantShutdown {
				t.Errorf("ShutdownGrace = %v, want %v", cfg.ShutdownGrace, tt.wantShutdown)
			}
			if tt.name == "lease-duration and renew-deadline" {
				if cfg.LeaseDuration != 30*time.Second {
					t.Errorf("LeaseDuration = %v, want 30s", cfg.LeaseDuration)
				}
				if cfg.RenewDeadline != 12*time.Second {
					t.Errorf("RenewDeadline = %v, want 12s", cfg.RenewDeadline)
				}
				if cfg.RetryPeriod != 4*time.Second {
					t.Errorf("RetryPeriod = %v, want 4s", cfg.RetryPeriod)
				}
			}
		})
	}
}

func TestParseArgsLeaseNameFromEnv(t *testing.T) {
	t.Setenv(envLeaseName, "from-env-lease")
	cfg, err := ParseArgs([]string{"--", "true"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LeaseName != "from-env-lease" {
		t.Fatalf("LeaseName = %q, want from-env-lease", cfg.LeaseName)
	}
}

func TestValidateGraceInvariant(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "defaults ok",
			cfg: Config{
				LeaseName:     "x",
				LeaseDuration: defaultLeaseDuration,
				RenewDeadline: defaultRenewDeadline,
				LostGrace:     defaultLostGrace,
				Command:       []string{"sleep"},
			},
		},
		{
			name: "lost-grace just under max",
			cfg: Config{
				LeaseName:     "x",
				LeaseDuration: 20 * time.Second,
				RenewDeadline: 10 * time.Second,
				LostGrace:     9 * time.Second,
				Command:       []string{"sleep"},
			},
		},
		{
			name: "lost-grace equal rejected",
			cfg: Config{
				LeaseName:     "x",
				LeaseDuration: 20 * time.Second,
				RenewDeadline: 10 * time.Second,
				LostGrace:     10 * time.Second,
				Command:       []string{"sleep"},
			},
			wantErr: "--lost-grace",
		},
		{
			name: "lost-grace above max rejected",
			cfg: Config{
				LeaseName:     "x",
				LeaseDuration: 20 * time.Second,
				RenewDeadline: 10 * time.Second,
				LostGrace:     11 * time.Second,
				Command:       []string{"sleep"},
			},
			wantErr: "--lost-grace",
		},
		{
			name: "lease-duration not greater than renew-deadline",
			cfg: Config{
				LeaseName:     "x",
				LeaseDuration: 10 * time.Second,
				RenewDeadline: 10 * time.Second,
				LostGrace:     1 * time.Second,
				Command:       []string{"sleep"},
			},
			wantErr: "lease-duration",
		},
		{
			name: "missing lease name",
			cfg: Config{
				LeaseDuration: defaultLeaseDuration,
				RenewDeadline: defaultRenewDeadline,
				LostGrace:     defaultLostGrace,
				Command:       []string{"sleep"},
			},
			wantErr: "--lease-name is required",
		},
		{
			name: "missing command",
			cfg: Config{
				LeaseName:     "x",
				LeaseDuration: defaultLeaseDuration,
				RenewDeadline: defaultRenewDeadline,
				LostGrace:     defaultLostGrace,
			},
			wantErr: "command is required after --",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.cfg.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}
