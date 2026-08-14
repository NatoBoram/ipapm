package apt

import (
	"bufio"
	"bytes"
	"compress/bzip2"
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

// Packages is an entire Packages file.
type Packages []Package

// Package is a single package entry in a [Packages] file.
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

type PackagesBytes struct {
	Packages  Packages
	FileBytes FileBytes
}

func (c *Client) Packages(
	ctx context.Context, uri *url.URL, suite string, component Component,
) (PackagesBytes, error) {
	downloads := make([]FileByte, 0, len(component.Files))

	for _, file := range component.Files {
		downloaded, err := c.file(ctx, uri, suite, file)
		if err != nil {
			return PackagesBytes{}, fmt.Errorf("couldn't download Packages: %w", err)
		}

		downloads = append(downloads, downloaded)
	}

	uncompressed, err := UncompressPackages(downloads)
	if err != nil {
		return PackagesBytes{}, fmt.Errorf("couldn't uncompress Packages: %w", err)
	}

	parsed, err := ParsePackages(uncompressed)
	if err != nil {
		return PackagesBytes{}, fmt.Errorf("couldn't parse Packages: %w", err)
	}

	return PackagesBytes{
		FileBytes: downloads,
		Packages:  parsed,
	}, nil
}

func UncompressPackages(downloads []FileByte) (io.Reader, error) {
	for _, download := range downloads {
		if path.Base(download.Hashes.Filename) != "Packages" {
			continue
		}

		return bytes.NewReader(download.Bytes), nil
	}

	for _, download := range downloads {
		base := path.Base(download.Hashes.Filename)

		switch base {
		case "Packages.gz":
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
		case "Packages.bz2":
			r := bytes.NewReader(download.Bytes)
			u := bzip2.NewReader(r)

			uncompressed, err := io.ReadAll(u)
			if err != nil {
				return nil, fmt.Errorf("failed to decompress bzip2 stream: %w", err)
			}

			return bytes.NewReader(uncompressed), nil
		}
	}

	return nil, errors.New("no supported Packages file found")
}

type packagesSection string

const (
	packagesSectionDescription packagesSection = "Description"
	packagesSectionRoot        packagesSection = "Root"
)

func ParsePackages(r io.Reader) (Packages, error) {
	index := []Package{}
	scanner := bufio.NewScanner(r)

	current := Package{}
	section := packagesSectionRoot

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
			section = packagesSectionRoot
			continue
		}

		// Sections
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			if section == packagesSectionDescription {
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

			continue
		}

		section = packagesSectionRoot

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
			section = packagesSectionDescription
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
