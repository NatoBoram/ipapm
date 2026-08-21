package kubo

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"net/url"
	"path"
)

type FilesStat struct {
	Blocks         int    `json:"Blocks"`
	CumulativeSize uint64 `json:"CumulativeSize"`
	Hash           string `json:"Hash"`
	Local          bool   `json:"Local"`
	Mode           uint32 `json:"Mode"`
	Mtime          int64  `json:"Mtime"`
	MtimeNsecs     int    `json:"MtimeNsecs"`
	Size           uint64 `json:"Size"`
	SizeLocal      uint64 `json:"SizeLocal"`
	Type           string `json:"Type"`
	WithLocality   bool   `json:"WithLocality"`
}

func (k *Client) RepoRoot(ctx context.Context, uri *url.URL) (FilesStat, error) {
	target := path.Join(k.MFS, uri.Hostname(), uri.EscapedPath())
	if !path.IsAbs(target) {
		target = "/" + target
	}

	return k.filesStat(ctx, target)
}

// filesStat gets the [FilesStat] of a path in MFS.
func (k *Client) filesStat(ctx context.Context, target string) (FilesStat, error) {
	resp, err := k.Request("files/stat", target).Send(ctx)
	if err != nil {
		return FilesStat{}, fmt.Errorf("requesting files/stat %q: %w", target, err)
	}
	defer resp.Close()

	if resp.Error != nil {
		return FilesStat{}, fmt.Errorf("kubo error (%d) %q", resp.Error.Code, resp.Error.Message)
	}

	var out FilesStat
	if err := json.UnmarshalRead(resp.Output, &out); err != nil {
		return FilesStat{}, fmt.Errorf("parsing files/stat: %w", err)
	}

	return out, nil
}
