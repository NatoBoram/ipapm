package kubo

import (
	"context"
	"fmt"
	"net/url"
	"path"

	"github.com/NatoBoram/ipapm/apt"
)

func (k *Client) RemovePackage(ctx context.Context, uri *url.URL, suite string, file apt.Package) error {
	target := path.Join(k.MFS, uri.Hostname(), uri.EscapedPath(), file.Filename)
	return k.filesRm(ctx, target)
}

func (k *Client) RemoveSource(ctx context.Context, uri *url.URL, suite string, file apt.Source) error {
	files, err := file.FileHashes()
	if err != nil {
		return fmt.Errorf("couldn't get file hashes for source %s: %w", file.Directory, err)
	}

	for _, f := range files {
		target := path.Join(k.MFS, uri.Hostname(), uri.EscapedPath(), f.Filename)
		err := k.filesRm(ctx, target)
		if err != nil {
			return fmt.Errorf("couldn't remove source file %s: %w", f.Filename, err)
		}
	}

	return nil
}

func (k *Client) RemoveComponent(ctx context.Context, uri *url.URL, suite string, component apt.Component) error {
	if component.Architecture == "source" {
		if err := k.removeComponentSources(ctx, uri, suite, component); err != nil {
			return fmt.Errorf("couldn't remove sources: %w", err)
		}
	} else {
		if err := k.removeComponentPackages(ctx, uri, suite, component); err != nil {
			return fmt.Errorf("couldn't remove packages: %w", err)
		}
	}

	for _, file := range component.Files {
		target := path.Join(k.MFS, uri.Host, uri.EscapedPath(), "dists", suite, file.Filename)
		err := k.filesRm(ctx, target)
		if err != nil {
			return fmt.Errorf("couldn't remove file %s: %w", file.Filename, err)
		}
	}

	return nil
}

func (k *Client) removeComponentSources(ctx context.Context, uri *url.URL, suite string, component apt.Component) error {
	sources, err := k.Sources(ctx, uri, suite, component)
	if err != nil {
		return fmt.Errorf("couldn't get Sources for component %s: %w", component.Name, err)
	}

	for _, source := range sources.Sources {
		err = k.RemoveSource(ctx, uri, suite, source)
		if err != nil {
			return fmt.Errorf("couldn't remove source %s: %w", source.Directory, err)
		}
	}

	return nil
}

func (k *Client) removeComponentPackages(ctx context.Context, uri *url.URL, suite string, component apt.Component) error {
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
