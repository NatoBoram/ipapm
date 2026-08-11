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

// Sources is an entire [Sources] file.
type Sources []Source

// Source is a single entry in a [Sources] file.
type Source struct {
	Package          string
	Format           string
	Binary           string
	Architecture     string
	Version          string
	Maintainer       string
	StandardsVersion string
	BuildDepends     string
	Homepage         *url.URL
	VcsBrowser       *url.URL
	VcsGit           *url.URL
	Directory        string
	PackageList      []string
	Files            []SourceSum
	ChecksumsSha1    []SourceSum
	ChecksumsSha256  []SourceSum
	ChecksumsSha512  []SourceSum

	Warnings []error
}

// SourceSum is a single checksum entry in a [Source] definition.
type SourceSum struct {
	Hash string
	Size uint
	Name string
}

type SourcesBytes struct {
	Sources   Sources
	FileBytes FileBytes
}

// Sources downloads and parses Sources index files from a remote HTTP mirror.
func (c *Client) Sources(
	ctx context.Context, uri *url.URL, suite string, component Component,
) (SourcesBytes, error) {
	downloads := make([]FileByte, 0, len(component.Files))

	for name, file := range component.Files {
		base := path.Base(name)
		if !strings.HasPrefix(base, "Sources") {
			continue
		}

		downloaded, err := c.file(ctx, uri, suite, file)
		if err != nil {
			return SourcesBytes{}, fmt.Errorf("couldn't download Sources: %w", err)
		}

		downloads = append(downloads, downloaded)
	}

	uncompressed, err := UncompressSources(downloads)
	if err != nil {
		return SourcesBytes{}, fmt.Errorf("couldn't uncompress Sources: %w", err)
	}

	parsed, err := ParseSources(uncompressed)
	if err != nil {
		return SourcesBytes{}, fmt.Errorf("couldn't parse Sources: %w", err)
	}

	return SourcesBytes{
		FileBytes: downloads,
		Sources:   parsed,
	}, nil
}

// UncompressSources selects an uncompressed or gzipped Sources index from a slice of FileByte.
func UncompressSources(downloads []FileByte) (io.Reader, error) {
	for _, download := range downloads {
		if path.Base(download.Hashes.Filename) == "Sources" {
			return bytes.NewReader(download.Bytes), nil
		}
	}

	for _, download := range downloads {
		base := path.Base(download.Hashes.Filename)
		if base == "Sources.gz" {
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

	return nil, errors.New("no supported Sources file found")
}

type sourcesSection string

const (
	sourcesSectionPackageList     sourcesSection = "Package-List"
	sourcesSectionFiles           sourcesSection = "Files"
	sourcesSectionChecksumsSha1   sourcesSection = "Checksums-Sha1"
	sourcesSectionChecksumsSha256 sourcesSection = "Checksums-Sha256"
	sourcesSectionChecksumsSha512 sourcesSection = "Checksums-Sha512"
	sourcesSectionRoot            sourcesSection = "Root"
)

func ParseSources(r io.Reader) (Sources, error) {
	index := Sources{}
	scanner := bufio.NewScanner(r)

	current := Source{Warnings: []error{}}
	section := sourcesSectionRoot

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip
		if trimmed == "" {
			// Minimum fields to make this actionable
			if current.Package != "" && current.Directory != "" {
				index = append(index, current)
			}

			current = Source{Warnings: []error{}}
			section = sourcesSectionRoot
			continue
		}

		// Sections
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			switch section {
			case sourcesSectionPackageList:
				current.PackageList = append(current.PackageList, trimmed)

			case sourcesSectionFiles:
				sum, err := parseSourceSum(trimmed)
				if err != nil {
					current.Warnings = append(current.Warnings, err)
					continue
				}
				current.Files = append(current.Files, sum)

			case sourcesSectionChecksumsSha1:
				sum, err := parseSourceSum(trimmed)
				if err != nil {
					current.Warnings = append(current.Warnings, err)
					continue
				}
				current.ChecksumsSha1 = append(current.ChecksumsSha1, sum)

			case sourcesSectionChecksumsSha256:
				sum, err := parseSourceSum(trimmed)
				if err != nil {
					current.Warnings = append(current.Warnings, err)
					continue
				}
				current.ChecksumsSha256 = append(current.ChecksumsSha256, sum)

			case sourcesSectionChecksumsSha512:
				sum, err := parseSourceSum(trimmed)
				if err != nil {
					current.Warnings = append(current.Warnings, err)
					continue
				}
				current.ChecksumsSha512 = append(current.ChecksumsSha512, sum)
			}

			continue
		}

		section = sourcesSectionRoot

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
		case "Format":
			current.Format = value
		case "Binary":
			current.Binary = value
		case "Architecture":
			current.Architecture = value
		case "Version":
			current.Version = value
		case "Maintainer":
			current.Maintainer = value
		case "Standards-Version":
			current.StandardsVersion = value
		case "Build-Depends":
			current.BuildDepends = value
		case "Directory":
			current.Directory = value
		case "Homepage":
			parsed, err := url.Parse(value)
			if err != nil {
				current.Warnings = append(
					current.Warnings,
					fmt.Errorf("error parsing Homepage for source %s %s: %w", current.Package, current.Version, err),
				)
				continue
			}
			current.Homepage = parsed
		case "Vcs-Browser":
			parsed, err := url.Parse(value)
			if err != nil {
				current.Warnings = append(
					current.Warnings,
					fmt.Errorf("error parsing Vcs-Browser for source %s %s: %w", current.Package, current.Version, err),
				)
				continue
			}
			current.VcsBrowser = parsed
		case "Vcs-Git":
			parsed, err := url.Parse(value)
			if err != nil {
				current.Warnings = append(
					current.Warnings,
					fmt.Errorf("error parsing Vcs-Git for source %s %s: %w", current.Package, current.Version, err),
				)
				continue
			}
			current.VcsGit = parsed
		case "Package-List":
			section = sourcesSectionPackageList
		case "Files":
			section = sourcesSectionFiles
		case "Checksums-Sha1":
			section = sourcesSectionChecksumsSha1
		case "Checksums-Sha256":
			section = sourcesSectionChecksumsSha256
		case "Checksums-Sha512":
			section = sourcesSectionChecksumsSha512
		}
	}

	if err := scanner.Err(); err != nil {
		return Sources{}, fmt.Errorf("failed to scan Sources file: %w", err)
	}

	if current.Package != "" && current.Directory != "" {
		index = append(index, current)
	}

	return index, nil
}

// parseSourceSum parses lines in multiline checksum blocks (e.g. "8df5a8cb... 1635 adw-gtk3.dsc").
func parseSourceSum(line string) (SourceSum, error) {
	fields := strings.Fields(line)
	if len(fields) != 3 {
		return SourceSum{}, fmt.Errorf("invalid checksum line: %q", line)
	}

	size, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return SourceSum{}, fmt.Errorf("invalid size %q in checksum line: %w", fields[1], err)
	}

	return SourceSum{
		Hash: fields[0],
		Size: uint(size),
		Name: fields[2],
	}, nil
}
