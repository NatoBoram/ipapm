package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/NatoBoram/ipapm/apt"
	"github.com/NatoBoram/ipapm/kubo"
	"github.com/NatoBoram/ipapm/progress"
	slogctx "github.com/veqryn/slog-context"
)

func syncAll(
	ctx context.Context, kubo *kubo.Client, client *apt.Client,
	config apt.Config, suite string, next apt.InRelease,
	bar *progress.Bar,
) error {
	slog.InfoContext(ctx, "Syncing all files")

	files, err := next.FileHashes()
	if err != nil {
		return fmt.Errorf("getting files from InRelease: %w", err)
	}

	components := next.ByComponents(files)
	for _, component := range components {
		ctx := slogctx.Prepend(
			ctx,
			"component", component.Name,
			"architecture", component.Architecture,
		)

		if component.Architecture == "source" {
			err := syncAllSources(ctx, kubo, client, config, suite, component, bar)
			if err != nil {
				return fmt.Errorf("syncing sources: %w", err)
			}
		} else {
			err := syncAllPackages(ctx, kubo, client, config, suite, component, bar)
			if err != nil {
				return fmt.Errorf("syncing packages: %w", err)
			}
		}
	}

	bar.AddTotal(len(files))
	for _, file := range files {
		ctx := slogctx.Prepend(
			ctx,
			"file", file.Filename,
		)

		slog.InfoContext(ctx, "Committing file")

		r, err := client.StreamFile(ctx, config.URI, suite, file)
		if err != nil {
			return fmt.Errorf("fetching file %q: %w", file.Filename, err)
		}

		err = kubo.WriteFile(ctx, config.URI, suite, file, r)
		r.Close()
		if err != nil {
			return fmt.Errorf("writing %q to MFS: %w", file.Filename, err)
		}

		bar.Increment()
	}

	return nil
}

func syncAllSources(
	ctx context.Context, kubo *kubo.Client, client *apt.Client,
	config apt.Config, suite string, component apt.Component,
	bar *progress.Bar,
) error {
	slog.DebugContext(ctx, "Source")
	sources, err := client.Sources(ctx, config.URI, suite, component)
	if err != nil {
		return fmt.Errorf("downloading Sources file: %w", err)
	}

	slog.InfoContext(
		ctx, "Got Sources files",
		"sources", len(sources.Sources),
		"files", len(sources.FileBytes),
	)

	err = upsertSources(ctx, kubo, client, config, sources.Sources, bar)
	if err != nil {
		return fmt.Errorf("upserting sources: %w", err)
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
			return fmt.Errorf("writing %q to MFS: %w", f.Hashes.Filename, err)
		}
	}

	return nil
}

func syncAllPackages(
	ctx context.Context, kubo *kubo.Client, client *apt.Client,
	config apt.Config, suite string, component apt.Component,
	bar *progress.Bar,
) error {
	slog.InfoContext(ctx, "Getting Packages files")
	packages, err := client.Packages(ctx, config.URI, suite, component)
	if err != nil {
		return fmt.Errorf("fetching Packages: %w", err)
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

	err = upsertPackages(ctx, kubo, client, config, packages.Packages, bar)
	if err != nil {
		return fmt.Errorf("upserting packages: %w", err)
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
			return fmt.Errorf("writing %q to MFS: %w", f.Hashes.Filename, err)
		}
	}

	return nil
}
