package apt

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type FileByte struct {
	hashes FileHash
	bytes  []byte
}

func (c *Client) File(
	ctx context.Context, uri *url.URL, suite string,
	component Component, file FileHash,
) (FileByte, error) {
	target := uri.JoinPath("dists", suite, file.Path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return FileByte{}, fmt.Errorf("failed to build request for %s: %w", target, err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return FileByte{}, fmt.Errorf("failed to fetch %s: %w", target, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return FileByte{}, fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, target)
	}

	// Stream and hash simultaneously
	md5Hasher := md5.New()
	sha1Hasher := sha1.New()
	sha256Hasher := sha256.New()
	sha512Hasher := sha512.New()
	multiHasher := io.MultiWriter(md5Hasher, sha1Hasher, sha256Hasher, sha512Hasher)

	body, err := io.ReadAll(io.TeeReader(resp.Body, multiHasher))
	if err != nil {
		return FileByte{}, fmt.Errorf("failed to read response body for %s: %w", target, err)
	}

	if uint(len(body)) != file.Size {
		return FileByte{}, fmt.Errorf("size mismatch for %s: expected %d, got %d", file.Path, file.Size, len(body))
	}

	if file.MD5Sum != "" {
		checksum := hex.EncodeToString(md5Hasher.Sum(nil))
		if checksum != file.MD5Sum {
			return FileByte{}, fmt.Errorf("MD5 mismatch for %s: expected %s, got %s", file.Path, file.MD5Sum, checksum)
		}
	}

	if file.SHA1 != "" {
		checksum := hex.EncodeToString(sha1Hasher.Sum(nil))
		if checksum != file.SHA1 {
			return FileByte{}, fmt.Errorf("SHA1 mismatch for %s: expected %s, got %s", file.Path, file.SHA1, checksum)
		}
	}

	if file.SHA256 != "" {
		checksum := hex.EncodeToString(sha256Hasher.Sum(nil))
		if checksum != file.SHA256 {
			return FileByte{}, fmt.Errorf("SHA256 mismatch for %s: expected %s, got %s", file.Path, file.SHA256, checksum)
		}
	}

	if file.SHA512 != "" {
		checksum := hex.EncodeToString(sha512Hasher.Sum(nil))
		if checksum != file.SHA512 {
			return FileByte{}, fmt.Errorf("SHA512 mismatch for %s: expected %s, got %s", file.Path, file.SHA512, checksum)
		}
	}

	return FileByte{bytes: body, hashes: file}, nil
}
