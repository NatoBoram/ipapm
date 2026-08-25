package kubo

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"net/url"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/ipfs/boxo/mfs"
	"golang.org/x/sync/errgroup"
)

type FilesLs struct {
	Entries []FilesLsEntry `json:"Entries"`
}

type FilesLsEntry struct {
	Hash string       `json:"Hash"`
	Name string       `json:"Name"`
	Size int64        `json:"Size"`
	Type mfs.NodeType `json:"Type"`

	Parent   string
	Filename string
}

func (e FilesLsEntry) Path() string {
	return path.Join(e.Parent, e.Name)
}

func (k *Client) Walk(ctx context.Context, uri *url.URL) ([]FilesLsEntry, error) {
	target := path.Join(k.MFS, uri.Host, uri.EscapedPath())

	dirs := new([]string{target})
	files := new([]FilesLsEntry)
	mu := new(sync.Mutex)
	cond := sync.NewCond(mu)
	pending := new(1)

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(runtime.NumCPU())

	stop := context.AfterFunc(ctx, func() {
		mu.Lock()
		cond.Broadcast()
		mu.Unlock()
	})
	defer stop()

	for range runtime.NumCPU() {
		g.Go(func() error { return k.filesWalk(ctx, mu, cond, pending, dirs, files, target) })
	}

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("walking files: %w", err)
	}

	return *files, nil
}

func (k *Client) filesWalk(
	ctx context.Context, mu *sync.Mutex, cond *sync.Cond, pending *int,
	dirs *[]string, files *[]FilesLsEntry,
	prefix string,
) error {
	for {
		mu.Lock()
		for ctx.Err() == nil && len(*dirs) == 0 && *pending > 0 {
			cond.Wait()
		}

		if err := ctx.Err(); err != nil || (len(*dirs) == 0 && *pending == 0) {
			cond.Broadcast()
			mu.Unlock()
			return err
		}

		dir := (*dirs)[0]
		*dirs = (*dirs)[1:]
		mu.Unlock()

		ls, err := k.filesLs(ctx, dir, prefix)
		if err != nil {
			mu.Lock()
			*pending--
			cond.Broadcast()
			mu.Unlock()
			return fmt.Errorf("listing files: %w", err)
		}

		mu.Lock()
		for _, entry := range ls.Entries {
			switch entry.Type {
			case mfs.TDir:
				*dirs = append(*dirs, entry.Path())
				*pending++
			case mfs.TFile:
				*files = append(*files, entry)
			}
		}

		*pending--
		cond.Broadcast()
		mu.Unlock()
	}
}

func (k *Client) filesLs(ctx context.Context, target string, prefix string) (FilesLs, error) {
	resp, err := k.Request("files/ls", target).
		Option("long", true).
		Option("U", true).
		Send(ctx)
	if err != nil {
		return FilesLs{}, fmt.Errorf("listing files at %q in MFS: %w", target, err)
	}
	defer resp.Close()

	if resp.Error != nil {
		return FilesLs{}, errorf("kubo error", resp.Error)
	}

	var out FilesLs
	if err := json.UnmarshalRead(resp.Output, &out); err != nil {
		return FilesLs{}, fmt.Errorf("parsing files/ls: %w", err)
	}

	for i, entry := range out.Entries {
		entry.Parent = target
		entry.Filename = strings.TrimPrefix(path.Join(target, entry.Name), prefix+"/")
		out.Entries[i] = entry
	}

	return out, nil
}
