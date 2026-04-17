# 🧭 Overview: relevant tools

Tools that are very similar, with a focus on:

- CLI / CI suitability
- multi-platform support
- deterministic behavior
- Markdown → ToC / anchor logic

## Comparison table 1

| Tool                                       | Language | CLI / CI     | Anchor strategy                         | Special characteristics     | Link                                                                                            |
|--------------------------------------------|----------|--------------|-----------------------------------------|-----------------------------|-------------------------------------------------------------------------------------------------|
| **markdown-toc**                           | Node.js  | ✅            | configurable (`slugify`)                | very widespread, simple     | [Open GitHub repo](https://github.com/jonschlinkert/markdown-toc?utm_source=chatgpt.com)       |
| **md-toc**                                 | Python   | ✅            | renderer-compatible (GitHub/GitLab etc.) | deliberately standards-compliant | [Open GitHub repo](https://github.com/frnmst/md-toc?utm_source=chatgpt.com)               |
| **bitdowntoc**                             | Kotlin   | ✅            | profiles (GitHub, GitLab, etc.)         | idempotent + marker-based   | [Open GitHub repo](https://github.com/derlin/bitdowntoc?utm_source=chatgpt.com)                |
| **doctoc (Python variant)**                | Python   | ✅            | classic slug links                      | simple, CLI                 | [Open GitHub repo](https://github.com/ktechhub/doctoc?utm_source=chatgpt.com)                  |
| **Generic Markdown TOC Generator (Tools)** | various  | ⚠️ mostly web | GitHub-compatible slugs                 | more reference behavior     | [View example tool](https://toolsbox.io/code/markdown-toc-generator?utm_source=chatgpt.com)    |
| **MkDocs (indirect)**                      | Python   | ✅            | clean slugs                             | full pipeline approach      | [MkDocs info](https://en.wikipedia.org/wiki/MkDocs?utm_source=chatgpt.com)                     |

> What these tools have in common
> 
> All relevant tools:
> 
> parse `#` headings
> build a hierarchical ToC
> generate anchors deterministically
> ignore code blocks
> are CI-capable
> 
> 👉 Exactly your problem space.
> 
> 🔍 Important differences (crucial for your design)
> 1. Anchor strategy
> dumeng (your current state)
> almost no transformation
> not URL-safe
> inconsistent
> modern tools (standard)
> 
> 👉 almost all of them generate slugs
> 
> Typical:
> 
> Übergrößenträger → ubergrossentrager
> naïve façade → naive-facade
> 
> ➡️ URL-safe
> ➡️ stable
> ➡️ renderer-compatible
> 
> 2. Renderer compatibility
> 
> Some tools (for example `md-toc`) go even further:
> 
> 👉 they adapt behavior to the platform:
> 
> GitHub
> GitLab
> Bitbucket
> 
> 3. Idempotence (very important)
> 
> Best example:
> 
> 👉 bitdowntoc
> 
> detects an existing ToC
> overwrites it cleanly
> can run any number of times
> 
> 👉 exactly your goal.
> 
> 🧠 What changes for your project (without a dumeng constraint)
> 
> Now it gets interesting:
> 
> Option 1 — “State of the Art” (recommended)
> 
> 👉 You switch to slug-based anchors
> 
> Advantages
> compatible with:
> GitHub
> GitLab
> MkDocs
> stable links
> no broken links like now
> Option 2 — “Platform mode”
> 
> 👉 like `md-toc`:
> 
> mdtoc generate --anchor-style=github
> mdtoc generate --anchor-style=gitlab
> Option 3 — “Minimal (your approach)”
> 
> 👉 simple parser + dumeng-like
> 
> → but you wanted to move away from that
> 
> 🚀 My clear recommendation (for you)
> 
> 👉 Go with:
> 
> ✅ Default: slug-based anchors
> 
> Why:
> 
> standard in almost all tools
> simpler than dumeng (!) in the long run
> easier to test
> no VS Code issues
> 📐 Concrete target definition for `mdtoc`
> 
> You could define:
> 
> Default anchor-style = slug
> 
> Optional:
> 
> --anchor-style=dumeng
> --anchor-style=github
> 🧩 How your tool is positioned
> 
> Based on your current development, `mdtoc` would be:
> 
> 👉 a mix of:
> 
> bitdowntoc (idempotence + markers)
> markdown-toc (CLI simplicity)
> md-toc (renderer awareness)
> 
> That is a very good position.
> 
> ⚠️ Important point (design decision!)
> 
> I would write this question into the spec:
> 
> > Design decision:
> > By default, `mdtoc` uses slug-based anchor IDs,
> > because they are stable, URL-safe, and compatible with common Markdown renderers.
> 🧾 Short conclusion
> 
> You are now at a turning point:
> 
> Variant	Bewertung
> dumeng-compatible	simple, but error-prone
> slug-based	best practice
> renderer-specific	maximally compatible, more complex
> 
> 👉 I would clearly say:
> 
> Switch to slugs – that is the professional setup.

# Open-source alternatives to `mdtoc` for the Trice project

## Scope

This overview contains **only freely usable tools** that are suitable for OSS workflows and can fundamentally be used in **CLI/CI pipelines**.  
I did **not** include **UI** as a table attribute because all listed candidates are **CLI-first** and **do not** provide a classic desktop GUI.  
**dumeng-toc compatibility** was **no longer** a criterion.

## Important preliminary remarks

- **A `slug` is not automatically universally unique.** It is stable only if the algorithm, Unicode rules, and duplicate handling are precisely defined. That is exactly why parser-/renderer-aware tools such as **`md-toc`** or **`mdformat-toc`** are interesting.
- **Most mature tools solve only ToC + anchor behavior**, but **not** your full target picture of **ToC + heading numbering + strip + state/check**.
- If you want to stay **as close as possible to your planned `mdtoc`**, the following candidates are especially relevant:
  - **`kubernetes-sigs/mdtoc`**: very close to the idea of a small Go CLI
  - **`md-toc`**: strongest in renderer/slug compatibility
  - **`hoylen/markdown_toc`**: rare support for numbering, but a much smaller project

## Not in the main table, although thematically related

- **BitDownToc**: deliberately **not** included because the maintainer describes the **Common Clause** as a licensing decision in the repo; for a strict “free for OSS / open license” selection, that is not a clean candidate.
- **`gh-md-toc` (Shell)**: not in the main list because the project itself mainly mentions **Ubuntu/macOS** and explicitly refers Windows users to the Go implementation.
- **`remark-toc`**: strong in subject matter, but more of a **plugin building block** in the `remark` ecosystem than a direct, small `mdtoc` replacement with its own marker/check workflow.
- **`codegourmet/markdown-toc`**: not included because the author explicitly speaks of a **“quick and dirty writeup”**, **missing documentation**, and **missing tests**.

## Short ranking

1. **mdtoc (Kubernetes SIGs)** — 91/100
2. **doctoc** — 88/100
3. **md-toc** — 86/100
4. **mdformat-toc** — 82/100
5. **markdown-toc** — 78/100
6. **toc (ycd)** — 76/100
7. **markdown-toc-gen** — 74/100
8. **github-markdown-toc.go** — 69/100
9. **markdown_toc (Dart)** — 63/100

## Table 1 — tools horizontally, attributes vertically

| Attribute | mdtoc (Kubernetes SIGs) | doctoc | markdown-toc | md-toc | mdformat-toc | toc (ycd) | github-markdown-toc.go | markdown-toc-gen | markdown_toc (Dart) |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Recommendation | 91 | 88 | 78 | 86 | 82 | 76 | 69 | 74 | 63 |
| Features | Go CLI; in-place, dry-run, stdin, glob patterns; GFM-only; marker `<!-- toc --> ... <!-- /toc -->` | Node CLI; auto-update between comments; dry-run/stdout/update-only; suitable for git hooks; title/header/footer; min/max level | Node CLI and API; uses Remarkable; simple ToC generation; easy to integrate as a library | Python CLI/library; offline; marker-based; diff check; multiple parsers/profiles (GitHub, GitLab, CommonMark, Redcarpet …); pre-commit hook | Python plugin for `mdformat`; ToC at marker line; GitHub/GitLab slug function; optional HTML anchors; min/max level | Go CLI; `<!--toc-->` marker; optional bulleted/numbered list; skip/depth; stdout or append | Go CLI; GitHub-specific ToC generation; multiple files; parallel processing; local/remote files; token support | TypeScript/Node CLI; insert/update/dry-run/check; batch globs; GFM- and pandoc-compatible navigation; ignores code blocks | Dart CLI/package; generates ToC and numbered headings; binary or Dart usage; can remove ToC/numbering |
| Language | Go | JavaScript (Node.js) | JavaScript (Node.js) | Python | Python | Go | Go | TypeScript / Node.js | Dart |
| License | Apache-2.0 | MIT | MIT | GPL-3.0-or-later | MIT | Apache-2.0 | MIT | MIT | BSD-3-Clause |
| Internals | Single-purpose Go tool; normalized indentation; explicit tag-based update workflow | Comment pragmas for updates; marker-based in-place updates | Parser-/library-oriented; also generates slugs/JSON via API | Profile-/parser-aware ToC rules; reverse-engineering of renderer rules; many packaging channels | Plugin model on top of mdformat; HTML anchors plus GitHub slug by default; very small wheels | Zero-config approach; published binaries for Windows/macOS/Linux | Go port of `gh-md-toc`; no shell dependencies; internet-dependent | Small CLI with CI-oriented `check`; focus on Prettier-compatible output | Rare tool with numbering + ToC; pub package and compilable executable |
| Download | [Repo](https://github.com/kubernetes-sigs/mdtoc) / [Releases](https://github.com/kubernetes-sigs/mdtoc/releases) | [Repo](https://github.com/thlorenz/doctoc) / [npm](https://www.npmjs.com/package/doctoc) | [Repo](https://github.com/jonschlinkert/markdown-toc) / [npm](https://www.npmjs.com/package/markdown-toc) | [Repo](https://github.com/frnmst/md-toc) / [PyPI](https://pypi.org/project/md-toc/) | [Repo](https://github.com/hukkin/mdformat-toc) / [PyPI](https://pypi.org/project/mdformat-toc/) | [Repo](https://github.com/ycd/toc) / [Releases](https://github.com/ycd/toc/releases) | [Repo](https://github.com/ekalinin/github-markdown-toc.go) / [Releases](https://github.com/ekalinin/github-markdown-toc.go/releases) | [Repo](https://github.com/thesilk-tux/markdown-toc-gen) / [npm](https://www.npmjs.com/package/markdown-toc-gen) | [Repo](https://github.com/hoylen/markdown_toc) / [Pub](https://pub.dev/packages/markdown_toc) |
| Review | [DeepWiki Overview](https://deepwiki.com/kubernetes-sigs/mdtoc/1-overview) | [DeepWiki](https://deepwiki.com/thlorenz/doctoc) / [LibHunt comparison](https://www.libhunt.com/compare-doctoc-vs-vscode-markdown-pdf) | [LibHunt](https://www.libhunt.com/r/markdown-toc) / [Comparison](https://www.libhunt.com/compare-markdown-toc-vs-github-markdown-toc) | [LibHunt](https://www.libhunt.com/r/frnmst/md-toc) / [Docs](https://docs.franco.net.eu.org/md-toc/) | [PyPI project page](https://pypi.org/project/mdformat-toc/) / [mdformat plugin docs](https://mdformat.readthedocs.io/en/stable/users/plugins.html) | [LibHunt comparison](https://www.libhunt.com/compare-ycd--toc-vs-doctoc) | [Libraries.io](https://libraries.io/homebrew/githubmarkdowntoc) | [GitHub README](https://github.com/thesilk-tux/markdown-toc-gen) / [Release Notes](https://github.com/thesilk-tux/markdown-toc-gen/blob/master/RELEASE_NOTES.md) | [Pub package](https://pub.dev/packages/markdown_toc) / [API docs](https://pub.dev/documentation/markdown_toc/latest/) |
| Adoption | 47 GitHub stars; 8 releases; Kubernetes SIG release context | 4.4k GitHub stars; npm package with 83 dependents | 1.7k GitHub stars; npm reports 205 additional projects; official README names NASA/openmct, Prisma, Prettier, Docusaurus, among others | 38 GitHub stars; search snippet names 405 dependent repos; Debian/nix/Anaconda/PyPI badges | 27 GitHub stars; part of the mdformat ecosystem | 98 GitHub stars; 7 releases | 522 GitHub stars; 22 releases | 5 GitHub stars; 14 tags; npm package updated most recently in 2025 | 0 GitHub stars; pub metadata shows 23 downloads |
| OS | Go-based; Linux binary documented, `go install` available | multi-platform via Node.js; optional Docker workflow documented | multi-platform via Node.js | multi-platform via Python | multi-platform via Python >=3.9 | explicit binaries for Windows 32/64, macOS 64, Linux 32/64/ARM64 | explicitly cross-platform, including Windows/Mac/Linux | multi-platform via Node.js | Dart-based; usable as a Dart script or compiled executable |
| Lightweight | high – Go CLI, standalone binary documented | medium – npm global/local install; very easy in CI | medium – npm tool, but very small and easy to script | medium – Python runtime required, but offline and very flexible | high – PyPI wheel ~9.7 kB, but only really useful together with mdformat | very high – small Go tool with direct binary use | high – Go binary, no external Unix tools required | high – small npm tool, good CI commands | medium – small codebase, but Dart toolchain or own build required |
| LargeUsers | Kubernetes SIG Release / Kubernetes documentation environment | broad OSS usage; many repo/README workflows | NASA/openmct, Prisma, Prettier, Docusaurus, Joi, Mocha, among others | according to the project mainly blogs/docs/README workflows; 405 dependent repos in the badge/snippet | no prominently named individual projects; benefits from the mdformat ecosystem | no prominently named individual projects | strong GitHub README/wiki usage, but GitHub-centered | no prominently named individual projects | no prominently named individual projects |
| Verdict | Best direct replacement if you want a small, deterministic CLI with CI check; drawback: GFM only and no heading numbering. | Very robust ToC generator with a mature CLI; very good replacement if heading numbering is not mandatory. | Strong reference candidate for API/CLI and adoption; weaker as a replacement for strict marker/check workflows. | One of the best candidates for renderer-compatible slugs/links; excellent if behavior compatibility matters more than minimalism. | Technically clean and small; very good if you already use mdformat or want to couple formatting + ToC. | Slim and portable; good replacement for simple ToC workflows, but significantly simpler than your target picture. | Interesting if Trice strictly needs GitHub-compatible anchors; as a generic `mdtoc` replacement only a second choice because of internet dependence. | Very usable for modern CI pipelines; low adoption is the biggest drawback. | Highly relevant for your specification because of numbering, but due to very low adoption and documented code-block limitations it is not a worry-free standard replacement. |

## Table 2 — tools vertically, attributes horizontally

| Tool | Recommendation | Features | Language | License | Internals | Download | Review | Adoption | OS | Lightweight | LargeUsers | Verdict |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| mdtoc (Kubernetes SIGs) | 91 | Go CLI; in-place, dry-run, stdin, glob patterns; GFM-only; marker `<!-- toc --> ... <!-- /toc -->` | Go | Apache-2.0 | Single-purpose Go tool; normalized indentation; explicit tag-based update workflow | [Repo](https://github.com/kubernetes-sigs/mdtoc) / [Releases](https://github.com/kubernetes-sigs/mdtoc/releases) | [DeepWiki Overview](https://deepwiki.com/kubernetes-sigs/mdtoc/1-overview) | 47 GitHub stars; 8 releases; Kubernetes SIG release context | Go-based; Linux binary documented, `go install` available | high – Go CLI, standalone binary documented | Kubernetes SIG Release / Kubernetes documentation environment | Best direct replacement if you want a small, deterministic CLI with CI check; drawback: GFM only and no heading numbering. |
| doctoc | 88 | Node CLI; auto-update between comments; dry-run/stdout/update-only; suitable for git hooks; title/header/footer; min/max level | JavaScript (Node.js) | MIT | Comment pragmas for updates; marker-based in-place updates | [Repo](https://github.com/thlorenz/doctoc) / [npm](https://www.npmjs.com/package/doctoc) | [DeepWiki](https://deepwiki.com/thlorenz/doctoc) / [LibHunt comparison](https://www.libhunt.com/compare-doctoc-vs-vscode-markdown-pdf) | 4.4k GitHub stars; npm package with 83 dependents | multi-platform via Node.js; optional Docker workflow documented | medium – npm global/local install; very easy in CI | broad OSS usage; many repo/README workflows | Very robust ToC generator with a mature CLI; very good replacement if heading numbering is not mandatory. |
| markdown-toc | 78 | Node CLI and API; uses Remarkable; simple ToC generation; easy to integrate as a library | JavaScript (Node.js) | MIT | Parser-/library-oriented; also generates slugs/JSON via API | [Repo](https://github.com/jonschlinkert/markdown-toc) / [npm](https://www.npmjs.com/package/markdown-toc) | [LibHunt](https://www.libhunt.com/r/markdown-toc) / [Comparison](https://www.libhunt.com/compare-markdown-toc-vs-github-markdown-toc) | 1.7k GitHub stars; npm reports 205 additional projects; official README names NASA/openmct, Prisma, Prettier, Docusaurus | multi-platform via Node.js | medium – npm tool, but very small and easy to script | NASA/openmct, Prisma, Prettier, Docusaurus, Joi, Mocha, among others | Strong reference candidate for API/CLI and adoption; weaker as a replacement for strict marker/check workflows. |
| md-toc | 86 | Python CLI/library; offline; marker-based; diff check; multiple parsers/profiles (GitHub, GitLab, CommonMark, Redcarpet …); pre-commit hook | Python | GPL-3.0-or-later | Profile-/parser-aware ToC rules; reverse-engineering of renderer rules; many packaging channels | [Repo](https://github.com/frnmst/md-toc) / [PyPI](https://pypi.org/project/md-toc/) | [LibHunt](https://www.libhunt.com/r/frnmst/md-toc) / [Docs](https://docs.franco.net.eu.org/md-toc/) | 38 GitHub stars; search snippet names 405 dependent repos; Debian/nix/Anaconda/PyPI badges | multi-platform via Python | medium – Python runtime required, but offline and very flexible | according to the project mainly blogs/docs/README workflows; 405 dependent repos in the badge/snippet | One of the best candidates for renderer-compatible slugs/links; excellent if behavior compatibility matters more than minimalism. |
| mdformat-toc | 82 | Python plugin for `mdformat`; ToC at marker line; GitHub/GitLab slug function; optional HTML anchors; min/max level | Python | MIT | Plugin model on top of mdformat; HTML anchors plus GitHub slug by default; very small wheels | [Repo](https://github.com/hukkin/mdformat-toc) / [PyPI](https://pypi.org/project/mdformat-toc/) | [PyPI project page](https://pypi.org/project/mdformat-toc/) / [mdformat plugin docs](https://mdformat.readthedocs.io/en/stable/users/plugins.html) | 27 GitHub stars; part of the mdformat ecosystem | multi-platform via Python >=3.9 | high – PyPI wheel ~9.7 kB, but only really useful together with mdformat | no prominently named individual projects; benefits from the mdformat ecosystem | Technically clean and small; very good if you already use mdformat or want to couple formatting + ToC. |
| toc (ycd) | 76 | Go CLI; `<!--toc-->` marker; optional bulleted/numbered list; skip/depth; stdout or append | Go | Apache-2.0 | Zero-config approach; published binaries for Windows/macOS/Linux | [Repo](https://github.com/ycd/toc) / [Releases](https://github.com/ycd/toc/releases) | [LibHunt comparison](https://www.libhunt.com/compare-ycd--toc-vs-doctoc) | 98 GitHub stars; 7 releases | explicit binaries for Windows 32/64, macOS 64, Linux 32/64/ARM64 | very high – small Go tool with direct binary use | no prominently named individual projects | Slim and portable; good replacement for simple ToC workflows, but significantly simpler than your target picture. |
| github-markdown-toc.go | 69 | Go CLI; GitHub-specific ToC generation; multiple files; parallel processing; local/remote files; token support | Go | MIT | Go port of `gh-md-toc`; no shell dependencies; internet-dependent | [Repo](https://github.com/ekalinin/github-markdown-toc.go) / [Releases](https://github.com/ekalinin/github-markdown-toc.go/releases) | [Libraries.io](https://libraries.io/homebrew/githubmarkdowntoc) | 522 GitHub stars; 22 releases | explicitly cross-platform, including Windows/Mac/Linux | high – Go binary, no external Unix tools required | strong GitHub README/wiki usage, but GitHub-centered | Interesting if Trice strictly needs GitHub-compatible anchors; as a generic `mdtoc` replacement only a second choice because of internet dependence. |
| markdown-toc-gen | 74 | TypeScript/Node CLI; insert/update/dry-run/check; batch globs; GFM- and pandoc-compatible navigation; ignores code blocks | TypeScript / Node.js | MIT | Small CLI with CI-oriented `check`; focus on Prettier-compatible output | [Repo](https://github.com/thesilk-tux/markdown-toc-gen) / [npm](https://www.npmjs.com/package/markdown-toc-gen) | [GitHub README](https://github.com/thesilk-tux/markdown-toc-gen) / [Release Notes](https://github.com/thesilk-tux/markdown-toc-gen/blob/master/RELEASE_NOTES.md) | 5 GitHub stars; 14 tags; npm package updated most recently in 2025 | multi-platform via Node.js | high – small npm tool, good CI commands | no prominently named individual projects | Very usable for modern CI pipelines; low adoption is the biggest drawback. |
| markdown_toc (Dart) | 63 | Dart CLI/package; generates ToC and numbered headings; binary or Dart usage; can remove ToC/numbering | Dart | BSD-3-Clause | Rare tool with numbering + ToC; pub package and compilable executable | [Repo](https://github.com/hoylen/markdown_toc) / [Pub](https://pub.dev/packages/markdown_toc) | [Pub package](https://pub.dev/packages/markdown_toc) / [API docs](https://pub.dev/documentation/markdown_toc/latest/) | 0 GitHub stars; pub metadata shows 23 downloads | Dart-based; usable as a Dart script or compiled executable | medium – small codebase, but Dart toolchain or own build required | no prominently named individual projects | Highly relevant for your specification because of numbering, but due to very low adoption and documented code-block limitations it is not a worry-free standard replacement. |

## Reading key for the recommendation level

- **90–100**: very good replacement candidate
- **80–89**: strong, but with a clear limitation
- **70–79**: usable if the tool profile matches your workflow
- **60–69**: interesting only for special edge cases
- **<60**: more of an inspiration source than a real replacement

## My conclusion

If you are looking for **a small, clean, cross-platform CLI reference**, **`kubernetes-sigs/mdtoc`** is the best reference.  
If you prioritize **renderer/slug compatibility**, **`frnmst/md-toc`** is the most interesting in terms of content.  
If **heading numbering** is important to you, **`hoylen/markdown_toc`** is relevant, but more as a source of ideas than as a standard tool.
