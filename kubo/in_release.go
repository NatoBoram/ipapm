package kubo

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"

	"github.com/NatoBoram/ipapm/apt"
)

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
	defer resp.Close()
	if resp.Error != nil {
		if resp.Error.Message == fmt.Sprintf("%s: %s", target, os.ErrNotExist) {
			return apt.InRelease{}, fmt.Errorf("%s: %w", target, os.ErrNotExist)
		}

		return apt.InRelease{}, fmt.Errorf("kubo error (%d) \"%s\"", resp.Error.Code, resp.Error.Message)
	}

	return apt.ParseInRelease(resp.Output)
}
