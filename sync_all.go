package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/NatoBoram/ipapm/apt"
	"github.com/NatoBoram/ipapm/kubo"
	slogctx "github.com/veqryn/slog-context"
)

func syncAll(
	ctx context.Context, kubo *kubo.Client, client *apt.Client,
	config apt.Config, suite string, next apt.InRelease,
) error {
	slog.InfoContext(ctx, "Syncing all files")

	files, err := next.FileHashes()
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
			err := syncAllSources(ctx, kubo, client, config, suite, component)
			if err != nil {
				return fmt.Errorf("error while syncing sources: %w", err)
			}
		} else {
			err := syncAllPackages(ctx, kubo, client, config, suite, component)
			if err != nil {
				return fmt.Errorf("error while syncing packages: %w", err)
			}
		}
	}

	slog.InfoContext(ctx, "Committing InRelease file")
	err = kubo.WriteInRelease(ctx, config.URI, suite, next)
	if err != nil {
		return fmt.Errorf("error while writing InRelease to MFS: %w", err)
	}

	return nil
}

func syncAllSources(
	ctx context.Context, kubo *kubo.Client, client *apt.Client,
	config apt.Config, suite string,
	component apt.Component,
) error {
	slog.DebugContext(ctx, "Source")
	sources, err := client.Sources(ctx, config.URI, suite, component)
	if err != nil {
		return fmt.Errorf("couldn't download Sources file: %w", err)
	}

	slog.InfoContext(
		ctx, "Got Sources files",
		"sources", len(sources.Sources),
		"files", len(sources.FileBytes),
	)

	err = upsertSources(ctx, kubo, client, config, sources.Sources)
	if err != nil {
		return fmt.Errorf("error while upserting sources: %w", err)
	}

	// Commit sources. Since there's no previous Sources file, we commit
	// everything.
	for _, f := range sources.FileBytes {
		ctx := slogctx.Prepend(
			ctx,
			"Sources", f.Hashes.Filename,
		)

		slog.InfoContext(ctx, "Committing Sources file")

		err = kubo.WriteSources(ctx, config.URI, suite, f)
		if err != nil {
			return fmt.Errorf("error while writing %s to MFS: %w", f.Hashes.Filename, err)
		}
	}

	return nil
}

func syncAllPackages(
	ctx context.Context, kubo *kubo.Client, client *apt.Client,
	config apt.Config, suite string,
	component apt.Component,
) error {
	slog.InfoContext(ctx, "Getting Packages files")
	packages, err := client.Packages(ctx, config.URI, suite, component)
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
		"packages", len(packages.Packages),
		"files", len(packages.FileBytes),
	)

	err = upsertPackages(ctx, kubo, client, config, packages.Packages)
	if err != nil {
		return fmt.Errorf("error while upserting packages: %w", err)
	}

	// Commit packages. Since there's no previous Packages file, we commit
	// everything.
	for _, f := range packages.FileBytes {
		ctx := slogctx.Prepend(
			ctx,
			"Packages", f.Hashes.Filename,
		)

		slog.InfoContext(ctx, "Committing Packages file")

		err = kubo.WritePackages(ctx, config.URI, suite, f)
		if err != nil {
			return fmt.Errorf("error while writing %s to MFS: %w", f.Hashes.Filename, err)
		}
	}

	return nil
}
