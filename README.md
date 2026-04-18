# mdtoc

<!-- mdtoc -->
* [1. Why mdtoc?](#why-mdtoc)
* [2. Install](#install)
  * [2.1. Releases](#releases)
  * [2.2. Build from source](#build-from-source)
* [3. Usage](#usage)
  * [3.1. Inspect the CLI](#inspect-the-cli)
  * [3.2. Use this README as example](#use-this-readme-as-example)
* [4. Managed Structure](#managed-structure)
* [5. Scope](#scope)
* [6. Documentation](#documentation)
  * [6.1. Specification](#specification)
  * [6.2. Comparison](#comparison)
* [7. Status](#status)
<!-- mdtoc-config
numbering=on
min-level=2
max-level=4
anchors=on
toc=on
state=generated
-->
<!-- /mdtoc -->

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

Deterministic Markdown ToC manager for single files.

`mdtoc` generates and validates a managed table of contents, heading numbering, and stable anchors without turning your Markdown into a moving target.

## 1. <a id="why-mdtoc"></a>Why mdtoc?

* one small CLI for ToC, numbering, anchors, stripping, and CI checks
* deterministic and idempotent output
* anchors are derived from the semantic heading title, not from generated numbers
* fenced code blocks are ignored safely while parsing headings and markers
* generated content stays clearly separated from authored content

## 2. <a id="install"></a>Install

### 2.1. <a id="releases"></a>Releases

Download a prebuilt binary from [GitHub Releases](https://github.com/rokath/mdtoc/releases).

### 2.2. <a id="build-from-source"></a>Build from source

```bash
go build ./cmd/mdtoc
```

## 3. <a id="usage"></a>Usage

### 3.1. <a id="inspect-the-cli"></a>Inspect the CLI

```bash
mdtoc --version # show version information
mdtoc --help    # show CLI usage and commands
```

### 3.2. <a id="use-this-readme-as-example"></a>Use this README as example

```bash
mdtoc generate -f README.md -a off -toc off # rewrite headings only, keep anchors and ToC disabled
cat README.md | mdtoc strip > README.md     # remove managed artifacts via Unix pipe and write clean Markdown back
mdtoc generate -f README.md                 # generate the managed container, numbering, anchors, and ToC
mdtoc check -f README.md                    # fail in CI when README.md differs from the reconstructed target state
```

## 4. <a id="managed-structure"></a>Managed Structure

`mdtoc` uses an explicit container so generated content is easy to spot and safe to regenerate:

```md
<!-- mdtoc -->
* [About](#about)
<!-- mdtoc-config
numbering=on
min-level=2
max-level=4
anchors=on
toc=on
state=generated
-->
<!-- /mdtoc -->
```

The heading title stays the source of truth. Numbers, anchors, and ToC entries are derived from it.

## 5. <a id="scope"></a>Scope

`mdtoc` is intentionally small:

* processes one Markdown file at a time
* supports file input and Unix pipes
* supports ATX headings (`#` to `######`)
* is not a site generator and not a full Markdown formatter

## 6. <a id="documentation"></a>Documentation

### 6.1. <a id="specification"></a>Specification

* [mdtoc spec](./docs/mdtoc-spec.md)

### 6.2. <a id="comparison"></a>Comparison

* [mdtoc tools comparison](./docs/mdtoc-tools-comparison.md)

## 7. <a id="status"></a>Status

```diff
+ READY TO USE +
```
