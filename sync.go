package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/NatoBoram/ipapm/apt"
	"github.com/NatoBoram/ipapm/config"
	"github.com/NatoBoram/ipapm/env"
	"github.com/NatoBoram/ipapm/kubo"
	slogctx "github.com/veqryn/slog-context"
)

func sync(ctx context.Context, env env.Env, kubo *kubo.Client, client *apt.Client) error {
	config, err := config.Load(config.Env{CONFIG_DIR: env.CONFIG_DIR})
	if err != nil {
		return fmt.Errorf("couldn't load config: %w", err)
	}
	kubo.MFS = config.Kubo.MFS

	mapped, err := apt.MapSources(config.Sources)
	if err != nil {
		return fmt.Errorf("couldn't map sources: %w", err)
	}

	for _, source := range mapped {
		for suite := range source.Suites {
			ctx = slogctx.Prepend(
				ctx,
				"source", source.URI.String(),
				"suite", suite,
			)

			// Get InRelease file
			next, err := client.InRelease(ctx, source.URI, suite)
			if err != nil {
				slog.ErrorContext(
					ctx, "Error while fetching InRelease",
					slog.Any("error", err),
				)
				continue
			}
			for _, err := range next.Warnings {
				slog.WarnContext(
					ctx, "Errors while parsing InRelease",
					slog.Any("error", err),
				)
			}
			slog.InfoContext(ctx, "Got InRelease file")

			// Verify signature
			err = verifyPgp(source.SignedBy, next.Raw)
			if err != nil {
				slog.ErrorContext(
					ctx, "Failed to verify PGP signature",
					slog.Any("error", err),
				)
				continue
			}
			slog.InfoContext(ctx, "Verified PGP signature")

			// Get InRelease from MFS
			previous, err := kubo.InRelease(ctx, source.URI, suite)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				slog.ErrorContext(
					ctx, "Error while fetching InRelease from MFS",
					slog.Any("error", err),
				)
				continue
			}

			if errors.Is(err, os.ErrNotExist) {
				slog.InfoContext(ctx, "No previous InRelease file in MFS")
				err := syncAll(ctx, kubo, client, source, suite, next)
				if err != nil {
					slog.ErrorContext(
						ctx, "Error while syncing everything",
						slog.Any("error", err),
					)
				}
				continue
			}

			slog.InfoContext(ctx, "Got previous InRelease file from MFS")
			err = syncDiff(ctx, kubo, client, source, suite, next, previous)
			if err != nil {
				slog.ErrorContext(
					ctx, "Error while syncing diff",
					slog.Any("error", err),
				)
				continue
			}
		}
	}

	return nil
}

func syncAll(
	ctx context.Context, kubo *kubo.Client, client *apt.Client,
	source apt.Source, suite string, next apt.InRelease,
) error {
	slog.InfoContext(ctx, "All")
	return nil
}

func syncDiff(
	ctx context.Context, kubo *kubo.Client, client *apt.Client,
	source apt.Source, suite string, next apt.InRelease,
	previous apt.InRelease,
) error {
	diff, err := previous.Diff(next)
	if err != nil {
		return fmt.Errorf("couldn't get the diff between previous and next InRelease files: %s", err)
	}

	slog.InfoContext(
		ctx, "Diff",
		slog.Int("added", len(diff.Added)),
		slog.Int("changed", len(diff.Changed)),
		slog.Int("removed", len(diff.Removed)),
	)

	return nil
}
