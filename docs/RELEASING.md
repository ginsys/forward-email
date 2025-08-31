# Releasing

Use semantic version tags. GoReleaser publishes prereleases automatically when the tag has a prerelease suffix.

## Simple Workflow
- Only one entrypoint: `make release <type>`
- `VERSION` in the repo reflects the next tag to be created for prereleases; release commands tag the current version then bump for ongoing development when appropriate.

### Commands
- `make release <type>`: Dry-run; prints the exact steps without changing files.
- `make release-do <type>`: Executes the printed steps.

Types:
- `alpha`:
  - Tags the current `VERSION` if it’s `-alpha.N`; otherwise starts at `-alpha.0` and tags it.
  - Always bumps `VERSION` to the next `-alpha.N+1` after tagging.

- `beta`:
  - Same as `alpha`, but for beta.

- `stable`:
  - Makes a formal release for the base version (drops prerelease): tags `vX.Y.Z`.
  - Bumps `VERSION` to the next patch prerelease: `vX.Y.(Z+1)-alpha.0`.

- `bump {minor|major}`:
  - Updates `VERSION` to start a new development cycle: `vX.(Y+1).0-alpha.0` or `v(X+1).0.0-alpha.0`.
  - Commits the change (no tag created).

Notes:
- Pre-1.0 (`v0.x`) may contain breaking changes.
- Prerelease tags (`-alpha.N`, `-beta.N`) are GitHub prereleases. Stable tags (no suffix) are normal releases.

## Publish (CI)
Push the tag created locally:
```
git push origin vX.Y.Z[-pre.N]
```
The CI workflow runs GoReleaser to build and attach artifacts. Prerelease vs stable is detected from the tag.

## First Release Examples
- First alpha on a new line (dry-run then execute):
  - `echo v0.1.0-alpha.0 > VERSION && git commit -am 'chore: start v0.1.0-alpha.0'`
  - `make release alpha` (preview)
  - `make release-do alpha` → tags `v0.1.0-alpha.0`, bumps to `-alpha.1`.
- Move to a new minor line for development:
  - `make release bump minor` → sets `VERSION` to `v0.2.0-alpha.0`.

## Rollback
If a wrong tag was created:
```
git push --delete origin vWRONG || true
git tag -d vWRONG || true
```
Then correct locally and re‑push.
