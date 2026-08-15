package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"sync"

	"github.com/NatoBoram/ipapm/apt"
	"github.com/NatoBoram/ipapm/config"
	"github.com/NatoBoram/ipapm/env"
	"github.com/NatoBoram/ipapm/kubo"
	"github.com/NatoBoram/ipapm/progress"
	slogctx "github.com/veqryn/slog-context"
	"golang.org/x/sync/errgroup"
)

func syncConfig(ctx context.Context, env env.Env, kubo *kubo.Client, client *apt.Client, pool *progress.Pool) error {
	config, err := config.Load(config.Env{CONFIG_DIR: env.CONFIG_DIR, Name: env.Name})
	if err != nil {
		return fmt.Errorf("couldn't load config: %w", err)
	}
	kubo.MFS = config.Kubo.MFS

	mapped, err := apt.MapConfigs(config.Sources)
	if err != nil {
		return fmt.Errorf("couldn't map sources: %w", err)
	}

	jobs := make(chan apt.Config, len(mapped))
	wg := new(sync.WaitGroup)

	for range min(len(mapped), 4) {
		wg.Go(func() {
			for source := range jobs {
				bar := pool.NewBar(path.Join(source.URI.Host, source.URI.Path))
				err := syncSource(ctx, env, kubo, client, source, bar)
				bar.Finish()

				if err != nil {
					slog.WarnContext(ctx, "Error while syncing source", slog.Any("error", err))
				}
			}
		})
	}

	for _, source := range mapped {
		jobs <- source
	}
	close(jobs)

	wg.Wait()
	return nil
}

func syncSource(
	ctx context.Context, env env.Env,
	kubo *kubo.Client, client *apt.Client, config apt.Config,
	bar *progress.Bar,
) error {
	err := syncSuites(ctx, kubo, client, config, bar)
	if err != nil {
		return fmt.Errorf("error while syncing suites: %w", err)
	}

	// Publish to IPNS
	stat, err := kubo.RepoRoot(ctx, config.URI)
	if err != nil {
		return fmt.Errorf("error while getting suite stat: %w", err)
	}

	keyName := env.Name + ":" + path.Join(config.URI.Hostname(), config.URI.EscapedPath())
	name, err := kubo.NamePublish(ctx, keyName, stat.Hash)
	if err != nil {
		return fmt.Errorf("error while publishing to IPNS: %w", err)
	}
	slog.InfoContext(
		ctx, "Published to IPNS",
		"cid", name.Cid().String(), "key", name.Peer().String(), "name", name.String(),
	)

	return nil
}

func syncSuites(
	ctx context.Context,
	kubo *kubo.Client, client *apt.Client, config apt.Config,
	bar *progress.Bar,
) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(4)

	for suite := range config.Suites {
		g.Go(func() error { return syncSuite(ctx, kubo, client, config, suite, bar) })
	}

	return g.Wait()
}

func syncSuite(
	ctx context.Context,
	kubo *kubo.Client, client *apt.Client, config apt.Config, suite string,
	bar *progress.Bar,
) error {
	ctx = slogctx.Prepend(
		ctx,
		"uri", config.URI.String(),
		"suite", suite,
	)

	// Get InRelease file
	next, err := client.InRelease(ctx, config.URI, suite)
	if err != nil {
		return fmt.Errorf("error while fetching InRelease file: %w", err)
	}
	for _, err := range next.Warnings {
		slog.WarnContext(
			ctx, "Errors while parsing InRelease",
			slog.Any("error", err),
		)
	}
	slog.InfoContext(ctx, "Got InRelease file")

	// Verify signature
	err = verifyPgp(config.SignedBy, string(next.Raw))
	if err != nil {
		return fmt.Errorf("failed to verify PGP signature: %w", err)
	}
	slog.InfoContext(ctx, "Verified PGP signature")

	// Get InRelease from MFS
	previous, err := kubo.InRelease(ctx, config.URI, suite)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("error while fetching InRelease from MFS: %w", err)
	}

	// Sync
	if errors.Is(err, os.ErrNotExist) {
		slog.InfoContext(ctx, "No previous InRelease file in MFS")

		err = syncAll(ctx, kubo, client, config, suite, next, bar)
		if err != nil {
			return fmt.Errorf("error while syncing all files: %w", err)
		}
	} else {
		slog.InfoContext(ctx, "Got previous InRelease file from MFS")

		err = syncDiff(ctx, kubo, client, config, suite, next, previous, bar)
		if err != nil {
			return fmt.Errorf("error while syncing diff: %w", err)
		}
	}

	slog.InfoContext(ctx, "Committing InRelease file")
	err = kubo.WriteInRelease(ctx, config.URI, suite, next)
	if err != nil {
		return fmt.Errorf("error while writing InRelease to MFS: %w", err)
	}

	return nil
}

func upsertSources(
	ctx context.Context, kubo *kubo.Client, client *apt.Client,
	source apt.Config, sources apt.Sources,
	bar *progress.Bar,
) error {
	for _, s := range sources {
		ctx := slogctx.Prepend(
			ctx,
			"package", s.Package,
			"version", s.Version,
		)

		files, err := s.FileHashes()
		if err != nil {
			return fmt.Errorf("couldn't get files from Source: %w", err)
		}

		bar.AddTotal(len(files))
		for _, f := range files {
			ctx := slogctx.Prepend(
				ctx,
				"filename", f.Filename,
			)

			slog.InfoContext(ctx, "Streaming source...")

			r, err := client.StreamSource(ctx, source.URI, f)
			if err != nil {
				return fmt.Errorf("error while fetching source %s: %w", f.Filename, err)
			}

			err = kubo.WriteSource(ctx, source.URI, f, r)
			r.Close()
			if err != nil {
				return fmt.Errorf("error while writing %s to MFS: %w", f.Filename, err)
			}

			bar.Increment()
		}
	}

	return nil
}

func upsertPackages(
	ctx context.Context, kubo *kubo.Client, client *apt.Client,
	source apt.Config, packages apt.Packages,
	bar *progress.Bar,
) error {
	bar.AddTotal(len(packages))

	for _, p := range packages {
		ctx := slogctx.Prepend(
			ctx,
			"package", p.Package,
			"version", p.Version,
		)

		slog.InfoContext(ctx, "Streaming package...")

		r, err := client.StreamPackage(ctx, source.URI, p.FileHash())
		if err != nil {
			return fmt.Errorf("error while fetching package %s: %w", p.Filename, err)
		}

		err = kubo.WritePackage(ctx, source.URI, p.FileHash(), r)
		r.Close()
		if err != nil {
			return fmt.Errorf("error while writing %s to MFS: %w", p.Filename, err)
		}

		bar.Increment()
	}

	return nil
}
