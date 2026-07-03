# Releasing

How a new exoscale/cli version ships. The flow is intentionally small: one
release-prep PR, one tag push, everything else is `.github/workflows/release.yml`.

## Steps

1. Land the release-prep PR. It moves the `## Unreleased` block of `CHANGELOG.md`
   into a dated section for the next version (e.g. `## 1.95.5 - 2026-05-28`)
   and clears `## Unreleased` for the cycle after. See PR #862 for the
   current example.

2. Push the version tag from the resulting `master` HEAD:

       git tag vX.Y.Z master
       git push origin vX.Y.Z

   The tag must point at a commit that already contains the changelog entry.
   Tagging the prep PR before it merges means the released commit ships
   without its `CHANGELOG.md` row, and the row only shows up in the next
   release.

3. `release.yml` runs. It is triggered by `push` of a tag matching
   `v[0-9]+\.[0-9]+\.[0-9]+` (see `.github/workflows/release.yml`). It does,
   in order:

   - `community-docs`: triggers `exoscale/community-ng`'s `gen-cli.yaml` with
     the new version.
   - `goreleaser`: builds per `.goreleaser.yml` (Linux/Darwin/Windows/OpenBSD
     on amd64/arm/arm64, plus a macOS universal binary), GPG-signs artifacts,
     then publishes to: GitHub Releases (binaries, source tarball, signed
     checksums), the Homebrew tap (`exoscale/homebrew-tap`), Docker images
     (`exoscale/cli:latest`, `:MAJOR`, `:MAJOR.MINOR`, `:MAJOR.MINOR.PATCH`),
     the in-repo Scoop bucket (`bucket/`), and nfpm deb/rpm to SOS buckets
     via `go.mk/scripts/publish-*-artifact-to-sos.sh`.
   - `archrelease`: bumps the three AUR packages (`exoscale-cli`,
     `exoscale-cli-bin`, `exoscale-cli-git`) in parallel.

No bot or backoffice tool tags the release. Tag step is manual.

## Major vs patch releases

For patch and minor releases, the steps above are everything. For major
releases or anything with user-visible impact worth announcing, ping `#ridge`
to coordinate external comms before tagging.

## Worked example: v1.94.2

Full cycle on 2026-04-27. Useful as a reference for what each step leaves
behind in the history.

| step                                        | time (UTC)        | sha / event                                     |
|---------------------------------------------|-------------------|-------------------------------------------------|
| 1. prepare-release commit                   | 14:07:48          | `09680feb` "prepare release" (CHANGELOG +5/-1) |
| 2. tag `v1.94.2` pushed (lightweight, same sha as step 1) | ~14:08:3x | tag points at `09680feb` (lightweight: no own timestamp) |
| 3. release workflow created (event=`push`, head=`v1.94.2`) | 14:08:39 | tag-push dispatch lands within seconds of the push |
| 4. GitHub release published                 | 14:18:24          | goreleaser finished                             |
| 5. Scoop update auto-commit (goreleaser)    | 14:18:26          | `279c8357` bucket/exoscale-cli.json +3/-3      |

Links:

- prepare-release commit: https://github.com/exoscale/cli/commit/09680febec2f45c501c6d479832bd2e41908181c
- tag page: https://github.com/exoscale/cli/releases/tag/v1.94.2
- release (workflow output, same URL): https://github.com/exoscale/cli/releases/tag/v1.94.2
- Scoop auto-commit: https://github.com/exoscale/cli/commit/279c83574f8c3d21a3c204f5c50075fe00be8d08

Total wall time: ~11 minutes from tag push to release published; Scoop
auto-commit lands ~2 seconds after the release. Lightweight tags carry no
separate timestamp, so "14:08:3x" is bounded by the commit time (14:07:48)
and the workflow `created_at` (14:08:39), not an exact push second.

## Open follow-ups

The process works but is implicit. Worth tracking separately:

- Tag push is manual. Easy to forget, easy to push against the wrong commit.
  A release-drafter or tag-bot would remove a class of mistakes.
- No documented rollback path if the workflow fails midway (partial
  publishing across GitHub/Homebrew/Docker/Scoop/AUR).
