package apt

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (c *Client) Package(ctx context.Context, uri *url.URL, file FileHash) (io.ReadCloser, error) {
	target := uri.JoinPath(file.Filename)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request for %s: %w", target, err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", target, err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, target)
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
			return n, fmt.Errorf("size exceeded for %s: expected max %d bytes, got at least %d", r.file.Filename, r.file.Size, r.read)
		}

		if _, werr := r.hashers.writer.Write(p[:n]); werr != nil {
			return n, fmt.Errorf("failed to write stream to hashers: %w", werr)
		}
	}

	if err == io.EOF && !r.verified {
		r.verified = true

		if r.file.Size > 0 && r.read != r.file.Size {
			return n, fmt.Errorf("size mismatch for %s: expected %d, got %d", r.file.Filename, r.file.Size, r.read)
		}

		for _, h := range r.hashers.hashers {
			checksum := hex.EncodeToString(h.writer.Sum(nil))
			if checksum != h.sum {
				return n, fmt.Errorf("%s mismatch for %s: expected %s, got %s", h.kind, r.file.Filename, h.sum, checksum)
			}
		}
	}

	return n, err
}

func (r *hashReader) Close() error {
	return r.r.Close()
}
