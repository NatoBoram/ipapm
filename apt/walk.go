package apt

import (
	"context"
	"fmt"
	"path"

	slogctx "github.com/veqryn/slog-context"
)

// Tree lists all files reachable from a config.
type Tree struct {
	Suites []TreeSuite
}

type TreeSuite struct {
	Suite      string
	InRelease  InRelease
	Components []TreeComponent
}

type TreeComponent struct {
	Component Component
	Packages  PackagesBytes
	Sources   SourcesBytes
}

func (t Tree) Flat() (FileHashes, error) {
	flat := make(FileHashes)

	for _, suite := range t.Suites {

		in := FileHash{
			Filename: path.Join("dists", suite.Suite, "InRelease"),
			Size:     uint(len(suite.InRelease.Raw)),
		}
		flat[in.Filename] = in

		files, err := suite.InRelease.FileHashes()
		if err != nil {
			return nil, fmt.Errorf("getting file hashes for suite %s: %w", suite.Suite, err)
		}

		for _, file := range files {
			target := path.Join("dists", suite.Suite, file.Filename)
			flat[target] = file
		}

		for _, component := range suite.Components {
			if component.Component.Architecture == "source" {
				for _, source := range component.Sources.Sources {
					files, err := source.FileHashes()
					if err != nil {
						return nil, fmt.Errorf("getting file hashes for source %s %s: %w", source.Package, source.Version, err)
					}

					for _, file := range files {
						target := path.Join(file.Filename)
						flat[target] = file
					}
				}
			} else {
				for _, file := range component.Packages.Packages {
					target := path.Join(file.Filename)
					flat[target] = file.FileHash()
				}
			}
		}
	}

	return flat, nil
}

func (c *Client) Walk(ctx context.Context, config Config) (Tree, error) {
	var tree Tree

	for suite := range config.Suites {
		ctx := slogctx.Prepend(ctx, "suite", suite)
		var treeSuite TreeSuite
		treeSuite.Suite = suite

		// InRelease
		in, err := c.InRelease(ctx, config.URI, suite)
		if err != nil {
			return Tree{}, fmt.Errorf("walking suite: %w", err)
		}

		err = VerifyPgp(config.SignedBy, string(in.Raw))
		if err != nil {
			return Tree{}, fmt.Errorf("verifying PGP signature for suite %s: %w", suite, err)
		}

		treeSuite.InRelease = in

		// Contents
		files, err := in.FileHashes()
		if err != nil {
			return Tree{}, fmt.Errorf("getting file hashes for suite: %w", err)
		}

		// Components & architectures
		components := in.ByComponents(files)
		for _, component := range components {
			ctx := slogctx.Prepend(ctx, "component", component.Name, "architecture", component.Architecture)
			var treeComponent TreeComponent
			treeComponent.Component = component

			if component.Architecture == "source" {
				sources, err := c.Sources(ctx, config.URI, suite, component)
				if err != nil {
					return Tree{}, fmt.Errorf("getting sources for component: %w", err)
				}
				treeComponent.Sources = sources
			} else {
				packages, err := c.Packages(ctx, config.URI, suite, component)
				if err != nil {
					return Tree{}, fmt.Errorf("getting packages for component: %w", err)
				}
				treeComponent.Packages = packages
			}

			treeSuite.Components = append(treeSuite.Components, treeComponent)
		}

		tree.Suites = append(tree.Suites, treeSuite)
	}

	return tree, nil
}
