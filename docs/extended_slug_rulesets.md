# mdtoc Slug Rulesets: Profiles, Sub-Profiles, and Implementation Fixtures

Date: 2026-05-09  
Version: v4-en

This document describes the **21 distinct base profiles** that `mdtoc` should support as independent slug rulesets. Every profile name and every alias written in backticks is intended as a future value accepted by `slug=`. In addition, this document defines **sub-profiles** such as `github-anyascii`: these are mdtoc-derived combinations of a base profile plus a Unicode/transliteration modifier.

## 1. Basic model

`mdtoc` should treat slug generation as a pipeline:

1. **Parse the Markdown heading.** This step handles ATX closing hashes, optionally inline markup, and optionally manual ID attributes.
2. **Determine the effective heading text.** Example: `#### Closed␠␠ATX␠␠␠####` becomes the title text `Closed␠␠ATX` for slug generation. The three spaces directly before `####` belong to the closing sequence and must not create an additional separator.
3. **Apply the profile rules.** This includes Unicode handling, lowercasing, punctuation, separators, and collapse rules.
4. **Resolve collisions.** Most profiles use `-1`, `-2`; Python-Markdown stacks use `_1`, `_2`.

Important: CommonMark/GFM does not normatively define automatic heading IDs. Heading slugs are behavior of the respective renderer or plugin.

## 2. `slug=` naming strategy

### 2.1 Base profiles

The 21 base profiles are the visible primary values of `slug=`. Aliases are listed only in the profile sections, not in the comparison matrices.

### 2.2 Sub-profiles

Sub-profiles are formed with this pattern:

```text
slug=<canonical-profile>-<modifier>
```

Examples:

```text
slug=github
slug=github-anyascii
slug=github-strip
slug=gitlab-current-anyascii
slug=pandoc-unicode
```

A sub-profile is **not an exact upstream emulation** anymore. `github-anyascii` means: GitHub-like punctuation, whitespace, and duplicate rules, but Unicode is first converted to ASCII with AnyAscii. Internally, `mdtoc` should not implement this as a new slugger, but as a composition:

```text
base_profile = github
unicode_modifier = anyascii
```

Recommended modifiers:

| Modifier | Meaning | Implementation status |
| --- | --- | --- |
| `-unicode` | Preserve Unicode, even when the base profile is otherwise ASCII-oriented. | recommended |
| `-strip` | Remove non-ASCII code points before applying the slug rule. | recommended |
| `-anyascii` | Transliterate Unicode to ASCII with AnyAscii. | recommended; generic default for new ASCII sub-profiles |
| `-unidecode` | Transliterate Unicode to ASCII with Unidecode/Text::Unidecode. | optional; the version must be pinned |
| `-icu` | ICU/CLDR transliteration, for example `Any-Latin; Latin-ASCII`. | optional/expert; pin ICU/CLDR version and transform ID |

Aliases should **not automatically** be combined with suffixes. `github-slugger` is an alias for `github`; `github-slugger-anyascii` should only be accepted if it is explicitly registered. This keeps the visible surface area small.

## 3. Transliteration: rules, sources, and implementation

For Unicode-preserving profiles such as `github`, `gitlab-current`, `pandoc`, `crossnote`, and `mdbook`, **no transliteration** is normally correct. Transliteration is relevant only for ASCII profiles or mdtoc sub-profiles.

### 3.1 Directly recommended implementation rules

1. `-unicode`: no transliteration and no ASCII reduction; then apply the normal base profile rule.
2. `-strip`: remove all code points `> U+007F`; then apply the normal base profile rule.
3. `-anyascii`: apply a pinned AnyAscii table; then apply the normal base profile rule. AnyAscii works character-by-character and without context, but covers a very large number of Unicode blocks and is available for several languages.
4. `-unidecode`: only as an optional expert profile; Unidecode is lossy, context-free, and can produce linguistically unsuitable results.
5. `-icu`: only with a fixed transform ID. The recommended default would be `Any-Latin; Latin-ASCII`, but without locale data ICU remains only an approximation for CJK and ambiguous characters as well.

### 3.2 Known transliteration tables and rulesets

| Name | Link | Role for mdtoc |
| --- | --- | --- |
| AnyAscii | https://github.com/anyascii/anyascii | Recommended generic `-anyascii` modifier. |
| Unidecode / Text::Unidecode | https://pypi.org/project/Unidecode/ | Optional `-unidecode` modifier; pin the version. |
| ICU Transforms | https://unicode-org.github.io/icu/userguide/transforms/general/ | Optional `-icu` modifier; pin transform ID and ICU/CLDR version. |
| Unicode CLDR Transliteration Guidelines | https://cldr.unicode.org/index/cldr-spec/transliteration-guidelines | Normative background source for ICU/CLDR transforms. |
| python-slugify / text-unidecode | https://pypi.org/project/python-slugify/ | Common Python slugify variant; mainly a comparison source. |
| npm `transliteration` | https://github.com/yf-hk/transliteration | JavaScript ecosystem; evaluate as a possible later dependency. |
| kramdown/Stringex | https://github.com/rsl/stringex | Relevant for `kramdown-transliterated`. |
| ISO 9 | https://www.iso.org/standard/3589.html | Cyrillic-to-Latin; standardized, but not universal. |
| ALA-LC Romanization Tables | https://www.loc.gov/catdir/cpso/roman.html | Library/cataloging rules for many scripts. |
| BGN/PCGN Romanization Systems | https://www.gov.uk/government/publications/romanization-systems | Geographic names; not generic for Markdown. |
| Hanyu Pinyin | https://www.iso.org/standard/61420.html | Mandarin Chinese; not generic without language segmentation. |
| Hepburn Romanization | https://www.loc.gov/catdir/cpso/romanization/japanese.pdf | Japanese; language-specific, not a generic slugger. |

For `mdtoc`, the initial set is enough: `-unicode`, `-strip`, `-anyascii`. `-unidecode` and `-icu` should only be implemented when concrete user requirements arise.

## 4. Comparison matrix of the 21 base profiles

The matrix contains only comparison values. Sources and aliases are listed in the profile sections.

| # | Profile | Non-ASCII | Case | Whitespace | Punctuation | `-`/`_`/`.` | Duplicates | Manual ID | Why distinct? |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | `github` | preserved | Unicode lowercase | literal Space -> `-`; no collapse | remove punctuation | preserve `-`, `_`; remove `.` | `-1` | no `{#id}` attributes; HTML anchors separately | GitHub is the most important de facto profile for README and GFM-compatible anchors. |
| 2 | `gitlab-current` | preserved | lowercase | Spaces -> `-`; no collapse in the current example | remove non-word/punctuation | preserve `-`, `_`; remove `.` | `-1` | no documented `{#id}` attributes | GitLab 17+ changed heading ID generation; the current example preserves repeated `-`. |
| 3 | `gitlab-legacy` | preserved | lowercase | collapse separators | remove / separator | collapse repeated `-` | `-1` | no documented `{#id}` attributes | This profile is relevant only for old GitLab installations and editor compatibility. |
| 4 | `crossnote` | preserved | lowercase | internally Whitespace -> `~`; then `~` -> `-` | remove | literal `-` runs collapse; `_` remains | `-1` | `{#id}` attributes possible in Mume/Crossnote environments | Crossnote/Mume uses custom preprocessing plus `uslug`; it is neither GitHub nor plain `uslug`. |
| 5 | `pandoc` | preserved | lowercase | Pandoc AST normalizes whitespace; Spaces/Newlines -> `-` | remove except `_`, `-`, `.` | preserve `.` | `-1` | `{#id}` wins | Pandoc is normatively documented and preserves periods, unlike GitHub. |
| 6 | `pandoc-gfm` | preserved | lowercase | Pandoc/GFM; whitespace normalized from AST | remove except `-`, `_`; emoji names | remove `.` | `-1` | no `{#id}` attributes in the GFM reader | Pandoc-GFM is close to GitHub, but not identical because of the Pandoc parser and emoji handling. |
| 7 | `pandoc-ascii` | ASCII-reduced | lowercase | like `pandoc` | like `pandoc` | preserve `.` | `-1` | `{#id}` wins | `ascii_identifiers` handles non-Latin headings fundamentally differently from Unicode-preserving profiles. |
| 8 | `kramdown` | preserved | lowercase | non-letters/non-digits -> `-` | separator/remove | `.`/`_` are not preserved as word characters | `-1` | `{#id}` wins | Important for Ruby/Jekyll/kramdown compatibility and because of the `section` fallback. |
| 9 | `kramdown-transliterated` | transliterated to ASCII | lowercase | like `kramdown` | like `kramdown` | like `kramdown` | `-1` | `{#id}` wins | kramdown can create ASCII-only IDs; mdtoc needs a pinned transliteration for this. |
| 10 | `blackfriday` | preserved | lowercase | invalid runs -> one `-` | separator | collapse runs | `-1` | heading-ID extension possible | Relevant for old Hugo/Go/Gitea stacks. |
| 11 | `hugo-github-ascii` | ASCII-reduced | lowercase | GitHub/Goldmark-like | GitHub/Goldmark-like | GitHub/Goldmark-like | `-1` | Goldmark heading attributes possible | Hugo provides `github-ascii` in addition to `github`; it is no longer GitHub emulation. |
| 12 | `python-markdown` | ASCII-reduced | lowercase | separator runs -> one `-` | remove | preserve `_`; remove `.` | `_1` | `attr_list` optional | Important for Python-Markdown and MkDocs; duplicates use `_1` instead of `-1`. |
| 13 | `python-markdown-unicode` | preserved | lowercase | separator runs -> one `-` | remove | preserve `_`; remove `.` | `_1` | `attr_list` optional | Python-Markdown provides a Unicode-preserving slugify variant. |
| 14 | `pymdownx` | preserved/configurable | configurable; mdtoc default lowercase | configurable; here Spaces -> `-` | remove | preserve `-`, `_`; remove `.` | typically `_1` in the Python-Markdown stack | `attr_list`/TOC stack | PyMdownX is a configurable slug family; mdtoc needs a stable default. |
| 15 | `markdown-it-anchor` | percent-encoded | lowercase | whitespace runs -> one `-` | often remains as `%XX` | preserve `.`; `+` -> `%2B` | `-1` | no `{#id}` in plugin default | Many JavaScript documentation systems use markdown-it-anchor or derivatives. |
| 16 | `vscode` | percent-encoded | lowercase | whitespace runs -> one `-` | remove defined punctuators | remove `_`; trim `-` | `-1` | no `{#id}` in outline slugger | VS Code generates its own UI/outline anchors and is often perceived as the expected editor behavior. |
| 17 | `azure-devops` | preserved/renderer-dependent | lowercase | spaces and many special characters -> `-` | mostly separator | repeated `-` possible | `-1` | HTML anchors possible; `{#id}` not documented as a core rule | Azure DevOps often treats punctuation as separators and differs clearly from GitHub. |
| 18 | `bitbucket-cloud` | preserved | lowercase | GitHub/legacy-like, usually collapsed | remove | prefix `markdown-header-` | `-1` | not documented as a `{#id}` core rule | The `markdown-header-` prefix is slug-semantic and requires a separate profile. |
| 19 | `zola-on` | ASCII/transliterated | lowercase | whitespace -> `-` | slugified/removed | ASCII slug | `-1` | check Zola/Markdown attribute support separately | Zola slugifies anchors by default; Unicode is not preserved like GitHub. |
| 20 | `zola-safe` | preserved | preserve case | spaces -> `_` | mostly safely preserved | preserve `_`, `-`, `.` | `-1` | check Zola/Markdown attribute support separately | Zola `safe`/`off` uses a different URL-stability model: spaces become `_`, not `-`. |
| 21 | `mdbook` | preserved | lowercase; verify Unicode case against Rust version | spaces -> `-`; no GitHub dash-collapse rule | remove | preserve `-`, `_`; remove `.` | `-1` | check custom heading attributes depending on mdBook/renderer version | mdBook has its own Rust implementation and is widely used for Rust documentation. |

## 5. Test fixtures and expectation matrix

`␠` marks visible spaces in this document. The real fixture does not use this character; it is only for readability. Values with `†` are **pinned mdtoc target behavior** or source-derived values that should be confirmed with a live/source test before final implementation.

### 5.1 Fixture legend

| ID | Heading | Meaning |
| --- | --- | --- |
| F1 | `# A+B` | Punctuation without whitespace: remove, separator, or percent-encode? |
| F2 | `# Version 3.5` | Period handling. |
| F3 | `# Привет 你好 & TEST` | Unicode, case folding, punctuation between spaces. |
| F4 | `# foo␠␠bar\-\-\-baz` | Two spaces plus three literal hyphens; backslashes prevent smart punctuation. |
| F5 | `# Manual Über 汉字 {#Fix_ID-42}` | Pandoc/kramdown/Goldmark-style manual ID attribute; otherwise literal text. |
| F6 | `#### Closed␠␠ATX␠␠␠####` | Closed ATX: two spaces in the title, three spaces before closing hashes; the closing sequence is not part of the title. |
| F7 | second `# Repeat` | Duplicate handling; the first slug would be `repeat` or profile-dependent `Repeat`. |
| F8 | `# 123` | Purely numeric heading text and fallback rules. |

### 5.2 Expected values by profile

| Profile | F1<br>`# A+B` | F2<br>`# Version 3.5` | F3<br>`# Привет 你好 & TEST` | F4<br>`# foo␠␠bar\-\-\-baz` | F5<br>`# Manual Über 汉字 {#Fix_ID-42}` | F6<br>`#### Closed␠␠ATX␠␠␠####` | F7<br>second `# Repeat` | F8<br>`# 123` |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `github` | `ab` | `version-35` | `привет-你好--test` | `foo--bar---baz` | `manual-über-汉字-fix_id-42` | `closed--atx` | `repeat-1` | `123` |
| `gitlab-current` | `ab` | `version-35` | `привет-你好--test` | `foo--bar---baz` | `manual-über-汉字-fix_id-42` | `closed--atx` | `repeat-1` | `123` |
| `gitlab-legacy` | `ab` | `version-35` | `привет-你好-test` | `foo-bar-baz` | `manual-über-汉字-fix_id-42` | `closed-atx` | `repeat-1` | `anchor-123` |
| `crossnote` | `ab` | `version-35` | `привет-你好--test` | `foo--bar-baz` | `Fix_ID-42` | `closed--atx` | `repeat-1` | `123` |
| `pandoc` | `ab` | `version-3.5` | `привет-你好-test` | `foo-bar---baz` | `Fix_ID-42` | `closed-atx` | `repeat-1` | `section` |
| `pandoc-gfm` | `ab` | `version-35` | `привет-你好--test` | `foo-bar---baz` | `manual-über-汉字-fix_id-42` | `closed-atx` | `repeat-1` | `123` |
| `pandoc-ascii` | `ab` | `version-3.5` | `test` | `foo-bar---baz` | `Fix_ID-42` | `closed-atx` | `repeat-1` | `section` |
| `kramdown` | `ab` | `version-35` | `привет-你好-test` | `foo-bar-baz` | `Fix_ID-42` | `closed-atx` | `repeat-1` | `section` |
| `kramdown-transliterated` | `ab` | `version-35` | `privet-ni-hao-test` † | `foo-bar-baz` | `Fix_ID-42` | `closed-atx` | `repeat-1` | `section` |
| `blackfriday` | `a-b` | `version-3-5` | `привет-你好-test` | `foo-bar-baz` | `Fix_ID-42` | `closed-atx` | `repeat-1` | `123` |
| `hugo-github-ascii` | `ab` | `version-35` | `test` † | `foo--bar---baz` | `Fix_ID-42` | `closed--atx` | `repeat-1` | `123` |
| `python-markdown` | `ab` | `version-35` | `test` | `foo-bar-baz` | `Fix_ID-42` † | `closed-atx` | `repeat_1` | `123` |
| `python-markdown-unicode` | `ab` | `version-35` | `привет-你好-test` | `foo-bar-baz` | `Fix_ID-42` † | `closed-atx` | `repeat_1` | `123` |
| `pymdownx` | `ab` | `version-35` | `привет-你好--test` | `foo--bar---baz` | `Fix_ID-42` † | `closed--atx` | `repeat_1` | `123` |
| `markdown-it-anchor` | `a%2Bb` | `version-3.5` | `%D0%BF%D1%80%D0%B8%D0%B2%D0%B5%D1%82-%E4%BD%A0%E5%A5%BD-%26-test` | `foo-bar---baz` | `manual-%C3%BCber-%E6%B1%89%E5%AD%97-%7B%23fix_id-42%7D` | `closed-atx` | `repeat-1` | `123` |
| `vscode` | `ab` | `version-35` | `%D0%BF%D1%80%D0%B8%D0%B2%D0%B5%D1%82-%E4%BD%A0%E5%A5%BD--test` | `foo-bar---baz` | `manual-%C3%BCber-%E6%B1%89%E5%AD%97-fixid-42` | `closed-atx` | `repeat-1` | `123` |
| `azure-devops` | `a-b` | `version-3-5` | `привет-你好--test` † | `foo--bar---baz` | `manual-über-汉字--fix-id-42` † | `closed--atx` | `repeat-1` | `123` |
| `bitbucket-cloud` | `markdown-header-ab` | `markdown-header-version-35` | `markdown-header-привет-你好-test` | `markdown-header-foo-bar-baz` | `markdown-header-manual-über-汉字-fix_id-42` | `markdown-header-closed-atx` | `markdown-header-repeat-1` | `markdown-header-123` |
| `zola-on` | `ab` | `version-35` | `privet-ni-hao-test` † | `foo-bar-baz` | `Fix_ID-42` † | `closed-atx` | `repeat-1` | `123` |
| `zola-safe` | `A+B` | `Version_3.5` | `Привет_你好_&_TEST` | `foo__bar---baz` | `Fix_ID-42` † | `Closed__ATX` | `Repeat-1` | `123` |
| `mdbook` | `ab` | `version-35` | `привет-你好--test` † | `foo--bar---baz` | `Fix_ID-42` † | `closed--atx` | `repeat-1` | `123` |

## 6. Profile sections

### 1. GitHub heading anchors - `github` (aliases: `gfm`, `github-slugger`, `github-heading`, `github-readme`)

**Why this profile exists:** GitHub is the most important de facto profile for README and GFM-compatible anchors.

**Short rule:** Lowercase, remove GitHub punctuation, convert normal spaces to `-`, preserve Unicode, do not collapse repeated spaces or existing repeated `-`, duplicates with `-1`.

**Manual IDs:** no `{#id}` attributes; HTML anchors separately.

**Sources:**
- GitHub Docs: Section links and custom anchors: https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax
- github-slugger source: https://raw.githubusercontent.com/Flet/github-slugger/master/index.js

### 2. GitLab current heading IDs - `gitlab-current` (aliases: `gitlab`, `glfm`, `gitlab-17`, `gitlab-current-heading`)

**Why this profile exists:** GitLab 17+ changed heading ID generation; the current example preserves repeated `-`.

**Short rule:** Lowercase, remove non-word/punctuation, convert spaces to `-`, do not collapse repeated `-`; preserve Unicode; duplicates with `-1`. The GitLab rule text about dash collapse currently contradicts the official example.

**Manual IDs:** no documented `{#id}` attributes.

**Sources:**
- GitLab Docs: Heading IDs and links: https://docs.gitlab.com/user/markdown/#heading-ids-and-links
- GitLab issue #440733: https://gitlab.com/gitlab-org/gitlab/-/issues/440733

### 3. GitLab legacy / Redcarpet-like heading IDs - `gitlab-legacy` (aliases: `gitlab-old`, `glfm-legacy`, `gitlab-redcarpet`, `gitlab-pre-17`)

**Why this profile exists:** This profile is relevant only for old GitLab installations and editor compatibility.

**Short rule:** Lowercase, remove punctuation, convert spaces/separators to `-`, collapse repeated `-`; purely numeric slugs may be handled as `anchor-N`; duplicates with `-1`.

**Manual IDs:** no documented `{#id}` attributes.

**Sources:**
- GitLab issue #440733: https://gitlab.com/gitlab-org/gitlab/-/issues/440733
- Markdown All in One slugify.ts: https://raw.githubusercontent.com/yzhang-gh/vscode-markdown/master/src/util/slugify.ts

### 4. Crossnote / Mume heading IDs - `crossnote` (aliases: `mume`, `markdown-preview-enhanced`, `crossnote-uslug`)

**Why this profile exists:** Crossnote/Mume uses custom preprocessing plus `uslug`; it is neither GitHub nor plain `uslug`.

**Short rule:** Trim heading; remove `~`, `|`, `。`; convert whitespace to `~`; apply `uslug`; then convert `~` to `-`; duplicates with `-1`.

**Manual IDs:** `{#id}` attributes possible in Mume/Crossnote environments.

**Sources:**
- Crossnote HeadingIdGenerator: https://raw.githubusercontent.com/shd101wyy/crossnote/master/src/markdown-engine/heading-id-generator.ts
- uslug source: https://raw.githubusercontent.com/jeremys/uslug/master/lib/uslug.js

### 5. Pandoc auto_identifiers - `pandoc` (aliases: `pandoc-auto`, `pandoc-default`, `auto_identifiers`)

**Why this profile exists:** Pandoc is normatively documented and preserves periods, unlike GitHub.

**Short rule:** Remove formatting/links/footnotes; remove everything except alphanumeric characters, `_`, `-`, `.`; convert spaces/newlines to `-`; lowercase; remove everything before the first letter; empty slug becomes `section`; duplicates with `-1`.

**Manual IDs:** `{#id}` wins.

**Sources:**
- Pandoc Manual: Headings and sections: https://pandoc.org/MANUAL.html#headings-and-sections

### 6. Pandoc GFM auto identifiers - `pandoc-gfm` (aliases: `gfm_auto_identifiers`, `pandoc-gfm-auto`, `pandoc-github`)

**Why this profile exists:** Pandoc-GFM is close to GitHub, but not identical because of the Pandoc parser and emoji handling.

**Short rule:** GitHub-like Pandoc mode: spaces to dashes, uppercase to lowercase, remove punctuation except `-` and `_`, convert emoji to names; duplicates with `-1`.

**Manual IDs:** no `{#id}` attributes in the GFM reader.

**Sources:**
- Pandoc Manual: gfm_auto_identifiers: https://pandoc.org/MANUAL.html#extension-gfm_auto_identifiers

### 7. Pandoc ASCII identifiers - `pandoc-ascii` (aliases: `ascii_identifiers`, `pandoc-ascii-identifiers`)

**Why this profile exists:** `ascii_identifiers` handles non-Latin headings fundamentally differently from Unicode-preserving profiles.

**Short rule:** Like `pandoc`, but non-ASCII is reduced to ASCII; accents are removed and many non-Latin characters are dropped; empty slug becomes `section`.

**Manual IDs:** `{#id}` wins.

**Sources:**
- Pandoc Manual: ascii_identifiers: https://pandoc.org/MANUAL.html#extension-ascii_identifiers

### 8. kramdown auto IDs - `kramdown` (aliases: `kramdown-auto`, `kramdown-auto_ids`, `jekyll-kramdown`)

**Why this profile exists:** Important for Ruby/Jekyll/kramdown compatibility and because of the `section` fallback.

**Short rule:** Plain heading text; keep only letters, numbers, spaces, and dashes; remove everything before the first letter; convert everything except letters/numbers to `-`; lowercase; empty slug becomes `section`; duplicates with `-1`.

**Manual IDs:** `{#id}` wins.

**Sources:**
- kramdown HTML converter: header IDs: https://kramdown.gettalong.org/converter/html.html#toc-header-ids
- kramdown options: https://kramdown.gettalong.org/options.html

### 9. kramdown transliterated header IDs - `kramdown-transliterated` (aliases: `kramdown-ascii`, `kramdown-translit`, `kramdown-transliterated_header_ids`)

**Why this profile exists:** kramdown can create ASCII-only IDs; mdtoc needs a pinned transliteration for this.

**Short rule:** Like `kramdown`, but header IDs are transliterated so that only ASCII remains in automatically generated IDs. Exact values depend on kramdown/stringex and must be pinned.

**Manual IDs:** `{#id}` wins.

**Sources:**
- kramdown transliterated_header_ids: https://kramdown.gettalong.org/options.html#option-transliterated-header-ids
- Stringex repository: https://github.com/rsl/stringex

### 10. Blackfriday sanitized anchor names - `blackfriday` (aliases: `blackfriday-v2`, `hugo-blackfriday`, `gitea-blackfriday`)

**Why this profile exists:** Relevant for old Hugo/Go/Gitea stacks.

**Short rule:** Unicode letters and numbers remain; invalid character runs between valid characters become exactly one `-`; invalid characters at the beginning/end are removed; duplicates with `-1`.

**Manual IDs:** heading-ID extension possible.

**Sources:**
- Blackfriday v2 docs: https://pkg.go.dev/github.com/russross/blackfriday/v2

### 11. Hugo Goldmark github-ascii - `hugo-github-ascii` (aliases: `github-ascii`, `hugo-ascii`, `goldmark-github-ascii`)

**Why this profile exists:** Hugo provides `github-ascii` in addition to `github`; it is no longer GitHub emulation.

**Short rule:** GitHub-like anchors, but after ASCII reduction. Hugo/Goldmark version and AutoIDType must be pinned because the exact Unicode reduction is upstream behavior.

**Manual IDs:** Goldmark heading attributes possible.

**Sources:**
- Hugo markup configuration: https://gohugo.io/configuration/markup/
- Hugo goldmark_config AutoIDType: https://pkg.go.dev/github.com/gohugoio/hugo/markup/goldmark/goldmark_config

### 12. Python-Markdown TOC default - `python-markdown` (aliases: `python-markdown-ascii`, `py-markdown`, `mkdocs`, `mkdocs-default`)

**Why this profile exists:** Important for Python-Markdown and MkDocs; duplicates use `_1` instead of `-1`.

**Short rule:** Normalize NFKD, encode to ASCII and lose non-ASCII; remove characters outside `\w`, whitespace, and `-`; trim/lowercase; collapse separator runs to one `-`; duplicates `_1`, `_2`.

**Manual IDs:** `attr_list` optional.

**Sources:**
- Python-Markdown TOC extension: https://python-markdown.github.io/extensions/toc/
- Python-Markdown toc.py: https://raw.githubusercontent.com/Python-Markdown/markdown/master/markdown/extensions/toc.py

### 13. Python-Markdown Unicode slugify - `python-markdown-unicode` (aliases: `slugify_unicode`, `python-markdown-unicode-slugify`, `mkdocs-unicode`)

**Why this profile exists:** Python-Markdown provides a Unicode-preserving slugify variant.

**Short rule:** Like `python-markdown`, but without ASCII loss; Unicode word characters remain; separator runs collapse; duplicates `_1`, `_2`.

**Manual IDs:** `attr_list` optional.

**Sources:**
- Python-Markdown TOC: slugify_unicode: https://python-markdown.github.io/extensions/toc/
- Python-Markdown toc.py: https://raw.githubusercontent.com/Python-Markdown/markdown/master/markdown/extensions/toc.py

### 14. PyMdownX slugify - `pymdownx` (aliases: `pymdownx-slugify`, `pymdown-extensions`, `mkdocs-material-pymdownx`)

**Why this profile exists:** PyMdownX is a configurable slug family; mdtoc needs a stable default.

**Short rule:** For mdtoc: preserve Unicode, lowercase, remove invalid characters, replace spaces with `-`, no general dash collapse; duplicates in the Python-Markdown stack are typically `_1`.

**Manual IDs:** `attr_list`/TOC stack.

**Sources:**
- PyMdown Extensions: Slugs: https://facelessuser.github.io/pymdown-extensions/extras/slugs/
- PyMdownX slugs.py: https://raw.githubusercontent.com/facelessuser/pymdown-extensions/main/pymdownx/slugs.py

### 15. markdown-it-anchor default - `markdown-it-anchor` (aliases: `markdown-it`, `vitepress`, `vuepress`, `markdown-it-anchor-default`)

**Why this profile exists:** Many JavaScript documentation systems use markdown-it-anchor or derivatives.

**Short rule:** Trim string, lowercase, convert whitespace runs to one `-`, then apply `encodeURIComponent`; punctuation is therefore encoded instead of deleted.

**Manual IDs:** no `{#id}` in the plugin default.

**Sources:**
- markdown-it-anchor source: https://raw.githubusercontent.com/valeriangalliat/markdown-it-anchor/master/index.js

### 16. Visual Studio Code Markdown Outline - `vscode` (aliases: `visual-studio-code`, `vscode-markdown`, `markdown-all-in-one-vscode`)

**Why this profile exists:** VS Code generates its own UI/outline anchors and is often perceived as the expected editor behavior.

**Short rule:** Convert inline Markdown to plain text; trim/lowercase; whitespace runs to `-`; remove defined punctuation, including `_`; trim leading/trailing `-`; URL-encode with `encodeURI`; duplicates `-1`.

**Manual IDs:** no `{#id}` in the outline slugger.

**Sources:**
- VS Code slugify source: https://github.com/microsoft/vscode/blob/main/extensions/markdown-language-features/src/slugify.ts
- Markdown All in One slugify.ts: https://raw.githubusercontent.com/yzhang-gh/vscode-markdown/master/src/util/slugify.ts

### 17. Azure DevOps Wiki anchors - `azure-devops` (aliases: `azure`, `azure-wiki`, `ado`, `devops-wiki`)

**Why this profile exists:** Azure DevOps often treats punctuation as separators and differs clearly from GitHub.

**Short rule:** Lowercase; convert spaces and many special characters/punctuation marks to `-`; repeated hyphens may remain visible; for edge cases, check the rendered HTML ID.

**Manual IDs:** HTML anchors possible; `{#id}` not documented as a core rule.

**Sources:**
- Azure DevOps Markdown guidance: https://learn.microsoft.com/en-us/azure/devops/project/wiki/markdown-guidance?view=azure-devops

### 18. Bitbucket Cloud heading anchors - `bitbucket-cloud` (aliases: `bitbucket`, `bitbucket-cloud-markdown`, `bitbucket-header`)

**Why this profile exists:** The `markdown-header-` prefix is slug-semantic and requires a separate profile.

**Short rule:** GitHub/legacy-like slugification with prefix `markdown-header-`; separators collapse; duplicates with `-1`.

**Manual IDs:** not documented as a `{#id}` core rule.

**Sources:**
- Markdown All in One slugify.ts: Bitbucket branch: https://raw.githubusercontent.com/yzhang-gh/vscode-markdown/master/src/util/slugify.ts

### 19. Zola slugify anchors = on - `zola-on` (aliases: `zola`, `zola-default`, `zola-slugify-on`)

**Why this profile exists:** Zola slugifies anchors by default; Unicode is not preserved like GitHub.

**Short rule:** Zola default for `slugify.anchors = "on"`: slugified, ASCII-like anchors. Exact Unicode transliteration must be pinned against the Zola version.

**Manual IDs:** check Zola/Markdown attribute support separately.

**Sources:**
- Zola configuration: https://www.getzola.org/documentation/getting-started/configuration/

### 20. Zola slugify anchors = safe/off - `zola-safe` (aliases: `zola-safe`, `zola-off`, `zola-slugify-safe`, `zola-slugify-off`)

**Why this profile exists:** Zola `safe`/`off` uses a different URL-stability model: spaces become `_`, not `-`.

**Short rule:** Unicode and many characters remain; spaces become `_`; case is not necessarily changed; duplicates with `-1`.

**Manual IDs:** check Zola/Markdown attribute support separately.

**Sources:**
- Zola configuration: https://www.getzola.org/documentation/getting-started/configuration/

### 21. mdBook heading IDs - `mdbook` (aliases: `mdbook-rust`, `rust-mdbook`, `mdbook-default`)

**Why this profile exists:** mdBook has its own Rust implementation and is widely used for Rust documentation.

**Short rule:** mdtoc target behavior: preserve Unicode, lowercase, remove punctuation, convert spaces to `-`, repeated spaces as repeated `-`, duplicates with `-1`. Pin live fixtures against the current mdBook version.

**Manual IDs:** check custom heading attributes depending on mdBook/renderer version.

**Sources:**
- mdBook Markdown format: https://rust-lang.github.io/mdBook/format/markdown.html

## 7. Hugo/Goldmark in more detail

Hugo uses Goldmark as its default Markdown renderer. In Hugo configuration, `markup.goldmark.parser.autoHeadingID = true` and `autoIDType = github` are the defaults. This means the Hugo default can later be mapped as an alias to `github`, provided no other configuration is set.

`mdtoc` still needs separate profiles for:

- `hugo-github-ascii`: Goldmark/Hugo `autoIDType = github-ascii`; GitHub-like rules, but ASCII-oriented Unicode handling.
- `blackfriday`: old Hugo configurations or projects that want to emulate the earlier Blackfriday renderer.

`goldmark` by itself is therefore not a separate `slug=` value. The relevant setting is the AutoID type: `github`, `github-ascii`, or an older/different renderer configuration.

## 8. Machine-readable specification

The following blocks are intentionally included in this Markdown document so that documentation and test basis do not diverge. They can later be extracted into YAML files.

### 8.1 Profile registry

```yaml
profiles:
  - id: github
    name: "GitHub heading anchors"
    aliases: [gfm, github-slugger, github-heading, github-readme]
    duplicate_suffix: "-1"
    manual_id_policy: "no {#id} attributes; HTML anchors separately"
    unicode_policy: "preserved"
    implementation_status: "source-derived"
    sources:
      - "GitHub Docs: Section links and custom anchors: https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax"
      - "github-slugger source: https://raw.githubusercontent.com/Flet/github-slugger/master/index.js"
  - id: gitlab-current
    name: "GitLab current heading IDs"
    aliases: [gitlab, glfm, gitlab-17, gitlab-current-heading]
    duplicate_suffix: "-1"
    manual_id_policy: "no documented {#id} attributes"
    unicode_policy: "preserved"
    implementation_status: "source-derived"
    sources:
      - "GitLab Docs: Heading IDs and links: https://docs.gitlab.com/user/markdown/#heading-ids-and-links"
      - "GitLab issue #440733: https://gitlab.com/gitlab-org/gitlab/-/issues/440733"
  - id: gitlab-legacy
    name: "GitLab legacy / Redcarpet-like heading IDs"
    aliases: [gitlab-old, glfm-legacy, gitlab-redcarpet, gitlab-pre-17]
    duplicate_suffix: "-1"
    manual_id_policy: "no documented {#id} attributes"
    unicode_policy: "preserved"
    implementation_status: "source-derived"
    sources:
      - "GitLab issue #440733: https://gitlab.com/gitlab-org/gitlab/-/issues/440733"
      - "Markdown All in One slugify.ts: https://raw.githubusercontent.com/yzhang-gh/vscode-markdown/master/src/util/slugify.ts"
  - id: crossnote
    name: "Crossnote / Mume heading IDs"
    aliases: [mume, markdown-preview-enhanced, crossnote-uslug]
    duplicate_suffix: "-1"
    manual_id_policy: "{#id} attributes possible in Mume/Crossnote environments"
    unicode_policy: "preserved"
    implementation_status: "source-derived"
    sources:
      - "Crossnote HeadingIdGenerator: https://raw.githubusercontent.com/shd101wyy/crossnote/master/src/markdown-engine/heading-id-generator.ts"
      - "uslug source: https://raw.githubusercontent.com/jeremys/uslug/master/lib/uslug.js"
  - id: pandoc
    name: "Pandoc auto_identifiers"
    aliases: [pandoc-auto, pandoc-default, auto_identifiers]
    duplicate_suffix: "-1"
    manual_id_policy: "{#id} wins"
    unicode_policy: "preserved"
    implementation_status: "source-derived"
    sources:
      - "Pandoc Manual: Headings and sections: https://pandoc.org/MANUAL.html#headings-and-sections"
  - id: pandoc-gfm
    name: "Pandoc GFM auto identifiers"
    aliases: [gfm_auto_identifiers, pandoc-gfm-auto, pandoc-github]
    duplicate_suffix: "-1"
    manual_id_policy: "no {#id} attributes in the GFM reader"
    unicode_policy: "preserved"
    implementation_status: "source-derived"
    sources:
      - "Pandoc Manual: gfm_auto_identifiers: https://pandoc.org/MANUAL.html#extension-gfm_auto_identifiers"
  - id: pandoc-ascii
    name: "Pandoc ASCII identifiers"
    aliases: [ascii_identifiers, pandoc-ascii-identifiers]
    duplicate_suffix: "-1"
    manual_id_policy: "{#id} wins"
    unicode_policy: "ASCII-reduced"
    implementation_status: "source-derived"
    sources:
      - "Pandoc Manual: ascii_identifiers: https://pandoc.org/MANUAL.html#extension-ascii_identifiers"
  - id: kramdown
    name: "kramdown auto IDs"
    aliases: [kramdown-auto, kramdown-auto_ids, jekyll-kramdown]
    duplicate_suffix: "-1"
    manual_id_policy: "{#id} wins"
    unicode_policy: "preserved"
    implementation_status: "source-derived"
    sources:
      - "kramdown HTML converter: header IDs: https://kramdown.gettalong.org/converter/html.html#toc-header-ids"
      - "kramdown options: https://kramdown.gettalong.org/options.html"
  - id: kramdown-transliterated
    name: "kramdown transliterated header IDs"
    aliases: [kramdown-ascii, kramdown-translit, kramdown-transliterated_header_ids]
    duplicate_suffix: "-1"
    manual_id_policy: "{#id} wins"
    unicode_policy: "transliterated to ASCII"
    implementation_status: "source-derived"
    sources:
      - "kramdown transliterated_header_ids: https://kramdown.gettalong.org/options.html#option-transliterated-header-ids"
      - "Stringex repository: https://github.com/rsl/stringex"
  - id: blackfriday
    name: "Blackfriday sanitized anchor names"
    aliases: [blackfriday-v2, hugo-blackfriday, gitea-blackfriday]
    duplicate_suffix: "-1"
    manual_id_policy: "heading-ID extension possible"
    unicode_policy: "preserved"
    implementation_status: "source-derived"
    sources:
      - "Blackfriday v2 docs: https://pkg.go.dev/github.com/russross/blackfriday/v2"
  - id: hugo-github-ascii
    name: "Hugo Goldmark github-ascii"
    aliases: [github-ascii, hugo-ascii, goldmark-github-ascii]
    duplicate_suffix: "-1"
    manual_id_policy: "Goldmark heading attributes possible"
    unicode_policy: "ASCII-reduced"
    implementation_status: "source-derived"
    sources:
      - "Hugo markup configuration: https://gohugo.io/configuration/markup/"
      - "Hugo goldmark_config AutoIDType: https://pkg.go.dev/github.com/gohugoio/hugo/markup/goldmark/goldmark_config"
  - id: python-markdown
    name: "Python-Markdown TOC default"
    aliases: [python-markdown-ascii, py-markdown, mkdocs, mkdocs-default]
    duplicate_suffix: "_1"
    manual_id_policy: "attr_list optional"
    unicode_policy: "ASCII-reduced"
    implementation_status: "source-derived"
    sources:
      - "Python-Markdown TOC extension: https://python-markdown.github.io/extensions/toc/"
      - "Python-Markdown toc.py: https://raw.githubusercontent.com/Python-Markdown/markdown/master/markdown/extensions/toc.py"
  - id: python-markdown-unicode
    name: "Python-Markdown Unicode slugify"
    aliases: [slugify_unicode, python-markdown-unicode-slugify, mkdocs-unicode]
    duplicate_suffix: "_1"
    manual_id_policy: "attr_list optional"
    unicode_policy: "preserved"
    implementation_status: "source-derived"
    sources:
      - "Python-Markdown TOC: slugify_unicode: https://python-markdown.github.io/extensions/toc/"
      - "Python-Markdown toc.py: https://raw.githubusercontent.com/Python-Markdown/markdown/master/markdown/extensions/toc.py"
  - id: pymdownx
    name: "PyMdownX slugify"
    aliases: [pymdownx-slugify, pymdown-extensions, mkdocs-material-pymdownx]
    duplicate_suffix: "typically `_1` in the Python-Markdown stack"
    manual_id_policy: "attr_list/TOC stack"
    unicode_policy: "preserved/configurable"
    implementation_status: "source-derived"
    sources:
      - "PyMdown Extensions: Slugs: https://facelessuser.github.io/pymdown-extensions/extras/slugs/"
      - "PyMdownX slugs.py: https://raw.githubusercontent.com/facelessuser/pymdown-extensions/main/pymdownx/slugs.py"
  - id: markdown-it-anchor
    name: "markdown-it-anchor default"
    aliases: [markdown-it, vitepress, vuepress, markdown-it-anchor-default]
    duplicate_suffix: "-1"
    manual_id_policy: "no {#id} in plugin default"
    unicode_policy: "percent-encoded"
    implementation_status: "source-derived"
    sources:
      - "markdown-it-anchor source: https://raw.githubusercontent.com/valeriangalliat/markdown-it-anchor/master/index.js"
  - id: vscode
    name: "Visual Studio Code Markdown Outline"
    aliases: [visual-studio-code, vscode-markdown, markdown-all-in-one-vscode]
    duplicate_suffix: "-1"
    manual_id_policy: "no {#id} in the outline slugger"
    unicode_policy: "percent-encoded"
    implementation_status: "source-derived"
    sources:
      - "VS Code slugify source: https://github.com/microsoft/vscode/blob/main/extensions/markdown-language-features/src/slugify.ts"
      - "Markdown All in One slugify.ts: https://raw.githubusercontent.com/yzhang-gh/vscode-markdown/master/src/util/slugify.ts"
  - id: azure-devops
    name: "Azure DevOps Wiki anchors"
    aliases: [azure, azure-wiki, ado, devops-wiki]
    duplicate_suffix: "-1"
    manual_id_policy: "HTML anchors possible; {#id} not documented as a core rule"
    unicode_policy: "preserved/renderer-dependent"
    implementation_status: "source-derived"
    sources:
      - "Azure DevOps Markdown guidance: https://learn.microsoft.com/en-us/azure/devops/project/wiki/markdown-guidance?view=azure-devops"
  - id: bitbucket-cloud
    name: "Bitbucket Cloud heading anchors"
    aliases: [bitbucket, bitbucket-cloud-markdown, bitbucket-header]
    duplicate_suffix: "-1"
    manual_id_policy: "not documented as a {#id} core rule"
    unicode_policy: "preserved"
    implementation_status: "source-derived"
    sources:
      - "Markdown All in One slugify.ts: Bitbucket branch: https://raw.githubusercontent.com/yzhang-gh/vscode-markdown/master/src/util/slugify.ts"
  - id: zola-on
    name: "Zola slugify anchors = on"
    aliases: [zola, zola-default, zola-slugify-on]
    duplicate_suffix: "-1"
    manual_id_policy: "check Zola/Markdown attribute support separately"
    unicode_policy: "ASCII/transliterated"
    implementation_status: "source-derived"
    sources:
      - "Zola configuration: https://www.getzola.org/documentation/getting-started/configuration/"
  - id: zola-safe
    name: "Zola slugify anchors = safe/off"
    aliases: [zola-safe, zola-off, zola-slugify-safe, zola-slugify-off]
    duplicate_suffix: "-1"
    manual_id_policy: "check Zola/Markdown attribute support separately"
    unicode_policy: "preserved"
    implementation_status: "source-derived"
    sources:
      - "Zola configuration: https://www.getzola.org/documentation/getting-started/configuration/"
  - id: mdbook
    name: "mdBook heading IDs"
    aliases: [mdbook-rust, rust-mdbook, mdbook-default]
    duplicate_suffix: "-1"
    manual_id_policy: "check custom heading attributes depending on mdBook/renderer version"
    unicode_policy: "preserved"
    implementation_status: "source-derived"
    sources:
      - "mdBook Markdown format: https://rust-lang.github.io/mdBook/format/markdown.html"
```

### 8.2 Sub-profiles and transliteration

```yaml
subprofiles:
  syntax: "<canonical-profile>-<modifier>"
  parsing_rule: "parse by longest matching canonical profile id; then parse one suffix modifier"
  generated_slug_values: "For every canonical profile id in profiles[], mdtoc may accept <id>-unicode, <id>-strip, <id>-anyascii, <id>-unidecode, and <id>-icu. Alias names are accepted only without generated suffix unless explicitly registered."
  exact_emulation_warning: "A sub-profile is an mdtoc-derived profile, not exact upstream emulation of the base renderer."
  modifiers:
    - suffix: unicode
      meaning: "force Unicode preservation before the base slug rules"
      implementation: "Do not transliterate or drop non-ASCII before slugification; base profile filtering still applies."
      recommended: true
    - suffix: strip
      meaning: "drop non-ASCII code points"
      implementation: "Remove all code points > U+007F before case/space/punctuation processing, unless the base profile requires earlier parsing of manual IDs."
      recommended: true
    - suffix: anyascii
      meaning: "context-free Unicode-to-ASCII transliteration using AnyAscii"
      implementation: "Apply a pinned AnyAscii table before base profile case/space/punctuation processing. This is the recommended generic transliteration mode."
      recommended: true
      source: "https://github.com/anyascii/anyascii"
    - suffix: unidecode
      meaning: "Unicode-to-ASCII transliteration using Unidecode/Text::Unidecode"
      implementation: "Optional. Apply a pinned Unidecode version before base profile processing. Do not use as default because results are lossy and language-insensitive."
      recommended: false
      source: "https://pypi.org/project/Unidecode/"
    - suffix: icu
      meaning: "ICU/CLDR transliteration"
      implementation: "Optional expert mode. Pin ICU/CLDR version and transform id; default transform should be 'Any-Latin; Latin-ASCII'."
      recommended: false
      source: "https://unicode-org.github.io/icu/userguide/transforms/general/"
```

### 8.3 Fixtures and expected values

```yaml
fixtures:
  - id: F1
    markdown: "# A+B"
    meaning: "Punctuation without whitespace: remove, separator, or percent-encode?"
  - id: F2
    markdown: "# Version 3.5"
    meaning: "Period handling."
  - id: F3
    markdown: "# Привет 你好 & TEST"
    meaning: "Unicode, case folding, punctuation between spaces."
  - id: F4
    markdown: "# foo␠␠bar\\-\\-\\-baz"
    meaning: "Two spaces plus three literal hyphens; backslashes prevent smart punctuation."
  - id: F5
    markdown: "# Manual Über 汉字 {#Fix_ID-42}"
    meaning: "Pandoc/kramdown/Goldmark-style manual ID attribute; otherwise literal text."
  - id: F6
    markdown: "#### Closed␠␠ATX␠␠␠####"
    meaning: "Closed ATX: two spaces in the title, three spaces before closing hashes; the closing sequence is not part of the title."
  - id: F7
    markdown: "second # Repeat"
    meaning: "Duplicate handling; the first slug would be `repeat` or profile-dependent `Repeat`."
  - id: F8
    markdown: "# 123"
    meaning: "Purely numeric heading text and fallback rules."
expected_by_profile:
  github:
    F1:
      slug: "ab"
      status: expected
    F2:
      slug: "version-35"
      status: expected
    F3:
      slug: "привет-你好--test"
      status: expected
    F4:
      slug: "foo--bar---baz"
      status: expected
    F5:
      slug: "manual-über-汉字-fix_id-42"
      status: expected
    F6:
      slug: "closed--atx"
      status: expected
    F7:
      slug: "repeat-1"
      status: expected
    F8:
      slug: "123"
      status: expected
  gitlab-current:
    F1:
      slug: "ab"
      status: expected
    F2:
      slug: "version-35"
      status: expected
    F3:
      slug: "привет-你好--test"
      status: expected
    F4:
      slug: "foo--bar---baz"
      status: expected
    F5:
      slug: "manual-über-汉字-fix_id-42"
      status: expected
    F6:
      slug: "closed--atx"
      status: expected
    F7:
      slug: "repeat-1"
      status: expected
    F8:
      slug: "123"
      status: expected
  gitlab-legacy:
    F1:
      slug: "ab"
      status: expected
    F2:
      slug: "version-35"
      status: expected
    F3:
      slug: "привет-你好-test"
      status: expected
    F4:
      slug: "foo-bar-baz"
      status: expected
    F5:
      slug: "manual-über-汉字-fix_id-42"
      status: expected
    F6:
      slug: "closed-atx"
      status: expected
    F7:
      slug: "repeat-1"
      status: expected
    F8:
      slug: "anchor-123"
      status: expected
  crossnote:
    F1:
      slug: "ab"
      status: expected
    F2:
      slug: "version-35"
      status: expected
    F3:
      slug: "привет-你好--test"
      status: expected
    F4:
      slug: "foo--bar-baz"
      status: expected
    F5:
      slug: "Fix_ID-42"
      status: expected
    F6:
      slug: "closed--atx"
      status: expected
    F7:
      slug: "repeat-1"
      status: expected
    F8:
      slug: "123"
      status: expected
  pandoc:
    F1:
      slug: "ab"
      status: expected
    F2:
      slug: "version-3.5"
      status: expected
    F3:
      slug: "привет-你好-test"
      status: expected
    F4:
      slug: "foo-bar---baz"
      status: expected
    F5:
      slug: "Fix_ID-42"
      status: expected
    F6:
      slug: "closed-atx"
      status: expected
    F7:
      slug: "repeat-1"
      status: expected
    F8:
      slug: "section"
      status: expected
  pandoc-gfm:
    F1:
      slug: "ab"
      status: expected
    F2:
      slug: "version-35"
      status: expected
    F3:
      slug: "привет-你好--test"
      status: expected
    F4:
      slug: "foo-bar---baz"
      status: expected
    F5:
      slug: "manual-über-汉字-fix_id-42"
      status: expected
    F6:
      slug: "closed-atx"
      status: expected
    F7:
      slug: "repeat-1"
      status: expected
    F8:
      slug: "123"
      status: expected
  pandoc-ascii:
    F1:
      slug: "ab"
      status: expected
    F2:
      slug: "version-3.5"
      status: expected
    F3:
      slug: "test"
      status: expected
    F4:
      slug: "foo-bar---baz"
      status: expected
    F5:
      slug: "Fix_ID-42"
      status: expected
    F6:
      slug: "closed-atx"
      status: expected
    F7:
      slug: "repeat-1"
      status: expected
    F8:
      slug: "section"
      status: expected
  kramdown:
    F1:
      slug: "ab"
      status: expected
    F2:
      slug: "version-35"
      status: expected
    F3:
      slug: "привет-你好-test"
      status: expected
    F4:
      slug: "foo-bar-baz"
      status: expected
    F5:
      slug: "Fix_ID-42"
      status: expected
    F6:
      slug: "closed-atx"
      status: expected
    F7:
      slug: "repeat-1"
      status: expected
    F8:
      slug: "section"
      status: expected
  kramdown-transliterated:
    F1:
      slug: "ab"
      status: expected
    F2:
      slug: "version-35"
      status: expected
    F3:
      slug: "privet-ni-hao-test"
      status: pinned-mdtoc-target
    F4:
      slug: "foo-bar-baz"
      status: expected
    F5:
      slug: "Fix_ID-42"
      status: expected
    F6:
      slug: "closed-atx"
      status: expected
    F7:
      slug: "repeat-1"
      status: expected
    F8:
      slug: "section"
      status: expected
  blackfriday:
    F1:
      slug: "a-b"
      status: expected
    F2:
      slug: "version-3-5"
      status: expected
    F3:
      slug: "привет-你好-test"
      status: expected
    F4:
      slug: "foo-bar-baz"
      status: expected
    F5:
      slug: "Fix_ID-42"
      status: expected
    F6:
      slug: "closed-atx"
      status: expected
    F7:
      slug: "repeat-1"
      status: expected
    F8:
      slug: "123"
      status: expected
  hugo-github-ascii:
    F1:
      slug: "ab"
      status: expected
    F2:
      slug: "version-35"
      status: expected
    F3:
      slug: "test"
      status: pinned-mdtoc-target
    F4:
      slug: "foo--bar---baz"
      status: expected
    F5:
      slug: "Fix_ID-42"
      status: expected
    F6:
      slug: "closed--atx"
      status: expected
    F7:
      slug: "repeat-1"
      status: expected
    F8:
      slug: "123"
      status: expected
  python-markdown:
    F1:
      slug: "ab"
      status: expected
    F2:
      slug: "version-35"
      status: expected
    F3:
      slug: "test"
      status: expected
    F4:
      slug: "foo-bar-baz"
      status: expected
    F5:
      slug: "Fix_ID-42"
      status: pinned-mdtoc-target
    F6:
      slug: "closed-atx"
      status: expected
    F7:
      slug: "repeat_1"
      status: expected
    F8:
      slug: "123"
      status: expected
  python-markdown-unicode:
    F1:
      slug: "ab"
      status: expected
    F2:
      slug: "version-35"
      status: expected
    F3:
      slug: "привет-你好-test"
      status: expected
    F4:
      slug: "foo-bar-baz"
      status: expected
    F5:
      slug: "Fix_ID-42"
      status: pinned-mdtoc-target
    F6:
      slug: "closed-atx"
      status: expected
    F7:
      slug: "repeat_1"
      status: expected
    F8:
      slug: "123"
      status: expected
  pymdownx:
    F1:
      slug: "ab"
      status: expected
    F2:
      slug: "version-35"
      status: expected
    F3:
      slug: "привет-你好--test"
      status: expected
    F4:
      slug: "foo--bar---baz"
      status: expected
    F5:
      slug: "Fix_ID-42"
      status: pinned-mdtoc-target
    F6:
      slug: "closed--atx"
      status: expected
    F7:
      slug: "repeat_1"
      status: expected
    F8:
      slug: "123"
      status: expected
  markdown-it-anchor:
    F1:
      slug: "a%2Bb"
      status: expected
    F2:
      slug: "version-3.5"
      status: expected
    F3:
      slug: "%D0%BF%D1%80%D0%B8%D0%B2%D0%B5%D1%82-%E4%BD%A0%E5%A5%BD-%26-test"
      status: expected
    F4:
      slug: "foo-bar---baz"
      status: expected
    F5:
      slug: "manual-%C3%BCber-%E6%B1%89%E5%AD%97-%7B%23fix_id-42%7D"
      status: expected
    F6:
      slug: "closed-atx"
      status: expected
    F7:
      slug: "repeat-1"
      status: expected
    F8:
      slug: "123"
      status: expected
  vscode:
    F1:
      slug: "ab"
      status: expected
    F2:
      slug: "version-35"
      status: expected
    F3:
      slug: "%D0%BF%D1%80%D0%B8%D0%B2%D0%B5%D1%82-%E4%BD%A0%E5%A5%BD--test"
      status: expected
    F4:
      slug: "foo-bar---baz"
      status: expected
    F5:
      slug: "manual-%C3%BCber-%E6%B1%89%E5%AD%97-fixid-42"
      status: expected
    F6:
      slug: "closed-atx"
      status: expected
    F7:
      slug: "repeat-1"
      status: expected
    F8:
      slug: "123"
      status: expected
  azure-devops:
    F1:
      slug: "a-b"
      status: expected
    F2:
      slug: "version-3-5"
      status: expected
    F3:
      slug: "привет-你好--test"
      status: pinned-mdtoc-target
    F4:
      slug: "foo--bar---baz"
      status: expected
    F5:
      slug: "manual-über-汉字--fix-id-42"
      status: pinned-mdtoc-target
    F6:
      slug: "closed--atx"
      status: expected
    F7:
      slug: "repeat-1"
      status: expected
    F8:
      slug: "123"
      status: expected
  bitbucket-cloud:
    F1:
      slug: "markdown-header-ab"
      status: expected
    F2:
      slug: "markdown-header-version-35"
      status: expected
    F3:
      slug: "markdown-header-привет-你好-test"
      status: expected
    F4:
      slug: "markdown-header-foo-bar-baz"
      status: expected
    F5:
      slug: "markdown-header-manual-über-汉字-fix_id-42"
      status: expected
    F6:
      slug: "markdown-header-closed-atx"
      status: expected
    F7:
      slug: "markdown-header-repeat-1"
      status: expected
    F8:
      slug: "markdown-header-123"
      status: expected
  zola-on:
    F1:
      slug: "ab"
      status: expected
    F2:
      slug: "version-35"
      status: expected
    F3:
      slug: "privet-ni-hao-test"
      status: pinned-mdtoc-target
    F4:
      slug: "foo-bar-baz"
      status: expected
    F5:
      slug: "Fix_ID-42"
      status: pinned-mdtoc-target
    F6:
      slug: "closed-atx"
      status: expected
    F7:
      slug: "repeat-1"
      status: expected
    F8:
      slug: "123"
      status: expected
  zola-safe:
    F1:
      slug: "A+B"
      status: expected
    F2:
      slug: "Version_3.5"
      status: expected
    F3:
      slug: "Привет_你好_&_TEST"
      status: expected
    F4:
      slug: "foo__bar---baz"
      status: expected
    F5:
      slug: "Fix_ID-42"
      status: pinned-mdtoc-target
    F6:
      slug: "Closed__ATX"
      status: expected
    F7:
      slug: "Repeat-1"
      status: expected
    F8:
      slug: "123"
      status: expected
  mdbook:
    F1:
      slug: "ab"
      status: expected
    F2:
      slug: "version-35"
      status: expected
    F3:
      slug: "привет-你好--test"
      status: pinned-mdtoc-target
    F4:
      slug: "foo--bar---baz"
      status: expected
    F5:
      slug: "Fix_ID-42"
      status: pinned-mdtoc-target
    F6:
      slug: "closed--atx"
      status: expected
    F7:
      slug: "repeat-1"
      status: expected
    F8:
      slug: "123"
      status: expected
```

## 9. Source index for multiple profiles

This section contains only sources that apply to multiple profiles, sub-profiles, or implementation axes. Profile-specific sources are also listed directly in the profile sections.

- GitHub Docs, GitLab Docs, and the Pandoc Manual are references for GFM-like and Markdown renderer profiles: https://docs.github.com/ , https://docs.gitlab.com/user/markdown/ , https://pandoc.org/MANUAL.html
- Hugo/Goldmark configuration explains the relationship between `github`, `hugo-github-ascii`, and older renderer options: https://gohugo.io/configuration/markup/
- VS Code/Markdown All in One sources are useful because they compare or emulate several platform sluggers: https://raw.githubusercontent.com/yzhang-gh/vscode-markdown/master/src/util/slugify.ts
- Unicode/ICU/CLDR explain the general transliteration layer for sub-profiles: https://unicode-org.github.io/icu/userguide/transforms/general/ and https://cldr.unicode.org/index/cldr-spec/transliteration-guidelines
- AnyAscii and Unidecode are the most important generic Unicode-to-ASCII sources for mdtoc sub-profiles: https://github.com/anyascii/anyascii and https://pypi.org/project/Unidecode/

## 10. Implementation recommendation

For the first robust mdtoc implementation:

1. Implement the 21 base profiles as a registry.
2. Accept `slug=<canonical>` and all alias values.
3. Generate sub-profiles only for canonical profiles: `<canonical>-unicode`, `<canonical>-strip`, `<canonical>-anyascii`; `-unidecode` and `-icu` are optional.
4. Internally always normalize to `base_profile` and `unicode_modifier`.
5. Use the fixtures from section 8.3 as regression tests.
6. Before release, confirm values with `status: pinned-mdtoc-target` through live renderer tests or source-code tests, or consciously document them as mdtoc target behavior.
