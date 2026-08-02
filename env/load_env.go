package env

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Env contains all the environment variables that are supported by this
// program.
type Env struct {
	GO_ENV        Environment
	CONFIG_DIR    string
	KUBO_API_AUTH string
	KUBO_API_URL  string
}

// loadEnv loads the environment variables from the .env files.
func LoadEnv() (Env, error) {
	environment := getEnvironment()

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

	return Env{
		GO_ENV:        getEnvironment(),
		CONFIG_DIR:    os.Getenv("CONFIG_DIR"),
		KUBO_API_AUTH: os.Getenv("KUBO_API_AUTH"),
		KUBO_API_URL:  os.Getenv("KUBO_API_URL"),
	}, nil
}
