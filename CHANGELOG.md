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
