package apt

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	h "github.com/NatoBoram/ipapm/http"
)

// InRelease is a parsed InRelease file from a Debian Repository.
//
// See https://wiki.debian.org/DebianRepository/Format#A.22Release.22_files.
type InRelease struct {
	Hash          string
	Architectures []string
	Codename      string
	Components    []string
	Date          string
	Description   string
	Label         string
	Origin        string
	Suite         string
	Version       string
	AcquireByHash string

	MD5Sum []InReleaseSum
	SHA1   []InReleaseSum
	SHA256 []InReleaseSum
	SHA512 []InReleaseSum

	Raw      []byte
	Warnings []error
}

// InReleaseSum is a single checksum entry in an [InRelease] file.
type InReleaseSum struct {
	Hash string
	Size uint
	Path string
}

// InRelease gets an [InRelease] file from a Debian Repository and parses it.
func (c *Client) InRelease(ctx context.Context, uri *url.URL, suite string) (InRelease, error) {
	target := uri.JoinPath("dists", suite, "InRelease")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return InRelease{}, fmt.Errorf("building request for %q: %w", target, err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return InRelease{}, fmt.Errorf("requesting %q: %w", target, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return InRelease{}, fmt.Errorf("unexpected status for %q: %w", target, h.ErrNotFound)
	}
	if resp.StatusCode != http.StatusOK {
		return InRelease{}, fmt.Errorf("unexpected status %s for %q", resp.Status, target)
	}

	return ParseInRelease(resp.Body)
}

type inReleaseSection string

const (
	inReleaseSectionMD5Sum inReleaseSection = "MD5Sum"
	inReleaseSectionSHA1   inReleaseSection = "SHA1"
	inReleaseSectionSHA256 inReleaseSection = "SHA256"
	inReleaseSectionSHA512 inReleaseSection = "SHA512"
	inReleaseSectionPGP    inReleaseSection = "PGP"
	inReleaseSectionRoot   inReleaseSection = "Root"
)

// ParseInRelease parses an [InRelease] file from a Debian Repository.
func ParseInRelease(r io.Reader) (InRelease, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return InRelease{}, fmt.Errorf("reading InRelease file: %w", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(raw))
	manifest := InRelease{
		MD5Sum:   []InReleaseSum{},
		SHA1:     []InReleaseSum{},
		SHA256:   []InReleaseSum{},
		SHA512:   []InReleaseSum{},
		Warnings: []error{},
		Raw:      raw,
	}

	sectionMap := map[inReleaseSection]*[]InReleaseSum{
		inReleaseSectionMD5Sum: &manifest.MD5Sum,
		inReleaseSectionSHA1:   &manifest.SHA1,
		inReleaseSectionSHA256: &manifest.SHA256,
		inReleaseSectionSHA512: &manifest.SHA512,
	}

	stringMap := map[string]*string{
		"Hash":            &manifest.Hash,
		"Codename":        &manifest.Codename,
		"Date":            &manifest.Date,
		"Description":     &manifest.Description,
		"Label":           &manifest.Label,
		"Origin":          &manifest.Origin,
		"Suite":           &manifest.Suite,
		"Version":         &manifest.Version,
		"Acquire-By-Hash": &manifest.AcquireByHash,
	}

	stringsMap := map[string]*[]string{
		"Architectures": &manifest.Architectures,
		"Components":    &manifest.Components,
	}

	section := inReleaseSectionRoot
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip
		if trimmed == "" || strings.HasPrefix(trimmed, "-----BEGIN PGP SIGNED MESSAGE-----") {
			continue
		}

		// Sections
		switch trimmed {
		case "MD5Sum:":
			section = inReleaseSectionMD5Sum
			continue
		case "SHA1:":
			section = inReleaseSectionSHA1
			continue
		case "SHA256:":
			section = inReleaseSectionSHA256
			continue
		case "SHA512:":
			section = inReleaseSectionSHA512
			continue
		case "-----BEGIN PGP SIGNATURE-----":
			section = inReleaseSectionPGP
			continue
		case "-----END PGP SIGNATURE-----":
			section = inReleaseSectionRoot
			continue
		}

		// Checksum fields
		if ptr, ok := sectionMap[section]; ok && (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")) {
			parts := strings.Fields(trimmed)
			if len(parts) != 3 {
				continue
			}

			hash := strings.TrimSpace(parts[0])
			path := strings.TrimSpace(parts[2])

			size, err := strconv.ParseUint(parts[1], 10, 64)
			if err != nil {
				manifest.Warnings = append(
					manifest.Warnings,
					fmt.Errorf(
						"parsing package size for %q in %s: %w",
						path, section, err,
					),
				)
				continue
			}

			*ptr = append(*ptr, InReleaseSum{
				Hash: hash,
				Size: uint(size),
				Path: path,
			})

			continue
		}

		// PGP signatures may contain a version, which would conflict with
		// InRelease's version. Example:
		// Version: BSN Pgp v1.0.0.0
		if section == inReleaseSectionPGP {
			continue
		}

		// Key-values
		if split := strings.SplitN(trimmed, ":", 2); len(split) == 2 {
			key := strings.TrimSpace(split[0])

			if ptr, ok := stringMap[key]; ok {
				*ptr = strings.TrimSpace(split[1])
				continue
			}

			if ptr, ok := stringsMap[key]; ok {
				*ptr = strings.Fields(strings.TrimSpace(split[1]))
				continue
			}
		}

	}

	if err := scanner.Err(); err != nil {
		return InRelease{}, fmt.Errorf("scanning InRelease file: %w", err)
	}

	return manifest, nil
}
