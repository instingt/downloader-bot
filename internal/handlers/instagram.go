package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"bot-downloader/internal/telegram"
)

type InstagramHandler struct {
	ytDlpBinary     string
	cookiesFilePath string
	logger          *slog.Logger
}

func NewInstagramHandler(ytDlpBinary string, cookiesFilePath string, logger *slog.Logger) *InstagramHandler {
	return &InstagramHandler{
		ytDlpBinary:     ytDlpBinary,
		cookiesFilePath: cookiesFilePath,
		logger:          logger,
	}
}

func (h *InstagramHandler) Matcher(u *url.URL) bool {
	if u == nil {
		return false
	}

	scheme := strings.ToLower(strings.TrimSpace(u.Scheme))
	if scheme != "https" {
		return false
	}

	host := strings.ToLower(strings.TrimSpace(u.Hostname()))
	if host == "" {
		return false
	}

	if host != "instagram.com" && !strings.HasSuffix(host, ".instagram.com") {
		return false
	}

	path := strings.TrimSpace(u.EscapedPath())
	return strings.HasPrefix(path, "/reel/")
}

func (h *InstagramHandler) Handle(ctx context.Context, tg telegram.Client, u *url.URL, replyChatID int64) error {
	if tg == nil {
		return errors.New("telegram client is nil")
	}
	if u == nil {
		return errors.New("url is nil")
	}
	if h.ytDlpBinary == "" {
		return errors.New("yt-dlp binary path is empty")
	}
	if h.cookiesFilePath == "" {
		return errors.New("instagram cookies file path is empty")
	}
	if h.logger == nil {
		return errors.New("logger is nil")
	}

	tmpDir, err := os.MkdirTemp("", "bot-instagram-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer func() {
		if cleanupErr := os.RemoveAll(tmpDir); cleanupErr != nil {
			h.logger.Warn("failed to cleanup temp dir", "dir", tmpDir, "error", cleanupErr)
		}
	}()

	h.logger.Info("instagram reel download started", "chat_id", replyChatID, "url", u.String())
	filePath, err := h.downloadVideo(u, tmpDir)
	if err != nil {
		return err
	}
	h.logger.Info("instagram reel download finished", "chat_id", replyChatID, "file", filePath)

	if err := tg.SendVideoFile(ctx, replyChatID, filePath); err != nil {
		if docErr := tg.SendDocumentFile(ctx, replyChatID, filePath); docErr != nil {
			return fmt.Errorf("send video failed: %v; send document fallback failed: %w", err, docErr)
		}
		h.logger.Info("instagram reel sent as document", "chat_id", replyChatID, "file", filePath)
		return nil
	}
	h.logger.Info("instagram reel sent successfully", "chat_id", replyChatID, "file", filePath)

	return nil
}

func (h *InstagramHandler) downloadVideo(u *url.URL, tmpDir string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		h.ytDlpBinary,
		"--no-playlist",
		"--restrict-filenames",
		"--merge-output-format", "mp4",
		"--cookies", h.cookiesFilePath,
		"--print", "after_move:filepath",
		"--paths", tmpDir,
		u.String(),
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("run yt-dlp: %w (stderr: %s)", err, strings.TrimSpace(stderr.String()))
	}

	filePath, err := resolveDownloadedFilePath(tmpDir)
	if err != nil {
		return "", fmt.Errorf("resolve downloaded file path: %w", err)
	}

	return filePath, nil
}
