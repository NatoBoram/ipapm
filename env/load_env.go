package env

import (
	"errors"
	"fmt"
	"os"
	"path"
	"runtime/debug"
	"strings"

	"github.com/joho/godotenv"
)

// Env contains all the environment variables and metadata that are supported by
// this program.
type Env struct {
	GO_ENV Environment

	CONFIG_DIR    string
	KUBO_API_AUTH string
	KUBO_API_URL  string

	Name    string
	Version string
}

// loadEnv loads the environment variables from the .env files.
func LoadEnv() (Env, error) {
	environment := GetEnvironment()

	files := []string{
		fmt.Sprintf(".env.%s.local", environment),
		fmt.Sprintf(".env.%s", environment),
		".env.local",
		".env",
	}

	for _, file := range files {
		if err := godotenv.Load(file); err != nil && !os.IsNotExist(err) {
			return Env{}, fmt.Errorf("failed to load environment variables from %q: %w", file, err)
		}
	}

	name, version, err := readBuildInfo()
	if err != nil {
		return Env{}, fmt.Errorf("failed to extract build info: %w", err)
	}

	CONFIG_DIR, err := envConfigDir(name)
	if err != nil {
		return Env{}, fmt.Errorf("failed to determine config dir: %w", err)
	}

	KUBO_API_URL := os.Getenv("KUBO_API_URL")
	if KUBO_API_URL == "" {
		KUBO_API_URL = "http://localhost:5001"
	}

	return Env{
		GO_ENV: environment,

		CONFIG_DIR:    CONFIG_DIR,
		KUBO_API_AUTH: os.Getenv("KUBO_API_AUTH"),
		KUBO_API_URL:  KUBO_API_URL,

		Name:    name,
		Version: version,
	}, nil
}

func readBuildInfo() (string, string, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", "", errors.New("failed to read build info")
	}

	parts := strings.Split(info.Main.Path, "/")
	if len(parts) == 0 {
		return "", "", errors.New("failed to parse build info")
	}

	name := parts[len(parts)-1]

	version := info.Main.Version
	if version == "(devel)" {
		version = "0.0.0"
	}

	return name, version, nil
}

func envConfigDir(name string) (string, error) {
	CONFIG_DIR := os.Getenv("CONFIG_DIR")
	if CONFIG_DIR != "" {
		return CONFIG_DIR, nil
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %w", err)
	}

	return path.Join(dir, name), nil
}
