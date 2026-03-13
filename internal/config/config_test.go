package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAppEnv(t *testing.T) {
	tests := []struct {
		name       string
		appEnv     string
		wantEnv    string
		wantErr    bool
		errContain string
	}{
		{name: "default development", appEnv: "", wantEnv: "development"},
		{name: "explicit development", appEnv: "development", wantEnv: "development"},
		{name: "production", appEnv: "production", wantEnv: "production"},
		{name: "invalid", appEnv: "staging", wantErr: true, errContain: "APP_ENV must be one of"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			binDir := t.TempDir()
			ytDlpPath := filepath.Join(binDir, "yt-dlp")
			if err := os.WriteFile(ytDlpPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
				t.Fatalf("write fake yt-dlp: %v", err)
			}

			t.Setenv("PATH", binDir)
			t.Setenv("TELEGRAM_BOT_TOKEN", "token")
			t.Setenv("ALLOWED_TELEGRAM_USER_IDS", "123")
			t.Setenv("ALLOWED_TELEGRAM_CHAT_IDS", "-1001")
			t.Setenv("APP_ENV", tc.appEnv)

			cfg, err := Load()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				if tc.errContain != "" && !strings.Contains(err.Error(), tc.errContain) {
					t.Fatalf("expected error to contain %q, got %q", tc.errContain, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.AppEnv != tc.wantEnv {
				t.Fatalf("expected AppEnv %q, got %q", tc.wantEnv, cfg.AppEnv)
			}
		})
	}
}
