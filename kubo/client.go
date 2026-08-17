package kubo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/blang/semver/v4"
	ipfs "github.com/ipfs/kubo"
	"github.com/ipfs/kubo/client/rpc"
)

type Client struct {
	*rpc.HttpApi

	MFS string
}

func (k *Client) Version(ctx context.Context) (*semver.Version, error) {
	resp, err := k.Request("version").Send(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting version from Kubo: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("kubo error (%d) %q", resp.Error.Code, resp.Error.Message)
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
