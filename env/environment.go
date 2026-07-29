package env

import (
	"os"
	"testing"
)

// Environment represents the current environment (development, test,
// production).
type Environment string

const (
	Development Environment = "development"
	Test        Environment = "test"
	Production  Environment = "production"
)

func (e Environment) String() string {
	return string(e)
}

// toEnvironment converts a string to an Environment.
func toEnvironment(s string) Environment {
	switch s {
	case "development":
		return Development
	case "test":
		return Test
	case "production":
		return Production
	}

	if testing.Testing() {
		os.Setenv("GO_ENV", "test")
		return "test"
	}

	return Development
}

// getEnvironment returns the current environment.
func getEnvironment() Environment {
	environment := toEnvironment(os.Getenv("GO_ENV"))
	os.Setenv("GO_ENV", environment.String())
	return environment
}
