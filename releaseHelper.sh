#!/bin/sh

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$SCRIPT_DIR"

if [ "$#" -ne 1 ]; then
  echo "Usage: ./releaseHelper.sh <version|vversion>" >&2
  exit 1
fi

RAW_VERSION=$1
VERSION=${RAW_VERSION#v}
TAG="v$VERSION"
CHANGELOG_HEADER="## <a id='${TAG}-changes'></a>${TAG} Changes"

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

if ! grep -Fq "$CHANGELOG_HEADER" CHANGELOG.md; then
  echo
  echo "CHANGELOG.md is not ready for $TAG." >&2
  echo "Required next steps:" >&2
  echo "  1. Add the section header: $CHANGELOG_HEADER" >&2
  echo "  2. Move the current release notes from Unreleased Changes into that $TAG section" >&2
  echo "  3. Reset Unreleased Changes so it only covers later work" >&2
  echo "  4. Commit the changelog update on main" >&2
  echo "  5. Re-run: ./releaseHelper.sh $TAG" >&2
  exit 1
fi

git fetch origin main --tags >/dev/null 2>&1 || true

CURRENT_BRANCH=$(git branch --show-current)
HEAD_SHA=$(git rev-parse --short HEAD)
ORIGIN_MAIN_SHA=$(git rev-parse --short origin/main 2>/dev/null || printf '%s' "missing")
LATEST_LOCAL_TAG=$(git tag --sort=version:refname | tail -n 1 || true)
TARGET_LOCAL_TAG_SHA=$(git rev-list -n 1 "$TAG" 2>/dev/null || true)
TARGET_REMOTE_TAG_SHA=$(git ls-remote --tags --refs origin "refs/tags/$TAG" 2>/dev/null | awk '{print $1}')

PACKAGE_VERSION=$(node -p "require('./extension/package.json').version")
LOCK_VERSION=$(node -p "require('./extension/package-lock.json').version")

echo "Release helper status"
echo "  target tag: $TAG"
echo "  branch: $CURRENT_BRANCH"
echo "  local HEAD: $HEAD_SHA"
echo "  origin/main: $ORIGIN_MAIN_SHA"
echo "  latest local tag: ${LATEST_LOCAL_TAG:-none}"
echo "  extension package version: $PACKAGE_VERSION"
echo "  extension lock version: $LOCK_VERSION"
if [ -n "$TARGET_LOCAL_TAG_SHA" ]; then
  printf '  local tag %s: %s\n' "$TAG" "$(printf '%.7s' "$TARGET_LOCAL_TAG_SHA")"
else
  echo "  local tag $TAG: missing"
fi
if [ -n "$TARGET_REMOTE_TAG_SHA" ]; then
  printf '  remote tag %s: %s\n' "$TAG" "$(printf '%.7s' "$TARGET_REMOTE_TAG_SHA")"
else
  echo "  remote tag $TAG: missing"
fi

if [ -n "$(git status --porcelain)" ]; then
  echo
  echo "Repository is dirty. Commit, stash, or discard local changes before running ./releaseHelper.sh." >&2
  git status --short
  exit 1
fi

if [ "$CURRENT_BRANCH" != "main" ]; then
  echo
  echo "Release preparation may be started from dev or main, but tagging itself must run from local main." >&2
  echo "Next steps:" >&2
  echo "  1. Merge dev into main on GitHub" >&2
  echo "  2. git switch main" >&2
  echo "  3. git pull --ff-only origin main" >&2
  echo "  4. ./releaseHelper.sh $TAG" >&2
  exit 1
fi

if ! git merge-base --is-ancestor origin/main HEAD; then
  echo
  echo "Local main is behind or diverged from origin/main. Sync main first." >&2
  echo "Suggested command: git pull --ff-only origin main" >&2
  exit 1
fi

if [ -n "$TARGET_REMOTE_TAG_SHA" ]; then
  if [ -z "$TARGET_LOCAL_TAG_SHA" ]; then
    echo
    echo "Remote tag $TAG already exists, but the local tag is missing." >&2
    echo "Suggested command: git fetch --tags origin" >&2
    exit 1
  fi
  if [ "$TARGET_REMOTE_TAG_SHA" != "$TARGET_LOCAL_TAG_SHA" ]; then
    echo
    echo "Local and remote tag $TAG point to different commits." >&2
    exit 1
  fi
  echo
  echo "Tag $TAG already exists locally and on origin."
  echo "Next step: start the GitHub goreleaser workflow for $TAG."
  exit 0
fi

if [ -n "$TARGET_LOCAL_TAG_SHA" ] && [ "$TARGET_LOCAL_TAG_SHA" != "$(git rev-parse HEAD)" ]; then
  echo
  echo "Local tag $TAG already exists but does not point to HEAD." >&2
  exit 1
fi

if [ "$PACKAGE_VERSION" != "$VERSION" ] || [ "$LOCK_VERSION" != "$VERSION" ]; then
  (
    cd extension
    npm run release:prepare-version -- "$VERSION"
  )
fi

if [ -n "$(git status --porcelain -- extension/package.json extension/package-lock.json)" ]; then
  git add extension/package.json extension/package-lock.json
  git commit -m "release: prepare $TAG extension version"
  HEAD_SHA=$(git rev-parse --short HEAD)
  echo
  echo "Created extension version commit at $HEAD_SHA."
fi

if [ -z "$TARGET_LOCAL_TAG_SHA" ]; then
  git tag "$TAG"
  echo "Created local tag $TAG."
else
  echo "Local tag $TAG already exists on HEAD."
fi

echo
echo "Next steps:"
echo "  1. git push origin main"
echo "  2. git push origin $TAG"
echo "  3. Start the GitHub goreleaser workflow for $TAG"
