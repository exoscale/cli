# Releasing

> Note: requires Docker.

> Note: If you need to do a release from your local machine for some reason, you need to deactivate(comment out or remove) the release workflow trigger from the [workflow definition](https://github.com/exoscale/cli/blob/2de70996c852c1cb9fe5523beadc7e87e994adbc/.github/workflows/release.yml#L3-L6). Once this has been committed and pushed, you can follow all the release steps below and run `make release` locally at the end.

1. Update the `CHANGELOG.md` file with a new entry describing the latest
   changes up to the version of the release
1. Commit the previous steps' changes with a message `Prepare release`
1. Make sure you have the latest git tags: `git fetch --tags`
1. Run `make git-tag` and select the appropriate version
1. Run `git push` to push your commit
1. Run `git push --tags` to push the tag you just created. This will trigger the `release` [GitHub Action workflow](https://github.com/exoscale/cli/actions/workflows/release.yml). Once it has completed successfully, the release is done.
