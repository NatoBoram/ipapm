package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NatoBoram/ipapm/api"
	"github.com/NatoBoram/ipapm/config"
	"github.com/NatoBoram/ipapm/env"
	"github.com/NatoBoram/ipapm/kubo"
)

const timeout time.Duration = 5 * time.Second

func run(ctx context.Context) error {
	env, err := env.LoadEnv()
	if err != nil {
		return fmt.Errorf("couldn't load environment variables: %w", err)
	}

	config, err := config.Load(config.Env{CONFIG_DIR: env.CONFIG_DIR})
	if err != nil {
		return fmt.Errorf("couldn't load config: %w", err)
	}

	kubo, err := kubo.New(
		kubo.Config{
			KUBO_API_AUTH: env.KUBO_API_AUTH,
			KUBO_API_URL:  env.KUBO_API_URL,
		},
		&http.Client{Timeout: timeout},
	)
	if err != nil {
		return fmt.Errorf("couldn't create Kubo client: %w", err)
	}

	v, err := kubo.Version(ctx)
	if err != nil {
		log.Printf("Couldn't connect to Kubo: %s", err)
	}
	log.Printf("Connected to Kubo %s", v.String())

	server := http.Server{
		Addr:        fmt.Sprintf(":%d", config.Port),
		BaseContext: func(net.Listener) context.Context { return ctx },
		Handler:     api.New(api.Config{Kubo: kubo}),
	}

	served := make(chan error, 1)
	go func() { served <- server.ListenAndServe() }()
	log.Printf("Starting server at http://localhost:%d", config.Port)

	select {
	case <-ctx.Done():

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt, syscall.SIGTERM,
	)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Println("couldn't run service:", err)
		os.Exit(1)
	}
}
