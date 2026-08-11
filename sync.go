package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"

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

	mapped, err := apt.MapConfigs(config.Sources)
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

				err = syncAll(ctx, kubo, client, source, suite, next)
				if err != nil {
					slog.ErrorContext(
						ctx, "Error while syncing all files",
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
	source apt.Config, suite string, next apt.InRelease,
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

		err = upsertPackages(ctx, kubo, client, source, suite, packages.Packages)
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
	source apt.Config, suite string, next, previous apt.InRelease,
) error {
	nfiles, err := next.Files()
	if err != nil {
		return fmt.Errorf("couldn't get files from next InRelease: %w", err)
	}
	ncomponents := next.ByComponents(nfiles)

	pfiles, err := previous.Files()
	if err != nil {
		return fmt.Errorf("couldn't get files from previous InRelease: %w", err)
	}
	pcomponents := previous.ByComponents(pfiles)

	diffComponents := pcomponents.Diff(ncomponents)
	upsertComponents := slices.Concat(diffComponents.Added, diffComponents.Changed)

	for _, component := range upsertComponents {
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
		npackages, err := client.Packages(ctx, source.URI, suite, component)
		if err != nil {
			return fmt.Errorf("error while fetching Packages: %w", err)
		}

		for _, p := range npackages.Packages {
			for _, err := range p.Warnings {
				slog.WarnContext(ctx, "Warning while parsing packages", "error", err)
			}
		}

		slog.InfoContext(
			ctx, "Got next Packages",
			"packages", len(npackages.Packages), "files", len(npackages.FileBytes),
		)

		ppackages, err := kubo.Packages(ctx, source.URI, suite, component)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("error while fetching previous Packages: %w", err)
		}

		slog.InfoContext(
			ctx, "Got previous packages",
			"packages", len(ppackages.Packages), "files", len(ppackages.FileBytes),
		)

		diff := ppackages.Packages.Diff(npackages.Packages)

		// Upsert packages
		upsert := slices.Concat(diff.Added, diff.Changed)
		err = upsertPackages(ctx, kubo, client, source, suite, upsert)
		if err != nil {
			return fmt.Errorf("error while upserting packages: %w", err)
		}

		// Remove packages
		for _, removed := range diff.Removed {
			ctx := slogctx.Prepend(
				ctx,
				"package", removed.Package,
				"version", removed.Version,
			)

			slog.InfoContext(ctx, "Removing package from MFS")
			err = kubo.RemovePackage(ctx, source.URI, suite, removed)
			if err != nil {
				return fmt.Errorf("error while removing %s from MFS: %w", removed.Filename, err)
			}
		}

		pkgDiff := ppackages.FileBytes.Diff(npackages.FileBytes)

		// Upsert Packages
		pkgUpsert := slices.Concat(pkgDiff.Added, pkgDiff.Changed)
		for _, f := range pkgUpsert {
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

		// Remove Packages
		for _, f := range pkgDiff.Removed {
			ctx := slogctx.Prepend(
				ctx,
				"Packages", f.Hashes.Filename,
			)

			slog.InfoContext(ctx, "Removing Packages file from MFS")
			err = kubo.RemovePackages(ctx, source.URI, suite, f)
			if err != nil {
				return fmt.Errorf("error while removing %s from MFS: %w", f.Hashes.Filename, err)
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
		err = kubo.RemoveComponent(ctx, source.URI, suite, component)
		if err != nil {
			return fmt.Errorf("error while removing %s/%s from MFS: %w", component.Name, component.Architecture, err)
		}
	}

	slog.InfoContext(ctx, "Committing InRelease file")
	err = kubo.WriteInRelease(ctx, source.URI, suite, next)
	if err != nil {
		return fmt.Errorf("error while writing InRelease to MFS: %w", err)
	}

	return nil
}

func upsertPackages(
	ctx context.Context, kubo *kubo.Client, client *apt.Client,
	source apt.Config, suite string, packages apt.Packages,
) error {
	for _, p := range packages {
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
		r.Close()
		if err != nil {
			return fmt.Errorf("error while writing %s to MFS: %w", p.Filename, err)
		}
	}

	return nil
}
