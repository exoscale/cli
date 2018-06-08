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
  exo [command]

Available Commands:
  affinitygroup Affinity groups management
  config        Generate config file for this cli
  eip           Elastic IPs management
  firewall      Security groups management
  help          Help about any command
  privnet       Private networks management
  ssh           SSH into a virtual machine instance
  sshkey        SSH keys pair management
  template      List all available templates
  vm            Virtual machines management
  zone          List all available zones

Flags:
      --config string   Specify an alternate config file [env CLOUDSTACK_CONFIG]
  -h, --help            help for exo
  -r, --region string   config ini file section name [env CLOUDSTACK_REGION] (default "cloudstack")

Use "exo [command] --help" for more information about a command.
$
```

### Documentation

To update the documentation, run the generator.

```
$ go run doc/main.go
```
