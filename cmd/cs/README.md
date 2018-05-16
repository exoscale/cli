# cs: Command-Line Interface for CloudStack

[![Build Status](https://travis-ci.org/exoscale/egoscale.svg?branch=master)](https://travis-ci.org/exoscale/egoscale)
[![GoDoc](https://godoc.org/github.com/exoscale/egoscale?status.svg)](https://godoc.org/github.com/exoscale/egoscale/cmd/cs)

Like the Pythonic [cs](https://pypi.python.org/pypi/cs) but in Go.

## Installation

Grab it from the release section.

or build it yourself, it requires somes dependencies though.

```console
$ go install github.com/exoscale/egoscale/cmd/cs
```

## Configuration

Create a config file `cloudstack.ini` or `$HOME/.cloudstack.ini`.

```ini
; Default region
[cloudstack]

; Exoscale credential
endpoint = https://api.exoscale.ch/compute
key = EXO...
secret = ...

theme = fruity


; Another region
[cloudstack:production]

endpoint = https://api.exoscale.ch/compute
key = EXO...
secret = ...

theme = vim


; global config for themes
[exoscale]
; Pygments theme, see: https://xyproto.github.io/splash/docs/
; dark
theme = monokai
; light
theme = tango
; no colors (only boldness is allowed)
theme = nocolors
```

### Themes

Thanks to [Alec Thomas](http://swapoff.org/)' [chroma](https://github.com/alecthomas/chroma), it supports many themes for your output:

- <https://xyproto.github.io/splash/docs/>
- <https://help.farbox.com/pygments.html>

## Usage

Some of the flags around a command.

```
$ cs (-h | --help)              global help
$ cs <command> (-h | --help)    help of a specific command
$ cs <command> (-d | --debug)   show the command and its expected output
$ cs <command> (-D | --dry-run) show the signed command
$ cs (-r | --region) <region>   specify a different region, default `cloudstack`
```
