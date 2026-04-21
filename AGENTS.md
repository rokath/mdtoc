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

* Use `README.md` and `./docs/mdtoc-spec.md` to understand intent and terminology.
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
* Before every `git push`, explicitly review `CHANGELOG.md`.
* Before every release tag, release creation, or release publication, explicitly review `CHANGELOG.md`.
* If the pushed commits affect release notes, unreleased notes, version sections, user-visible behavior, CI, docs, or shipped assets, `CHANGELOG.md` must be updated in the same push.
* Do not push commits first and postpone the `CHANGELOG.md` update for later.
* A release tag must not be created or published unless the tagged version already exists as its own section in `CHANGELOG.md`.
* When preparing a release, move the relevant notes from `Unreleased Changes` into the new versioned section and reset `Unreleased Changes` to only cover commits after the new tag.
* After tagging or before pushing tagged release commits, verify that the latest repository tag mentioned by `git tag` is present in `CHANGELOG.md` with the correct version header and git range.

## Tests

* Add or update tests when behavior changes.
* For Go tests in this repository, the standard library `testing` package is the default.

## Safety

* If a user message looks truncated or incomplete, ask a brief clarifying question before acting.
