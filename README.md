# mdtoc

<!-- mdtoc -->
* [1. About](#about)
* [2. Highlights](#highlights)
* [3. Build](#build)
* [4. Test](#test)
* [5. Why this tool exists](#why-this-tool-exists)
* [6. What makes mdtoc different](#what-makes-mdtoc-different)
* [7. Why not just use an existing tool?](#why-not-just-use-an-existing-tool)
* [8. Project goal](#project-goal)
* [9. In one sentence](#in-one-sentence)
<!-- mdtoc-config
numbering=on
min-level=2
max-level=4
anchors=on
toc=on
state=generated
-->
<!-- /mdtoc -->

## 1. <a id="about"></a>About

Deterministic Table of Contents (ToC) with Numbering and stabile Anchors including heading management for Markdown documents

This repository contains a Go reference implementation of the `mdtoc` specification (see [the specification](./docs/mdtoc-spec.md)

Alternatives: [replacement tools comparison](./docs/mdtoc-replacement-tools-comparison.md).

## 2. <a id="highlights"></a>Highlights

* deterministic container parsing
* stable heading numbering
* stable anchor IDs derived from unnumbered heading text
* `generate`, `strip`, `strip --raw`, and `check`
* unit tests for slug generation, parsing behavior, rendering, and CLI exit codes
* extensive comments in English throughout the source code

## 3. <a id="build"></a>Build

```bash
go build ./cmd/mdtoc
```

## 4. <a id="test"></a>Test

```bash
go test ./...
```


## 5. <a id="why-this-tool-exists"></a>Why this tool exists

Managing Markdown documents at scale sounds simple—until it isn’t.

<!--
Existing tools typically solve parts of the problem:

Generate a Table of Contents (ToC)
Generate anchor links
Sometimes number headings

But they often fail when you need all of the following at once:

Stable, deterministic output (CI-friendly)
Idempotent behavior (safe to run repeatedly)
Consistent heading numbering
Anchor IDs derived from the semantic title (not numbering)
Correct handling of Markdown edge cases (especially fenced code blocks)
Clean separation between source content and generated artifacts

In practice, combining multiple tools leads to:

Conflicting transformations
Broken anchor links
Non-reproducible results
Fragile CI pipelines
Design goals

mdtoc is built to address these issues with a single, coherent model.

Deterministic

Given the same input, mdtoc always produces the same output.

This is essential for:

CI pipelines
reproducible documentation builds
clean diffs in version control
-->

There are many tools that generate a table of contents. Some also generate anchors. A few can add heading numbers. But once a document needs all of these features together, the usual solutions become awkward:

* heading numbers change visible text
* anchors should stay based on the semantic title, not the number
* repeated runs should not keep changing the file
* code fences must not be mistaken for headings
* generated content must stay clearly separated from authored content

In practice, this often leads to fragile tool chains, broken links, noisy diffs, and CI checks that are hard to trust.

`mdtoc` is being developed to solve this as **one coherent problem**, not as a pile of loosely connected text transformations.

## 6. <a id="what-makes-mdtoc-different"></a>What makes mdtoc different

`mdtoc` is built around a simple idea:

> The heading title is the source of truth.  
> Everything else is derived from it.

That means:

* heading numbers are generated, not authored
* anchor IDs are generated from the unnumbered title
* the table of contents is generated from the same structure
* generated artifacts can be removed and recreated at any time

This keeps documents predictable, reviewable, and safe to process automatically.

## 7. <a id="why-not-just-use-an-existing-tool"></a>Why not just use an existing tool?

Because the existing tools we looked at solve only parts of the problem well.

Some are good at generating ToCs, but do not manage heading numbering.  
Some number headings, but are not robust around fenced code blocks.  
Some produce anchors, but not in a way that fits a deterministic, idempotent workflow.

For the intended use case, especially in CI and long-lived technical documentation, that is not enough.

## 8. <a id="project-goal"></a>Project goal

`mdtoc` is meant to be a small, reliable helper tool for Markdown documents that need:

* reproducible structure
* stable generated navigation
* deterministic heading numbering
* safe automation in CI

It is not meant to be a full Markdown processor or site generator.

## 9. <a id="in-one-sentence"></a>In one sentence

`mdtoc` exists because technical Markdown needs more than a ToC generator: it needs a deterministic structure manager.
