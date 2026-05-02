import { readFileSync, writeFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = dirname(fileURLToPath(import.meta.url));
const extensionDir = resolve(scriptDir, "..");
const packageJsonPath = resolve(extensionDir, "package.json");
const packageLockPath = resolve(extensionDir, "package-lock.json");

const rawVersion = process.argv[2];
if (!rawVersion) {
  console.error("Usage: npm run release:prepare-version -- <version>");
  process.exit(1);
}

const version = normalizeVersion(rawVersion);
if (!isValidSemver(version)) {
  console.error(`Invalid version: ${rawVersion}`);
  process.exit(1);
}

const packageJson = JSON.parse(readFileSync(packageJsonPath, "utf8"));
const packageLock = JSON.parse(readFileSync(packageLockPath, "utf8"));

packageJson.version = version;
packageLock.version = version;
if (packageLock.packages?.[""]) {
  packageLock.packages[""].version = version;
}

writeFileSync(packageJsonPath, `${JSON.stringify(packageJson, null, 2)}\n`);
writeFileSync(packageLockPath, `${JSON.stringify(packageLock, null, 2)}\n`);

console.log(`Set extension version to ${version}`);

function normalizeVersion(input) {
  return input.startsWith("v") ? input.slice(1) : input;
}

function isValidSemver(input) {
  return /^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$/.test(input);
}
