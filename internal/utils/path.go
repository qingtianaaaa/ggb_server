package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func FindRootPath() (string, error) {
	startDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current working directory: %w", err)
	}
	for {
		goModPath := filepath.Join(startDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return startDir, nil
		}

		parentDir := filepath.Dir(startDir)
		if parentDir == startDir {
			return "", fmt.Errorf("go.mod not found in or above %s", startDir)
		}
		startDir = parentDir
	}
}
