package config

import (
	"fmt"
	"os"
	"path"
)

func ConfigDir(env Env) (string, error) {
	path := env.CONFIG_DIR

	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return path, fmt.Errorf("failed to create config dir: %w", err)
	}

	return path, err
}

func ConfigFile(dir string) string {
	return path.Join(dir, "config.yaml")
}
