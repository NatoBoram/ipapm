package apt

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"strconv"
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

	Warnings []error
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

type PackagesSection string

const (
	PackagesSectionDescription PackagesSection = "Description"
	PackagesSectionRoot        PackagesSection = "Root"
)

func ParsePackages(r io.Reader) (Packages, error) {
	index := []Package{}
	scanner := bufio.NewScanner(r)

	current := Package{}
	section := PackagesSectionRoot

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip
		if trimmed == "" {
			// There's basically only one field I'll actually need to consider it
			// valid and actionable.
			if current.Filename != "" {
				index = append(index, current)
			}

			current = Package{}
			section = PackagesSectionRoot
			continue
		}

		// Sections
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			if section == PackagesSectionDescription {
				if trimmed == "." {
					current.Description += "\n\n"
					continue
				}

				if strings.HasSuffix(current.Description, "\n") {
					current.Description += trimmed
					continue
				}

				current.Description += " " + trimmed
				continue
			}
		}

		section = PackagesSectionRoot

		// Key-values
		split := strings.SplitN(trimmed, ":", 2)
		if len(split) != 2 {
			continue
		}

		key := strings.TrimSpace(split[0])
		value := strings.TrimSpace(split[1])

		switch key {
		case "Package":
			current.Package = value
		case "Version":
			current.Version = value
		case "Architecture":
			current.Architecture = value
		case "Section":
			current.Section = value
		case "Priority":
			current.Priority = value
		case "Installed-Size":
			size, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				current.Warnings = append(
					current.Warnings,
					fmt.Errorf(
						"error parsing %s for %s %s: %w",
						key, current.Package, current.Version, err,
					),
				)
				continue
			}
			current.InstalledSize = uint(size)
		case "Maintainer":
			current.Maintainer = value
		case "Description":
			section = PackagesSectionDescription
			current.Description = value
		case "Homepage":
			parsed, err := url.Parse(value)
			if err != nil {
				current.Warnings = append(
					current.Warnings,
					fmt.Errorf(
						"error parsing %s for %s %s: %w",
						key, current.Package, current.Version, err,
					),
				)
				continue
			}

			current.Homepage = parsed
		case "Conflicts":
			current.Conflicts = value
		case "Depends":
			current.Depends = value
		case "Recommends":
			current.Recommends = value
		case "Provides":
			current.Provides = value
		case "Replaces":
			current.Replaces = value
		case "MD5sum":
			current.MD5sum = value
		case "SHA1":
			current.SHA1 = value
		case "SHA256":
			current.SHA256 = value
		case "SHA512":
			current.SHA512 = value
		case "Size":
			size, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				current.Warnings = append(
					current.Warnings,
					fmt.Errorf(
						"error parsing %s for %s %s: %w",
						key, current.Package, current.Version, err,
					),
				)
				continue
			}
			current.Size = uint(size)
		case "Filename":
			current.Filename = value
		}
	}

	if err := scanner.Err(); err != nil {
		return Packages{}, fmt.Errorf("failed to scan Packages file: %w", err)
	}

	if current.Filename != "" {
		index = append(index, current)
	}

	return index, nil
}
