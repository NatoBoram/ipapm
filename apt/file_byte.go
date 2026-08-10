package apt

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"
	"net/url"
)

// FileByte is a file from a Debian Repository along with its hashes.
type FileByte struct {
	Hashes FileHash
	Bytes  []byte
}

type FileBytes []FileByte

func (c *Client) file(
	ctx context.Context, uri *url.URL, suite string,
	file FileHash,
) (FileByte, error) {
	target := uri.JoinPath("dists", suite, file.Filename)

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

	multiHasher := makeHashers(file)

	buf := bytes.NewBuffer(make([]byte, 0, file.Size))
	_, err = io.Copy(buf, io.TeeReader(resp.Body, multiHasher.writer))
	if err != nil {
		return FileByte{}, fmt.Errorf("failed to read response body for %s: %w", target, err)
	}
	body := buf.Bytes()

	if uint(len(body)) != file.Size {
		return FileByte{}, fmt.Errorf("size mismatch for %s: expected %d, got %d", file.Filename, file.Size, len(body))
	}

	for _, hasher := range multiHasher.hashers {
		checksum := hex.EncodeToString(hasher.writer.Sum(nil))
		if checksum != hasher.sum {
			return FileByte{}, fmt.Errorf("%s mismatch for %s: expected %s, got %s", hasher.kind, file.Filename, hasher.sum, checksum)
		}
	}

	return FileByte{Bytes: body, Hashes: file}, nil
}

type hashers struct {
	hashers []hasher
	writer  io.Writer
}

type hasher struct {
	sum    string
	writer hash.Hash
	kind   string
}

func makeHashers(file FileHash) hashers {
	list := make([]hasher, 0, 4)
	writers := make([]io.Writer, 0, 4)

	if file.MD5Sum != "" {
		writer := md5.New()
		list = append(list, hasher{sum: file.MD5Sum, writer: writer, kind: "MD5"})
		writers = append(writers, writer)
	}

	if file.SHA1 != "" {
		writer := sha1.New()
		list = append(list, hasher{sum: file.SHA1, writer: writer, kind: "SHA-1"})
		writers = append(writers, writer)
	}

	if file.SHA256 != "" {
		writer := sha256.New()
		list = append(list, hasher{sum: file.SHA256, writer: writer, kind: "SHA-256"})
		writers = append(writers, writer)
	}

	if file.SHA512 != "" {
		writer := sha512.New()
		list = append(list, hasher{sum: file.SHA512, writer: writer, kind: "SHA-512"})
		writers = append(writers, writer)
	}

	return hashers{
		hashers: list,
		writer:  io.MultiWriter(writers...),
	}
}
