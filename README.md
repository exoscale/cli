# exo cli

Manage easily your Exoscale infrastructure from the exo command-line


## Installation

We provide many alternatives on the [releases](https://github.com/exoscale/cli/releases) page.

### Manual compilation

```
$ git clone https://github.com/exoscale/cli
$ cd cli
$ go build -o exo
```

## Configuration

The CLI will guide you in the initial configuration.
The configuration file and all assets created by any `exo` command will be saved in the `~/.exoscale/` folder.

You can find your credentials in our [Exoscale Console](https://portal.exoscale.com/account/profile/api)

```shell
$ exo config
```

## Usage

```shell
$ exo --help
```
