# How to build a new release

**Disclaimer**: most of those instructions are intended for releasing the master
branch of https://github.com/exoscale/cli only. They are provided as a guideline
in the interest of transparency but should not be assumed to work on other
forks/branches.

1. Create a CHANGELOG.md entry for the release
2. Commit Changelog changes to master
3. Tag the release: `git tag -a vX.Y.Z`
4. Push the release tag to github `git push origin master --tags`
5. Run [goreleaser](https://goreleaser.com/) on the top-level directory: `GITHUB_TOKEN=<a github token with repo scope> goreleaser release --rm-dist`
   This step requires a GPG signing key configured in .goreleaser.yaml
6. Build the [snap](https://snapcraft.io/) package: on an Ubuntu system: `snapcraft; snapcraft push *.snap` Note the version number, then: `snapcraft release --stable <version number>`
