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
	"bot-downloader/internal/video"
)

type TikTokHandler struct {
	ytDlpBinary string
	encoder     video.Transcoder
	logger      *slog.Logger
}

func NewTiktokHandler(ytDlpBinary string, encoder video.Transcoder, logger *slog.Logger) *TikTokHandler {
	return &TikTokHandler{
		ytDlpBinary: ytDlpBinary,
		encoder:     encoder,
		logger:      logger,
	}
}

func (h *TikTokHandler) Matcher(u *url.URL) bool {
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

	return host == "tiktok.com" || strings.HasSuffix(host, ".tiktok.com")
}

func (h *TikTokHandler) Handle(ctx context.Context, tg telegram.Client, u *url.URL, replyChatID int64) error {
	if tg == nil {
		return errors.New("telegram client is nil")
	}
	if u == nil {
		return errors.New("url is nil")
	}
	if h.ytDlpBinary == "" {
		return errors.New("yt-dlp binary path is empty")
	}
	if h.encoder == nil {
		return errors.New("video encoder is nil")
	}
	if h.logger == nil {
		return errors.New("logger is nil")
	}

	tmpDir, err := os.MkdirTemp("", "bot-tiktok-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer func() {
		if cleanupErr := os.RemoveAll(tmpDir); cleanupErr != nil {
			h.logger.Warn("failed to cleanup temp dir", "dir", tmpDir, "error", cleanupErr)
		}
	}()

	h.logger.Info("tiktok download started", "chat_id", replyChatID, "url", u.String())
	filePath, err := h.downloadVideo(u, tmpDir)
	if err != nil {
		return err
	}
	h.logger.Info("tiktok download finished", "chat_id", replyChatID, "file", filePath)

	encodedPath, err := h.encoder.Transcode(ctx, filePath)
	if err != nil {
		return fmt.Errorf("encode downloaded video: %w", err)
	}
	h.logger.Info("tiktok video encoded", "chat_id", replyChatID, "input_file", filePath, "output_file", encodedPath)

	err = tg.SendVideoWithDocumentFallback(ctx, replyChatID, filePath)
	if err != nil {
		return fmt.Errorf("send tiktok video: %w", err)
	}

	return nil
}

func (h *TikTokHandler) downloadVideo(u *url.URL, tmpDir string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		h.ytDlpBinary,
		"--no-playlist",
		"--restrict-filenames",
		"--merge-output-format", "mp4",
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
