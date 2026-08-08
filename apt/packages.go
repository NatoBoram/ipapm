package apt

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
)

type Packages []Package

type Package struct {
	Package       string
	Version       string
	Architecture  string
	Section       string
	Priority      string
	InstalledSize uint
	Maintainer    string
	Description   string
	Homepage      *url.URL
	Conflicts     string
	Depends       string
	Recommends    string
	Provides      string
	Replaces      string
	MD5sum        string
	SHA1          string
	SHA256        string
	SHA512        string
	Size          uint
	Filename      string
}

type ComponentBytes struct {
	Component     Component
	Packages      Packages
	PackagesBytes []FileByte
}

func (c *Client) Packages(
	ctx context.Context, uri *url.URL, suite string, component Component,
) (ComponentBytes, error) {
	downloads := make([]FileByte, 0, len(component.Files))

	for name, file := range component.Files {
		base := path.Base(name)
		if !strings.HasPrefix(base, "Packages") {
			continue
		}

		downloaded, err := c.File(ctx, uri, suite, component, file)
		if err != nil {
			return ComponentBytes{}, fmt.Errorf("couldn't download Packages: %w", err)
		}

		downloads = append(downloads, downloaded)
	}

	uncompressed, err := uncompressPackages(downloads)
	if err != nil {
		return ComponentBytes{}, fmt.Errorf("couldn't uncompress Packages: %w", err)
	}

	parsed, err := ParsePackages(uncompressed)
	if err != nil {
		return ComponentBytes{}, fmt.Errorf("couldn't parse Packages: %w", err)
	}

	return ComponentBytes{
		Component:     component,
		Packages:      parsed,
		PackagesBytes: downloads,
	}, nil
}

func uncompressPackages(downloads []FileByte) (io.Reader, error) {
	for _, download := range downloads {
		if path.Base(download.Hashes.Path) != "Packages" {
			continue
		}

		return bytes.NewReader(download.Bytes), nil
	}

	for _, download := range downloads {
		base := path.Base(download.Hashes.Path)
		if base == "Packages.gz" {
			r := bytes.NewReader(download.Bytes)
			u, err := gzip.NewReader(r)
			if err != nil {
				return nil, fmt.Errorf("failed to open gzip stream: %w", err)
			}
			defer u.Close()

			uncompressed, err := io.ReadAll(u)
			if err != nil {
				return nil, fmt.Errorf("failed to decompress gzip stream: %w", err)
			}

			return bytes.NewReader(uncompressed), nil
		}
	}

	return nil, errors.New("no supported Packages file found")
}

func ParsePackages(r io.Reader) (Packages, error) {
	return []Package{}, errors.New("not implemented")
}
