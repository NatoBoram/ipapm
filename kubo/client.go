package kubo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"path"
	"time"

	"github.com/NatoBoram/ipapm/apt"
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

func (k *Client) WriteInRelease(ctx context.Context, uri *url.URL, suite string, body apt.InRelease) error {
	target := path.Join(k.MFS, uri.Hostname(), uri.EscapedPath(), "dists", suite, "InRelease")
	return k.filesWrite(ctx, target, bytes.NewReader(body.Raw))
}

func (k *Client) WritePackages(ctx context.Context, uri *url.URL, suite string, file apt.FileByte) error {
	target := path.Join(k.MFS, uri.Hostname(), uri.EscapedPath(), "dists", suite, file.Hashes.Filename)
	return k.filesWrite(ctx, target, bytes.NewReader(file.Bytes))
}

func (k *Client) WritePackage(ctx context.Context, uri *url.URL, suite string, file apt.FileHash, body io.Reader) error {
	target := path.Join(k.MFS, uri.Hostname(), uri.EscapedPath(), file.Filename)
	return k.filesWrite(ctx, target, body)
}

// filesWrite writes to the MFS.
//
// See https://docs.ipfs.tech/reference/kubo/rpc#api-v0-files-write.
func (k *Client) filesWrite(ctx context.Context, fileName string, body io.Reader) error {
	req := k.Request("files/write").
		Arguments(fileName).
		Option("create", true).
		Option("parents", true).
		Option("truncate", true).
		FileBody(body)

	resp, err := req.Send(ctx)
	if err != nil {
		return fmt.Errorf("failed to write %s to MFS: %w", fileName, err)
	}
	defer resp.Close()
	if resp.Error != nil {
		return fmt.Errorf("kubo error (%d) \"%s\"", resp.Error.Code, resp.Error.Message)
	}

	return nil
}
