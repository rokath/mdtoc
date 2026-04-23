# `mdtoc` Alternatives

This page provides a compact, reader-facing overview of alternative tools relative to `mdtoc`.

The focus here is quick comparison against `mdtoc`, not a full review of each project.

## Reading guide

Use this table for fast selection:

* closest small-CLI reference: `mdtoc (Kubernetes SIGs)`
* strongest renderer-compatibility reference: `md-toc`
* strongest adoption among simple ToC tools: `doctoc` and `markdown-toc`
* closest match if numbered headings matter: `markdown_toc (Dart)`

## Selection scope

This overview intentionally focuses on tools that are suitable for open-source documentation workflows and meaningful as CLI or CI references.

* included by default: tools with a real CLI workflow or a directly relevant formatter/plugin workflow
* not treated as a main criterion: desktop UI support
* not treated as a main criterion: compatibility with older `dumeng`-style anchor behavior

## Comparison notes

* Most mature alternatives solve ToC generation and anchor behavior, but not the full `mdtoc` target picture of ToC, heading numbering, strip, and state/check handling in one tool.
* A `slug` is not automatically unique or portable across platforms; practical stability depends on the exact renderer rules, Unicode handling, and duplicate-heading behavior.
* Renderer-aware tools such as `md-toc` and `mdformat-toc` are therefore especially relevant when GitHub or GitLab compatibility matters more than minimal tool size.
* `Ignore regions` in the table only means Markdown regions that are skipped while detecting headings or building the ToC.
* `Exclusion regions` only means explicit user-controlled opt-out regions beyond ordinary Markdown parsing.
* `Idempotent` only means that repeating the tool on already managed input is intended to produce the same managed result.

## Short glossary

| Term | Meaning |
| --- | --- |
| `CI` | Continuous Integration, meaning automated checks in a repository or build workflow. |
| `Documented` | Explicitly described in the compared material as a supported feature or behavior. |
| `Explicit exclusion` | A user-controlled region that is intentionally left out of ToC or heading processing. |
| `GFM` | GitHub Flavored Markdown, the Markdown dialect used by GitHub. |
| `git hook` | An automatically triggered Git action, for example before a commit is created. |
| `in-place update` | Modifies the existing file directly instead of only writing output to `stdout`. |
| `marker-based` | Updates content inside explicit marker comments in a Markdown file. |
| `Not documented` | Not stated clearly enough in the currently compared source material to claim support. |
| `parser-defined` | Determined by the Markdown parser or formatter behavior instead of by a small explicit region list in this document. Example: fenced code blocks handled through parser logic. |
| `Partial` | Supported in a limited or narrower sense than the full `mdtoc` feature. |
| `pre-commit` | A framework and workflow for running checks before a commit is created. |
| `renderer-aware` | Behavior intentionally matched to a specific Markdown renderer such as GitHub or GitLab. |
| `SIGs` | Special Interest Groups, here meaning subgroups within the Kubernetes project. |
| `slug` | A normalized anchor or URL fragment derived from heading text. |
| `stdout` | Standard output, meaning terminal output that can be piped or redirected. |

## Comparison overview

| Tool | License | Primary scope | Language | Runtime / install model | ToC update model | Anchor / renderer focus | Numbered headings | Repeated headings | Selectable bullets | Ignore regions | Exclusion regions | Idempotent | CI / check workflow | Adoption signal | Platform reach | Relative strengths vs `mdtoc` | Relative limits vs `mdtoc` | Links |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| [mdtoc](https://github.com/rokath/mdtoc) | MIT | Managed Markdown CLI for ToC, heading numbering, strip, regen, and check workflows | Go | Standalone Go binary or `go install`, Homebrew | Explicit managed container with generate, strip, regen, and check commands | GitHub and GitLab or custom anchor profiles, or anchors off | Yes | Yes | Yes | Documented: code fences, inline code spans, HTML comments | Explicit | Yes | Strong CLI/CI fit with explicit state validation | This project itself; primary comparison baseline | Cross-platform Go CLI with release binaries and package artifacts | Combines managed ToC, numbered headings, strip/regen, state checking, selectable bullets, and explicit exclusion regions in one focused tool | Newer and smaller than the most established alternatives; more specialized than general formatter ecosystems | [Repo](https://github.com/rokath/mdtoc), [Releases](https://github.com/rokath/mdtoc/releases) |
| [mdtoc (Kubernetes SIGs)](https://github.com/kubernetes-sigs/mdtoc) | Apache-2.0 | Small CLI for managed Markdown ToCs | Go | Standalone Go binary or `go install` | Marker-based in-place update, dry-run, stdin, glob patterns | GitHub-Flavored Markdown focused | No | Yes | Not documented | Documented: code fences | Not documented | Yes | Good CLI/CI fit | Smaller project, but tied to Kubernetes SIG release workflows | Cross-platform Go install, Linux binary documented | Closest overall profile to `mdtoc`; small Go CLI; deterministic workflow | GFM-only; no heading numbering; less explicit control over exclusions and bullets | [Repo](https://github.com/kubernetes-sigs/mdtoc), [Releases](https://github.com/kubernetes-sigs/mdtoc/releases) |
| [doctoc](https://github.com/thlorenz/doctoc) | MIT | Mature CLI for inserting and updating ToCs | Node.js | npm package, optional Docker workflow | Comment-marker update with stdout/update modes | Classic slug links, practical GitHub-style usage | No | Yes | Not documented | Documented: code fences | Not documented | Yes | Good CLI/git-hook/CI fit | High OSS adoption and long-lived npm presence | Multi-platform through Node.js | Mature ecosystem adoption; flexible update modes; easy repository integration | Focused on ToC generation rather than `mdtoc`'s wider managed-state model | [Repo](https://github.com/thlorenz/doctoc), [npm](https://www.npmjs.com/package/doctoc) |
| [markdown-toc](https://github.com/jonschlinkert/markdown-toc) | MIT | CLI and library for Markdown ToC generation | Node.js | npm package | Simple ToC generation; library/API-oriented workflows | Configurable slug generation via `slugify` | No | Yes | Not documented | Documented: code fences | Not documented | Yes | Usable in CI, but less check-oriented | Broad usage and strong package ecosystem visibility | Multi-platform through Node.js | Strong adoption; easy to embed as a library; flexible slug handling | Weaker fit for explicit marker/check/state workflows | [Repo](https://github.com/jonschlinkert/markdown-toc), [npm](https://www.npmjs.com/package/markdown-toc) |
| [md-toc](https://github.com/frnmst/md-toc) | GPL-3.0-or-later | Renderer-aware ToC CLI and library | Python | PyPI and multiple packaging channels | Marker-based update plus diff/check workflow | Strong profile support for GitHub, GitLab, CommonMark, Redcarpet, and more | No | Yes | Not documented | Documented, parser-defined | Not documented | Yes | Strong CI/pre-commit/check support | Smaller project, but unusually rich renderer-compatibility scope | Multi-platform through Python packaging | Best renderer-compatibility reference; explicit profile handling; offline | Python runtime; no heading numbering; broader tool surface than `mdtoc` | [Repo](https://github.com/frnmst/md-toc), [PyPI](https://pypi.org/project/md-toc/) |
| [mdformat-toc](https://github.com/hukkin/mdformat-toc) | MIT | ToC plugin for the `mdformat` ecosystem | Python | PyPI plugin on top of `mdformat` | Marker-line insertion within a formatter pipeline | GitHub/GitLab slug function, optional HTML anchors | No | Yes | Not documented | Documented, parser-defined | Not documented | Yes | Good if formatting is already CI-managed | Modest standalone adoption, but benefits from the `mdformat` ecosystem | Multi-platform through Python >=3.9 | Clean and lightweight in formatter-based workflows | Not a standalone `mdtoc`-style CLI workflow; depends on `mdformat` | [Repo](https://github.com/hukkin/mdformat-toc), [PyPI](https://pypi.org/project/mdformat-toc/) |
| [toc (ycd)](https://github.com/ycd/toc) | Apache-2.0 | Very small CLI for simple ToC insertion | Go | Published Go binaries | `<!--toc-->` marker with stdout or append modes | Simple anchor behavior; less renderer-focused | No | Not documented | Yes | Not documented | Not documented | Partial | Basic CLI/CI fit | Small utility project with modest adoption | Explicit binaries for Windows, macOS, and Linux | Very lightweight and portable; straightforward binary distribution | Considerably simpler scope than `mdtoc`; less explicit compatibility focus | [Repo](https://github.com/ycd/toc), [Releases](https://github.com/ycd/toc/releases) |
| [github-markdown-toc.go](https://github.com/ekalinin/github-markdown-toc.go) | MIT | GitHub-specific ToC generation | Go | Standalone Go binary | File and remote-source processing | GitHub-specific anchor behavior | No | Yes | Not documented | Documented: GitHub-rendered heading model | Not documented | Partial | Usable in CI for GitHub-centric docs | Well-known niche tool in GitHub README workflows | Explicit cross-platform binary story | Strong GitHub focus; no shell dependency; cross-platform binaries | Internet-dependent workflow; narrower use case than generic `mdtoc` document management | [Repo](https://github.com/ekalinin/github-markdown-toc.go), [Releases](https://github.com/ekalinin/github-markdown-toc.go/releases) |
| [markdown-toc-gen](https://github.com/thesilk-tux/markdown-toc-gen) | MIT | Modern CLI for insert/update/check workflows | TypeScript / Node.js | npm package | Insert, update, dry-run, and check commands | GFM- and Pandoc-compatible navigation | No | Yes | Not documented | Documented: code blocks | Not documented | Yes | Good check-oriented CI fit | Low adoption so far, but feature set is CI-friendly | Multi-platform through Node.js | Close to modern CI usage; explicit check command; ignores code blocks | Lower adoption; no heading numbering; Node.js runtime | [Repo](https://github.com/thesilk-tux/markdown-toc-gen), [npm](https://www.npmjs.com/package/markdown-toc-gen) |
| [markdown_toc (Dart)](https://github.com/hoylen/markdown_toc) | BSD-3-Clause | ToC plus numbered-heading generation | Dart | Dart package or compiled executable | Generates and can remove ToC / numbering | Basic Markdown heading handling | Yes | Not documented | Not documented | Not documented | Not documented | Not documented | CLI-capable, but smaller ecosystem footprint | Very low adoption and a comparatively small ecosystem | Dart-based portability, but less common in docs toolchains | Rare alternative that also covers heading numbering | Very low adoption; narrower ecosystem; less established as a standard CLI reference | [Repo](https://github.com/hoylen/markdown_toc), [Pub](https://pub.dev/packages/markdown_toc) |

## Related but excluded tools

These tools are related, but they are not in the main table because they are a weaker fit for a primary `mdtoc` alternatives overview.

| Tool | Why it is excluded from the main table |
| --- | --- |
| `BitDownToc` | Excluded because its licensing situation was intentionally treated as not clean enough for a strict OSS-only alternatives list. |
| `gh-md-toc` (Shell) | Excluded because its own project positioning is more Unix-shell-oriented and it is less suitable as a cross-platform baseline than the Go alternatives. |
| `remark-toc` | Excluded because it is better understood as a plugin building block inside the `remark` ecosystem than as a direct standalone `mdtoc` replacement. |
| `codegourmet/markdown-toc` | Excluded because the project presents itself as incomplete and lacks the level of documentation and testing expected for a primary reference. |
