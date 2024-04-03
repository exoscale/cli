[![Actions Status](https://github.com/exoscale/cli/workflows/CI/badge.svg?branch=master)](https://github.com/exoscale/cli/actions?query=workflow%3ACI+branch%3Amaster)

# Exoscale CLI

Manage your Exoscale infrastructure easily from the command-line with `exo`.


## Installation

### Debian and Red Hat based distributions

On Debian and Red Hat based distributions like Ubuntu and Fedora, we recommend using the installation script.

```shell
curl -fsSL https://raw.githubusercontent.com/exoscale/cli/master/install-latest.sh | sh
```

### Using pre-built releases

You can find pre-built releases of the CLI [here][releases].


### From sources

To build `exo` from sources, a Go compiler >= 1.16 is required.

```shell
$ git clone https://github.com/exoscale/cli
$ cd cli
$ make build
```

Upon successful compilation, the resulting `exo` binary is stored in the `bin/` directory.

### Using the scoop package manager on Windows

If you haven't installed scoop already, follow the instructions at [scoop.sh](https://scoop.sh) before installing `exo` with:

```shell
scoop bucket add exoscale-cli https://github.com/exoscale/cli
scoop install exoscale-cli
```

To update `exo` to the latest version:

```shell
scoop update
scoop update exoscale-cli
```

### From the AUR on Arch Linux

```shell
gpg --keyserver keys.openpgp.org --recv-key 7100E8BFD6199CE0374CB7F003686F8CDE378D41
git clone https://aur.archlinux.org/exoscale-cli-bin.git
cd exoscale-cli-bin/
makepkg --install
```

Alternatively there are two packages building from source https://aur.archlinux.org/exoscale-cli.git and https://aur.archlinux.org/exoscale-cli-git.git where the latter builds from the latest commit on the master branch and the former from the latest release commit.

### With brew on MacOS

```shell
tap "exoscale/tap"
brew install exoscale-cli
```

## Configuration

Running the `exo config` command will guide you through the initial configuration.

You can create and find API credentials in the *IAM* section of the [Exoscale Console](https://portal.exoscale.com/iam/api-keys).

The configuration file and all assets created during `exo` operations will be saved in the following location:

| OS | Location |
|:--|:--|
| GNU/Linux, *BSD | `$HOME/.config/exoscale/` |
| macOS | `$HOME/Library/Application Support/exoscale/` |
| Windows | `%USERPROFILE%\.exoscale\` |

The configuration parameters are then saved in a `exoscale.toml` file with the following minimum format:

```
defaultaccount = "account_name"

[[accounts]]
  account = "account_name"
  key = "API_KEY"
  name = "account_name"
  secret = "API_SECRET"
```

The current configuration and configuration file path can be shown with `exo config show`.

## Usage

The `exo` CLI contains documentation for all of its commands, you can explore them by running `exo help`.
Additional information and tutorials are available [on Exoscale's community website][communitydoc].


## Integrations

### Fig

When using [Fig](https://fig.io) you can run this command to output Fig completion spec:

```
exo integrations generate-fig-spec
```

## External contributions

- [setup-exoscale](https://github.com/marketplace/actions/setup-exoscale) GitHub action


[releases]: https://github.com/exoscale/cli/releases
[communitydoc]: https://community.exoscale.com/documentation/tools/exoscale-command-line-interface/
