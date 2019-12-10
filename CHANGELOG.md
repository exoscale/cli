1.8.0
-----

### New

* A new `exo sos show` command has been added to display object storage object properties (#204)
* Add support for `exo instancepool` as a top-level entity instead of as a lab feature (#211)

### Bug fixes

* Fixed Full-Control on object then adding a manual Grant rule. (#194)

### Changes

* Print back the SOS HTTP link when granting a canned `--public-read` or `--public-read-write` ACL (#208)
* `-z` is now available as a shorthand parameter wherever a `--zone` parameter is accepted (#209)

1.7.0
-----

### New

* Add new `exo config add` command to configure additional CLI accounts (#202)
* Add new `exo lab iam apikey operations` command to list supported IAM API key operations (#199)

### Bug fixes

* Allow IAM API key revocation by name in addition of by key (#195)
* Improve configuration account creation process when using restricted API keys (#195)

### Changes

* `exo vm show` now displays attached Private Networks (#196)

1.6.0
-----

### New

* Add support for Elastic IP descriptions (#191)
* Add support for IAM management preview in `exo lab` (#186)
* Add support for Instance Pools management in `exo lab` (#185)

### Bug fixes

* Fix panic when `$EXOSCALE_ACCOUNT` environment variable is set

### Changes

* `exo` now defaults to `$HOME` to look up configuration directory if `$XDG_CONFIG_HOME` is not set (#193)
* `exo vm create` now sets the service offering to *Medium* by default
* `exo sos create` now checks if user-specified zone exists (#183)
* `exo vm` lifecycle commands (`start`, `stop`...) are now more efficient with multiple instances (#134)
* On Windows, `exo sos` commands now require an external file containing the Exoscale SOS secure certificate chain. Use 
  the `exo sos --help` for more information regarding this issue.

1.5.1
-----

- Fix network retrieval by name (#175)
- `exo vm serviceoffering`: show the ID (#178)
- `exo zone`: honor command output formatting options (#179)
- `exo vm serviceoffering`: honor command output formatting options (#182)

1.5.0
-----

- Add new flag `--recursive` to the `sos delete` command to empty a bucket before deleting it (#172)
- Add "quiet" mode (#171)
- Fix `sos list` command panic if SOS returns bogus entries
- Fix `lab kube create` node instance upgrade stage (#166)
- Fix `affinitygroup delete` command confirmation prompt bug (#169)
- Fix `sos upload` issue with empty files (#173)
- Require protocol to be specified if a port is provided when adding a Security Group rule
- Require a user-data maximum length of 32Kb during instance creation (#168)

1.4.1
-----

- Disable logging by default in `sos upload` command (#160)
- Fix bug in `vm template list` command (#161)

1.4.0
-----

- Fix SOS upload large file corruption bug (#137)
- Add support for commands output customization (#150)
- Support template-filter in various commands (#151)
- Fix output bug in `network delete` command (#152)
- Display zone in `template (list|show)` commands (#153)
- Set a custom User-Agent (#154)
- Require confirmation for `vm stop`/`vm reboot` commands (#156)
- Update egoscale to 0.18.1

1.3.0
-----

- config: add support for client request custom HTTP headers
- vm: add support for *rescue profile* to `vm create`
- Various `exo * show` commands output normalization

1.2.0
-----

- Fix content-type sniffing on files < 512 bytes
- Add the registerCustomTemplate call
- exoscale/feat/list-template-filter
- exoscale/feat/deleteTemplate
- template list: add the templateFilter parameter
- templates: add the "exo vm template delete" subcommand
- exoscale/feat/updateIpAddress
- Add the `eip update` command
- exoscale/mcorbin/ch1915/eip-health-check
- eip_create/eip-show: support for healthchecks 

1.1.4
-----

- kube: calico/docker version
- vm: reset could accept a template parameter
- kube: force to accept the new conf of cloud-init
- api: make attach/detach ISO visible
- Pimp CMDs having this issue (issue #99) (pr #101)
- Allow VM instance security group modification

1.1.3
-----

- Fix #117
- makefile: build exoscale/cli:latest

1.1.2
-----

- config: panic on empty defaultZone
- fixup! config: improve life of people without config

1.1.1
-----

- config: improve life of people without config

1.1.0
-----

- Found a misspelling.
- Fix panic with env credentials
- CLI: show VMs in anti-affinity group
- api: highlight the output (stolen from go-cs)
- affinitygroup: enrich show and list
- lab: kube: add flag --version to create subcommand

1.0.9
-----

- feature: affinitygroup show
- fix: no panics when the config is made via env variables only

1.0.8
-----

- feature: What do now?
- feature: allow multiple EIP deletion
- feature: runstatus show page
- fix: runstatus reflect API changes

1.0.7
-----

- feature: spinners instead of fake loading bars
- feature: `api admin listVirtualMachines`
- feature: `sshkey delete --all`
- fix: `firewall ping6` protocol name
- fix: `firewall add --my-ip` to not create the default CIDR
- change: `firewall add` sets a CIDR by default

1.0.6
-----

- feature: runstatus
- feature: lab kube

1.0.5
-----

- feature: sos recursive upload
- feature: EXOSCALE_TRACE on the sos command
- feature: allow secrets to come from an external source
- feature: use XDG_CONFIG_HOME by default
- feature: dns remove asks for confirmation
- fix: `--my-ip` fix by @falzm

1.0.4
-----

- feature snapshot
- feature dns CAA record
- feature privnet `--cidrmask` as an alternative to `--netmask`
- manpage and bash autocompletion in binaries

1.0.3
-----

- feature exo status displaying the exoscale platform status
- feature new API call updateVmNicIp call
- feature sos download has a progress bar


1.0.2
-----

- feature sos listings `--short`
- fix change the account selection flag to `--use-account`
- fix version command do not require any config file

1.0.1
-----

- feature bump egoscale to v0.12.2

1.0.0
-----

- initial release
