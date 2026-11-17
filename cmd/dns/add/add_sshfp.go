package add

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/clicmd"
)

type addSSHFPCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"SSHFP"`

	Domain string `cli-arg:"#" cli-usage:"DOMAIN-NAME|ID"`

	Name            string `cli-short:"n" cli-usage:"Leave this blank to create a record for DOMAIN, You may use the '*' wildcard here."`
	Algorithm       int64  `cli-short:"a" cli-usage:"RSA(1) | DSA(2) | ECDSA(3) | ED25519(4)"`
	FingerprintType int64  `cli-flag:"fingerprint-type" cli-usage:"SHA1(1) | SHA256(2)"`
	Fingerprint     string `cli-short:"f" cli-usage:"Fingerprint"`
	TTL             int64  `cli-flag:"ttl" cli-short:"t" cli-usage:"The time in seconds to live (refresh rate) of the record."`
}

func (c *addSSHFPCmd) CmdAliases() []string { return nil }
func (c *addSSHFPCmd) CmdShort() string     { return "Add SSHFP record type to a domain" }
func (c *addSSHFPCmd) CmdLong() string {
	return `Edit an "SSHFP" record to share your SSH fingerprint with others.`
}

func (c *addSSHFPCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	if err := clicmd.CmdCheckRequiredFlags(cmd, []string{"algorithm", "fingerprint-type"}); err != nil {
		return err
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *addSSHFPCmd) CmdRun(_ *cobra.Command, _ []string) error {
	return addDomainRecord(
		c.Domain,
		c.Name,
		"SSHFP",
		fmt.Sprintf("%d %d %s", c.Algorithm, c.FingerprintType, c.Fingerprint),
		c.TTL,
		nil,
	)
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &addSSHFPCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
		TTL:                3600,
	}))
}
