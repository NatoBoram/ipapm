package kubo

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"

	"github.com/NatoBoram/ipapm/apt"
)

func (k *Client) Packages(ctx context.Context, uri *url.URL, suite string, component apt.Component) (apt.PackagesBytes, error) {
	downloads := make([]apt.FileByte, 0, len(component.Files))

	for _, file := range component.Files {
		target := path.Join(k.MFS, uri.Host, uri.EscapedPath(), "dists", suite, file.Filename)
		if !path.IsAbs(target) {
			target = "/" + target
		}

		downloaded, err := k.fileByte(ctx, target, file)
		if err != nil {
			return apt.PackagesBytes{}, fmt.Errorf("reading Packages: %w", err)
		}
		downloads = append(downloads, downloaded)
	}

	uncompressed, err := apt.UncompressPackages(downloads)
	if err != nil {
		return apt.PackagesBytes{}, fmt.Errorf("uncompressing Packages: %w", err)
	}

	parsed, err := apt.ParsePackages(uncompressed)
	if err != nil {
		return apt.PackagesBytes{}, fmt.Errorf("parsing Packages: %w", err)
	}

	return apt.PackagesBytes{
		FileBytes: downloads,
		Packages:  parsed,
	}, nil
}

func (k *Client) fileByte(ctx context.Context, target string, file apt.FileHash) (apt.FileByte, error) {
	resp, err := k.Request("files/read", target).Send(ctx)
	if err != nil {
		return apt.FileByte{}, fmt.Errorf("requesting file from MFS: %w", err)
	}
	defer resp.Close()
	if resp.Error != nil {
		return apt.FileByte{}, errorf("kubo error", resp.Error)
	}

	bytes, err := io.ReadAll(resp.Output)
	if err != nil {
		return apt.FileByte{}, fmt.Errorf("reading response body for %q: %w", target, err)
	}

	downloaded := apt.FileByte{
		Hashes: file,
		Bytes:  bytes,
	}
	return downloaded, nil
}

func (k *Client) Sources(ctx context.Context, uri *url.URL, suite string, component apt.Component) (apt.SourcesBytes, error) {
	downloads := make([]apt.FileByte, 0, len(component.Files))

	for _, file := range component.Files {
		target := path.Join(k.MFS, uri.Host, uri.EscapedPath(), "dists", suite, file.Filename)
		if !path.IsAbs(target) {
			target = "/" + target
		}

		downloaded, err := k.fileByte(ctx, target, file)
		if err != nil {
			return apt.SourcesBytes{}, fmt.Errorf("reading Sources: %w", err)
		}
		downloads = append(downloads, downloaded)
	}

	uncompressed, err := apt.UncompressSources(downloads)
	if err != nil {
		return apt.SourcesBytes{}, fmt.Errorf("uncompressing Sources: %w", err)
	}

	parsed, err := apt.ParseSources(uncompressed)
	if err != nil {
		return apt.SourcesBytes{}, fmt.Errorf("parsing Sources: %w", err)
	}

	return apt.SourcesBytes{
		FileBytes: downloads,
		Sources:   parsed,
	}, nil
}
