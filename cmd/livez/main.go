package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/NatoBoram/ipapm/config"
	"github.com/NatoBoram/ipapm/env"
)

func main() {
	env, err := env.LoadEnv()
	if err != nil {
		log.Fatalf("Couldn't load environment variables: %v", err)
	}

	config, err := config.Load(config.Env{
		CONFIG_DIR: env.CONFIG_DIR,
		Name:       env.Name,
	})
	if err != nil {
		log.Fatalf("Couldn't load config: %v", err)
	}

	self, err := url.Parse(fmt.Sprintf("http://localhost:%d/livez", config.Port))
	if err != nil {
		log.Fatalf("Couldn't parse URL: %v", err)
	}

	client := http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get(self.String())
	if err != nil || resp.StatusCode != http.StatusNoContent {
		os.Exit(1)
	}

	os.Exit(0)
}
