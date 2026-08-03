package apt

import (
	"net/http"
)

type Client struct {
	*http.Client
}

func New(client *http.Client) *Client {
	return new(Client{client})
}
