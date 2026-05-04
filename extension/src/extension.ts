import * as fs from "node:fs";
import * as path from "node:path";
import { spawn } from "node:child_process";
import * as vscode from "vscode";

const output = vscode.window.createOutputChannel("mdtoc");

type MdtocCommand = "generate" | "strip";
type ResolveMode = "bundled" | "custom";

interface ResolvedBinary {
  mode: ResolveMode;
  path: string;
}

interface CommandResult {
  stdout: string;
  stderr: string;
  exitCode: number;
}

export function activate(context: vscode.ExtensionContext): void {
  context.subscriptions.push(output);
  context.subscriptions.push(
    vscode.commands.registerCommand("mdtoc.generate", () => runMutatingCommand(context, "generate")),
    vscode.commands.registerCommand("mdtoc.strip", () => runMutatingCommand(context, "strip")),
  );
}

export function deactivate(): void {}

async function runMutatingCommand(
  context: vscode.ExtensionContext,
  command: MdtocCommand,
): Promise<void> {
  const editor = getActiveMarkdownEditor();
  if (!editor) {
    return;
  }

  const documentText = editor.document.getText();
  const resolved = await resolveBinary(context);
  if (!resolved) {
    return;
  }

  let result: CommandResult;
  try {
    result = await runMdtoc(resolved.path, commandArgs(command), documentText);
  } catch (error) {
    showRuntimeError(command, resolved.path, error);
    return;
  }

  if (result.exitCode !== 0) {
    showExecutionError(command, resolved, result);
    return;
  }

  const fullRange = fullDocumentRange(editor.document);
  const applied = await editor.edit((editBuilder) => {
    editBuilder.replace(fullRange, result.stdout);
  });

  if (!applied) {
    vscode.window.showErrorMessage(`mdtoc ${command} could not update the active document.`);
    return;
  }

  if (result.stderr.trim()) {
    output.appendLine(result.stderr.trim());
    output.show(true);
  }
}

function getActiveMarkdownEditor(): vscode.TextEditor | undefined {
  const editor = vscode.window.activeTextEditor;
  if (!editor || editor.document.languageId !== "markdown") {
    void vscode.window.showErrorMessage("mdtoc requires an active Markdown editor.");
    return undefined;
  }

  return editor;
}

async function resolveBinary(context: vscode.ExtensionContext): Promise<ResolvedBinary | undefined> {
  const config = vscode.workspace.getConfiguration("mdtoc");
  const customPath = config.get<string>("executable.customPath", "").trim();

  if (customPath) {
    return validateBinaryPath({
      mode: "custom",
      path: customPath,
    });
  }

  const bundledPath = bundledBinaryPath(context);
  if (!bundledPath) {
    void vscode.window.showErrorMessage(
      `No bundled mdtoc binary is available for ${process.platform}-${mapArch(process.arch)}.`,
    );
    return undefined;
  }

  return validateBinaryPath({
    mode: "bundled",
    path: bundledPath,
  });
}

function bundledBinaryPath(context: vscode.ExtensionContext): string | undefined {
  const target = `${process.platform}-${mapArch(process.arch)}`;
  const binaryName = process.platform === "win32" ? "mdtoc.exe" : "mdtoc";
  const candidate = path.join(context.extensionPath, "bin", target, binaryName);

  return fs.existsSync(candidate) ? candidate : undefined;
}

function commandArgs(command: MdtocCommand): string[] {
  if (command === "generate") {
    // Root mode lets the CLI decide between container-aware regeneration and
    // first-time generation with default settings.
    return [];
  }
  return [command];
}

function mapArch(arch: string): string {
  if (arch === "x64") {
    return "x64";
  }
  if (arch === "arm64") {
    return "arm64";
  }
  if (arch === "arm") {
    return "arm";
  }
  return arch;
}

async function validateBinaryPath(binary: ResolvedBinary): Promise<ResolvedBinary | undefined> {
  if (!path.isAbsolute(binary.path)) {
    void vscode.window.showErrorMessage(`mdtoc binary path must be absolute: ${binary.path}`);
    return undefined;
  }

  if (!fs.existsSync(binary.path)) {
    void vscode.window.showErrorMessage(`mdtoc binary not found: ${binary.path}`);
    return undefined;
  }

  try {
    await fs.promises.access(binary.path, fs.constants.X_OK);
  } catch {
    void vscode.window.showErrorMessage(`mdtoc binary is not executable: ${binary.path}`);
    return undefined;
  }

  let result: CommandResult;
  try {
    result = await runMdtoc(binary.path, ["--version"]);
  } catch (error) {
    showRuntimeError("--version", binary.path, error);
    return undefined;
  }

  if (result.exitCode !== 0) {
    showExecutionError("--version", binary, result);
    return undefined;
  }

  return binary;
}

function runMdtoc(binaryPath: string, args: string[], input?: string): Promise<CommandResult> {
  return new Promise((resolve, reject) => {
    const child = spawn(binaryPath, args, {
      stdio: "pipe",
    });

    let stdout = "";
    let stderr = "";

    child.stdout.setEncoding("utf8");
    child.stderr.setEncoding("utf8");

    child.stdout.on("data", (chunk: string) => {
      stdout += chunk;
    });

    child.stderr.on("data", (chunk: string) => {
      stderr += chunk;
    });

    child.on("error", (error) => {
      reject(error);
    });

    child.on("close", (code) => {
      resolve({
        stdout,
        stderr,
        exitCode: code ?? 1,
      });
    });

    if (input !== undefined) {
      child.stdin.write(input);
    }
    child.stdin.end();
  });
}

function fullDocumentRange(document: vscode.TextDocument): vscode.Range {
  const lastLine = document.lineCount > 0 ? document.lineAt(document.lineCount - 1) : undefined;
  const end = lastLine ? lastLine.range.end : new vscode.Position(0, 0);
  return new vscode.Range(new vscode.Position(0, 0), end);
}

function showExecutionError(
  command: string,
  resolved: ResolvedBinary,
  result: CommandResult,
): void {
  const message = result.stderr.trim() || `mdtoc ${command} failed with exit code ${result.exitCode}.`;
  output.appendLine(`[${resolved.mode}] ${resolved.path}`);
  output.appendLine(message);
  output.show(true);
  void vscode.window.showErrorMessage(`mdtoc: ${message}`);
}

function showRuntimeError(command: string, binaryPath: string, error: unknown): void {
  const message = error instanceof Error ? error.message : String(error);
  output.appendLine(`[runtime] ${binaryPath}`);
  output.appendLine(message);
  output.show(true);
  void vscode.window.showErrorMessage(`mdtoc ${command} failed to start: ${message}`);
}
