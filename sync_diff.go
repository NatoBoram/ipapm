package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"

	"github.com/NatoBoram/ipapm/apt"
	"github.com/NatoBoram/ipapm/http"
	"github.com/NatoBoram/ipapm/kubo"
	"github.com/NatoBoram/ipapm/progress"
	"github.com/NatoBoram/ipapm/wheel"
	slogctx "github.com/veqryn/slog-context"
)

func syncDiff(
	ctx context.Context, kubo *kubo.Client, client *apt.Client,
	config apt.Config, suite string, next, previous apt.InRelease,
	bar *progress.Bar,
) error {
	nfiles, err := next.FileHashes()
	if err != nil {
		return fmt.Errorf("getting files from next InRelease: %w", err)
	}
	ncomponents := next.ByComponents(nfiles)

	pfiles, err := previous.FileHashes()
	if err != nil {
		return fmt.Errorf("getting files from previous InRelease: %w", err)
	}
	pcomponents := previous.ByComponents(pfiles)

	diffComponents := pcomponents.Diff(ncomponents)
	slog.InfoContext(
		ctx, "Components diff",
		"added", len(diffComponents.Added), "changed", len(diffComponents.Changed),
		"removed", len(diffComponents.Removed),
	)
	upsertComponents := slices.Concat(diffComponents.Added, diffComponents.Changed)

	for _, component := range upsertComponents {
		ctx := slogctx.Prepend(
			ctx,
			"component", component.Name,
			"architecture", component.Architecture,
		)

		if component.Architecture == "source" {
			err := syncSources(ctx, kubo, client, config, suite, component, bar)
			if err != nil {
				return fmt.Errorf("syncing sources: %w", err)
			}
		} else {
			err := syncPackages(ctx, kubo, client, config, suite, component, bar)
			if err != nil {
				return fmt.Errorf("syncing packages: %w", err)
			}
		}
	}

	// Remove components
	for _, component := range diffComponents.Removed {
		ctx := slogctx.Prepend(
			ctx,
			"component", component.Name,
			"architecture", component.Architecture,
		)

		slog.InfoContext(ctx, "Removing component from MFS")
		err = kubo.RemoveComponent(ctx, config.URI, suite, component)
		if err != nil {
			return fmt.Errorf("removing %s/%s from MFS: %w", component.Name, component.Architecture, err)
		}
	}

	diff := pfiles.Diff(nfiles)
	bar.AddTotal(len(diff.Added) + len(diff.Changed) + len(diff.Removed))

	// Upsert files
	upsert := wheel.MergeMaps(diff.Added, diff.Changed)
	for _, f := range upsert {
		ctx := slogctx.Prepend(ctx, "file", f.Filename)

		slog.InfoContext(ctx, "Committing file")

		r, err := client.StreamFile(ctx, config.URI, suite, f)
		if err != nil {
			if errors.Is(err, http.ErrNotFound) {
				slog.WarnContext(ctx, "File not found, skipping", "error", err)

				err := kubo.RemoveFile(ctx, config.URI, suite, f)
				if err != nil {
					return fmt.Errorf("removing %q from MFS: %w", f.Filename, err)
				}

				bar.Increment()
				continue
			}

			return fmt.Errorf("fetching file %q: %w", f.Filename, err)
		}

		err = kubo.WriteFile(ctx, config.URI, suite, f, r)
		r.Close()
		if err != nil {
			return fmt.Errorf("writing %q to MFS: %w", f.Filename, err)
		}

		bar.Increment()
	}

	// Remove files
	for _, f := range diff.Removed {
		ctx := slogctx.Prepend(ctx, "file", f.Filename)

		slog.InfoContext(ctx, "Removing file from MFS")
		err = kubo.RemoveFile(ctx, config.URI, suite, f)
		if err != nil {
			return fmt.Errorf("removing %q from MFS: %w", f.Filename, err)
		}

		bar.Increment()
	}

	return nil
}

func syncSources(
	ctx context.Context, kubo *kubo.Client, client *apt.Client,
	config apt.Config, suite string, component apt.Component,
	bar *progress.Bar,
) error {
	sources, err := client.Sources(ctx, config.URI, suite, component)
	if err != nil {
		return fmt.Errorf("downloading Sources: %w", err)
	}

	slog.InfoContext(
		ctx, "Got next Sources",
		"sources", len(sources.Sources),
		"files", len(sources.FileBytes),
	)

	psources, err := kubo.Sources(ctx, config.URI, suite, component)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("fetching previous Sources: %w", err)
	}

	slog.InfoContext(
		ctx, "Got previous sources",
		"sources", len(psources.Sources), "files", len(psources.FileBytes),
	)

	diff, err := psources.Sources.Diff(sources.Sources)
	if err != nil {
		return fmt.Errorf("diffing sources: %w", err)
	}
	slog.InfoContext(
		ctx, "Sources diff",
		"added", len(diff.Added), "changed", len(diff.Changed),
		"removed", len(diff.Removed),
	)

	// Upsert sources
	upsert := slices.Concat(diff.Added, diff.Changed)
	err = upsertSources(ctx, kubo, client, config, upsert, bar)
	if err != nil {
		return fmt.Errorf("upserting sources: %w", err)
	}

	// Remove sources
	for _, removed := range diff.Removed {
		ctx := slogctx.Prepend(
			ctx,
			"package", removed.Package,
			"version", removed.Version,
		)

		slog.InfoContext(ctx, "Removing source from MFS")
		fhs, err := removed.FileHashes()
		if err != nil {
			return fmt.Errorf("getting file hashes for source %s %s: %w", removed.Package, removed.Version, err)
		}

		for _, fh := range fhs {
			err = kubo.RemoveSource(ctx, config.URI, fh)
			if err != nil {
				return fmt.Errorf("removing source %q from MFS: %w", fh.Filename, err)
			}
		}
	}

	srcDiff := psources.FileBytes.Diff(sources.FileBytes)

	// Upsert Sources
	upsertSrc := slices.Concat(srcDiff.Added, srcDiff.Changed)
	for _, f := range upsertSrc {
		ctx := slogctx.Prepend(ctx, "Sources", f.Hashes.Filename)

		slog.InfoContext(ctx, "Committing Sources file")

		err = kubo.WriteSources(ctx, config.URI, suite, f)
		if err != nil {
			return fmt.Errorf("writing %q to MFS: %w", f.Hashes.Filename, err)
		}
	}

	// Remove Sources
	for _, f := range srcDiff.Removed {
		ctx := slogctx.Prepend(ctx, "Sources", f.Hashes.Filename)

		slog.InfoContext(ctx, "Removing Sources file from MFS")
		err = kubo.RemoveSources(ctx, config.URI, suite, f)
		if err != nil {
			return fmt.Errorf("removing %q from MFS: %w", f.Hashes.Filename, err)
		}
	}

	return nil
}

func syncPackages(
	ctx context.Context, kubo *kubo.Client, client *apt.Client,
	config apt.Config, suite string, component apt.Component,
	bar *progress.Bar,
) error {
	slog.InfoContext(ctx, "Getting Packages files")
	npackages, err := client.Packages(ctx, config.URI, suite, component)
	if err != nil {
		return fmt.Errorf("fetching Packages: %w", err)
	}

	for _, p := range npackages.Packages {
		for _, err := range p.Warnings {
			slog.WarnContext(ctx, "Warning while parsing packages", "error", err)
		}
	}

	slog.InfoContext(
		ctx, "Got next Packages",
		"packages", len(npackages.Packages),
		"files", len(npackages.FileBytes),
	)

	ppackages, err := kubo.Packages(ctx, config.URI, suite, component)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("fetching previous Packages: %w", err)
	}

	slog.InfoContext(
		ctx, "Got previous packages",
		"packages", len(ppackages.Packages), "files", len(ppackages.FileBytes),
	)

	diff := ppackages.Packages.Diff(npackages.Packages)
	slog.InfoContext(
		ctx, "Packages diff",
		"added", len(diff.Added), "changed", len(diff.Changed),
		"removed", len(diff.Removed),
	)

	// Upsert packages
	upsert := slices.Concat(diff.Added, diff.Changed)
	err = upsertPackages(ctx, kubo, client, config, upsert, bar)
	if err != nil {
		return fmt.Errorf("upserting packages: %w", err)
	}

	// Remove packages
	for _, removed := range diff.Removed {
		ctx := slogctx.Prepend(
			ctx,
			"package", removed.Package,
			"version", removed.Version,
		)

		slog.InfoContext(ctx, "Removing package from MFS")
		err = kubo.RemovePackage(ctx, config.URI, removed)
		if err != nil {
			return fmt.Errorf("removing %q from MFS: %w", removed.Filename, err)
		}
	}

	pkgDiff := ppackages.FileBytes.Diff(npackages.FileBytes)

	// Upsert Packages
	pkgUpsert := slices.Concat(pkgDiff.Added, pkgDiff.Changed)
	for _, f := range pkgUpsert {
		ctx := slogctx.Prepend(ctx, "Packages", f.Hashes.Filename)

		slog.InfoContext(ctx, "Committing Packages file")

		err = kubo.WritePackages(ctx, config.URI, suite, f)
		if err != nil {
			return fmt.Errorf("writing %q to MFS: %w", f.Hashes.Filename, err)
		}
	}

	// Remove Packages
	for _, f := range pkgDiff.Removed {
		ctx := slogctx.Prepend(ctx, "Packages", f.Hashes.Filename)

		slog.InfoContext(ctx, "Removing Packages file from MFS")
		err = kubo.RemovePackages(ctx, config.URI, suite, f)
		if err != nil {
			return fmt.Errorf("removing %q from MFS: %w", f.Hashes.Filename, err)
		}
	}

	return nil
}
