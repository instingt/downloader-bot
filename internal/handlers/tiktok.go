package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TikTokHandler struct {
	ytDlpBinary string
}

func NewTiktokHandler(ytDlpBinary string) *TikTokHandler {
	return &TikTokHandler{
		ytDlpBinary: ytDlpBinary,
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

func (h *TikTokHandler) Handle(bot *tgbotapi.BotAPI, u *url.URL, replyChatID int64) error {
	if bot == nil {
		return errors.New("bot is nil")
	}
	if u == nil {
		return errors.New("url is nil")
	}
	if h.ytDlpBinary == "" {
		return errors.New("yt-dlp binary path is empty")
	}

	tmpDir, err := os.MkdirTemp("", "bot-tiktok-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath, err := h.downloadVideo(u, tmpDir)
	if err != nil {
		return err
	}

	videoMsg := tgbotapi.NewVideo(replyChatID, tgbotapi.FilePath(filePath))
	if _, err := bot.Send(videoMsg); err != nil {
		docMsg := tgbotapi.NewDocument(replyChatID, tgbotapi.FilePath(filePath))
		if _, docErr := bot.Send(docMsg); docErr != nil {
			return fmt.Errorf("send video failed: %v; send document fallback failed: %w", err, docErr)
		}
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

func resolveDownloadedFilePath(tmpDir string) (string, error) {
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return "", fmt.Errorf("read output directory: %w", err)
	}

	var selected string
	var selectedModTime time.Time

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.EqualFold(filepath.Ext(entry.Name()), ".mp4") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if selected == "" || info.ModTime().After(selectedModTime) {
			selected = entry.Name()
			selectedModTime = info.ModTime()
		}
	}

	if selected == "" {
		return "", errors.New("no downloaded mp4 file found in output directory")
	}

	return tmpDir + string(os.PathSeparator) + selected, nil
}
