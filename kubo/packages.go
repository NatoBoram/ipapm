package kubo

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/NatoBoram/ipapm/apt"
)

func (k *Client) Packages(ctx context.Context, uri *url.URL, suite string, component apt.Component) (apt.PackagesBytes, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	downloads := make([]apt.FileByte, 0, len(component.Files))

	for name, file := range component.Files {
		base := path.Base(name)
		if !strings.HasPrefix(base, "Packages") {
			continue
		}

		target := path.Join(k.MFS, uri.Host, uri.EscapedPath(), "dists", suite, file.Filename)
		if !path.IsAbs(target) {
			target = "/" + target
		}

		downloaded, err := k.fileByte(ctx, target, file)
		if err != nil {
			return apt.PackagesBytes{}, fmt.Errorf("couldn't read Packages: %w", err)
		}
		downloads = append(downloads, downloaded)
	}

	uncompressed, err := apt.UncompressPackages(downloads)
	if err != nil {
		return apt.PackagesBytes{}, fmt.Errorf("couldn't uncompress Packages: %w", err)
	}

	parsed, err := apt.ParsePackages(uncompressed)
	if err != nil {
		return apt.PackagesBytes{}, fmt.Errorf("couldn't parse Packages: %w", err)
	}

	return apt.PackagesBytes{
		FileBytes: downloads,
		Packages:  parsed,
	}, nil
}

func (k *Client) fileByte(ctx context.Context, target string, file apt.FileHash) (apt.FileByte, error) {
	resp, err := k.Request("files/read", target).Send(ctx)
	if err != nil {
		return apt.FileByte{}, fmt.Errorf("failed to request Packages file from MFS: %w", err)
	}
	defer resp.Close()
	if resp.Error != nil {
		if resp.Error.Message == fmt.Sprintf("%s: %s", target, os.ErrNotExist) {
			return apt.FileByte{}, fmt.Errorf("%s: %w", target, os.ErrNotExist)
		}

		return apt.FileByte{}, fmt.Errorf("kubo error (%d) \"%s\"", resp.Error.Code, resp.Error.Message)
	}

	bytes, err := io.ReadAll(resp.Output)
	if err != nil {
		return apt.FileByte{}, fmt.Errorf("failed to read response body for %s: %w", target, err)
	}

	downloaded := apt.FileByte{
		Hashes: file,
		Bytes:  bytes,
	}
	return downloaded, nil
}
