# Changelog

This file summarizes notable repository changes in a compact, release-oriented format.

## <a id='unreleased-changes'></a>Unreleased Changes

### <a id='unreleased-overview'></a>Unreleased Overview

* No unreleased changes recorded yet.

### <a id='unreleased-git-log'></a>Unreleased Git Log

Used git range: `v0.2.2..HEAD`

```txt
```

## <a id='v0.2.2-changes'></a>v0.2.2 Changes (2026-05-02)

### <a id='v0.2.2-overview'></a>v0.2.2 Overview

* VS Code extension groundwork was added:
  * a new `extension/` workspace now provides a desktop-only MVP scaffold with commands for `generate`, `regen`, `strip`, `check`, and `showVersion`
  * the extension resolves `mdtoc` via `customPath -> bundled binary`, applies document changes through `stdin`/`stdout`, and documents the MVP contract in a dedicated spec
  * the MVP spec link in `README.md` now points to a short English version in `docs/vscode-extension-mvp.md`, including the end-user `Install from VSIX...` step
  * a new `package:release` extension script now stages binaries from the GoReleaser `dist/` archives and packages all four MVP VSIX targets in one pass
  * the release workflows now build and upload per-platform VSIX artifacts for `darwin-x64`, `darwin-arm64`, `linux-x64`, and `win32-x64`
  * local developer ergonomics were improved with an `F5` launch configuration and explicit README guidance for testing through either the Extension Development Host or a locally installed VSIX
* CLI naming became a bit more forgiving:
  * `refresh` is now accepted as an alias for the `regen` subcommand
  * the documented CLI surface still keeps `regen` as the canonical command name
* Documentation paths were shortened and clarified:
  * the primary docs now live at `docs/spec.md`, `docs/alternatives.md`, and `docs/vscode-extension-mvp.md`
  * `README.md`, `index.md`, and current changelog references now point to the shorter document paths
  * the VS Code extension MVP document now expands `MVP` as `Minimum Viable Product`
* Release preparation rules were tightened:
  * `AGENTS.md` now requires explicit `CHANGELOG.md` review before relevant commits, not only before push or tag
  * release-preparation commits now require an already prepared version section, correct git range, and a reset `Unreleased Changes` section
  * extension build artifacts such as `extension/bin/`, `extension/out/`, and `extension/.tmp/` are now ignored

### <a id='v0.2.2-git-log'></a>v0.2.2 Git Log

Used git range: `v0.2.1..v0.2.2`

```txt
* 1d77840 2026-05-02 docs: shorten docs paths and retire German drafts
* 0241f58 2026-05-02 feat(extension): add VS Code MVP and VSIX release flow
* 0ffd65d 2026-05-02 feat(cli): add refresh alias and subcommand-flag hints
```

## <a id='v0.2.1-changes'></a>v0.2.1 Changes (2026-04-25)

### <a id='v0.2.1-overview'></a>v0.2.1 Overview

* CLI file handling became more flexible:
  * positional file arguments are now accepted alongside `--file/-f` for `generate`, `regen`, `strip`, and `check`
  * root mode now dispatches automatically to `regen` for valid managed input without generate overrides and to `generate` otherwise
  * root mode and explicit subcommands now reject conflicting input sources such as positional file plus `--file`, or file input plus piped stdin
  * file-backed commands now tolerate empty redirected `stdin` so CI shells can run `-f/--file` workflows without false mixed-input failures
  * new CLI workflow tests cover root-mode dispatch, positional file support, and input-source conflict handling with the virtual filesystem helpers
* CLI input validation and upgrade guidance were tightened:
  * commands now reject mixed input from piped stdin together with `--file/-f` instead of silently ignoring stdin
  * config length errors now hint that the document may require a newer `mdtoc` version when a versioned container format is detected
  * tests now cover mixed-input rejection for `generate`, `regen`, `strip`, and `check`
* GitLab anchor generation was implemented:
  * `--anchor gitlab` now derives heading IDs with a GitLab-specific profile instead of behaving like `github`
  * tests now cover documented GitHub/GitLab differences and a file-level generate workflow for a punctuation-sensitive heading
  * `README.md` and the specification now document the GitLab profile and link to a stable reference section
* The documentation set was consolidated:
  * the specification was brought back in sync with the current implementation, including config normalization, legacy v1 containers, bullet auto-detection, `regen`, and `strip --raw` fallback behavior
  * the alternatives page was reduced to a more compact comparison overview
  * README usage wording and extension planning notes were clarified

### <a id='v0.2.1-git-log'></a>v0.2.1 Git Log

Used git range: `v0.2.0..v0.2.1`

```txt
* 627cfd1 2026-04-23 docs(changelog): add missing v0.2.0 notes
* c22a117 2026-04-24 feat(anchor): implement GitLab slug profile
* 29a434d 2026-04-24 docs(alternatives): replace tools comparison pages
* a9902d2 2026-04-24 docs(spec): align behavior reference with implementation
* d5efef8 2026-04-24 Add usage instruction for mdtoc
* 752e00b 2026-04-24 Clarify GitLab heading ID behavior in README
* ded9549 2026-04-24 docs: add German VS Code extension notes
* e9449f8 2026-04-24 fix(cli): reject mixed stdin and file input
* 2b792a7 2026-04-25 feat(cli): add root dispatch and positional file input
* 55ab000 2026-04-25 fix(cli): allow --file with empty redirected stdin
* c306649 2026-04-25 docs(internal): clarify type field semantics
```

## <a id='v0.2.0-changes'></a>v0.2.0 Changes (2026-04-23)

### <a id='v0.2.0-overview'></a>v0.2.0 Overview

* Release notes formatting was corrected:
  * the GoReleaser footer now escapes the macOS `- OR -` separator so it renders as intended in published release notes
* Managed config handling was versioned and normalized:
  * new managed config blocks now write `container-version=v2`
  * legacy config blocks without a version remain readable as implicit `v1`
  * persisted boolean config values now normalize to `true|false`, while `anchor` normalizes to `github|gitlab|off`
  * the CLI accepts `on|off|true|false` for boolean flags and accepts `false` as an alias for `--anchor off`
  * the deprecated `--anchors` CLI flag was removed

### <a id='v0.2.0-git-log'></a>v0.2.0 Git Log

Used git range: `v0.1.8..v0.2.0`

```txt
* 498156a 2026-04-23 docs(changelog): record config normalization changes
* 579450c 2026-04-23 config: canonicalize bool and anchor values
* 03f44f9 2026-04-23 format: version managed config and refine anchor mode
* 74bddc1 2026-04-23 docs(release): escape footer separator
```

## <a id='v0.1.8-changes'></a>v0.1.8 Changes (2026-04-23)

### <a id='v0.1.8-overview'></a>v0.1.8 Overview

* Homebrew release auditing was corrected:
  * the version assertion now normalizes the release tag from `vX.Y.Z` to `X.Y.Z` before comparing it with `mdtoc --version`

### <a id='v0.1.8-git-log'></a>v0.1.8 Git Log

Used git range: `v0.1.7..v0.1.8`

```txt
* 048587c 2026-04-23 fix(ci): normalize Homebrew audit version check
```

## <a id='v0.1.7-changes'></a>v0.1.7 Changes (2026-04-23)

### <a id='v0.1.7-overview'></a>v0.1.7 Overview

* Homebrew release auditing was simplified:
  * the macOS Homebrew audit now uses the minimal user-facing path with `brew tap rokath/tap`
  * unnecessary cleanup steps were removed so the CI path stays aligned with `brew install rokath/tap/mdtoc`
* Release notes were refined:
  * the GoReleaser release footer now includes the Homebrew install command `brew install rokath/tap/mdtoc`
* Repository release-process guidance was tightened:
  * `AGENTS.md` now requires changelog reconciliation with `origin/main` before continuing from a structurally older release-preparation state
  * `CHANGELOG.md` was restored so `v0.1.5` and `v0.1.6` remain present as explicit version sections

### <a id='v0.1.7-git-log'></a>v0.1.7 Git Log

Used git range: `v0.1.6..v0.1.7`

```txt
* fb89a76 2026-04-23 docs(agents): require changelog sync with main
* 3c606ce 2026-04-23 docs(changelog): restore v0.1.5 and v0.1.6 sections
* 4229466 2026-04-23 fix(ci): simplify Homebrew release audit
```

## <a id='v0.1.6-changes'></a>v0.1.6 Changes (2026-04-23)

### <a id='v0.1.6-overview'></a>v0.1.6 Overview

* Homebrew release auditing was corrected:
  * the macOS Homebrew audit now uses the canonical tap command `brew tap rokath/tap`
  * this matches the documented user install path and avoids a false-negative audit failure

### <a id='v0.1.6-git-log'></a>v0.1.6 Git Log

Used git range: `v0.1.5..v0.1.6`

```txt
* 7760117 2026-04-22 fix(ci): use canonical Homebrew tap command
```

## <a id='v0.1.5-changes'></a>v0.1.5 Changes (2026-04-22)

### <a id='v0.1.5-overview'></a>v0.1.5 Overview

* `strip --raw` was hardened:
  * it now falls back to tolerant container removal when strict config parsing fails for malformed or future managed containers
  * fallback stripping still removes managed numbering and inline anchors from headings after the container is removed
  * new regression tests cover future-version config lines, unknown config keys, malformed config blocks, and a file-level CLI fallback path
* Repository testing guidance was tightened:
  * `AGENTS.md` now requires a file-level test by default for CLI file workflow and file-backed command changes
  * virtual filesystem test helpers should be preferred over OS-level files when feasible
* Homebrew distribution support was added:
  * GoReleaser now publishes a formula for `mdtoc` into `rokath/homebrew-tap`
  * the release workflow now documents the required `HOMEBREW_TAP_GITHUB_TOKEN` secret for cross-repository publishing
  * the README now documents the intended install command `brew install rokath/tap/mdtoc`
* Homebrew release auditing was added:
  * the manual `release-audit` workflow now runs a macOS Homebrew install audit against `brew install rokath/tap/mdtoc`
  * the audit verifies that the installed Homebrew binary matches the latest GitHub release tag and passes the shared smoke-test fixture

### <a id='v0.1.5-git-log'></a>v0.1.5 Git Log

Used git range: `v0.1.4..v0.1.5`

```txt
* e15ccd6 2026-04-22 ci: audit Homebrew installation path
* dc13519 2026-04-22 distribution: add Homebrew tap publishing path
* 8ca3e1e 2026-04-22 docs(agents): require file-level workflow tests by default
* cb57d2b 2026-04-22 fix(strip): harden raw strip fallback
```

## <a id='v0.1.4-changes'></a>v0.1.4 Changes (2026-04-22)

### <a id='v0.1.4-overview'></a>v0.1.4 Overview

* Pull request install verification was added:
  * a dedicated `install-checks` GitHub Actions workflow now builds release-style snapshot archives on pull requests
  * the workflow uploads deterministic Linux, macOS, and Windows artifacts for downstream validation jobs
  * separate Ubuntu, macOS, and Windows jobs now unpack the PR-built archives and verify that the shipped binary starts successfully
  * the install jobs now reuse a checked-in smoke-test fixture that exercises repeated headings, numbered headings, exclusions, fenced code, and `+` bullet auto-detection
  * managed ToC preservation now recognizes generated `*`, `-`, and `+` list entries consistently, and Go tests cover fixture-based `generate` plus `check` file workflows
* Published release auditing was added:
  * a separate manual `release-audit` workflow now downloads the latest published Linux and Windows release artifacts from GitHub Releases
  * the workflow verifies `mdtoc --version` plus the shared generate/check smoke-test flow against the checked-in install fixture
* Linux packaging support was expanded:
  * GoReleaser now emits `.deb`, `.rpm`, and `.apk` artifacts for the initial supported Linux package targets
  * Debian, RPM, and Alpine package metadata is now defined explicitly for `mdtoc`, including package name, homepage, license, install path, and shipped license file
  * pull request install checks now install the generated Debian, RPM, and Alpine packages and run the shared smoke-test flow against the installed `/usr/bin/mdtoc`
* README guidance was refined:
  * the feature list now calls out the single-binary, no-external-tools setup
  * usage examples now show safe pipe output to a different file and a simple stdin dry-run pattern

### <a id='v0.1.4-git-log'></a>v0.1.4 Git Log

Used git range: `v0.1.3..v0.1.4`

```txt
* fec53b0 2026-04-22 ci: test Alpine APK installation
* c1c78d0 2026-04-22 packaging: add Alpine APK package artifacts
* b7a5053 2026-04-22 ci: test RPM package installation
* 6eded44 2026-04-21 packaging: add RPM package artifacts
* 26a7148 2026-04-21 ci: test Debian package installation
* e70bba4 2026-04-21 packaging: add Debian package artifacts
* 9e35f05 2026-04-21 ci: add manual release audit workflow
* 358a3f4 2026-04-21 fix(toc): preserve generated bullet variants
* 418d925 2026-04-21 ci: add install smoke-test fixture
* 3adfb8f 2026-04-21 ci: add PR install-checks workflow
* 9a77d54 2026-04-21 docs(changelog): add missing v0.1.3 notes
* fe142b1 2026-04-21 docs(readme): refine features and usage examples
```

## <a id='v0.1.3-changes'></a>v0.1.3 Changes (2026-04-21)

### <a id='v0.1.3-overview'></a>v0.1.3 Overview

* ToC bullet handling was improved:
  * `generate` now auto-detects the dominant unordered-list bullet style from normal document content
  * supported bullet styles are `*`, `-`, and `+`
  * ties are resolved deterministically with `*` > `-` > `+`
  * `--bullets` and `-b` now allow forcing a specific ToC bullet style or keeping `auto`
  * workflow regression tests now verify that managed ToC bullets are not counted during `auto` detection and that legacy containers without `bullets=` stay on `*`
* Markdown heading exclusion support was added:
  * `<!-- mdtoc off -->` and `<!-- mdtoc on -->` now exclude heading regions from ToC generation and managed heading rewrites
  * a missing `<!-- mdtoc on -->` is accepted and keeps the exclusion active until end of file
* README guidance was refined:
  * handling of repeated headings was clarified
  * bullet-style detection wording was corrected and duplicate command wording was removed

### <a id='v0.1.3-git-log'></a>v0.1.3 Git Log

Used git range: `v0.1.2..v0.1.3`

```txt
* 4c4eeec 2026-04-21 Fix bullet style detection description in README and removed duplicate line about commands.
* 737e0af 2026-04-21 test(toc): expand bullet workflow regressions
* 76c01b7 2026-04-20 feat(toc): auto-detect unordered list bullets
* eedc98a 2026-04-20 feat(parser): support mdtoc off/on exclusions
* 70d34d3 2026-04-20 Fix capitalization in README.md
* 3cd401e 2026-04-20 Add note about handling repeated headings
```

## <a id='v0.1.2-changes'></a>v0.1.2 Changes (2026-04-20)

### <a id='v0.1.2-overview'></a>v0.1.2 Overview

* CLI capabilities expanded:
  * `regen` was added as an explicit command for rebuilding the generated state from persisted container config
  * `regen` now restores the generated state correctly even after `strip`
  * `mdtoc --verbose` now shows the long root help
  * `mdtoc <command> -v` now shows the long help for the selected command
* Test coverage and workflow safety improved:
  * new regression tests cover `generate`, `strip`, `regen`, and `check` as real command sequences
  * file-based CLI workflow tests now run against an in-memory filesystem to catch state-transition regressions
* README and project metadata were refined:
  * the README now documents `regen`, the new help behavior, and the persisted-config semantics more clearly
  * the README hero section, collapsible blocks, and usage examples were polished further
  * the README now highlights safe handling of fenced code blocks more prominently
  * a Coveralls coverage badge was added
* Project infrastructure and docs were completed:
  * a dedicated GitHub Actions coverage workflow now uploads Go coverage to Coveralls
  * an MIT license file was added
  * GitHub Pages rendering for `docs/` Markdown pages was stabilized
  * spec and comparison docs were normalized further for Markdown list formatting and links
  * `AGENTS.md` now requires reviewing and updating `CHANGELOG.md` before every push when relevant

### <a id='v0.1.2-git-log'></a>v0.1.2 Git Log

Used git range: `v0.1.1..v0.1.2`

```txt
* 3ca1ed6 2026-04-20 feat(cli): add regen workflows and verbose command help
* e6cdeea 2026-04-20 docs(readme): clarify generate vs stored config
* 6517855 2026-04-20 Enhance README formatting and content details
* ea7da43 2026-04-19 docs: tighten changelog-before-push rule
* 09973d8 2026-04-19 docs(readme): refresh mascot asset and intro
* ec73c52 2026-04-19 docs: fix pages rendering and update push guidance
* 01e9e1b 2026-04-19 Update markdown replacement for German tools comparison
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
