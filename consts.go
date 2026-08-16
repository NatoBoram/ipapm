package main

import "time"

const (
	// interval at which repositories should be updated.
	interval time.Duration = 4 * time.Hour
	// timeout for main actions, like closing the HTTP server.
	timeout time.Duration = 5 * time.Second
)
