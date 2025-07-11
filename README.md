[![Actions Status](https://github.com/exoscale/cli/workflows/CI/badge.svg?branch=master)](https://github.com/exoscale/cli/actions?query=workflow%3ACI+branch%3Amaster)

# Exoscale CLI

Manage your Exoscale infrastructure easily from the command-line with `exo`.


## Installation

Follow the steps for your platform on our [community docs](https://community.exoscale.com/tools/command-line-interface/#installation).


## Configuration

Running the `exo config` command will guide you through the initial configuration.

You can create and find API credentials in the *IAM* section of the [Exoscale Console](https://portal.exoscale.com/iam/keys).

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
[communitydoc]: https://community.exoscale.com/tools/command-line-interface/#configuration
