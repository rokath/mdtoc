# Repository Agent Rules

## Scope

* Only modify files that are directly relevant to the requested task.
* Do not refactor, rename, or reorganize unrelated parts of the repository.
* Avoid drive-by improvements unless the user explicitly asks for them.

## Editing

* Use `apply_patch` for manual text edits.
* Prefer minimal, reviewable diffs over broad rewrites.
* Do not reformat entire files unless formatting is part of the requested task.
* Preserve file encoding and line endings where practical.

## Project Structure

* The CLI entry point is `./cmd/mdtoc`.
* Core behavior lives in `./internal/mdtoc`.
* Release configuration lives in `./.goreleaser.yaml` and `./.github/workflows/goreleaser.yml`.
* GitHub Pages content is rooted at `README.md`, `index.md`, `./docs`, `./_config.yml`, and `./.github/workflows/pages.yml`.

## Documentation

* Use `README.md` and `./docs/spec.md` to understand intent and terminology.
* Do not change code only because the documentation suggests a nicer design.
* If code and documentation disagree, treat the current code and tests as the source of truth unless the user explicitly asks to align them.
* Repository-internal documentation links must stay relative.

## Build and Dependencies

* Do not introduce new dependencies without explicit user approval.
* Keep release and CI configuration simple and repository-specific. Remove copied settings that refer to other projects or unused tooling.
* Keep shell snippets portable across common CI environments when they are intended for workflows or docs.

## Communication

* Match the user's language for discussion.
* Write code comments, workflow comments, and repository file comments in English.
* If asked to draft issue text, write it in English unless the user explicitly requests another language.

## Commits

* Keep unrelated changes out of the same commit.
* If the work naturally splits into independent topics, prefer separate commits unless the user requests a single combined commit.
* Before every commit that affects release notes, unreleased notes, version sections, user-visible behavior, CI, docs, or shipped assets, explicitly review `CHANGELOG.md`.
* Before every `git push`, explicitly review `CHANGELOG.md`.
* Before every release tag, release creation, or release publication, explicitly review `CHANGELOG.md`.
* If the pushed commits affect release notes, unreleased notes, version sections, user-visible behavior, CI, docs, or shipped assets, `CHANGELOG.md` must be updated in the same push.
* If a task includes committing, pushing, or opening a PR for changes in those areas, treat the `CHANGELOG.md` update as a precondition and do not defer it.
* Do not create a release-preparation commit until `CHANGELOG.md` already has the intended version section, the correct git range for that version, and `Unreleased Changes` is reset to only later work.
* Do not push commits first and postpone the `CHANGELOG.md` update for later.
* Before editing `CHANGELOG.md` on a working branch, first reconcile it with the current `origin/main` state if any changelog or release-preparation PRs have merged in the meantime.
* Do not continue a structurally older `CHANGELOG.md` on `dev` after release-preparation work has already been merged into `main`; sync first, then add new unreleased notes.
* If `CHANGELOG.md` on the current branch differs structurally from `origin/main` around `Unreleased Changes` or recent version sections, fix that divergence before adding or moving entries.
* A release tag must not be created or published unless the tagged version already exists as its own section in `CHANGELOG.md`.
* When preparing a release, move the relevant notes from `Unreleased Changes` into the new versioned section and reset `Unreleased Changes` to only cover commits after the new tag.
* After tagging or before pushing tagged release commits, verify that the latest repository tag mentioned by `git tag` is present in `CHANGELOG.md` with the correct version header and git range.

## Tests

* Add or update tests when behavior changes.
* When a behavior change affects CLI file workflows, file mutation, or file-backed command paths, add or update at least one file-level test by default. Prefer the existing virtual filesystem test helpers over OS-level files when feasible.
* When changing CLI parsing, dispatch, or accepted flag forms, add explicit tests for:
  * the intended success path
  * conflicting or invalid input forms
  * argument-order variants that the active parser accepts
  * alternative spellings that the active parser accepts, including tolerated single-dash long options when applicable
* Do not claim CLI behavior is covered unless the relevant accepted spellings and argument orders are exercised by tests.
* For Go tests in this repository, the standard library `testing` package is the default.

## Safety

* If a user message looks truncated or incomplete, ask a brief clarifying question before acting.
