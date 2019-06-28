UNRELEASED
----------

- sos: disable upload logging by default (#160)
- vm template: fix bug in "list" command (#161)

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
