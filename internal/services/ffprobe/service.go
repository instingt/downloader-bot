// Package ffprobe provides a service for reading video metadata using the ffprobe tool.
package ffprobe

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"

	"bot-downloader/internal/video"
)

type FFProbeService struct {
	ffprobeBinaryPath string
}

func New(ffprobeBinaryPath string) (*FFProbeService, error) {
	if ffprobeBinaryPath == "" {
		return nil, fmt.Errorf("ffprobe binary path is empty")
	}

	return &FFProbeService{
		ffprobeBinaryPath: ffprobeBinaryPath,
	}, nil
}

func (f *FFProbeService) ReadMetadata(ctx context.Context, videoPath string) (video.VideoMetadata, error) {
	if strings.TrimSpace(videoPath) == "" {
		return video.VideoMetadata{}, errors.New("video path is empty")
	}

	cmd := exec.Command(
		f.ffprobeBinaryPath,
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height:format=duration",
		"-of", "json",
		videoPath,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return video.VideoMetadata{}, fmt.Errorf("run ffprobe: %w (stderr: %s)", err, strings.TrimSpace(stderr.String()))
	}

	var payload struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
		Format struct {
			Duration json.RawMessage `json:"duration"`
		} `json:"format"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		return video.VideoMetadata{}, fmt.Errorf("parse ffprobe output: %w", err)
	}

	if len(payload.Streams) == 0 {
		return video.VideoMetadata{}, errors.New("ffprobe output has no video streams")
	}
	if payload.Streams[0].Width <= 0 || payload.Streams[0].Height <= 0 {
		return video.VideoMetadata{}, errors.New("ffprobe output has invalid width or height")
	}

	duration, err := parseDuration(payload.Format.Duration)
	if err != nil {
		return video.VideoMetadata{}, fmt.Errorf("parse ffprobe duration: %w", err)
	}

	return video.VideoMetadata{
		Width:    payload.Streams[0].Width,
		Height:   payload.Streams[0].Height,
		Duration: int(math.Round(duration)),
	}, nil
}

func parseDuration(raw json.RawMessage) (float64, error) {
	if len(raw) == 0 {
		return 0, errors.New("duration is empty")
	}

	var durationStr string
	if err := json.Unmarshal(raw, &durationStr); err == nil {
		value, parseErr := strconv.ParseFloat(durationStr, 64)
		if parseErr != nil {
			return 0, fmt.Errorf("duration is not a float: %w", parseErr)
		}
		if value < 0 {
			return 0, errors.New("duration is negative")
		}
		return value, nil
	}

	var durationFloat float64
	if err := json.Unmarshal(raw, &durationFloat); err == nil {
		if durationFloat < 0 {
			return 0, errors.New("duration is negative")
		}
		return durationFloat, nil
	}

	return 0, fmt.Errorf("duration has unsupported format: %s", string(raw))
}
