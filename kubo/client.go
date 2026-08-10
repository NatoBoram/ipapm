package kubo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/blang/semver/v4"
	ipfs "github.com/ipfs/kubo"
	"github.com/ipfs/kubo/client/rpc"
)

const timeout time.Duration = time.Minute * 1

type Client struct {
	*rpc.HttpApi

	MFS string
}

func (k *Client) Version(ctx context.Context) (*semver.Version, error) {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(timeout))
	defer cancel()

	resp, err := k.Request("version").Send(ctx)
	if err != nil {
		return nil, fmt.Errorf("couldn't get version from Kubo: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("couldn't get version from Kubo: %w", resp.Error)
	}
	defer resp.Close()

	var out ipfs.VersionInfo
	if err := json.NewDecoder(resp.Output).Decode(&out); err != nil {
		return nil, err
	}

	remoteVersion, err := semver.New(out.Version)
	if err != nil {
		return nil, err
	}

	return remoteVersion, nil
}
