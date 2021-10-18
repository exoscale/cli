# Changelog

## 1.45.0

### Features

- New `exo dbaas` commands (#395)
- `exo compute sks`: add support for taints to Nodepools (#390)
- `exo compute instance start`: add new flag `--rescue-mode` (#389)
- `exo compute instance-template show`: output zone

### Bug Fixes

- `exo storage upload`: fix large file upload bug (#397)
- `exo compute instance`: raise operation timeout to 10mn (#391)


## 1.44.0

### Features

- New `exo compute instance snapshot revert` command


## 1.43.0

### Changes

- Commands `exo compute instance-pool (create|update)` flags `--keypair`/`--privnet` are now deprecated, replaced by `--ssh-key`/`--private-network`

### Features

- New `exo compute instance snapshot export` command


## 1.42.0

### Changes

- Command `exo compute instance-pool show` output template label `.ServiceOffering` has been renamed `.InstanceType`
- Commands `exo compute instance-pool (create|update)` flags `--disk`/`--service-offering` are now deprecated, replaced by `--disk-size`/`--instance-type`


## 1.41.1

### Bug Fixes

- Fix `exo compute instance private-network update-ip` command


## 1.41.0

### Changes

- All Compute-related commands have been relocated to the `exo compute`
  sub-section. Original top-level commands (e.g. `exo vm`, `exo firewall`, `exo
  privnet`...) are now deprecated, and will be removed in a future release.

### Features

- New `exo compute security-group` commands
- New `exo compute sks upgrade-service-level` command

### Bug Fixes

- Fix Bash shell completion destination file path


## 1.40.5

### Changes

- `exo lab db show`: change `.Users` output label format


## 1.40.4

### Changes

- Update experimental `exo lab db` commands


## 1.40.3

### Bug Fixes

- Fix `exo compute instance-template list` command ignoring the `--zone` flag


## 1.40.2

### Bug Fixes

- Report missing Compute instance snapshots size in `exo compute instance snapshot show` command


## 1.40.1

### Bug Fixes

- Fix the asynchronous operation spinner to output to *stderr* intead of *stdout*


## 1.40.0

### Features

- New `exo compute instance snapshot` commands


## 1.39.0

### Features

- New `exo compute instance (resize-disk|reset|scale)` commands
- New `exo compute ssh-key` commands
- `exo compute instance create` now creates a single-use SSH key by default if none specified (similar to `exo vm create`)

### Changes

- `exo vm deploytarget` commands have been relocated to `exo compute deploy-target`


## 1.38.0

### Features

- `exo sks create`: add `--nodepool-private-network` flag

### Bug Fix

- `exo compute instance create`: fix private networks attachment


## 1.37.0

### Features

- Add `exo compute instance-template` commands
- Add `exo compute instance-type` commands
- `exo sks nodepool`: add support for Private Networks


## 1.36.0

### Features

- `exo vm`: add support for reverse DNS management


## 1.35.1

### Changes

- `exo vm`: remove deprecation warning


## 1.35.0

### Features

- `exo compute instance`: add `private-network` commands
- `exo compute instance`: add `security-group` commands
- `exo compute instance`: add `reboot` command
- `exo compute instance`: add `ssh`/`scp` commands


## 1.34.0

### Features

- sks: add support for labels/auto-upgrade

### Bug Fixes

- Add missing IP address in `exo compute instance show` command output


## 1.33.0

### Features

- Add new `exo compute instance` commands

### Changes

- Removed deprecated `exo api` command
- Deprecated `exo vm` commands


## 1.32.2

### Bug Fixes

- Fix crash during `exo lab db types list|show`
- Fix Zsh completion file installation path


## 1.32.1

### Bug Fixes

- `exo lab db update`: fix `--termination-protection` flag handling when set to `false`


## 1.32.0

### Features

- New commands `exo lab db`

### Bug Fixes

- Fix output annotations for `exo deploytarget list` command
- Fix `exo sks create` command description

### Changes

- The `exo lab kube` commands have been removed


## 1.31.0

### Features

- Add autocompletion generation for more shells
- `exo nlb`: add support for labels

## Bug Fixes

- `exo limits`: add missing organization resource limits
- `exo storage upload`: detect content type before file upload
- `exo firewall`: support Security Group rules with ICMP code/type -1

## Changes

- `exo nlb service add`: the flag `--instance-pool-id` has been replaced by `--instance-pool` accepting either a name or ID


## 1.30.0

### Features

- `exo sks`: add support for Instance Prefix/Deploy Target to Nodepools

### Bug Fixes

- `exo instancepool`: fix a bug in the "evict" command


## 1.29.0

### Features

- `exo vm deploytarget`: add support for Deploy Target resources
- `exo instancepool`: add support for Elastic IPs, Deploy Targets and Instance Prefix
- `exo instancepool`: add `evict` command

### Changes

- `exo sks nodepool scale`: ask for confirmation (can be overridden via the `-f, --force` flag)
- `exo eip list`: remove instances list from the output (information available via `exo eip show`)


## 1.28.0

### Improvements

- `exo storage show`: display object URL (#333)
- `exo sks create`: deploy K8s Metrics Server add-on by default (#331)


## 1.27.2

### Bug Fixes

- `exo vm create`: invalid API request signature caused by cloud-init userdata (#330)
- Various `exo storage` bug fixes (#326)


## 1.27.1

### Bug Fixes

- Various `exo storage` bug fixes (#326)


## 1.27.0

### New

- `exo storage` commands (#319)

### Changes

- The `exo sos` commands are now deprecated and replaced by `exo storage` commands


## 1.26.0

### Bug Fixes

- Raise the timeout value for the `exo sks *` commands

### Improvements

- `exo sks kubeconfig`: add support for exec credential mode (#323)


## 1.25.0

### Features

- `exo sks`: add `authority-cert` command
- `exo sks`: add `rotate-ccm-credentials` command
- `exo sks nodepool`: add `list` command (#314)

### Bug Fixes

- Manpages are now rendered correctly

### Improvements

- `exo sks nodepool`: support Nodepools Security Groups/Anti-Affinity Groups updating

### Changes

- `exo sks kubeconfig`: use group `system:masters` by default if no groups are specified
- `exo sks create`: flag `--version` now defaults to `latest` (latest available version returned by `exo sks versions`)


## 1.24.0

### Features

- `exo sks nodepool`: add Anti-Affinity Groups support

### Improvements

- `exo sks nodepool`: prompt for confirmation before evict

### Bug Fixes

- `exo instancepool delete`: prevent deletion if still referenced (#310)
- `exo sks evict`: fix arguments parsing issue (#312)

### Changes

- Drop support for CloudStack configuration (#311)
- `exo sks create`: set default version to 1.20.2


## 1.23.0

### Features

- New command `exo sks versions`
- New command `exo sks upgrade`
- New command `exo sks nodepool evict`

### Improvements

- `exo vm firewall` commands now update the Security Group memberships without requiring stopping the Compute instance (#308)


## 1.22.2

### Improvements

- `exo sos upload`: always send content md5 (#304)


## 1.22.1

### Bug Fixes

- `exo eip`: fixed "Healthcheck TLS Skip Verify" property reset to `false` after update operation


## 1.22.0

### New

- Add support for SKS resources management (#299)
- Add support for Anti-Affinity Groups to Instance Pools (#302)

### Bug Fixes

- `exo limits`: incorrect custom templates reporting (#300)


## 1.21.0

### Improvements

- `exo vm create` now supports the global `-O|--output-format` flag (#297)

### Changes

- Switched default API endpoint to `https://api.exoscale.com/v1`


## 1.20.2

### Changes

- Command custom `--output text` mode doesn't add a trailing empty line anymore, since in a pipe usage this can generate bogus empty entries in line-based processing.


## 1.20.1

### Bug Fixes

- sos: fix endpoint construction (#295)


## 1.20.0

### New

* `exo lab coi` command (#292)

### Improvements

* Improved `exo sos list` command performance with large buckets (#293)


## 1.19.0

### New

* `exo sos acl add`: support for recursive ACL addition (#290)


## 1.18.0

### New

* `exo nlb`: support for HTTP health checking (#284)

### Bug Fixes

* sos: fix bucket location inferring logic (#285)


## 1.17.0

### New

* `exo instancepool`: support for disk size updating (#282)
* `exo instancepool`: support for IPv6 activation
* `exo eip`: support for HTTP health checking

### Changes

* Operations progress info/messages is now output to `stderr` (#280)


## 1.16.1

### Bug Fixes

* `vm template list`: don't de-dup custom templates (#277)

### Changes

* `privnet show` command now reports the Private Network description in output
* `vm template list` command now reports the full creation date in output
* Instead of returning an error when multiple templates match a same name, the CLI now uses the most recent template (#278)


## 1.16.0

### New

* `exo vm *`/`exo ssh` commands now support instance names shell autocompletion (Bash only) (#273)

### Changes

* `exo vm snapshot show`: `Instance` field has been replaced by 2 fields `Instance Name`/`Instance ID`, and 2 new fields `Template Name`/`Template ID` have been added (#274)


## 1.15.0

### New

* `exo vm template register`: new flag `--from-snapshot` allowing registration of a custom template directly from a Compute instance snapshot (#268)

### Bug Fixes

* `exo lab kube create`: bumped outdated software versions

### Changes

* The `exo vm template register` command now expects the template name to be specified as positional argument instead of `--name` flag.


## 1.14.0

### New

* `exo scp` command (#267)
* `exo vm template register`: new flag `--boot-mode` to register UEFI-based custom templates (#266)

### Changes

* The `--description` flag is now optional in `exo vm template register`
* `exo nlb show`: JSON output `services` key is now lowercase


## 1.13.3

### Bug Fixes

* Fixed `exo ssh` command that didn't detect SSH private key file properly (#264)

### Changes

* `exo nlb` commands now accept a resource name as well as an ID (#265)


## 1.13.2

### Bug Fixes

* Fixed subcommand config settings leaking (#260)
* Fixed unused configuration cache file generation (#261)


## 1.13.1

### Internal

* Updated egoscale library following API V2 changes


## 1.13.0

### New

* Add support for Network Load Balancer resources management (`exo nlb`)
* Command `exo vm snapshot export` can now download exported snapshots with flag `--download` (#249)
* Arbitrary SSH client options can now be passed to the `exo ssh` command with flag `--ssh-options` (#250)
* `exo help environment` displays information about supported environment variables (#253)
* New command `exo vm update` to allow Compute instance properties modification (#255)
* `exo config show` now displays the path to the currently used configuration file (#257)
* Command `exo sos download` can now overwrite the destination file with flag `--force`

### Bug Fixes

* Fixed Snapcraft packaging (#243)
* Fixed client User Agent setting (#248)
* Fixed handling issues with username-less templates (#257)
* Fixed configuration file detection on Windows (#259)

### Changes

* Improved SOS certificates handling on Windows (#244)
* `exo zones` now displays zones sorted alphabetically (#246)
* `exo sos list` now returns the buckets size (#252)
* Commands that require a zone to be specified now default to the current account's default zone setting (#258)


## 1.12.0

### New

* Add [`go.mk`](https://github.com/exoscale/go.mk) support for exo cli (#233)
* Add `exo vm snapshot export` command to export an instant snapshot of a volume (#234)
* Add `exo limits` command to show the safety limits currently enforced on your account (#232)
* Add support to run `exo` binary on arm architecture 32/64 bits (#230)

### Bug Fixes

* Fix account selector in `exo config` (#241)
* Fix panic when `--quiet` flag is used (#236)

### Changes

* The `--output-format|-O` flag is no longer required with the `--output-template` flag (#239)
* Improve `apikey` commands output UX (#231)


## 1.11.0

### New

* Add new `exo vm snapshot show` command to display a Compute instance snapshot details

### Bug Fixes

* Fix configuration file detection issue on Windows
* Fix Calico version error in `exo lab kube` (#225)

### Changes

* Configuration profiles management (`exo config`) has been improved (#221)
* The following commands now support output customization through the global `--output-format|-O` flag:
  * `exo affinitygroup create`
  * `exo privnet create`
  * `exo sshkey create`
  * `exo sshkey upload`
  * `exo vm snapshot create`
  * `exo vm template register`


## 1.10.0

### New

* Add support for resource-level IAM API keys creation (#219)


## 1.9.0

### New

* Add support for `exo iam` as a top-level entity instead of as a lab feature (#214)

### Bug fixes

* Fix bug when you use an API key with sos/* rights only (#217) 

### Changes

* Changes the number of requests to minio before returning an error in `exo sos` (#213) 
* Improves the output of the `exo iam apikey operations` command (#212) 


## 1.8.0

### New

* A new `exo sos show` command has been added to display object storage object properties (#204)
* Add support for `exo instancepool` as a top-level entity instead of as a lab feature (#211)

### Bug fixes

* Fixed Full-Control on object then adding a manual Grant rule. (#194)

### Changes

* Print back the SOS HTTP link when granting a canned `--public-read` or `--public-read-write` ACL (#208)
* `-z` is now available as a shorthand parameter wherever a `--zone` parameter is accepted (#209)


## 1.7.0

### New

* Add new `exo config add` command to configure additional CLI accounts (#202)
* Add new `exo lab iam apikey operations` command to list supported IAM API key operations (#199)

### Bug fixes

* Allow IAM API key revocation by name in addition of by key (#195)
* Improve configuration account creation process when using restricted API keys (#195)

### Changes

* `exo vm show` now displays attached Private Networks (#196)


## 1.6.0

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
  


## 1.5.1

- Fix network retrieval by name (#175)
- `exo vm serviceoffering`: show the ID (#178)
- `exo zone`: honor command output formatting options (#179)
- `exo vm serviceoffering`: honor command output formatting options (#182)


## 1.5.0

- Add new flag `--recursive` to the `sos delete` command to empty a bucket before deleting it (#172)
- Add "quiet" mode (#171)
- Fix `sos list` command panic if SOS returns bogus entries
- Fix `lab kube create` node instance upgrade stage (#166)
- Fix `affinitygroup delete` command confirmation prompt bug (#169)
- Fix `sos upload` issue with empty files (#173)
- Require protocol to be specified if a port is provided when adding a Security Group rule
- Require a user-data maximum length of 32Kb during instance creation (#168)


## 1.4.1

- Disable logging by default in `sos upload` command (#160)
- Fix bug in `vm template list` command (#161)


## 1.4.0

- Fix SOS upload large file corruption bug (#137)
- Add support for commands output customization (#150)
- Support template-filter in various commands (#151)
- Fix output bug in `network delete` command (#152)
- Display zone in `template (list|show)` commands (#153)
- Set a custom User-Agent (#154)
- Require confirmation for `vm stop`/`vm reboot` commands (#156)
- Update egoscale to 0.18.1


## 1.3.0

- config: add support for client request custom HTTP headers
- vm: add support for *rescue profile* to `vm create`
- Various `exo * show` commands output normalization


## 1.2.0

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


## 1.1.4

- kube: calico/docker version
- vm: reset could accept a template parameter
- kube: force to accept the new conf of cloud-init
- api: make attach/detach ISO visible
- Pimp CMDs having this issue (issue #99) (pr #101)
- Allow VM instance security group modification


## 1.1.3

- Fix #117
- makefile: build exoscale/cli:latest


## 1.1.2

- config: panic on empty defaultZone
- fixup! config: improve life of people without config


## 1.1.1

- config: improve life of people without config


## 1.1.0

- Found a misspelling.
- Fix panic with env credentials
- CLI: show VMs in anti-affinity group
- api: highlight the output (stolen from go-cs)
- affinitygroup: enrich show and list
- lab: kube: add flag --version to create subcommand


## 1.0.9

- feature: affinitygroup show
- fix: no panics when the config is made via env variables only


## 1.0.8

- feature: What do now?
- feature: allow multiple EIP deletion
- feature: runstatus show page
- fix: runstatus reflect API changes


## 1.0.7

- feature: spinners instead of fake loading bars
- feature: `api admin listVirtualMachines`
- feature: `sshkey delete --all`
- fix: `firewall ping6` protocol name
- fix: `firewall add --my-ip` to not create the default CIDR
- change: `firewall add` sets a CIDR by default


## 1.0.6

- feature: runstatus
- feature: lab kube


## 1.0.5

- feature: sos recursive upload
- feature: EXOSCALE_TRACE on the sos command
- feature: allow secrets to come from an external source
- feature: use XDG_CONFIG_HOME by default
- feature: dns remove asks for confirmation
- fix: `--my-ip` fix by @falzm


## 1.0.4

- feature snapshot
- feature dns CAA record
- feature privnet `--cidrmask` as an alternative to `--netmask`
- manpage and bash autocompletion in binaries


## 1.0.3

- feature exo status displaying the exoscale platform status
- feature new API call updateVmNicIp call
- feature sos download has a progress bar


## 1.0.2

- feature sos listings `--short`
- fix change the account selection flag to `--use-account`
- fix version command do not require any config file


## 1.0.1

- feature bump egoscale to v0.12.2


## 1.0.0

- initial release
