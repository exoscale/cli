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
; Exoscale credentials
[cloudstack]
endpoint = https://api.exoscale.ch/compute
key = EXO...
secret = ...

; Pygments theme, see: https://xyproto.github.io/splash/docs/
[exoscale]
; dark
theme = monokai
; light
theme = tango
; no colors (only boldness is allowed)
theme = nocolors
```

### Themes

Thanks to [Alec Thomas](http://swapoff.org/)' [chroma](https://github.com/alecthomas/chroma), it supports many themes for your output: <https://xyproto.github.io/splash/docs/>
