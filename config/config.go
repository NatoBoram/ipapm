package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"runtime/debug"
	"strings"
)

func AppName() (string, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", errors.New("failed to read the build info")
	}

	parts := strings.Split(info.Path, "/")
	if len(parts) == 0 {
		return "", errors.New("failed to parse the build info")
	}

	return parts[len(parts)-1], nil
}

func ConfigDir(env Env) (string, error) {
	if env.CONFIG_DIR != "" {
		path := env.CONFIG_DIR

		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return path, fmt.Errorf("failed to create config dir: %w", err)
		}

		return path, err
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %w", err)
	}

	name, err := AppName()
	if err != nil {
		return "", fmt.Errorf("failed to get app name: %w", err)
	}

	path := path.Join(dir, name)
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return path, fmt.Errorf("failed to create config dir: %w", err)
	}

	return path, err
}

func ConfigFile(dir string) string {
	return path.Join(dir, "config.yaml")
}
