package api

import (
	"context"
	"net/http"

	"github.com/blang/semver/v4"
)

type Kubo interface {
	Version(ctx context.Context) (*semver.Version, error)
}

type Config struct {
	Kubo Kubo
}

func New(config Config) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /livez", http.HandlerFunc(livez))
	mux.Handle("GET /readyz", http.HandlerFunc(readyz(config)))

	return mux
}
