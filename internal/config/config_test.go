package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setBaseEnv(t *testing.T) string {
	t.Helper()

	binDir := t.TempDir()
	ytDlpPath := filepath.Join(binDir, "yt-dlp")
	if err := os.WriteFile(ytDlpPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write fake yt-dlp: %v", err)
	}

	cookiesPath := filepath.Join(t.TempDir(), "instagram-cookies.txt")
	if err := os.WriteFile(cookiesPath, []byte("# Netscape HTTP Cookie File\n"), 0o644); err != nil {
		t.Fatalf("write fake cookies file: %v", err)
	}

	t.Setenv("PATH", binDir)
	t.Setenv("TELEGRAM_BOT_TOKEN", "token")
	t.Setenv("ALLOWED_TELEGRAM_USER_IDS", "123")
	t.Setenv("ALLOWED_TELEGRAM_CHAT_IDS", "-1001")
	t.Setenv("INSTAGRAM_COOKIES_FILE_PATH", cookiesPath)

	return cookiesPath
}

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
			setBaseEnv(t)
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

func TestLoadInstagramCookiesFilePath(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		wantPath := setBaseEnv(t)

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.InstagramCookiesFilePath != wantPath {
			t.Fatalf("expected InstagramCookiesFilePath %q, got %q", wantPath, cfg.InstagramCookiesFilePath)
		}
	})

	t.Run("missing env var", func(t *testing.T) {
		setBaseEnv(t)
		t.Setenv("INSTAGRAM_COOKIES_FILE_PATH", "")

		_, err := Load()
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), "INSTAGRAM_COOKIES_FILE_PATH is required") {
			t.Fatalf("expected missing cookies env error, got %q", err.Error())
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		setBaseEnv(t)
		t.Setenv("INSTAGRAM_COOKIES_FILE_PATH", filepath.Join(t.TempDir(), "missing.txt"))

		_, err := Load()
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), "instagram cookies file path is invalid") {
			t.Fatalf("expected invalid cookies file path error, got %q", err.Error())
		}
	})

	t.Run("directory path", func(t *testing.T) {
		setBaseEnv(t)
		t.Setenv("INSTAGRAM_COOKIES_FILE_PATH", t.TempDir())

		_, err := Load()
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), "instagram cookies file path must be a file") {
			t.Fatalf("expected file-only cookies path error, got %q", err.Error())
		}
	})
}
