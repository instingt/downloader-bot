package handlers

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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
