import { existsSync, mkdirSync, rmSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const scriptDir = dirname(fileURLToPath(import.meta.url));
const extensionDir = resolve(scriptDir, "..");
const repoRoot = resolve(extensionDir, "..");
const tmpRoot = join(repoRoot, ".tmp", "vscode-extension");

const targets = [
  {
    target: "darwin-x64",
    archive: "dist/mdtoc_darwin_amd64.tar.gz",
    source: "mdtoc_darwin_amd64/mdtoc",
    extract: ["tar", ["-xzf", archivePath("dist/mdtoc_darwin_amd64.tar.gz"), "-C", tmpRoot]],
  },
  {
    target: "darwin-arm64",
    archive: "dist/mdtoc_darwin_arm64.tar.gz",
    source: "mdtoc_darwin_arm64/mdtoc",
    extract: ["tar", ["-xzf", archivePath("dist/mdtoc_darwin_arm64.tar.gz"), "-C", tmpRoot]],
  },
  {
    target: "linux-x64",
    archive: "dist/mdtoc_linux_amd64.tar.gz",
    source: "mdtoc_linux_amd64/mdtoc",
    extract: ["tar", ["-xzf", archivePath("dist/mdtoc_linux_amd64.tar.gz"), "-C", tmpRoot]],
  },
  {
    target: "win32-x64",
    archive: "dist/mdtoc_windows_amd64.zip",
    source: "mdtoc_windows_amd64/mdtoc.exe",
    extract: ["unzip", ["-q", archivePath("dist/mdtoc_windows_amd64.zip"), "-d", tmpRoot]],
  },
];

rmSync(tmpRoot, { force: true, recursive: true });
mkdirSync(tmpRoot, { recursive: true });

for (const entry of targets) {
  const archive = join(repoRoot, entry.archive);
  if (!existsSync(archive)) {
    console.error(`Release archive missing for ${entry.target}: ${archive}`);
    process.exit(1);
  }

  run(entry.extract[0], entry.extract[1], repoRoot);

  const source = join(tmpRoot, entry.source);
  if (!existsSync(source)) {
    console.error(`Extracted binary missing for ${entry.target}: ${source}`);
    process.exit(1);
  }

  run(
    "npm",
    ["run", "stage:binaries"],
    extensionDir,
    {
      ...process.env,
      MDTOC_BINARY_SOURCE: source,
      MDTOC_VSCODE_TARGET_PLATFORM: entry.target,
    },
  );

  run(
    "npm",
    ["run", "package:target"],
    extensionDir,
    {
      ...process.env,
      MDTOC_VSCODE_TARGET_PLATFORM: entry.target,
    },
  );
}

function archivePath(relativePath) {
  return join(repoRoot, relativePath);
}

function run(command, args, cwd, env = process.env) {
  const result = spawnSync(command, args, {
    cwd,
    env,
    stdio: "inherit",
  });

  if (result.status !== 0) {
    process.exit(result.status ?? 1);
  }
}
