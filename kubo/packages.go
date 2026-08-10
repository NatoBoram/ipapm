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

func (k *Client) RemovePackage(ctx context.Context, uri *url.URL, suite string, file apt.Package) error {
	target := path.Join(k.MFS, uri.Hostname(), uri.EscapedPath(), file.Filename)
	return k.filesRm(ctx, target)
}

func (k *Client) filesRm(ctx context.Context, fileName string) error {
	req := k.Request("files/rm").
		Arguments(fileName).
		Option("recursive", true).
		Option("force", true)

	resp, err := req.Send(ctx)
	if err != nil {
		return fmt.Errorf("failed to remove %s from MFS: %w", fileName, err)
	}
	defer resp.Close()
	if resp.Error != nil {
		return fmt.Errorf("kubo error (%d) \"%s\"", resp.Error.Code, resp.Error.Message)
	}

	return nil
}

func (k *Client) RemoveComponent(ctx context.Context, uri *url.URL, suite string, component apt.Component) error {
	packages, err := k.Packages(ctx, uri, suite, component)
	if err != nil {
		return fmt.Errorf("couldn't get Packages for component %s/%s: %w", component.Name, component.Architecture, err)
	}

	for _, file := range packages.Packages {
		err = k.RemovePackage(ctx, uri, suite, file)
		if err != nil {
			return fmt.Errorf("couldn't remove package %s %s: %w", file.Package, file.Version, err)
		}
	}

	for _, file := range component.Files {
		target := path.Join(k.MFS, uri.Host, uri.EscapedPath(), "dists", suite, file.Filename)
		err = k.filesRm(ctx, target)
		if err != nil {
			return fmt.Errorf("couldn't remove file %s: %w", file.Filename, err)
		}
	}

	return nil
}

func (k *Client) RemovePackages(ctx context.Context, uri *url.URL, suite string, file apt.FileByte) error {
	target := path.Join(k.MFS, uri.Hostname(), uri.EscapedPath(), "dists", suite, file.Hashes.Filename)
	return k.filesRm(ctx, target)
}
