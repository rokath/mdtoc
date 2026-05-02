import { existsSync, mkdirSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const scriptDir = dirname(fileURLToPath(import.meta.url));
const extensionDir = resolve(scriptDir, "..");
const target = process.env.MDTOC_VSCODE_TARGET_PLATFORM;

if (!target) {
  console.error("Usage: MDTOC_VSCODE_TARGET_PLATFORM=<target> npm run package:target");
  process.exit(1);
}

const binaryName = target.startsWith("win32-") ? "mdtoc.exe" : "mdtoc";
const binaryPath = join(extensionDir, "bin", target, binaryName);

if (!existsSync(binaryPath)) {
  console.error(`Bundled binary missing for ${target}: ${binaryPath}`);
  process.exit(1);
}

const outDir = join(extensionDir, "out");
mkdirSync(outDir, { recursive: true });

const outFile = join(outDir, `mdtoc-vscode-${target}.vsix`);
const packageResult = spawnSync(
  "npx",
  ["@vscode/vsce", "package", "--target", target, "--out", outFile],
  {
    cwd: extensionDir,
    stdio: "inherit",
    env: process.env,
  },
);

if (packageResult.status !== 0) {
  process.exit(packageResult.status ?? 1);
}

console.log(`Created ${outFile}`);
