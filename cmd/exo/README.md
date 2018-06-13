# EXO CLI

Manage easily your Exoscale infrastructure from the exo command-line


## Installation

#### Simple

Download binary in [releases](https://github.com/exoscale/egoscale/releases) section 

#### Manual

Compile it

```
$ go get github.com/exoscale/egoscale

$ cd $GOPATH/src/github.com/exoscale/egoscale/egoscale/cmd/exo

$ dep ensure

$ go install
```

### Configuration

- create configuration file to connect exo to exoscale

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

```
$ exo config
[+] Compute API Endpoint [https://api.exoscale.ch/compute]: ...
[+] API Key [none]: EXO...
[+] Secret Key [none]: ...
```

## Getting started

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

## Documentation

To update the documentation, run the generator.

```
$ go run doc/main.go
```

## Example

- Deploy virtual machine instance with exo
- Connect to a virtual machine instance with exo

##### Create virtual machine

- `exo vm create` command 

```
$ exo vm create
Usage:
  exo vm create <vm name> [flags]

Flags:
  -a, --anti-affinity-group string   <name | id, name | id, ...>
  -f, --cloud-init-file string       Deploy instance with a cloud-init file
  -d, --disk int                     <disk size> (default 50)
  -h, --help                         help for create
  -6, --ipv6                         enable ipv6
  -k, --keypair string               <ssh keys name>
  -p, --privnet string               <name | id, name | id, ...>
  -s, --security-group string        <name | id, name | id, ...>
  -o, --service-offering string      <name | id> (micro|tiny|small|medium|large|extra-large|huge|mega|titan (default "Small")
  -t, --template string              <template name | id> (default "Linux Ubuntu 18.04")
  -z, --zone string                  <zone name | id | keyword> (ch-dk-2|ch-gva-2|at-vie-1|de-fra-1) (default "ch-dk-2")

Global Flags:
      --config string   Specify an alternate config file [env CLOUDSTACK_CONFIG]
  -r, --region string   config ini file section name [env CLOUDSTACK_REGION] (default "cloudstack")
$
```
##### deploy vm

```
$ exo vm create vmExample
Creating sshkey
Deploying....
┼───────────┼─────────────────┼──────────────────────────────────────┼
│   NAME    │       IP        │                  ID                  │
┼───────────┼─────────────────┼──────────────────────────────────────┼
│ vmExemple │ 159.100.245.105 │ 86ef3e19-81a8-4e71-b3ef-f9f287dc40e9 │
┼───────────┼─────────────────┼──────────────────────────────────────┼
$
```
##### Connect to this instance
```
$ exo ssh vmExemple
The authenticity of host '159.100.245.105 (159.100.245.105)' can't be established.
ECDSA key fingerprint is SHA256:XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX.
Are you sure you want to continue connecting (yes/no)? yes
Warning: Permanently added '159.100.245.105' (ECDSA) to the list of known hosts.
Welcome to Ubuntu 18.04 LTS (GNU/Linux 4.15.0-20-generic x86_64)

...

ubuntu@vmExemple:~$^C
$
```


Complete EXO [command documentation](https://exoscale.github.io/egoscale/cli/exo/)