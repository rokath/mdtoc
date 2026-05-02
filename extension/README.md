# mdtoc VS Code Extension

This extension runs the `mdtoc` CLI against the active Markdown document in VS Code Desktop.

## Release Preparation

Use the repository root script for release tagging:

```bash
./setReleaseTag.sh 0.2.3
```

It normalizes `0.2.3` or `v0.2.3`, updates both extension version files, creates a version commit when needed, and then creates the repository tag.

## Local Test In VS Code

### Option A: Extension Development Host

Use this while developing the extension itself.

1. Open the `extension/` directory in VS Code.
2. Run `npm install`.
3. Run `npm run build`.
4. Press `F5`.
5. In the new Extension Development Host window, open a Markdown file.
6. Press `Shift-Cmd-P` and type `mdtoc`.

You should then see:

* `mdtoc: Generate ToC`
* `mdtoc: Regenerate ToC`
* `mdtoc: Strip ToC`
* `mdtoc: Check ToC`
* `mdtoc: Show Version`

### Option B: Install A Local VSIX

Use this when you want to test the packaged extension like a normal user install.

1. Stage a bundled binary into `extension/bin/<platform>/`.
2. Build the extension with `npm run build`.
3. Package the VSIX.
4. In VS Code, run `Extensions: Install from VSIX...`.
5. Select the generated file from `extension/out/`.

For the current macOS Apple Silicon target, the intended packaging command is:

```bash
MDTOC_VSCODE_TARGET_PLATFORM=darwin-arm64 npm run package:target
```

If you already staged a `darwin-arm64` binary, you can also use:

```bash
npm run package:macos-arm64
```

## Commands

* `mdtoc: Generate ToC`
* `mdtoc: Regenerate ToC`
* `mdtoc: Strip ToC`
* `mdtoc: Check ToC`
* `mdtoc: Show Version`

## Binary Resolution

The extension uses:

1. `mdtoc.executable.customPath`, if set
2. the bundled platform binary otherwise

There is no automatic `PATH` lookup in the MVP.

## Current Packaging Model

The bundled `mdtoc` binary inside `bin/<platform>/` is not the extension itself.

The installable VS Code extension is the generated `.vsix` file in `out/`.
