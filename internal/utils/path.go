package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func FindProjectName() (string, error) {
	rootPath, err := FindRootPath()
	if err != nil {
		return "", fmt.Errorf("failed to find root path: %w", err)
	}

	goModPath := filepath.Join(rootPath, "go.mod")
	file, err := os.Open(goModPath)
	if err != nil {
		return "", fmt.Errorf("failed to open go.mod: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(line[len("module "):]), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error scanning go.mod: %w", err)
	}

	return "", fmt.Errorf("module declaration not found in go.mod")
}
