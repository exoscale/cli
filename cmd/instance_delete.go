package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

type InstanceDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Instances []string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *InstanceDeleteCmd) cmdAliases() []string { return gRemoveAlias }

func (c *InstanceDeleteCmd) cmdShort() string { return "Delete a Compute instance" }

func (c *InstanceDeleteCmd) cmdLong() string { return "" }

func (c *InstanceDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *InstanceDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(
		ctx,
		globalstate.EgoscaleV3Client,
		v3.ZoneName(c.Zone),
	)
	if err != nil {
		return err
	}

	instances, err := client.ListInstances(ctx)
	if err != nil {
		return err
	}

	instanceToDelete := []v3.UUID{}
	for _, i := range c.Instances {
		instance, err := instances.FindListInstancesResponseInstances(i)
		if err != nil {
			if !c.Force {
				return err
			}
			fmt.Fprintf(os.Stderr, "warning: %s not found.\n", i)

			continue
		}

		if !c.Force {
			if !askQuestion(fmt.Sprintf("Are you sure you want to delete instance %q?", i)) {
				return nil
			}
		}

		instanceToDelete = append(instanceToDelete, instance.ID)
	}

	var fns []func() error
	for _, i := range instanceToDelete {
		fns = append(fns, func() error {
			op, err := client.DeleteInstance(ctx, i)
			if err != nil {
				return err
			}
			_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
			return err
		})
	}

	err = decorateAsyncOperations(fmt.Sprintf("Deleting instance %q...", strings.Join(c.Instances, ", ")), fns...)
	if err != nil {
		return err
	}

	// Cleaning up resources created in create instance
	// https://github.com/exoscale/cli/blob/master/cmd/instance_create.go#L220
	for _, i := range instanceToDelete {
		instanceDir := path.Join(globalstate.ConfigFolder, "instances", i.String())
		if _, err := os.Stat(instanceDir); !os.IsNotExist(err) {
			if err := os.RemoveAll(instanceDir); err != nil {
				return fmt.Errorf("error deleting instance directory: %w", err)
			}
		}
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceCmd, &InstanceDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
