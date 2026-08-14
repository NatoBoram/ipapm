package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/NatoBoram/ipapm/apt"
	"github.com/NatoBoram/ipapm/config"
	"github.com/NatoBoram/ipapm/env"
	"github.com/NatoBoram/ipapm/kubo"
	slogctx "github.com/veqryn/slog-context"
)

func sync(ctx context.Context, env env.Env, kubo *kubo.Client, client *apt.Client) error {
	config, err := config.Load(config.Env{CONFIG_DIR: env.CONFIG_DIR, Name: env.Name})
	if err != nil {
		return fmt.Errorf("couldn't load config: %w", err)
	}
	kubo.MFS = config.Kubo.MFS

	mapped, err := apt.MapConfigs(config.Sources)
	if err != nil {
		return fmt.Errorf("couldn't map sources: %w", err)
	}

	for _, source := range mapped {
		err := syncSuites(ctx, env, kubo, client, source)
		if err != nil {
			slog.ErrorContext(ctx, "Error while syncing suites", slog.Any("error", err))
			continue
		}

		// Publish to IPNS
		stat, err := kubo.RepoRoot(ctx, source.URI)
		if err != nil {
			slog.ErrorContext(
				ctx, "Error while getting suite stat",
				slog.Any("error", err),
			)
			continue
		}

		keyName := env.Name + ":" + path.Join(source.URI.Hostname(), source.URI.EscapedPath())
		name, err := kubo.NamePublish(ctx, keyName, stat.Hash)
		if err != nil {
			slog.ErrorContext(
				ctx, "Error while publishing to IPNS",
				slog.Any("error", err),
			)
			continue
		}
		slog.InfoContext(
			ctx, "Published to IPNS",
			"cid", name.Cid().String(), "key", name.Peer().String(), "name", name.String(),
		)
	}

	return nil
}

func syncSuites(ctx context.Context, env env.Env, kubo *kubo.Client, client *apt.Client, source apt.Config) error {
	for suite := range source.Suites {
		ctx := slogctx.Prepend(
			ctx,
			"uri", source.URI.String(),
			"suite", suite,
		)

		// Get InRelease file
		next, err := client.InRelease(ctx, source.URI, suite)
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
		err = verifyPgp(source.SignedBy, string(next.Raw))
		if err != nil {
			return fmt.Errorf("failed to verify PGP signature: %w", err)
		}
		slog.InfoContext(ctx, "Verified PGP signature")

		// Get InRelease from MFS
		previous, err := kubo.InRelease(ctx, source.URI, suite)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("error while fetching InRelease from MFS: %w", err)
		}

		// Sync
		if errors.Is(err, os.ErrNotExist) {
			slog.InfoContext(ctx, "No previous InRelease file in MFS")

			err = syncAll(ctx, kubo, client, source, suite, next)
			if err != nil {
				return fmt.Errorf("error while syncing all files: %w", err)
			}

			continue
		} else {
			slog.InfoContext(ctx, "Got previous InRelease file from MFS")

			err = syncDiff(ctx, kubo, client, source, suite, next, previous)
			if err != nil {
				return fmt.Errorf("error while syncing diff: %w", err)
			}

			continue
		}
	}

	return nil
}

func upsertSources(
	ctx context.Context, kubo *kubo.Client, client *apt.Client,
	source apt.Config, sources apt.Sources,
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

		for _, f := range files {
			ctx := slogctx.Prepend(
				ctx,
				"filename", f.Filename,
			)

			slog.InfoContext(ctx, "Streaming source...")

			r, err := client.Stream(ctx, source.URI, f)
			if err != nil {
				return fmt.Errorf("error while fetching source %s: %w", f.Filename, err)
			}

			err = kubo.WriteSource(ctx, source.URI, f, r)
			r.Close()
			if err != nil {
				return fmt.Errorf("error while writing %s to MFS: %w", f.Filename, err)
			}
		}
	}

	return nil
}

func upsertPackages(
	ctx context.Context, kubo *kubo.Client, client *apt.Client,
	source apt.Config, packages apt.Packages,
) error {
	for _, p := range packages {
		ctx := slogctx.Prepend(
			ctx,
			"package", p.Package,
			"version", p.Version,
		)

		slog.InfoContext(ctx, "Streaming package...")

		r, err := client.Stream(ctx, source.URI, p.FileHash())
		if err != nil {
			return fmt.Errorf("error while fetching package %s: %w", p.Filename, err)
		}

		err = kubo.WritePackage(ctx, source.URI, p.FileHash(), r)
		r.Close()
		if err != nil {
			return fmt.Errorf("error while writing %s to MFS: %w", p.Filename, err)
		}
	}

	return nil
}
