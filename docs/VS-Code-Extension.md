# mdtoc VS Code Extension

This document defines the `mdtoc` extension for Visual Studio Code Desktop.

<img src="./VS-Code-Extension.png" width="960">

## Core rule

The Go CLI remains the canonical implementation. The extension is only a thin VS Code adapter around the `mdtoc` binary.

## Scope

Included:

* VS Code Desktop
* the active Markdown document
* manual command execution
* one bundled `mdtoc` binary per target platform
* an optional absolute-path override via setting

Excluded:

* `vscode.dev` and `github.dev`
  The extension targets VS Code Desktop only. Browser-based VS Code environments
  and the web extension runtime are not supported in the MVP.
* workspace-wide batch processing
  The extension only operates on the active Markdown document. It does not scan
  or modify multiple Markdown files across the workspace or repository.
* live diagnostics while typing
* `formatOnSave`

## Command model

The extension reads the active editor content, sends it to `mdtoc` over `stdin`, reads the result from `stdout`, and updates the open document through a VS Code edit.

Commands:

* `mdtoc.generate`
* `mdtoc.strip`

`generate` runs the CLI in root mode so the binary decides between first-time generation and regeneration from an existing valid managed container.

If the container is invalid, the extension must leave the document unchanged and surface the CLI error.

## Binary resolution

The extension uses exactly one runtime setting:

```json
{
  "mdtoc.executable.customPath": ""
}
```

Resolution rules:

* If `customPath` is set, use that exact path.
* If `customPath` is invalid, fail clearly and do not fall back.
* If `customPath` is empty, use the bundled binary for the current platform.
* Do not search `mdtoc` in `PATH` in the extension.

Before normal execution, the extension validates the chosen binary via `mdtoc --version`.

## Initial target platforms

The extension packages separate VSIX files for:

* `darwin-arm64`
* `darwin-x64`
* `linux-arm64`
* `linux-x64`
* `win32-x64`
* `win32-arm64`

Each VSIX contains:

* the TypeScript extension code
* exactly one matching bundled `mdtoc` binary

## Release flow

The intended release flow is:

1. GoReleaser builds the platform-specific `mdtoc` archives.
2. Extension packaging stages the matching binaries into `extension/bin/<target>/`.
3. The extension is packaged once per target as its own `.vsix` file.
4. Those VSIX artifacts are uploaded to the GitHub release.

## Package installation

Users install the packaged extension from a generated `.vsix` file:

1. Download the `.vsix` that matches the local platform.
2. In VS Code, run `Extensions: Install from VSIX...`.
3. Select the downloaded file.

For normal in-editor discovery and automatic updates, the extension also needs to be published to the VS Code Marketplace and, optionally, Open VSX.

## Later work

Likely follow-up work:

* more target platforms
* a separate architecture for browser or web environments
