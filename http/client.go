package http

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

const timeout = 10 * time.Second

type Transport struct {
	RoundTripper http.RoundTripper
	UserAgent    string
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	cloned := req.Clone(req.Context())
	cloned.Header.Set("User-Agent", t.UserAgent)

	return t.RoundTripper.RoundTrip(cloned)
}

func New(transport *Transport) *http.Client {
	return new(http.Client{
		Timeout:   timeout,
		Transport: transport,
	})
}

func UserAgent() (string, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", errors.New("failed to read build info")
	}

	parts := strings.Split(info.Path, "/")
	if len(parts) == 0 {
		return "", errors.New("failed to parse build info")
	}

	name := parts[len(parts)-1]

	version := info.Main.Version
	if version == "(devel)" {
		version = "0.0.0"
	}

	return fmt.Sprintf("%s/%s (+%s)", name, version, info.Path), nil
}
