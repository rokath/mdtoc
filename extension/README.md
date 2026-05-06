# mdtoc VS Code Extension

> `mdtoc` - generate and strip Markdown tables of contents
> ☰ with numbering and stable anchor links (configurable)

<img src="https://raw.githubusercontent.com/rokath/mdtoc/main/extension/mdtoc_mascot_1024.webp" width="420">

This is a thin VS Code extension around [mdtoc CLI](https://github.com/rokath/mdtoc) and updates the active Markdown document in place.

<!-- mdtoc -->
<!-- mdtoc-config
container-version=v2
numbering=true
min-level=2
max-level=4
anchor=off
toc=false
bullets=auto
state=generated
-->
<!-- /mdtoc -->

## 1. Features

* very easy to use with editor context menu:
  * right-click inside an open Markdown editor and choose `mdtoc: Generate ToC`
* highly configurable: edit the `mdtoc` config block values directly to match your needs
  * on/off for numbering, anchor, toc
    * ToC link targets stay unnumbered for inline-anchor profiles but follow rendered heading text when `anchor=off`
  * targets ATX headings (`#` to `######`)
  * auto-detects the dominant bullet style (`*`, `-`, `+`) for ToC
  * explicit **anchor profiles**: `github` (default), `gitlab`, or `off`
* ignores headings inside **fenced code blocks** safely
* ignores headings inside **HTML comments**: `<!-- ... ## Example -->`
* **exclusion regions**: `<!-- mdtoc off -->` ... `<!-- mdtoc on -->`
* **repeated headings** support
* generated content stays clearly separated from authored content
* deterministic and idempotent output
* keep the VS Code workflow aligned with the same CLI `mdtoc` binary in local scripts and CI, get it from https://github.com/rokath/mdtoc/releases
* Excluded:
  * no Setext heading support (`Heading` followed by `===` or `---`)
  * no HTML heading support (`<h2>Example</h2>`)
  * not a site generator
  * not a Markdown formatter

## 2. How to Use

Open a Markdown file in VS Code, then use one of these entry points:

* Command Palette: `Shift+Cmd+P` on macOS or `Ctrl+Shift+P` on Windows/Linux, then run `mdtoc: Generate ToC` or `mdtoc: Strip ToC`
* Editor context menu: right-click inside an open Markdown editor and choose `mdtoc: Generate ToC` or `mdtoc: Strip ToC`

The table of contents is initially created at the beginning of the document. You can then move the managed block to another place in the file and `mdtoc: Generate ToC` will update it there.

<img src="https://raw.githubusercontent.com/rokath/mdtoc/main/extension/Animation.gif" width="420" alt="Animated demo of generating and stripping a table of contents">

## 3. Additional Information

### 3.1. Behaviour

`Generate ToC` runs `mdtoc` in root mode:

* if the document has no managed container yet, `mdtoc` creates one with its default settings (generate)
* if the document already has a valid managed container, `mdtoc` renews it from the stored container config
* if the managed container is invalid, the document stays unchanged and the CLI error is shown
* if a managed container is broken, beyond repair, you can delete it and run `mdtoc: Generate ToC` again to create a fresh one

`Strip ToC` runs the explicit `strip` subcommand. If the CLI reports an error, the document also stays unchanged.

### 3.2. Configuration

The extension supports one runtime setting:

```json
{
  "mdtoc.executable.customPath": ""
}
```

If `mdtoc.executable.customPath` is set, the extension uses that absolute path. Otherwise it uses the bundled platform binary.

There is no automatic `PATH` lookup in the current extension.

### 3.3. Installation Alternative

Install the extension from a packaged `.vsix` file:

1. Download the `.vsix` that matches your platform.
2. In VS Code, run `Extensions: Install from VSIX...`.
3. Select the downloaded `.vsix` file.

### 3.4. Continuous Integration

The underlying `mdtoc` binary is not limited to VS Code. You can use it directly in shell workflows, scripts, and CI, for example with `mdtoc check README.md` to fail a pipeline when a managed Markdown file is out of date.

For CLI usage and the full feature set, see the repository [README](https://github.com/rokath/mdtoc/blob/main/README.md).
