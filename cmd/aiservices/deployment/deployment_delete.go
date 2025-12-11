package deployment

import (
	"fmt"
	"os"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type DeploymentDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Deployments []string    `cli-arg:"#" cli-usage:"NAME|ID..."`
	Force       bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone        v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *DeploymentDeleteCmd) CmdAliases() []string { return exocmd.GDeleteAlias }
func (c *DeploymentDeleteCmd) CmdShort() string     { return "Delete AI deployment" }
func (c *DeploymentDeleteCmd) CmdLong() string {
	return "This command deletes an AI deployment by ID or name."
}
func (c *DeploymentDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *DeploymentDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	// Resolve deployment IDs using the SDK helper
	list, err := client.ListDeployments(ctx)
	if err != nil {
		return err
	}

	deploymentsToDelete := []v3.UUID{}
	for _, deploymentStr := range c.Deployments {
		entry, err := list.FindListDeploymentsResponseEntry(deploymentStr)
		if err != nil {
			if !c.Force {
				return err
			}
			fmt.Fprintf(os.Stderr, "warning: %s not found.\n", deploymentStr)
			continue
		}

		if !c.Force {
			if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to delete deployment %q?", deploymentStr)) {
				return nil
			}
		}

		deploymentsToDelete = append(deploymentsToDelete, entry.ID)
	}

	var fns []func() error
	for _, id := range deploymentsToDelete {
		fns = append(fns, func() error {
			op, err := client.DeleteDeployment(ctx, id)
			if err != nil {
				return err
			}
			_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
			return err
		})
	}

	err = utils.DecorateAsyncOperations(fmt.Sprintf("Deleting deployment(s)..."), fns...)
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Fprintln(os.Stdout, "Deployment(s) deleted.")
	}
	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &DeploymentDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
