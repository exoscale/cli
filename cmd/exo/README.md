# exo cli

Manage easily your Exoscale infrastructure from the exo command-line


## Installation

We provide many alternatives on the [releases](https://github.com/exoscale/egoscale/releases) page.

### Manual compilation

```
$ go get -u github.com/golang/dep/cmd/dep
$ go get -d github.com/exoscale/egoscale/...

$ cd $GOPATH/src/github.com/exoscale/egoscale/
$ dep ensure -vendor-only

$ cd cmd/exo
$ dep ensure -vendor-only

$ go install
```

## Configuration

A configuration file holding your credentials is required.
You can generate one via a guided prompt.
The following parameters are requested:

- `API Key`
- `Secret Key`

You can find those in our [Exoscale Console](https://portal.exoscale.com/account/profile/api)

### Automatic configuration via prompt

The CLI will guide you in the initial configuration.
The configuration file and all assets created by any `exo` command will be saved in the `~/.exoscale/` folder.

```shell
$ exo config
Hi happy Exoscalian, some configuration is required to use exo

We now need some very important informations, find them there.
https://portal.exoscale.com/account/profile/api

[+] Account name [none]: Production
[+] API Key [none]: EXO...
[+] Secret Key [none]: ...
Choose [Production] default zone:
1: ch-gva-2
2: ch-dk-2
3: at-vie-1
4: de-fra-1
[+] Select [1]: 1
[+] Do you wish to add another account? [Yn]: n
```
If you have already a `cloudstack.ini` configuration file

### Automatic configuration `cloustack.ini` migration

```shell
$ exo config
We've found a "cloudstack.ini" configuration file with the following configurations:
- [production] EXO...
- [preprod] EXO...
- [fail] EXO...INVALIDE_KEY
[+] Do you wish to import them automagically? [All, some, none]: all
Choose [production] default zone:
1: ch-gva-2
2: ch-dk-2
3: at-vie-1
4: de-fra-1
[+] Select [1]: 1
[+] Make [production] your default profile? [Yn]: y
Choose [preprod] default zone:
1: ch-gva-2
2: ch-dk-2
3: at-vie-1
4: de-fra-1
[+] Select [1]: 1
[+] Make [preprod] your default profile? [Yn]: n
Account [fail]: unable to verify user credentials
[+] Do you want to keep this account? [Yn]: n
[+] Do you wish to add another account? [Yn]: n
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
