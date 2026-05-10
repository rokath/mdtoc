# mdtoc

[![Release](https://img.shields.io/github/v/release/rokath/mdtoc)](https://github.com/rokath/mdtoc/releases)
[![Commits Since Release](https://img.shields.io/github/commits-since/rokath/mdtoc/latest)](https://github.com/rokath/mdtoc/commits/main/)
[![GitHub Issues](https://img.shields.io/github/issues/rokath/mdtoc)](https://github.com/rokath/mdtoc/issues)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](https://makeapullrequest.com)
[![License](https://img.shields.io/github/license/rokath/mdtoc)](https://github.com/rokath/mdtoc)
[![Downloads](https://img.shields.io/github/downloads/rokath/mdtoc/total)](https://github.com/rokath/mdtoc/releases)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/rokath/mdtoc)](https://goreportcard.com/report/github.com/rokath/mdtoc)
[![Coverage](https://coveralls.io/repos/github/rokath/mdtoc/badge.svg?branch=main)](https://coveralls.io/github/rokath/mdtoc?branch=main)
[![Pages](https://img.shields.io/badge/docs-GitHub%20Pages-blue)](https://rokath.github.io/mdtoc/)
[![Pages Workflow](https://github.com/rokath/mdtoc/actions/workflows/pages.yml/badge.svg)](https://github.com/rokath/mdtoc/actions/workflows/pages.yml)

[View Github Pages](https://rokath.github.io/mdtoc/)

>`mdtoc`: Markdown Table of Contents ☰ with numbering and stable anchor links
>
>`generate`, `strip`, `regen`/`refresh`, `check` without turning your Markdown into a moving target.

<img src="./extension/mdtoc_mascot_1024.webp" width="600">

<details markdown="1"> <!-- parse this block as markdown -->
<summary><strong style="font-size: 1.25em;">Table of Contents</strong> <span style="font-size: 0.66em;">(click to expand)</span></summary>

<!-- mdtoc -->

* [1. Features](#1-features)
* [2. Install](#2-install)
  * [2.1. Releases](#2-1-releases)
  * [2.2. Build from source](#2-2-build-from-source)
* [3. Usage](#3-usage)
  * [3.1. Inspect the CLI](#3-1-inspect-the-cli)
  * [3.2. Use EXAMPLE.md as play ground](#3-2-use-example-md-as-play-ground)
* [4. Managed Structure](#4-managed-structure)
* [5. Limits](#5-limits)
* [6. Documentation](#6-documentation)
  * [6.1. Specification](#6-1-specification)
  * [6.2. Comparison](#62-comparison)
* [7. Status](#7-status)

<!-- numbering=true min=2 max=4 slug=github anchor=false link=true toc=true bullets=auto -->
<!-- /mdtoc -->
</details>

## 1. Features

* **easy** to use: `mdtoc MY.md`
* single binary, also as **vsCode** extension
* **configurable**: CLI or edit generated `<!-- numbering=true min=2 max=4 slug=github anchor=true link=true toc=true bullets=auto -->`
  * `on|off` for **numbering**, **anchor**, **link**, **toc**
  * targets ATX headings (**min** `#` to **max** `######`)
  * **slug** profiles: `github`, `gitlab`, `crossnote`
  * auto or explicit (`*`, `-`, `+`) ToC **bullets** style
* **move** the generated ToC with its container to any place - it will be re-generated there
* **protects** non-generated content inside ToC area
  * generated content stays clearly separated from authored content
* works with **files** and Unix **pipes**
* **repeated headings** support
* **ignores** headings:
  * in **Setext** stype
  * inside **fenced code blocks**
  * inside **HTML comments**: `<!-- ... ## Example -->`
  * between **exclusion regions**: `<!-- mdtoc off -->` ... `<!-- mdtoc on -->`
* deterministic and idempotent output

## 2. Install

### 2.1. Releases

Download a prebuilt binary from [GitHub Releases](https://github.com/rokath/mdtoc/releases).

Homebrew tap install:

```bash
brew install rokath/tap/mdtoc
```

### 2.2. Build from source

```bash
go build ./cmd/mdtoc
```

## 3. Usage

### 3.1. Inspect the CLI

```bash
mdtoc --help        # show compact CLI usage and commands
mdtoc --verbose     # show extended root help with command details
mdtoc <command> -v  # show the long help for one command
```

### 3.2. Use [EXAMPLE.md](./docs/EXAMPLE.md) as play ground

```bash
mdtoc EXAMPLE.md                                   # root mode: regen if managed, otherwise generate
cat EXAMPLE.md | mdtoc -n off -a off               # root mode on stdin: generate a dry-run ToC-only view
mdtoc EXAMPLE.md -a off --toc off                  # root mode: explicit generate because flags override regen
mdtoc generate EXAMPLE.md --slug gitlab            # explicit command with positional file input
cat EXAMPLE.md | mdtoc strip > EXAMPLE.stripped.md # remove managed artifacts via Unix pipe and write to a different file
mdtoc regen EXAMPLE.md                             # rebuild generated output from the stored container config
mdtoc refresh EXAMPLE.md                           # alias for regen
mdtoc generate EXAMPLE.md                          # generate with current CLI flags or defaults and rewrite the config block
mdtoc check EXAMPLE.md                             # fail in CI when EXAMPLE.md differs from the reconstructed target state
```

* `gitlab` follows GitLab heading IDs; punctuation-heavy titles can therefore differ from `github` (for example `3.5` -> `35`). See [GitLab slug profile](docs/spec.md#gitlab-slug-profile).
* Exactly one input source is allowed per invocation: positional file, `--file/-f`, or piped `stdin`.
* Small CLI note: the Go-style one-dash long form such as `-toc off` is currently tolerated, but `--toc off` remains the documented form.

## 4. Managed Structure

`mdtoc` uses an explicit container so generated content is easy to spot and safe to regenerate.

<details markdown="1">
<summary>(click to expand)</summary>

```md
<!-- mdtoc -->

* [About](#about)

<!-- numbering=true min=2 max=4 slug=github anchor=true link=true toc=true bullets=auto -->
<!-- /mdtoc -->
```

The heading title stays the source of truth. Numbers, inline anchors, and ToC entries are derived from it. With `anchor=false`, the ToC target follows the rendered heading text because no managed inline anchor exists. Use `slug=github|gitlab|crossnote` to control the link type generation.

The optional config block records the settings used for regeneration. `generate` uses current CLI flags or defaults and then rewrites that block when settings need to be persisted. `regen` reuses the stored container config, or defaults if the config block was deleted. `refresh` is a command alias for `regen`.

This means:

* the stored config is persisted generator input
* `regen` rebuilds generated output from that config
* `refresh` is an alias for `regen`
* `check` validates against regenerated output
* changing generation settings must go through generate, not through manual config edits

</details>

## 5. Limits

* repeated-heading links depend on occurrence order ([#8](https://github.com/rokath/mdtoc/issues/8))
  * Workaround: [example](./docs/EXAMPLE.md#chapter-a-about)
* The per default with `anchor=true` generated ToC links guaranty to work in any environment, but reduce the readability of the raw Markdown document. With `anchor=off numbering=off slug=crossnote` a good working setting is possible. But switching `numbering=on` breaks the link stability promise then. There is no generally best setting - you have to choose. See also [#94](https://github.com/rokath/mdtoc/issues/94).
* The `check` command does not detect duplicate link anchors. See [#97](https://github.com/rokath/mdtoc/issues/97).
* not a site generator
* not a Markdown formatter

## 6. Documentation

### 6.1. Specification

* [mdtoc spec](./docs/spec.md)

### 6.2. Comparison

* [mdtoc alternatives](./docs/alternatives.md)
* [mdtoc VS Code extension MVP](./docs/vscode-extension-mvp.md)

## 7. Status

```diff
+ READY TO USE +
```
