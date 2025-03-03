# Changelog

## Unreleased

## 1.84.1

### Features

- sks: move cluster creation to egoscale v3 + enable-kube-proxy flag
- sks: add rotate operators CA cmd; more related egoscale v3 migrations

### Bug fixes

- sks update: fix missing feature gate entry #676

## 1.84.0

### Bug fixes

- fixing bigfile upload stuck issue. #671

### Features

- storage: bucket replication support #668
- dbaas: added valkey
- dbaas: remove redis create
- SKS cluster: display enable kube proxy #675

## 1.83.1

### Improvements

- lint: add golangci-lint action #665

### Bug fixes
- Refactor sos.DownloadFiles & fix file rename #664

## 1.83.0

### Features

- storage: Adding recursive feature to the storage command #653
- Update help for instance protection #658
- dbaas: database management for mysql, pg #661

### Bug fixes

- instance update: fixing no err check after creating client #657
- fix(error): Improve error message on snapshot creation (#655)
- security-group show: fix empty response if an API endpoint is misbehaving #660
- Fix broken zone flag for sks update command #662

## 1.82.0

### Features

- dbaas: added commands for managing service users #654

### Bug fixes

- config: fixing bug sosEndpoint lost after user switch account #652

## 1.81.0

### Features

- Private Network: support for DHCP options (dns-server/ntp-server/router/domain-search) #644

### Improvements

- Private Network: related commands are migrated to egoscale v3
- refactor(iam-api-key): Update IAM API Key manipulation to egoscale v3 #643

### Bug fixes

- Storage: handle errors in batch objects delete action #627
- Instance: Fix instance protection flag update zone context #648
- Anti-affinity group: fix show command to print all the attached instances from different zones #649

## 1.80.0

### Features
- Instance pool: added min-available flag to exo compute #629
- dbaas: external endpoints and external integration commands and sub-commands

## 1.79.1

### Improvements
- dbaas: added commands for getting and updating datadog integration settings #635
- go.mk: upgrade to v2.0.3 #632

### Bug fixes

- Fix creation and update of blockstorage volumes/snapshots in non-default zones

## 1.79.0

### Features

- instance: add a protection flag to exo compute #608

### Improvements

- DBaaS: external endpoints and integrations commands #631
- go.mk: update to v2.0.2 #630

### Bug Fixes

-  Fix list of dependencies for archlinux builds #628

## 1.78.6

### Bug Fixes

- Set API timeout from ENV when credentials are specified in ENV #625

### Changes

- Remove IAM access-key commands #626

## 1.78.5

### Improvements

- dbaas: use dedicated reveal-password endpoint to fetch password and build URI #618
- Instance Create: Migrate to egoscale v3 and add multiple sshkeys #620
- Reword quota description for blockstorage quotas (#622)

## 1.78.4

### Improvements

- Compute Instance delete: Remove multiple entities by their IDs/Names #619

### Bug Fixes

- output template: use text/template #617

### Improvements

- egoscale/v3: use separate module v3.1.0 #621

## 1.78.3

### Improvements

- go.mk: lint with staticcheck #606
- Update deprecated goreleaser directives #607
- sks nodepool: show instance family #615
- Update exo x #616

### Bug Fixes

- dbaas opensearch: remove top-level max-index-count flag #611
- Fix instance/ipool key naming in json output #612

## 1.78.2

### Bug Fixes
- security-group: show instances from all zones #605

## 1.78.1

### Bug Fixes
- SKS: Fix nodepool taints format parsing #600

## 1.78.0

### Features
- blockstorage: implement updating volume and snapshot labels and names #601

## 1.77.2

### Improvements
- Block Storage: Show all quotas #591
- config: remind user that no default account was set #593

### Bug Fixes
- Block Storage: Fix volume show with snapshot #589
- storage presign: fix panic when parsing arg[0] #590
- dbaas migration show: fix panic #597
- sks: enable CSI addon on existing clusters #596

## 1.77.1

### Features
- SKS nodepool: allow specifying kubelet image gc parameters on creation #586

### Improvements
- Egoscale v3: Fix the exoscale trace output #587

### Deprecations
- Config: Remove unused field on config reload #585

## 1.77.0

### Features
- compute: Add Block Storage #574
- sks: flag for CSI addon #572

### Improvements
- Instance reset password: remove wrong "rm" alias #583

### Deprecations

- Removed Windows ARM targets from prebuilt binaries #582

## 1.76.2

### Features
- limits: get block storage volume limit #577

### Improvements
- x: make `exo api` an alias to `exo x` #579
- Update `iam org-policy reset` confirmation text #568
- Update `README.md` with MacOS installation instructions #571

## 1.76.1

### Improvements
- go.mk: use as a plain repo instead of a submodule #575

## 1.76.0

### Bug Fixes
- SOS download: output warning when no objects exist at prefix #563
- Fix the bug in `iam role create` description that made it required #569
- Fix creating role with empty policy #569

### Features
- Updated 'exo x' list-block-storage-volumes #562
- completion: Adding fish support

### Improvements
- Update `exo iam role create` pro tip #55
- `exo iam org-policy`  `replace` command renamed to `update` where `replace` is now alias #569

## 1.75.0

### Features

- iam: implement Org Policy management commands #553
- iam: implement Role management commands #558
- iam: implement API Key management commands #560

## 1.74.5

### Improvements

- README: document installation from AUR #557
- install-script: install rpms from SOS repo #556
- Updated `exo x` for blockstorage #559

## 1.74.4

### Improvements

- publish releases as rpm packages on SOS #555

## 1.74.3

### Improvements

- install script: install from SOS apt repo if possible #551

### Bug Fixes

- Allow executing commands when dbaas JSON schema cannot be loaded #554

## 1.74.2

- update go.mk

## 1.74.1

### Improvements

- release workflow: publish deb packages to SOS #544
- aur releases: skip pgp check #547

## 1.74.0

- publish cli releases as scoop packages #546
- install script: verify signatures before installation #540
- release: adapt AUR release script for signed packages #541
- Updated `exo x` #542

## 1.73.0

### Features

- compute instance: implement reset-password command #536
- Updated `exo x` #539

### Improvements

- release: create source tarball and sign all artifacts #538

## 1.72.2

### Improvements

- sks show: display whether auto-upgrade is enabled #534

### Bug Fixes

- config add: fix adding new config (#537)

## 1.72.1

### Deprecations

- Remove all the deprecated commands (#526) (Deprecated since v0.51.0)

### Bug fixes

- compute: instance reset default template now falling back to current instance template (#528)
- compute: remmove uri and tlssni fields when nlb service healthcheck is "tcp"
- cmd: fix panic if inexistent config file is given (#530)

### Improvements

- release: automate AUR releases for Arch Linux (#531)

## 1.72.0

### Changes

- remove **runstatus** commands
- **status** command shows new status page

## 1.71.2

### Bug Fixes

- Fixed panic in in dbaas type show when authorized is nil (#524)

### Improvements

- Updated alpine version in Dockerfile (#523)

## 1.71.1

### Improvements

- create release GH Action workflow (#522)

## 1.71.0

### Features

- sos: add flags for filtering by version number and ID (#521)
- storage list: allow listing versions of objects (#518)

### Improvements

- Don't fetch account info when adding new account (#520)

## 1.70.0

### Features

- compute instance show: display deploy-target (#512)
- dbaas show grafana: show additional data (#507)
- storage: commands to enable, suspend and get the status of the object versioning setting (#509)
- Script to install the latest version on Debian and Red Hat-based distros

### Improvements

- standardize CI with other Go repos (#506)
- New "Exoternal Contributions" section in README.md with first addition: GitHub Action!
- Update MacOS compiled unified binary name to be inline with others (#517)

### Bug Fixes

- compute instance ssh: don't try to connect to private instances (#514)
- dbaas update: ignore regex checks in Database Settings data (#515)

## 1.69.0

### Features

 - dbaas: add grafana (#503)

## 1.68.0

### Features

 - storage: add support for setting the object ownership(#498)
 - integrations: fig completion (#475)
 - zones: add at-vie-2 to the list of zones (#501)

### Bug Fixes

 - compute instance snapshot: remove hardcoded timeout and bump default timeout to 20 minutes (#493)
 - compute instance list: fix data races (#497)

## 1.67.0

### Features

- `exo compute instance reveal-password`: new command that prints the password of a Compute instance (#494)
- `exo compute security-group list`: added flag `--visibility` to chose between private and public security groups (#494)
- `exo compute security-group rule add`: support creating rules referencing public groups (#495)
- Updated `exo x`

## 1.66.0

### Features

- `exo compute elastic-ip show <elastic ip>`: show names of instances attached to EIP (#490)

## 1.65.0

### Features

- `exo dbaas migration stop`: new command to stop database migration (#487)
- `exo compute security-group`: show instances in security group (#489)
- `exo compute sks nodepool`: show addons (#488)

## 1.64.0

### Features

- SKS nodepool: add `storage-lvm` addon (#486)
- Instance Pool: Deprecates `--template-filter` in favor of `--template-visibility` (#485)
- Updated `exo x`

### Bug Fixes

- Don't panic on nil pointer in dbaas opensearch commands (#484)
- Improve search template by name (#485)

## 1.63.0

### Features

-  `compute private instance support`: new `--private-instance` flag (#483)

## 1.62.0

### Features

- `compute instance update`, `compute elastic-ip update`: add support for Reverse DNS using `--reverse-dns` flag (#482)

## 1.61.0

### Features

- `storage list`: using delimiter to speed up listing of objects (#479)
- New configuration parameter (`clientTimeout`) to set API timeout (#478)
- Updated `dbaas show` for ACL API changes (#480)
- Updated `exo x`

## 1.60.0

### Features

- `config`: allow specifying a default output format (#476)
- Update 'Not Found' error message to include search zone where relevant (#472)
- Updated `exo x`

## 1.59.3

### Bug Fixes

- Fix panic in nlb show if a NLB doesn't have an IP yet (#473)
- Remove SOS certs that were shipped as a workaround with Windows releases (#470)

## 1.59.1

### Bug Fixes

- Fix panic in nlb list if a NLB doesn't have an IP yet (#468)

## 1.59.0

### Features

- `exo compute elastic-ip`: added IPv6 support.
- `exo x`: update commands.

### Bug Fixes

- `exo dbaas show opensearch`: fixed panic on nil value in response.
- `exo compute instance list`: fixed panic when instance has no IP.

## 1.58.0

### Features

- New `exo dbaas` type: OpenSearch.
- Default instance template updated: Linux Ubuntu 22.04 LTS.
- `exo dns`: now uses exoscale v2 API.
- `exo sks`: new CA option `control-plane`.

## 1.57.0

### Features

- `exo compute instance-template register`: add `--build`, `--version` and `--maintainer` to set template metadata.
- `exo dbaas logs --help`: explain how to use `--offset`.

## 1.56.0

### Features

- `exo compute sks create`: add `--cni` to specify the CNI plugin to deploy (conflicts with `--no-cni`, default to 'calico').
- `exo compute instance-template register`: add `--timeout` to configure registration timeout (default to 1h).

## 1.55.0

### Features

- `exo dbaas type show`: add `--backup-config` to print backup configuration for service type and plan.

### Bug Fixes

- Fix request signature bug with unsafe characters in the URL path.

## 1.54.0

### Features

- `exo compute instance create`, `exo compute instance-pool create`: remove default Cloud-Init compression on Instance creation, add `--cloud-init-compress` to compress the Instance Cloud-Init user-data.

## 1.53.0

### Features

- `exo dbaas create`, `exo dbaas update`: add dbaas migration configuration, add `--mysql-binlog-retention-period` to configure binlog entries retention.

### Bug Fixes

- `exo dbaas create`: fix panic when using `--maintenance-dow` and `--maintenance-time`.
- `exo * list`: fix race condition in most list commands.

## 1.52.1

### Bug Fixes

- `exo compute instance-template register` with `--from-snapshot`: now handle correctly '--disable-password', '--disable-ssh-key', and '--username' flags.

### Changes

- Now built with go 1.17

## 1.52.0

### Features

- `exo x`: bump commands
- `exo compute sks nodepool add`: add `linbit` flag to allow a non-standard partitioning scheme on nodes

## 1.51.2

### Bug Fixes

- Fix panic while rendering the table output of some commands (#439)

## 1.51.0

### Features

- `exo compute sks deprecated-resources`: list deprecated resources that will be removed in a future version of Kubernetes

### Changes

- `exo compute sks upgrade`: now warns about deprecated resources if target version doesn't support them anymore.

## 1.50.0

### Features

- `exo dbaas migration status`: get the status of a dbaas migration


## 1.49.3

### Bug Fixes

- `exo compute`: fix to use defaultTemplate from current account

- `exo storage`: fix empty object upload and download

## 1.49.2

- `exo dbaas`: fix a crash in the `show` command


## 1.49.1

### Bug Fixes

- `exo compute instance-template register`: fix `--from-snapshot` flag


## 1.49.0

### Changes

- `exo compute sks create`: flag `--oidc-required-claim` value type is now string *stringToString* instead of *string*


## 1.48.2

### Bug Fixes

- `exo dbaas`: fix a crash in the `logs` command


## 1.48.1

### Bug Fixes

- `exo compute sks`: fix a crash in the `create` command


## 1.48.0

### Changes

- `exo iam apikey *` commands are now deprecated, replaced by `exo iam access-key *`

### Features

- New `exo iam access-key *` commands
- New `exo dbaas metrics` command
- New `exo dbaas metrics` command


## 1.47.2

### Bug Fixes

- `exo dbaas show`: fix a crash with `pg`-type services
- `exo limits`: add missing entry for NLBs


## 1.47.1

### Bug Fixes

- Fix a bug crashing deprecated commands
- Improve formatting of the "Available Versions" column for the `exo dbaas type list` command output


## 1.47.0

### Changes

- `exo dbaas type update (list|show)` commands output: the `LatestVersion` label has been replaced by `AvailableVersions`

### Features

- `exo compute sks create`: add support for OpenID Connect configuration via `--oidc-*` flags
- `exo compute security-group delete`: add `--delete-rules|-r` flag


## 1.46.0

### Changes

- `exo dbaas (create|update) --help`: all type-specific `--<TYPE>-*` flag help descriptions have been moved to `--help-<TYPE>`
- `exo dbaas type show`: plans are not displayed by default, use the `--plans` flag to display a detailed list of plans supported by type (#405)

### Features

- `exo compute instance-type list`: new flag `--verbose|-v` to display more details (# CPUs, memory) (#407)
- `exo dbaas create mysql`: add `--mysql-recovery-backup-time` flag
- `exo dbaas create pg`: add `--pg-recovery-backup-time` flag
- `exo dbaas create redis`: add `--redis-recovery-backup-name` flag
- `exo dbaas show`: output service software version (#402)

### Bug Fixes

- `exo dbaas show`: add missing version for types `mysql`/`pg` (#406)
- `exo dbaas (create|update)`: improve maintenance-related flags handling (#404)


## 1.45.2

### Bug Fixes

- `exo compute instance create`: fixed a bug causing the CLI to crash when the `--private-network` flag is specified (#401)


## 1.45.1

### Bug Fixes

- `exo compute instance-template register`: fixed a bug preventing the use of the command without passing `--disable-(password|ssh-key)` flags (#399)


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
