package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NatoBoram/ipapm/api"
	"github.com/NatoBoram/ipapm/apt"
	"github.com/NatoBoram/ipapm/config"
	"github.com/NatoBoram/ipapm/env"
	h "github.com/NatoBoram/ipapm/http"
	"github.com/NatoBoram/ipapm/kubo"
	"github.com/NatoBoram/ipapm/log"
)

func start(ctx context.Context, env env.Env, kubo *kubo.Client, apt *apt.Client) {
	// Wait for a very small amount of time for the initial start then reset using
	// the full interval
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			err := sync(ctx, env, kubo, apt)
			if err != nil {
				slog.WarnContext(ctx, "Error during sync", slog.Any("error", err))
			}
		}

		timer.Reset(interval)
	}
}

func run(ctx context.Context) error {
	env, err := env.LoadEnv()
	if err != nil {
		return fmt.Errorf("couldn't load environment variables: %w", err)
	}

	slog.InfoContext(
		ctx, "Loading config",
		slog.String("CONFIG_DIR", env.CONFIG_DIR),
	)
	config, err := config.Load(config.Env{CONFIG_DIR: env.CONFIG_DIR})
	if err != nil {
		return fmt.Errorf("couldn't load config: %w", err)
	}

	userAgent, err := h.UserAgent()
	if err != nil {
		return fmt.Errorf("couldn't get user agent: %w", err)
	}
	slog.DebugContext(ctx, "User-Agent", slog.String("User-Agent", userAgent))

	client := h.New(new(h.Transport{
		RoundTripper: http.DefaultTransport,
		UserAgent:    userAgent,
	}))

	kubo, err := kubo.New(
		kubo.Config{
			KUBO_API_AUTH: env.KUBO_API_AUTH,
			KUBO_API_URL:  env.KUBO_API_URL,
			MFS:           config.Kubo.MFS,
		},
		client,
	)
	if err != nil {
		return fmt.Errorf("couldn't create Kubo client: %w", err)
	}

	v, err := kubo.Version(ctx)
	if err != nil {
		slog.WarnContext(ctx, "Couldn't connect to Kubo", slog.Any("error", err))
	}
	slog.InfoContext(ctx, "Connected to Kubo", slog.String("version", v.String()))

	apt := apt.New(client)

	server := http.Server{
		Addr:        fmt.Sprintf(":%d", config.Port),
		BaseContext: func(net.Listener) context.Context { return ctx },
		Handler:     api.New(api.Config{Kubo: kubo}),
	}
	served := make(chan error, 1)
	go func() { served <- server.ListenAndServe() }()
	slog.InfoContext(
		ctx, "Starting server",
		slog.String("url", fmt.Sprintf("http://localhost:%d", config.Port)),
	)

	go start(ctx, env, kubo, apt)

	select {
	case <-ctx.Done():

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// To put `^C` on its own line.
		fmt.Println("")

		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("couldn't shutdown server: %w", err)
		}

		if err := <-served; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("couldn't close the server: %w", err)
		}

		return nil

	case err := <-served:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("server error: %w", err)
	}
}

func main() {
	GO_ENV := env.GetEnvironment()
	logger := log.New(log.Config{GO_ENV: GO_ENV})
	slog.SetDefault(logger)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt, syscall.SIGTERM,
	)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.ErrorContext(ctx, "Couldn't run service", slog.Any("error", err))
		os.Exit(1)
	}
}
