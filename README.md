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
>`generate`, `strip`, `regen`, `check` without turning your Markdown into a moving target.

<img src="./docs/mdtoc_mascot_1024.webp" width="600">

<h2>Table of Contents</h2><!-- TABLE OF CONTENTS START -->

<details markdown="1"> <!-- parse this block as markdown -->
<summary>(click to expand)</summary>

<!-- mdtoc -->
* [1. Features](#features)
* [2. Install](#install)
  * [2.1. Releases](#releases)
  * [2.2. Build from source](#build-from-source)
* [3. Usage](#usage)
  * [3.1. Inspect the CLI](#inspect-the-cli)
  * [3.2. Use this README as example](#use-this-readme-as-example)
* [4. Managed Structure](#managed-structure)
* [5. Limits](#limits)
* [6. Documentation](#documentation)
  * [6.1. Specification](#specification)
  * [6.2. Comparison](#comparison)
* [7. Status](#status)
<!-- mdtoc-config
container-version=v2
numbering=on
min-level=2
max-level=4
anchor=github
toc=on
bullets=auto
state=generated
-->
<!-- /mdtoc -->
</details>

## 1. <a id="features"></a>Features

* single binary, no external tools required
* auto-detects the dominant bullet style (`*`, `-`, `+`) for ToC
* works with files and Unix pipes
* targets ATX headings (`#` to `######`)
* ignores headings inside fenced code blocks safely
* exclusion regions: `<!-- mdtoc off -->` ... `<!-- mdtoc on -->`
* explicit anchor profiles: `github` (default), `gitlab`, or `off`
* slug link anchors from heading titles, not numbers
* works with repeated headings
* generated content stays clearly separated from authored content
* deterministic and idempotent output

## 2. <a id="install"></a>Install

### 2.1. <a id="releases"></a>Releases

Download a prebuilt binary from [GitHub Releases](https://github.com/rokath/mdtoc/releases).

Homebrew tap install:

```bash
brew install rokath/tap/mdtoc
```

### 2.2. <a id="build-from-source"></a>Build from source

```bash
go build ./cmd/mdtoc
```

## 3. <a id="usage"></a>Usage

### 3.1. <a id="inspect-the-cli"></a>Inspect the CLI

```bash
mdtoc --help        # show compact CLI usage and commands
mdtoc --verbose     # show extended root help with command details
mdtoc <command> -v  # show the long help for one command
```

### 3.2. <a id="use-this-readme-as-example"></a>Use this README as example

```bash
cat MY_IMPORTANT_DOCUMENT.md | mdtoc generate    # do a dry-run
mdtoc generate -f README.md -a false -toc off    # rewrite headings only, keep inline anchors and ToC disabled
mdtoc generate -f README.md -a gitlab            # persist the GitLab anchor profile explicitly
cat README.md | mdtoc strip > README.stripped.md # remove managed artifacts via Unix pipe and write to a different file 
mdtoc regen -f README.md                         # rebuild the generated state from the stored container config
mdtoc generate -f README.md                      # generate with current CLI flags or defaults and rewrite the config block
mdtoc check -f README.md                         # fail in CI when README.md differs from the reconstructed target state
```

## 4. <a id="managed-structure"></a>Managed Structure

`mdtoc` uses an explicit container so generated content is easy to spot and safe to regenerate.

<details markdown="1">
<summary>(click to expand)</summary>

```md
<!-- mdtoc -->
* [About](#about)
<!-- mdtoc-config
container-version=v2
numbering=on
min-level=2
max-level=4
anchor=github
toc=on
bullets=auto
state=generated
-->
<!-- /mdtoc -->
```

The heading title stays the source of truth. Numbers, anchors, and ToC entries are derived from it.

The config block records the settings that produced the current managed state. `generate` always uses current CLI flags or defaults and then rewrites that block. `regen` reuses the stored container config and rebuilds the generated state from it.

This means:

* the stored config is persisted output state
* `regen` rebuilds the generated state from that persisted config
* `check` uses that persisted state
* changing generation settings currently goes through `generate`, not through manual config edits

</details>

## 5. <a id="limits"></a>Limits

* no Setext heading support (`Heading` followed by `===` or `---`)
* repeated-heading links depend on occurrence order ([issue #8](https://github.com/rokath/mdtoc/issues/8))
* not a site generator
* not a Markdown formatter

## 6. <a id="documentation"></a>Documentation

### 6.1. <a id="specification"></a>Specification

* [mdtoc spec](./docs/mdtoc-spec.md)

### 6.2. <a id="comparison"></a>Comparison

* [mdtoc tools comparison](./docs/mdtoc-tools-comparison.md)

## 7. <a id="status"></a>Status

```diff
+ READY TO USE +
```
