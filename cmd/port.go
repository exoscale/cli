package cmd

import (
	"strings"
)

//go:generate stringer -type=port

type port uint16

const (
	daytime   port = 13
	ftp       port = 21
	ssh       port = 22
	telnet    port = 23
	smtp      port = 25
	time      port = 37
	whois     port = 43
	dns       port = 53
	tftp      port = 69
	gopher    port = 70
	http      port = 80
	kerberos  port = 88
	nic       port = 101
	sftp      port = 115
	ntp       port = 123
	imap      port = 143
	snmp      port = 161
	irc       port = 194
	https     port = 443
	rdp       port = 3389
	minecraft port = 25565
)

func (i port) StringFormatted() string {
	res := i.String()
	if strings.HasPrefix(res, "port(") {
		return ""
	}
	return res
}
