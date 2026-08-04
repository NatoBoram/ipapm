package kubo

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/ipfs/kubo/client/rpc"
)

type Config struct {
	KUBO_API_AUTH string
	KUBO_API_URL  string

	MFS string
}

func New(config Config, c *http.Client) (*Client, error) {
	kubo, err := rpc.NewURLApiWithClient(config.KUBO_API_URL, c)
	client := new(Client{HttpApi: kubo, MFS: config.MFS})
	if err != nil {
		return client, fmt.Errorf("couldn't create a new api client: %w", err)
	}

	if config.KUBO_API_AUTH == "" {
		return client, err
	}

	secret, err := DecodeAuthSecret(config.KUBO_API_AUTH)
	if err != nil {
		return client, fmt.Errorf("couldn't decode auth secret: %w", err)
	}
	if secret.Token != "" {
		client.Headers.Add("Authorization", fmt.Sprintf("Bearer %s", secret.Token))
		return client, err
	}

	if secret.Username != "" && secret.Password != "" {
		combined := fmt.Appendf([]byte(""), "%s:%s", secret.Username, secret.Password)
		encoded := base64.StdEncoding.EncodeToString(combined)
		client.Headers.Add("Authorization", fmt.Sprintf("Basic %s", encoded))
		return client, err
	}

	return client, fmt.Errorf("invalid KUBO_API_AUTH format")
}
