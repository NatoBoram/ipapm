package kubo

import (
	"context"
	"fmt"
	"net/url"
	"path"

	"github.com/NatoBoram/ipapm/apt"
)

func (k *Client) RemovePackage(ctx context.Context, uri *url.URL, file apt.Package) error {
	target := path.Join(k.MFS, uri.Hostname(), uri.EscapedPath(), file.Filename)
	return k.filesRm(ctx, target)
}

func (k *Client) RemoveSource(ctx context.Context, uri *url.URL, file apt.FileHash) error {
	target := path.Join(k.MFS, uri.Hostname(), uri.EscapedPath(), file.Filename)
	return k.filesRm(ctx, target)
}

func (k *Client) RemoveComponent(ctx context.Context, uri *url.URL, suite string, component apt.Component) error {
	if component.Architecture == "source" {
		if err := k.removeComponentSources(ctx, uri, suite, component); err != nil {
			return fmt.Errorf("removing sources: %w", err)
		}
	} else {
		if err := k.removeComponentPackages(ctx, uri, suite, component); err != nil {
			return fmt.Errorf("removing packages: %w", err)
		}
	}

	for _, file := range component.Files {
		target := path.Join(k.MFS, uri.Host, uri.EscapedPath(), "dists", suite, file.Filename)
		err := k.filesRm(ctx, target)
		if err != nil {
			return fmt.Errorf("removing file %q: %w", file.Filename, err)
		}
	}

	return nil
}

func (k *Client) removeComponentSources(ctx context.Context, uri *url.URL, suite string, component apt.Component) error {
	sources, err := k.Sources(ctx, uri, suite, component)
	if err != nil {
		return fmt.Errorf("getting Sources for component %q: %w", component.Name, err)
	}

	for _, source := range sources.Sources {
		fhs, err := source.FileHashes()
		if err != nil {
			return fmt.Errorf("getting file hashes for source %s %s: %w", source.Package, source.Version, err)
		}

		for _, fh := range fhs {
			err = k.RemoveSource(ctx, uri, fh)
			if err != nil {
				return fmt.Errorf("removing source for %s %s: %w", source.Package, source.Version, err)
			}
		}
	}

	return nil
}

func (k *Client) removeComponentPackages(ctx context.Context, uri *url.URL, suite string, component apt.Component) error {
	packages, err := k.Packages(ctx, uri, suite, component)
	if err != nil {
		return fmt.Errorf("getting Packages for component %s/%s: %w", component.Name, component.Architecture, err)
	}

	for _, file := range packages.Packages {
		err = k.RemovePackage(ctx, uri, file)
		if err != nil {
			return fmt.Errorf("removing package %s %s: %w", file.Package, file.Version, err)
		}
	}

	return nil
}

func (k *Client) RemovePackages(ctx context.Context, uri *url.URL, suite string, file apt.FileByte) error {
	target := path.Join(k.MFS, uri.Hostname(), uri.EscapedPath(), "dists", suite, file.Hashes.Filename)
	return k.filesRm(ctx, target)
}

func (k *Client) RemoveSources(ctx context.Context, uri *url.URL, suite string, file apt.FileByte) error {
	target := path.Join(k.MFS, uri.Hostname(), uri.EscapedPath(), "dists", suite, file.Hashes.Filename)
	return k.filesRm(ctx, target)
}

func (k *Client) RemoveFile(ctx context.Context, uri *url.URL, suite string, file apt.FileHash) error {
	target := path.Join(k.MFS, uri.Hostname(), uri.EscapedPath(), "dists", suite, file.Filename)
	return k.filesRm(ctx, target)
}

func (k *Client) filesRm(ctx context.Context, fileName string) error {
	req := k.Request("files/rm").
		Arguments(fileName).
		Option("recursive", true).
		Option("force", true)

	resp, err := req.Send(ctx)
	if err != nil {
		return fmt.Errorf("removing %q from MFS: %w", fileName, err)
	}
	defer resp.Close()
	if resp.Error != nil {
		return fmt.Errorf("kubo error (%d) %q", resp.Error.Code, resp.Error.Message)
	}

	return nil
}
