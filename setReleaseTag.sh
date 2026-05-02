#!/bin/sh

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$SCRIPT_DIR"

if [ "$#" -ne 1 ]; then
  echo "Usage: ./setReleaseTag.sh <version|vversion>" >&2
  exit 1
fi

RAW_VERSION=$1
VERSION=${RAW_VERSION#v}
TAG="v$VERSION"

case "$VERSION" in
  ''|*[!0-9A-Za-z.+-]*)
    echo "Invalid version: $RAW_VERSION" >&2
    exit 1
    ;;
esac

if ! printf '%s\n' "$VERSION" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?([+][0-9A-Za-z.-]+)?$'; then
  echo "Invalid version: $RAW_VERSION" >&2
  exit 1
fi

if [ -n "$(git status --porcelain)" ]; then
  echo "Repository is dirty. Commit, stash, or discard local changes before running $0." >&2
  exit 1
fi

if git rev-parse -q --verify "refs/tags/$TAG" >/dev/null 2>&1; then
  echo "Tag already exists: $TAG" >&2
  exit 1
fi

if ! grep -Fq "## <a id='${TAG}-changes'></a>${TAG} Changes" CHANGELOG.md; then
  echo "CHANGELOG.md does not contain the required section for $TAG." >&2
  exit 1
fi

(
  cd extension
  npm run release:prepare-version -- "$VERSION"
)

if [ -n "$(git status --porcelain -- extension/package.json extension/package-lock.json)" ]; then
  git add extension/package.json extension/package-lock.json
  git commit -m "release: prepare $TAG extension version"
fi

git tag "$TAG"

echo "Prepared extension version and created tag $TAG."
echo "Next steps:"
echo "  git push"
echo "  git push --tags"
echo "  Run the goreleaser workflow for $TAG"
