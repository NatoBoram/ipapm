package apt

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (c *Client) InRelease(ctx context.Context, uri *url.URL, suite string) (InRelease, error) {
	target := uri.JoinPath("dists", suite, "InRelease")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return InRelease{}, fmt.Errorf("failed to build request for %s: %w", target, err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return InRelease{}, fmt.Errorf("failed to fetch %s: %w", target, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return InRelease{}, fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, target)
	}

	return ParseInRelease(resp.Body)
}

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

	MD5Sum []Checksum
	SHA1   []Checksum
	SHA256 []Checksum
	SHA512 []Checksum

	Raw string
}

type Checksum struct {
	Hash string
	Size uint
	Path string
}

type InReleaseSection string

const (
	InReleaseSectionMD5Sum InReleaseSection = "MD5Sum"
	InReleaseSectionSHA1   InReleaseSection = "SHA1"
	InReleaseSectionSHA256 InReleaseSection = "SHA256"
	InReleaseSectionSHA512 InReleaseSection = "SHA512"
	InReleaseSectionPGP    InReleaseSection = "PGP"
	InReleaseSectionNone   InReleaseSection = ""
)

func ParseInRelease(r io.Reader) (InRelease, error) {
	scanner := bufio.NewScanner(r)
	manifest := InRelease{
		MD5Sum: []Checksum{},
		SHA1:   []Checksum{},
		SHA256: []Checksum{},
		SHA512: []Checksum{},
	}

	sectionMap := map[InReleaseSection]*[]Checksum{
		InReleaseSectionMD5Sum: &manifest.MD5Sum,
		InReleaseSectionSHA1:   &manifest.SHA1,
		InReleaseSectionSHA256: &manifest.SHA256,
		InReleaseSectionSHA512: &manifest.SHA512,
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

	section := InReleaseSectionNone
	for scanner.Scan() {
		line := scanner.Text()
		manifest.Raw += line + "\n"
		trimmed := strings.TrimSpace(line)

		// Skip
		if trimmed == "" || strings.HasPrefix(trimmed, "-----BEGIN PGP SIGNED MESSAGE-----") {
			continue
		}

		// Sections
		switch trimmed {
		case "MD5Sum:":
			section = InReleaseSectionMD5Sum
			continue
		case "SHA1:":
			section = InReleaseSectionSHA1
			continue
		case "SHA256:":
			section = InReleaseSectionSHA256
			continue
		case "SHA512:":
			section = InReleaseSectionSHA512
			continue
		case "-----BEGIN PGP SIGNATURE-----":
			section = InReleaseSectionPGP
			continue
		case "-----END PGP SIGNATURE-----":
			section = InReleaseSectionNone
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
				log.Printf("Error while parsing package size for %s: %s", path, err)
				continue
			}

			*ptr = append(*ptr, Checksum{
				Hash: hash,
				Size: uint(size),
				Path: path,
			})

			continue
		}

		// PGP signatures may contain a version, which would conflict with
		// InRelease's version. Example:
		// Version: BSN Pgp v1.0.0.0
		if section == InReleaseSectionPGP {
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
		return InRelease{}, fmt.Errorf("failed to scan InRelease file: %w", err)
	}

	return manifest, nil
}
