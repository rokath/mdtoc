# mdtoc VS Code Extension

> `mdtoc` - generate and strip Markdown tables of contents
> ☰ with numbering and stable anchor links (configurable)

<img src="https://raw.githubusercontent.com/rokath/mdtoc/main/extension/mdtoc_mascot_1024.webp" width="420">

This is a thin VS Code extension around [mdtoc](https://github.com/rokath/mdtoc) and updates the active Markdown document in place.

<details markdown="1"> <!-- parse this block as markdown -->
<summary><strong style="font-size: 1.25em;">Table of Contents</strong> <span style="font-size: 0.66em;">(click to expand)</span></summary>

<!-- mdtoc -->

* [1. Features](#1-features)
* [2. How to Use](#2-how-to-use)
* [3. Additional Information](#3-additional-information)
  * [3.1. Behaviour](#31-behaviour)
  * [3.2. Configuration](#32-configuration)
  * [3.3. Installation Alternative](#33-installation-alternative)
  * [3.4. Continuous Integration](#34-continuous-integration)

<!-- numbering=true min=2 max=4 slug=gitlab anchor=false link=true toc=true bullets=auto -->
<!-- /mdtoc -->

</details>

## 1. Features

* **easy** to use with editor context menu:
  * right-click inside an open Markdown editor and choose `mdtoc: Generate ToC`
* **configurable**: edit the generated `mdtoc` config block values directly to match your needs
  * `on|off` for **numbering**, **anchor**, **link**, **toc**
  * targets ATX headings (**min** `#` to **max** `######`)
  * **slug** profiles: `github`, `gitlab`, `crossnote`
  * auto or explicit (`*`, `-`, `+`) ToC **bullets** style
  * **delete** line `<!-- numbering=true min=2 max=4 slug=github anchor=true link=true toc=true bullets=auto -->` for **defaults** only
* **repeated headings** support
* **protects** non-generated content inside ToC area
  * generated content stays clearly separated from authored content
* deterministic and idempotent output
* keep the VS Code **workflow aligned** with the same [mdtoc CLI binary](https://github.com/rokath/mdtoc/releases) in local scripts and CI
* **Intentionally ignored headings**:
  * as **Setext headings**:
  
    ```md
    IGNORED Heading 1
    =========
    
    IGNORED Heading 2
    ---------
    ```

  * in **fenced code blocks**:

    ````md
    ``` 
    ## IGNORED Heading 3
    ```
    ````

  * in **HTML comments**:
  
    ```md
    <!-- 
    ## IGNORED Heading 4
    -->
    ```

  * as **HTML syntax**:
  
    ```md
    <h4>IGNORED Heading 5</h4>
    ```

  * between **exclusion regions**:
  
    ```md
    <!-- mdtoc off -->
    ## IGNORED Heading 6
    <!-- mdtoc on -->
    ```

  * with **starting space(s)**:

    ```md
     ## IGNORED Heading 7
    ```

## 2. How to Use

Open a Markdown file in VS Code, then use one of these entry points:

* Command Palette: `Shift+Cmd+P` on macOS or `Ctrl+Shift+P` on Windows/Linux, then run `mdtoc: Generate ToC` or `mdtoc: Strip ToC`
* Editor context menu: right-click inside an open Markdown editor and choose `mdtoc: Generate ToC` or `mdtoc: Strip ToC`

The table of contents is initially created at the beginning of the document. You can then move the managed block to another place in the file and `mdtoc: Generate ToC` will update it there.

<img src="https://raw.githubusercontent.com/rokath/mdtoc/main/extension/Animation.gif" width="420" alt="Animated demo of generating and stripping a table of contents">

## 3. Additional Information

### 3.1. Behaviour

`Generate ToC` runs `mdtoc` in root mode:

* if the document has no managed container yet, `mdtoc` creates one with its default settings (generate):
  * `<!-- numbering=true min=2 max=4 slug=github anchor=true link=true toc=true bullets=auto -->`
* if the document already has a valid managed container, `mdtoc` regenerates it from the stored container config
  * If the defaults ok for you can delete the config block line
  * break this config block line and you get re-generated:

  ```md
  <!--
  numbering=true
  min=2
  max=4
  slug=gitlab
  anchor=false
  link=true
  toc=true
  bullets=auto
  -->
  ```

* the document stays unchanged and the CLI error is shown if
  * more than one managed container exists
  * the managed container is invalid
* if a managed container is broken beyond repair, you can delete it and run `mdtoc: Generate ToC` again to create a fresh one

`Strip ToC` runs the explicit `strip` subcommand and is implicit executed on each re-generate. If the CLI reports an error, the document stays unchanged. If inside the ToC area non-generated lines found according to the current config block rules, these are saved inside aditional regions (**protection**).

```md
<!-- preserved by mdtoc
  * [3.1. Behaviour](#31-behaviour) accidently entered text here
-->
```

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

1. [Download](https://github.com/rokath/mdtoc/releases) the `.vsix` that matches your platform.
2. In VS Code, run `Extensions: Install from VSIX...`.
3. Select the downloaded `.vsix` file.

### 3.4. Continuous Integration

The underlying `mdtoc` binary is not limited to VS Code. You can use it directly in shell workflows, scripts, and CI, for example with `mdtoc check README.md` to fail a pipeline when a managed Markdown file is out of date. For CLI usage and the full feature set, see the repository [README](https://github.com/rokath/mdtoc/blob/main/README.md).
