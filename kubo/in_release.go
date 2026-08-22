package kubo

import (
	"context"
	"fmt"
	"net/url"
	"path"

	"github.com/NatoBoram/ipapm/apt"
)

func (k *Client) InRelease(ctx context.Context, uri *url.URL, suite string) (apt.InRelease, error) {
	target := path.Join(k.MFS, uri.Host, uri.EscapedPath(), "dists", suite, "InRelease")
	if !path.IsAbs(target) {
		target = "/" + target
	}

	resp, err := k.Request("files/read", target).Send(ctx)
	if err != nil {
		return apt.InRelease{}, fmt.Errorf("requesting InRelease file from MFS: %w", err)
	}
	defer resp.Close()
	if resp.Error != nil {
		return apt.InRelease{}, errorf("kubo error", resp.Error)
	}

	return apt.ParseInRelease(resp.Output)
}
