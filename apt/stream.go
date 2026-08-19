package apt

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"

	h "github.com/NatoBoram/ipapm/http"
)

func (c *Client) StreamFile(ctx context.Context, uri *url.URL, suite string, file FileHash) (io.ReadCloser, error) {
	target := uri.JoinPath("dists", suite, file.Filename)
	return c.stream(ctx, file, target)
}

func (c *Client) StreamSource(ctx context.Context, uri *url.URL, file FileHash) (io.ReadCloser, error) {
	target := uri.JoinPath(file.Filename)
	return c.stream(ctx, file, target)
}

func (c *Client) StreamPackage(ctx context.Context, uri *url.URL, file FileHash) (io.ReadCloser, error) {
	target := uri.JoinPath(file.Filename)
	return c.stream(ctx, file, target)
}

func (c *Client) stream(ctx context.Context, file FileHash, target *url.URL) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("building request for %q: %w", target, err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching %q: %w", target, err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("unexpected status for %q: %w", target, h.ErrNotFound)
		}

		return nil, fmt.Errorf("unexpected status %s for %q", resp.Status, target)
	}

	return newHashReader(resp.Body, file), nil
}

func newHashReader(r io.ReadCloser, file FileHash) io.ReadCloser {
	return new(hashReader{file: file, hashers: makeHashers(file), r: r})
}

type hashReader struct {
	r        io.ReadCloser
	file     FileHash
	hashers  hashers
	read     uint
	verified bool
}

func (r *hashReader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	if n > 0 {
		r.read += uint(n)

		if r.file.Size > 0 && r.read > r.file.Size {
			return n, fmt.Errorf("size exceeded for %q: expected max %d bytes, got at least %d", r.file.Filename, r.file.Size, r.read)
		}

		if _, werr := r.hashers.writer.Write(p[:n]); werr != nil {
			return n, fmt.Errorf("writing stream to hashers: %w", werr)
		}
	}

	if err == io.EOF && !r.verified {
		r.verified = true

		if r.file.Size > 0 && r.read != r.file.Size {
			return n, fmt.Errorf("size mismatch for %q: expected %d, got %d", r.file.Filename, r.file.Size, r.read)
		}

		for _, h := range r.hashers.hashers {
			checksum := hex.EncodeToString(h.writer.Sum(nil))
			if checksum != h.sum {
				return n, fmt.Errorf("%s mismatch for %q: expected %s, got %s", h.kind, r.file.Filename, h.sum, checksum)
			}
		}
	}

	return n, err
}

func (r *hashReader) Close() error {
	return r.r.Close()
}
