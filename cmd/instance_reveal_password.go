package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type instanceRevealCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"reveal-password"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`
	Zone     string `cli-short:"z" cli-usage:"instance zone"`
}

type instanceRevealOutput struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}

func (o *instanceRevealOutput) Type() string { return "Compute instance" }
func (o *instanceRevealOutput) ToJSON()      { output.JSON(o) }
func (o *instanceRevealOutput) ToText()      { output.Text(o) }
func (o *instanceRevealOutput) ToTable()     { output.Table(o) }

func (c *instanceRevealCmd) cmdAliases() []string { return nil }

func (c *instanceRevealCmd) cmdShort() string { return "Reveal the password of a Compute instance" }

func (c *instanceRevealCmd) cmdLong() string { return "" }

func (c *instanceRevealCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceRevealCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	instance, err := globalstate.EgoscaleClient.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	pwd, err := globalstate.EgoscaleClient.RevealInstancePassword(ctx, c.Zone, instance)
	if err != nil {
		return err
	}

	out := instanceRevealOutput{
		ID:       *instance.ID,
		Password: pwd,
	}
	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceCmd, &instanceRevealCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
