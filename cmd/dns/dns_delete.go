package dns

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type dnsDeleteCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Domain string `cli-arg:"#" cli-usage:"DOMAIN-NAME|ID"`

	Force bool `cli-short:"f" cli-usage:"attempt to perform the operation without prompting for confirmation"`
}

func (c *dnsDeleteCmd) CmdAliases() []string { return clicmd.GDeleteAlias }
func (c *dnsDeleteCmd) CmdShort() string     { return "Delete a domain" }
func (c *dnsDeleteCmd) CmdLong() string      { return "" }

func (c *dnsDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dnsDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
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
		if errors.Is(err, v3.ErrNotFound) {
			return fmt.Errorf("resource not found")
		}
		return err
	}

	if !c.Force && !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to delete %q domain?", domain.UnicodeName)) {
		return nil
	}

	op, err := client.DeleteDNSDomain(ctx, domain.ID)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Deleting DNS domain %q...", domain.UnicodeName), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Printf("Domain %q was deleted successfully\n", domain.UnicodeName)
	}

	return nil
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &dnsDeleteCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
