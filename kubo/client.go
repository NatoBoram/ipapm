package kubo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/NatoBoram/ipapm/apt"
	"github.com/blang/semver/v4"
	ipfs "github.com/ipfs/kubo"
	"github.com/ipfs/kubo/client/rpc"
)

const timeout time.Duration = time.Second * 30

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

func (k *Client) InRelease(ctx context.Context, uri *url.URL, suite string) (apt.InRelease, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	target := path.Join(k.MFS, uri.Host, uri.EscapedPath(), "dists", suite, "InRelease")
	if !path.IsAbs(target) {
		target = "/" + target
	}

	resp, err := k.Request("files/read", target).Send(ctx)
	if err != nil {
		return apt.InRelease{}, fmt.Errorf("failed to request InRelease file from MFS: %w", err)
	}
	if resp.Error != nil {
		if resp.Error.Message == fmt.Sprintf("%s: %s", target, os.ErrNotExist) {
			return apt.InRelease{}, fmt.Errorf("%s: %w", target, os.ErrNotExist)
		}

		return apt.InRelease{}, fmt.Errorf("kubo error (%d) \"%s\"", resp.Error.Code, resp.Error.Message)
	}
	defer resp.Close()

	return apt.ParseInRelease(resp.Output)
}
