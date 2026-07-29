package api

import (
	"net/http"

	"github.com/NatoBoram/ipapm/kubo"
)

type Config struct {
	Kubo kubo.Client
}

func New(config Config) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /livez", http.HandlerFunc(livez))
	mux.Handle("GET /readyz", http.HandlerFunc(readyz(config)))

	return mux
}
