package dns

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type dnsRemoveCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"remove"`

	Domain string `cli-arg:"#" cli-usage:"DOMAIN-NAME|ID"`
	Record string `cli-arg:"#" cli-usage:"RECORD-NAME|ID"`

	Force bool `cli-short:"f" cli-usage:"attempt to perform the operation without prompting for confirmation"`
}

func (c *dnsRemoveCmd) CmdAliases() []string { return clicmd.GRemoveAlias }
func (c *dnsRemoveCmd) CmdShort() string     { return "Remove a domain record" }
func (c *dnsRemoveCmd) CmdLong() string      { return "" }

func (c *dnsRemoveCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dnsRemoveCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	domainsList, err := client.ListDNSDomains(ctx)
	if err != nil {
		return err
	}
	domain, err := domainsList.FindDNSDomain(c.Domain)
	if err != nil {
		return err
	}

	domainRecordsList, err := client.ListDNSDomainRecords(ctx, domain.ID)
	if err != nil {
		return err
	}
	record, err := domainRecordsList.FindDNSDomainRecord(c.Record)
	if err != nil {
		return err
	}

	if !c.Force && !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to delete record %q?", record.ID)) {
		return nil
	}

	op, err := client.DeleteDNSDomainRecord(ctx, domain.ID, record.ID)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Deleting DNS record %q...", domain.UnicodeName), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Printf("Record %q removed successfully from %q\n", record.ID, domain.UnicodeName)
	}

	return nil
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &dnsRemoveCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
