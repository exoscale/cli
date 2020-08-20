[![Actions Status](https://github.com/exoscale/cli/workflows/CI/badge.svg)](https://github.com/exoscale/cli/actions?query=workflow%3ACI)

# Exoscale CLI

Manage easily your Exoscale infrastructure from the command-line with `exo`.


## Installation

### Using pre-built releases (recommended)

You can find pre-built releases of the CLI [here][releases].


### From sources

To build `exo` from sources, a Go compiler >= 1.13 is required.

```shell
$ git clone https://github.com/exoscale/cli
$ cd cli
$ git submodule update --init --recursive go.mk
$ make build
```

Upon successful compilation, the resulting `exo` binary is stored in the `bin/` directory.


## Configuration

Running the `exo config` command will guide you through the initial configuration.

You can find your API credentials in the *Account* section of the [Exoscale Console](https://portal.exoscale.com/account/profile/api).

The configuration file and all assets created during `exo` operations will be saved in the following location:

| OS | Location |
|:--|:--|
| GNU/Linux, *BSD | `$HOME/.exoscale` |
| macOS | `$HOME/Library/Application Support/exoscale` |
| Windows | `%USERPROFILE%\.exoscale` |


## Usage

The `exo` CLI contains documentation for all of its commands, you can explore them by running `exo help`.
Additional information and tutorials are available [on Exoscale's community website][communitydoc].


[releases]: https://github.com/exoscale/cli/releases
[communitydoc]: https://community.exoscale.com/documentation/tools/exoscale-command-line-interface/
