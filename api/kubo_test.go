package api_test

import (
	"context"

	"github.com/blang/semver/v4"
)

type Kubo struct {
	version *semver.Version
	err     error
}

func (k *Kubo) Version(ctx context.Context) (*semver.Version, error) {
	return k.version, k.err
}
