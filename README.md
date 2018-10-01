# Exoscale CLI

[![Build Status](https://travis-ci.org/exoscale/cli.svg?branch=master)](https://travis-ci.org/exoscale/cli) [![Go Report Card](https://goreportcard.com/badge/github.com/exoscale/cli)](https://goreportcard.com/report/github.com/exoscale/cli)

Manage easily your Exoscale infrastructure from the command-line with `exo`.


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

You can find your credentials in our [Exoscale Console](https://portal.exoscale.com/account/profile/api) (having or creating an account is required).

```shell
$ exo config
```

## Usage

```shell
$ exo --help
```
