import * as fs from "node:fs";
import * as path from "node:path";
import { spawn } from "node:child_process";
import * as vscode from "vscode";

const output = vscode.window.createOutputChannel("mdtoc");
const extensionCommands = [
  "mdtoc.generate",
  "mdtoc.regen",
  "mdtoc.strip",
  "mdtoc.check",
  "mdtoc.showVersion",
] as const;

type MdtocCommand = "generate" | "regen" | "strip" | "check";
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
    vscode.commands.registerCommand("mdtoc.regen", () => runMutatingCommand(context, "regen")),
    vscode.commands.registerCommand("mdtoc.strip", () => runMutatingCommand(context, "strip")),
    vscode.commands.registerCommand("mdtoc.check", () => runCheckCommand(context)),
    vscode.commands.registerCommand("mdtoc.showVersion", () => showVersion(context)),
  );
}

export function deactivate(): void {}

async function runMutatingCommand(
  context: vscode.ExtensionContext,
  command: Exclude<MdtocCommand, "check">,
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
    result = await runMdtoc(resolved.path, [command], documentText);
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

async function runCheckCommand(context: vscode.ExtensionContext): Promise<void> {
  const editor = getActiveMarkdownEditor();
  if (!editor) {
    return;
  }

  const resolved = await resolveBinary(context);
  if (!resolved) {
    return;
  }

  let result: CommandResult;
  try {
    result = await runMdtoc(resolved.path, ["check"], editor.document.getText());
  } catch (error) {
    showRuntimeError("check", resolved.path, error);
    return;
  }

  const stderr = result.stderr.trim();

  if (result.exitCode === 0) {
    vscode.window.showInformationMessage("mdtoc: document matches its persisted state.");
    if (stderr) {
      output.appendLine(stderr);
      output.show(true);
    }
    return;
  }

  const message = stderr || "mdtoc check reported that the document differs from its persisted state.";
  output.appendLine(`[check] ${message}`);
  output.show(true);
  vscode.window.showWarningMessage(`mdtoc: ${message}`);
}

async function showVersion(context: vscode.ExtensionContext): Promise<void> {
  const resolved = await resolveBinary(context);
  if (!resolved) {
    return;
  }

  let current: CommandResult;
  try {
    current = await runMdtoc(resolved.path, ["--version"]);
  } catch (error) {
    showRuntimeError("--version", resolved.path, error);
    return;
  }

  if (current.exitCode !== 0) {
    showExecutionError("--version", resolved, current);
    return;
  }

  const lines = [
    `Extension version: ${context.extension.packageJSON.version ?? "unknown"}`,
    `mdtoc mode: ${resolved.mode}`,
    `mdtoc path: ${resolved.path}`,
    `mdtoc version: ${current.stdout.trim() || current.stderr.trim() || "unknown"}`,
  ];

  if (resolved.mode === "custom") {
    const bundled = bundledBinaryPath(context);
    if (bundled) {
      try {
        const bundledVersion = await runMdtoc(bundled, ["--version"]);
        if (bundledVersion.exitCode === 0) {
          lines.push(`bundled mdtoc version: ${bundledVersion.stdout.trim() || bundledVersion.stderr.trim()}`);
        }
      } catch {
        // Ignore bundled version lookup failures while reporting a custom binary.
      }
    }
  }

  output.clear();
  lines.forEach((line) => output.appendLine(line));
  output.show(true);
  void vscode.window.showInformationMessage("mdtoc version information written to the Output panel.");
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
