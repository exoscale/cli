# Egoscale Command-Line Interface

Like the Pythonic [cs](https://pypi.python.org/pypi/cs) but in Go.

## Installation

```console
$ go install github.com/exoscale/egoscale/cmd/cs
```

## Configuration

Create a `cloudstack.ini` file.

```ini
; Exoscale credentials
[cloudstack]
key = EXO...
secret = ...

; Pygments theme, see: http://pygments.org/demo/
[exoscale]
; dark
theme = monokai
; light
theme = tango
; no colors (only boldness)
theme = nocolors
```
