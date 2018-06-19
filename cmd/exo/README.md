# EXO CLI

Manage easily your Exoscale infrastructure from the exo command-line


## Installation

Download the [latest release](https://github.com/exoscale/egoscale/releases).

### Manual compilation

```
$ go get -u github.com/golang/dep/cmd/dep
$ go get github.com/exoscale/egoscale/...

$ cd $GOPATH/src/github.com/exoscale/egoscale/egoscale/
$ dep ensure -vendor

$ cd cmd/exo
$ dep ensure -vendor

$ go install
```

## Configuration

- create configuration file to connect `exo` to your Exoscale accounts.

#### Manual

Create a config file `cloudstack.ini` or `$HOME/.cloudstack.ini` ot `$HOME/.exoscale/cloudstack.ini`.

```ini
; Default region
[cloudstack]

; Exoscale credential
endpoint = https://api.exoscale.ch/compute
key = EXO...
secret = ...
```

#### Automatic

```shell
$ exo config
[+] Compute API Endpoint [https://api.exoscale.ch/compute]: ...
[+] API Key [none]: EXO...
[+] Secret Key [none]: ...
```

## Usage

```shell
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
```
