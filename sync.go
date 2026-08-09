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
			ctx := slogctx.Prepend(
				ctx,
				"uri", source.URI.String(),
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
			err = verifyPgp(source.SignedBy, string(next.Raw))
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
	slog.InfoContext(ctx, "Syncing all files")

	files, err := next.Files()
	if err != nil {
		return fmt.Errorf("couldn't get files from InRelease: %w", err)
	}

	components := next.ByComponents(files)
	for _, component := range components {
		ctx := slogctx.Prepend(
			ctx,
			"component", component.Name,
			"architecture", component.Architecture,
		)

		if component.Architecture == "source" {
			// Implement `client.Sources(component.Name)`
			slog.DebugContext(ctx, "Source")
			return errors.New("source not implemented")
		}

		slog.InfoContext(ctx, "Getting Packages files")
		packages, err := client.Packages(ctx, source.URI, suite, component)
		if err != nil {
			return fmt.Errorf("error while fetching Packages: %w", err)
		}

		for _, p := range packages.Packages {
			for _, err := range p.Warnings {
				slog.WarnContext(ctx, "Warning while parsing packages", "error", err)
			}
		}

		slog.InfoContext(
			ctx, "Got Packages",
			"packages", len(packages.Packages), "files", len(packages.FileBytes),
		)

		for _, p := range packages.Packages {
			ctx := slogctx.Prepend(
				ctx,
				"package", p.Package,
				"version", p.Version,
			)

			slog.InfoContext(ctx, "Streaming package...")

			r, err := client.Package(ctx, source.URI, p.FileHash())
			if err != nil {
				return fmt.Errorf("error while fetching package %s: %w", p.Filename, err)
			}

			err = kubo.WritePackage(ctx, source.URI, suite, p.FileHash(), r)
			if err != nil {
				return fmt.Errorf("error while writing %s to MFS: %w", p.Filename, err)
			}
		}

		for _, f := range packages.FileBytes {
			ctx := slogctx.Prepend(
				ctx,
				"Packages", f.Hashes.Filename,
			)

			slog.InfoContext(ctx, "Committing Packages file")

			err = kubo.WritePackages(ctx, source.URI, suite, f)
			if err != nil {
				return fmt.Errorf("error while writing %s to MFS: %w", f.Hashes.Filename, err)
			}
		}
	}

	slog.InfoContext(ctx, "Committing InRelease file")
	err = kubo.WriteInRelease(ctx, source.URI, suite, next)
	if err != nil {
		return fmt.Errorf("error while writing InRelease to MFS: %w", err)
	}

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

	slog.DebugContext(
		ctx, "Diff",
		slog.Int("added", len(diff.Added)),
		slog.Int("changed", len(diff.Changed)),
		slog.Int("removed", len(diff.Removed)),
	)

	components := next.ByComponents(diff.Added)
	for _, component := range components {
		ctx := slogctx.Prepend(
			ctx,
			"component", component.Name,
			"architecture", component.Architecture,
		)

		// Get the `Packages` and `Sources` files and add the entire thing.
		slog.DebugContext(ctx, "Added")
	}

	components = next.ByComponents(diff.Changed)
	for _, component := range components {
		ctx := slogctx.Prepend(
			ctx,
			"component", component.Name,
			"architecture", component.Architecture,
		)

		// Get the `Packages` and `Sources`, check the diff then
		// create/update/delete files as needed.
		slog.DebugContext(ctx, "Changed")
	}

	// For removed components and architectures, use `previous` to find all the
	// files that need to be removed.

	return nil
}
