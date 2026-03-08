// Package config loads and validates runtime configuration from environment variables.
package config

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Config struct {
	Token           string
	AllowedUserIDs  map[int64]struct{}
	AllowedChatIDs  map[int64]struct{}
	YtDlpBinaryPath string
}

func Load() (Config, error) {
	token := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if token == "" {
		return Config{}, errors.New("TELEGRAM_BOT_TOKEN is required")
	}

	userIDs, err := parseIDSet("ALLOWED_TELEGRAM_USER_IDS")
	if err != nil {
		return Config{}, err
	}

	chatIDs, err := parseIDSet("ALLOWED_TELEGRAM_CHAT_IDS")
	if err != nil {
		return Config{}, err
	}

	ytDlpBinaryPath, err := exec.LookPath("yt-dlp")
	if err != nil {
		return Config{}, fmt.Errorf("yt-dlp binary not found in PATH: %w", err)
	}

	return Config{
		Token:           token,
		AllowedUserIDs:  userIDs,
		AllowedChatIDs:  chatIDs,
		YtDlpBinaryPath: ytDlpBinaryPath,
	}, nil
}

func parseIDSet(envVar string) (map[int64]struct{}, error) {
	raw := strings.TrimSpace(os.Getenv(envVar))
	if raw == "" {
		return nil, fmt.Errorf("%s is required", envVar)
	}

	parts := strings.Split(raw, ",")
	ids := make(map[int64]struct{}, len(parts))

	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v == "" {
			return nil, fmt.Errorf("%s contains an empty value", envVar)
		}

		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%s has invalid id %q: %w", envVar, v, err)
		}

		ids[id] = struct{}{}
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("%s must include at least one id", envVar)
	}

	return ids, nil
}
