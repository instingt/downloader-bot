// Package logging configures application logging.
package logging

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

// New returns a logger configured for the provided app environment.
func New(appEnv string) (*slog.Logger, error) {
	switch appEnv {
	case "development":
		handler := tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelInfo,
			TimeFormat: "15:04:05",
			NoColor:    false,
		})
		return slog.New(handler), nil
	case "production":
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		return slog.New(handler), nil
	default:
		return nil, fmt.Errorf("unsupported APP_ENV %q, expected development or production", appEnv)
	}
}
