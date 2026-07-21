package leaderelect

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func clearLeaderElectEnv(t *testing.T) {
	t.Helper()
	t.Setenv(envLeaseName, "")
	t.Setenv(envLeaseDuration, "")
	t.Setenv(envRenewDeadline, "")
	t.Setenv(envRetryPeriod, "")
	t.Setenv(envShutdownGrace, "")
	t.Setenv(envLostGrace, "")
}

func TestParseArgs(t *testing.T) {
	clearLeaderElectEnv(t)
	t.Setenv(envLeaseName, "foo")

	tests := []struct {
		name    string
		args    []string
		wantCmd []string
		wantErr string
	}{
		{
			name:    "double-dash splits command",
			args:    []string{"--", "/usr/local/bin/node", "core_init.js"},
			wantCmd: []string{"/usr/local/bin/node", "core_init.js"},
		},
		{
			name:    "command without double-dash",
			args:    []string{"sleep", "100"},
			wantCmd: []string{"sleep", "100"},
		},
		{
			name:    "missing command",
			args:    []string{"--"},
			wantErr: "command is required after --",
		},
		{
			name:    "empty args",
			args:    nil,
			wantErr: "command is required after --",
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
			if cfg.LeaseName != "foo" {
				t.Errorf("LeaseName = %q, want foo", cfg.LeaseName)
			}
			if !reflect.DeepEqual(cfg.Command, tt.wantCmd) {
				t.Errorf("Command = %#v, want %#v", cfg.Command, tt.wantCmd)
			}
			if cfg.LostGrace != defaultLostGrace {
				t.Errorf("LostGrace = %v, want %v", cfg.LostGrace, defaultLostGrace)
			}
			if cfg.ShutdownGrace != defaultShutdownGrace {
				t.Errorf("ShutdownGrace = %v, want %v", cfg.ShutdownGrace, defaultShutdownGrace)
			}
		})
	}
}

func TestParseArgsMissingLeaseName(t *testing.T) {
	clearLeaderElectEnv(t)
	_, err := ParseArgs([]string{"--", "true"})
	if err == nil || !strings.Contains(err.Error(), envLeaseName) {
		t.Fatalf("expected error mentioning %s, got %v", envLeaseName, err)
	}
}

func TestParseArgsTimingsFromEnv(t *testing.T) {
	t.Setenv(envLeaseName, "lease")
	t.Setenv(envLeaseDuration, "30s")
	t.Setenv(envRenewDeadline, "12s")
	t.Setenv(envRetryPeriod, "4s")
	t.Setenv(envShutdownGrace, "15s")
	t.Setenv(envLostGrace, "5s")

	cfg, err := ParseArgs([]string{"--", "true"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LeaseDuration != 30*time.Second {
		t.Errorf("LeaseDuration = %v, want 30s", cfg.LeaseDuration)
	}
	if cfg.RenewDeadline != 12*time.Second {
		t.Errorf("RenewDeadline = %v, want 12s", cfg.RenewDeadline)
	}
	if cfg.RetryPeriod != 4*time.Second {
		t.Errorf("RetryPeriod = %v, want 4s", cfg.RetryPeriod)
	}
	if cfg.ShutdownGrace != 15*time.Second {
		t.Errorf("ShutdownGrace = %v, want 15s", cfg.ShutdownGrace)
	}
	if cfg.LostGrace != 5*time.Second {
		t.Errorf("LostGrace = %v, want 5s", cfg.LostGrace)
	}
}

func TestParseArgsLostGraceInvariantFromEnv(t *testing.T) {
	t.Setenv(envLeaseName, "lease")
	t.Setenv(envLeaseDuration, "20s")
	t.Setenv(envRenewDeadline, "10s")
	t.Setenv(envLostGrace, "10s")

	_, err := ParseArgs([]string{"--", "true"})
	if err == nil || !strings.Contains(err.Error(), envLostGrace) {
		t.Fatalf("expected error mentioning %s, got %v", envLostGrace, err)
	}
}

func TestDurationFromEnvInvalidFallsBack(t *testing.T) {
	t.Setenv(envLeaseDuration, "not-a-duration")
	if got := durationFromEnv(envLeaseDuration, defaultLeaseDuration); got != defaultLeaseDuration {
		t.Fatalf("durationFromEnv invalid = %v, want %v", got, defaultLeaseDuration)
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
				RetryPeriod:   defaultRetryPeriod,
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
				RetryPeriod:   3 * time.Second,
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
				RetryPeriod:   3 * time.Second,
				LostGrace:     10 * time.Second,
				Command:       []string{"sleep"},
			},
			wantErr: envLostGrace,
		},
		{
			name: "lost-grace above max rejected",
			cfg: Config{
				LeaseName:     "x",
				LeaseDuration: 20 * time.Second,
				RenewDeadline: 10 * time.Second,
				RetryPeriod:   3 * time.Second,
				LostGrace:     11 * time.Second,
				Command:       []string{"sleep"},
			},
			wantErr: envLostGrace,
		},
		{
			name: "lease-duration not greater than renew-deadline",
			cfg: Config{
				LeaseName:     "x",
				LeaseDuration: 10 * time.Second,
				RenewDeadline: 10 * time.Second,
				RetryPeriod:   3 * time.Second,
				LostGrace:     1 * time.Second,
				Command:       []string{"sleep"},
			},
			wantErr: envLeaseDuration,
		},
		{
			name: "retry-period equal renew-deadline rejected",
			cfg: Config{
				LeaseName:     "x",
				LeaseDuration: 20 * time.Second,
				RenewDeadline: 10 * time.Second,
				RetryPeriod:   10 * time.Second,
				LostGrace:     1 * time.Second,
				Command:       []string{"sleep"},
			},
			wantErr: envRenewDeadline,
		},
		{
			name: "renew-deadline equal retry-period*1.2 rejected",
			cfg: Config{
				LeaseName:     "x",
				LeaseDuration: 30 * time.Second,
				RenewDeadline: 12 * time.Second,
				RetryPeriod:   10 * time.Second, // 10s * 1.2 = 12s
				LostGrace:     1 * time.Second,
				Command:       []string{"sleep"},
			},
			wantErr: envRenewDeadline,
		},
		{
			name: "renew-deadline just above retry-period*1.2 ok",
			cfg: Config{
				LeaseName:     "x",
				LeaseDuration: 30 * time.Second,
				RenewDeadline: 13 * time.Second,
				RetryPeriod:   10 * time.Second, // 10s * 1.2 = 12s
				LostGrace:     1 * time.Second,
				Command:       []string{"sleep"},
			},
		},
		{
			name: "retry-period zero rejected",
			cfg: Config{
				LeaseName:     "x",
				LeaseDuration: defaultLeaseDuration,
				RenewDeadline: defaultRenewDeadline,
				RetryPeriod:   0,
				LostGrace:     defaultLostGrace,
				Command:       []string{"sleep"},
			},
			wantErr: envRetryPeriod,
		},
		{
			name: "missing lease name",
			cfg: Config{
				LeaseDuration: defaultLeaseDuration,
				RenewDeadline: defaultRenewDeadline,
				RetryPeriod:   defaultRetryPeriod,
				LostGrace:     defaultLostGrace,
				Command:       []string{"sleep"},
			},
			wantErr: envLeaseName,
		},
		{
			name: "missing command",
			cfg: Config{
				LeaseName:     "x",
				LeaseDuration: defaultLeaseDuration,
				RenewDeadline: defaultRenewDeadline,
				RetryPeriod:   defaultRetryPeriod,
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

func TestRecommendedTerminationGracePeriod(t *testing.T) {
	t.Parallel()
	// defaultShutdownGrace(25s) + childSigkillWait(5s) + leaseReleaseMargin(10s)
	if got, want := RecommendedTerminationGracePeriod(0), int64(40); got != want {
		t.Fatalf("RecommendedTerminationGracePeriod(0) = %d, want %d", got, want)
	}
	if got, want := RecommendedTerminationGracePeriod(30*time.Second), int64(45); got != want {
		t.Fatalf("RecommendedTerminationGracePeriod(30s) = %d, want %d", got, want)
	}
}
