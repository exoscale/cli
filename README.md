# Egoscale-cli

A simple CLI to use CloudStack using egoscale lib

### Prerequisites

- You need your clousctack.ini configuration file or you can generate it with CLI

### Installing

#### Simple installation

Download binary in releases section

#### Manual installation

Compile it

```
$ go get github.com/exoscale/egoscale
```

```
$ cd $GOPATH/src/github.com/exoscale/egoscale/egoscale/cmd/exo
```
```
$ dep ensure
```

```
$ go install
```
```
$ exo
A simple CLI to use CloudStack using egoscale lib

Usage:
  egoscale-cli [command]

Available Commands:
  affinitygroup Affinity group management
  config        Generate config file for this cli
  eip           Elastic IPs management
  firewall      Security group management
  help          Help about any command
  privnet       Pivate network management
  ssh           ssh into a virtual machine instance
  sshkey        SSH keys pairs management
  template      list all available templates
  vm            Virtual machines management
  zone          List all available zones

Flags:
      --config string   Specify an alternate config file (default: "~/.cloudstack.ini")
  -h, --help            help for egoscale-cli

Use "egoscale-cli [command] --help" for more information about a command.
$
```


