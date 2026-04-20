# Changelog

This file summarizes notable repository changes in a compact, release-oriented format.

## <a id='unreleased-changes'></a>Unreleased Changes

### <a id='unreleased-overview'></a>Unreleased Overview

* Documentation and presentation were refined further:
  * the README now includes a hero image and a collapsible table of contents
  * the README mascot asset and intro text were refreshed
  * the README now explains explicitly that `generate` does not reuse stored container config automatically
  * the tools-comparison link in the README was corrected
  * specification documents were normalized further for list-marker consistency
  * the obsolete `AI-log.md` file was removed from the repository
  * GitHub Pages rendering for the `docs/` Markdown pages was stabilized via Jekyll defaults and homepage link rewrites
* Licensing and metadata were completed:
  * an MIT license file was added to the repository
  * the README gained a Coveralls coverage badge
* CI coverage reporting was added:
  * a dedicated GitHub Actions workflow now runs Go coverage and uploads it to Coveralls
* Repository workflow guidance was tightened:
  * `AGENTS.md` now requires reviewing and updating `CHANGELOG.md` before every push when relevant
* Future CLI behavior was clarified:
  * issue [#5](https://github.com/rokath/mdtoc/issues/5) proposes an explicit `regen` command for reusing persisted container config
  * changing config values is still intended to go through `generate`

### <a id='unreleased-git-log'></a>Unreleased Git Log

Used git range: `v0.1.1..HEAD`

```txt
* 09973d8 2026-04-19 docs(readme): refresh mascot asset and intro
* 1cd6d2f 2026-04-19 docs(changelog): update unreleased notes
* a726d95 2026-04-19 docs(readme): add hero image and collapsible toc
* 48db36f 2026-04-18 Update link for tools comparison in README
* d3b9b65 2026-04-18 Changed list elements marker from dash - to start *.
* 23aeb4e 2026-04-18 docs: remove AI log
* b3a2453 2026-04-18 docs(license): add MIT license
* ca51634 2026-04-18 ci(coverage): add coveralls workflow
* d0d4bc5 2026-04-18 docs(readme): add coverage badge
```

## <a id='v0.1.1-changes'></a>v0.1.1 Changes (2026-04-18)

### <a id='v0.1.1-overview'></a>v0.1.1 Overview

* CLI usability improved for interactive use:
  * `generate`, `strip`, and `check` now fail fast when no `--file` is given and no input is piped via `stdin`
  * this resolves confusing blocking behavior described in [#4](https://github.com/rokath/mdtoc/issues/4)
* Test coverage and regression protection were expanded substantially:
  * overall statement coverage was raised above 90%
  * `cmd/mdtoc` is now test-covered
  * additional parser, config, CLI, process, and helper branches are verified directly
* Code and test documentation were normalized:
  * missing comments were added for exported and non-exported functions, structs, helper functions, and test functions
* README and project docs were refined:
  * badges were refreshed and aligned to `mdtoc`
  * the README now doubles as a nested ToC example and usage demo
  * comparison docs were renamed from `replacement-tools-comparison` to `tools-comparison`
  * specification docs were normalized for Markdown list spacing and marker consistency
  * an initial project changelog was added
* CI compatibility was updated:
  * the GoReleaser workflow now uses `goreleaser/goreleaser-action@v7`
  * this avoids the Node 20 deprecation warning on GitHub Actions runners

### <a id='v0.1.1-git-log'></a>v0.1.1 Git Log

Used git range: `v0.1.0..v0.1.1`

```txt
* 4623bb6 2026-04-18 ci(goreleaser): use goreleaser-action v7
* 28e1c85 2026-04-18 docs(spec): normalize list spacing and markers
* 33995a8 2026-04-18 docs(changelog): add initial changelog
* 8b39226 2026-04-18 test: raise coverage and document helpers
* ec07245 2026-04-18 feat(cli): fail fast on missing interactive input
* b4d12c8 2026-04-18 docs(readme): refresh badges and comparison docs
* 937d9cb 2026-04-18 Fix mdtoc command examples in README.md
* 8950547 2026-04-18 docs(readme): shorten usage and scope
* f5a60b4 2026-04-18 docs(readme): restructure README as usage example
* 53dbe8c 2026-04-18 Example added
* 2a66916 2026-04-18 Fix link typo and add usage examples in README
* 76c6d5e 2026-04-18 State chapter added
* 2e7c06e 2026-04-18 docs: translate german markdown references
* 184d4ec 2026-04-18 ci: update goreleaser workflow
```

## <a id='v0.1.0-changes'></a>v0.1.0 Changes (2026-04-18)

### <a id='v0.1.0-overview'></a>v0.1.0 Overview

* First tagged release of `mdtoc`
* Core functionality introduced:
  * deterministic generation of a managed Markdown ToC container
  * stable heading numbering
  * stable anchors derived from the semantic heading title
  * `generate`, `strip`, `strip --raw`, and `check`
* Initial repository and delivery setup:
  * GitHub Pages configuration
  * GoReleaser setup for release artifacts
  * README and specification docs in English and German
* Early issues were addressed before the first tag:
  * fixes for issues 1, 2, and 3

### <a id='v0.1.0-git-log'></a>v0.1.0 Git Log

Used git range: repository start..`v0.1.0`

```txt
* 03dbf86 2026-04-16 Initial commit
* 51f1a67 2026-04-16 README.md, mdtoc-spec.md initial update
* 49a2f64 2026-04-16 Compare list added
* 5d7d6b3 2026-04-16 mdtoc-spec fiinalized (german) and translated into English
* 3c0e753 2026-04-17 AI generated code added
* b05584b 2026-04-17 fix: resolve issues 1 2 3
* 5201e05 2026-04-17 build: add pages and goreleaser setup
* d6f96d0 2026-04-17 ci: update github pages workflow
* 19a368f 2026-04-18 ci: fix pages workflow conditions
```
