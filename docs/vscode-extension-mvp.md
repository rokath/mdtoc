# mdtoc VS Code Extension MVP (Minimum Viable Product)

This document defines the MVP (Minimum Viable Product) for a first `mdtoc` extension for Visual Studio Code Desktop.

## Core rule

The Go CLI remains the canonical implementation. The extension is only a thin VS Code adapter around the `mdtoc` binary.

## MVP scope

Included:

* VS Code Desktop
* the active Markdown document
* manual command execution
* one bundled `mdtoc` binary per target platform
* an optional absolute-path override via setting

Excluded:

* `vscode.dev` and `github.dev`
* `formatOnSave`
* workspace-wide batch processing
* live diagnostics while typing
* auto-update logic for external custom binaries
* automatic `PATH` lookup

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
* Do not search `mdtoc` in `PATH` in the MVP.

Before normal execution, the extension validates the chosen binary via `mdtoc --version`.

## Initial target platforms

The MVP packages separate VSIX files for:

* `darwin-arm64`
* `darwin-x64`
* `linux-x64`
* `win32-x64`

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

Likely follow-up work after the MVP:

* more target platforms
* optional `formatOnSave`
* a separate architecture for browser or web environments
