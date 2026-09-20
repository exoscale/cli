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

The configuration parameters are then saved in an `exoscale.toml` file with the following minimum format:

```toml
defaultaccount = "account_name"

[[accounts]]
  key = "API_KEY"
  name = "account_name"
  secret = "API_SECRET"
```

The top-level `defaultaccount` parameter selects an entry from `accounts` by
name. Each account supports these parameters:

| Parameter | Description |
|:--|:--|
| `name` | Account profile name. |
| `key` | API key. |
| `secret` | API secret. Use this or `secretCommand`. |
| `secretCommand` | Command and arguments that print the API secret, for example `["pass", "show", "exoscale"]`. |
| `defaultZone` | Default zone for commands that accept `--zone`. |
| `defaultSSHKey` | Default SSH key name for instance and instance pool creation. |
| `defaultTemplate` | Default instance template name. |
| `defaultOutputFormat` | Default output format: `table`, `json`, or `text`. |
| `endpoint` | Override the Exoscale API endpoint. |
| `sosendpoint` | Override the SOS endpoint. The value may contain a `{zone}` placeholder. |
| `customHeaders` | Map of HTTP headers added to API requests. |
| `clientTimeout` | Legacy API timeout in minutes. |
| `environment` | Legacy API environment name. |

The current configuration and configuration file path can be shown with `exo config show`.

### Environment variables

Command-line flags take precedence over environment variables. Environment
variables take precedence over values from the selected configuration profile.

| Variable | Description |
|:--|:--|
| `EXOSCALE_ACCOUNT` | Account profile to use. Equivalent to `--use-account`. |
| `EXOSCALE_API_ENDPOINT` | Override the Exoscale API endpoint. |
| `EXOSCALE_API_ENVIRONMENT` | Override the legacy API environment name. |
| `EXOSCALE_API_KEY` | Override the API key. Must be set with `EXOSCALE_API_SECRET`. |
| `EXOSCALE_API_SECRET` | Override the API secret. Must be set with `EXOSCALE_API_KEY`. |
| `EXOSCALE_API_TIMEOUT` | Override the legacy API timeout in minutes. |
| `EXOSCALE_CONFIG` | Path to an alternate configuration file. Equivalent to `--config`. |
| `EXOSCALE_STORAGE_API_ENDPOINT` | Override the SOS endpoint. |
| `EXOSCALE_TIMEOUT` | Per-zone timeout for list operations, for example `15s` or `1m`. Set to `-1s` to disable it. |
| `EXOSCALE_TRACE` | Enable HTTP request and response tracing. Unset it to disable tracing. Trace output may contain sensitive information. |
| `EXOSCALE_ZONE` | Override the default zone. |

For compatibility, the API key can also be read from `EXOSCALE_KEY`,
`CLOUDSTACK_KEY`, or `CLOUDSTACK_API_KEY`. The API secret can also be read from
`EXOSCALE_SECRET`, `EXOSCALE_SECRET_KEY`, `CLOUDSTACK_SECRET`, or
`CLOUDSTACK_SECRET_KEY`. `EXOSCALE_SOS_ENDPOINT` is an alias for
`EXOSCALE_STORAGE_API_ENDPOINT`.

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
