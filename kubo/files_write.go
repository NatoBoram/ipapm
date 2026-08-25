package kubo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"path"

	"github.com/NatoBoram/ipapm/apt"
)

func (k *Client) WriteInRelease(ctx context.Context, uri *url.URL, suite string, body apt.InRelease) error {
	target := path.Join(k.MFS, uri.Host, uri.EscapedPath(), "dists", suite, "InRelease")
	return k.filesWrite(ctx, target, bytes.NewReader(body.Raw))
}

func (k *Client) WritePackages(ctx context.Context, uri *url.URL, suite string, file apt.FileByte) error {
	target := path.Join(k.MFS, uri.Host, uri.EscapedPath(), "dists", suite, file.Hashes.Filename)
	return k.filesWrite(ctx, target, bytes.NewReader(file.Bytes))
}

func (k *Client) WriteSources(ctx context.Context, uri *url.URL, suite string, file apt.FileByte) error {
	target := path.Join(k.MFS, uri.Host, uri.EscapedPath(), "dists", suite, file.Hashes.Filename)
	return k.filesWrite(ctx, target, bytes.NewReader(file.Bytes))
}

func (k *Client) WritePackage(ctx context.Context, uri *url.URL, file apt.FileHash, body io.Reader) error {
	target := path.Join(k.MFS, uri.Host, uri.EscapedPath(), file.Filename)
	return k.filesWrite(ctx, target, body)
}

func (k *Client) WriteSource(ctx context.Context, uri *url.URL, file apt.FileHash, body io.Reader) error {
	target := path.Join(k.MFS, uri.Host, uri.EscapedPath(), file.Filename)
	return k.filesWrite(ctx, target, body)
}

func (k *Client) WriteFile(ctx context.Context, uri *url.URL, suite string, file apt.FileHash, body io.Reader) error {
	target := path.Join(k.MFS, uri.Host, uri.EscapedPath(), "dists", suite, file.Filename)
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
		return fmt.Errorf("writing %s to MFS: %w", fileName, err)
	}
	defer resp.Close()
	if resp.Error != nil {
		return errorf("kubo error", resp.Error)
	}

	return nil
}
