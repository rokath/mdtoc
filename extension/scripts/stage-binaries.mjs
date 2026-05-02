import { copyFileSync, existsSync, mkdirSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const rootDir = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const extensionDir = resolve(rootDir, "..", "extension");
const platform = process.env.MDTOC_VSCODE_TARGET_PLATFORM;
const source = process.env.MDTOC_BINARY_SOURCE;

if (!platform || !source) {
  console.error(
    "Usage: MDTOC_VSCODE_TARGET_PLATFORM=<platform> MDTOC_BINARY_SOURCE=<file> npm run stage:binaries",
  );
  process.exit(1);
}

if (!existsSync(source)) {
  console.error(`Binary source not found: ${source}`);
  process.exit(1);
}

const targetName = platform.startsWith("win32-") ? "mdtoc.exe" : "mdtoc";
const targetPath = join(extensionDir, "bin", platform, targetName);

mkdirSync(dirname(targetPath), { recursive: true });
copyFileSync(source, targetPath);
console.log(`Staged ${source} -> ${targetPath}`);
