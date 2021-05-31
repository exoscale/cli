package cmd

import (
	"errors"
	"fmt"
	"strings"

	exov2 "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var sksNodepoolResetFields = []string{
	"anti-affinity-groups",
	"deploy-target",
	"description",
	"security-groups",
}

type sksNodepoolUpdateCmd struct {
	_ bool `cli-cmd:"update"`

	Cluster  string `cli-arg:"#" cli-usage:"CLUSTER-NAME|ID"`
	Nodepool string `cli-arg:"#" cli-usage:"NODEPOOL-NAME|ID"`

	AntiAffinityGroups []string `cli-flag:"anti-affinity-group" cli-usage:"Nodepool Anti-Affinity Group NAME|ID (can be specified multiple times)"`
	DeployTarget       string   `cli-usage:"Nodepool Deploy Target NAME|ID"`
	Description        string   `cli-usage:"Nodepool description"`
	DiskSize           int64    `cli-usage:"Nodepool Compute instances disk size"`
	InstancePrefix     string   `cli-usage:"string to prefix Nodepool member names with"`
	InstanceType       string   `cli-usage:"Nodepool Compute instances type"`
	Name               string   `cli-usage:"Nodepool name"`
	ResetFields        []string `cli-flag:"reset" cli-usage:"properties to reset to default value"`
	SecurityGroups     []string `cli-flag:"security-group" cli-usage:"Nodepool Security Group NAME|ID (can be specified multiple times)"`
	Zone               string   `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksNodepoolUpdateCmd) cmdAliases() []string { return nil }

func (c *sksNodepoolUpdateCmd) cmdShort() string { return "Update an SKS cluster Nodepool" }

func (c *sksNodepoolUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command updates an SKS Nodepool.

Supported output template annotations: %s

Support values for --reset flag: %s`,
		strings.Join(outputterTemplateAnnotations(&sksNodepoolShowOutput{}), ", "),
		strings.Join(sksNodepoolResetFields, ", "),
	)
}

func (c *sksNodepoolUpdateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksNodepoolUpdateCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	var (
		nodepool *exov2.SKSNodepool
		updated  bool
	)

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	cluster, err := cs.FindSKSCluster(ctx, c.Zone, c.Cluster)
	if err != nil {
		return err
	}

	for _, n := range cluster.Nodepools {
		if n.ID == c.Nodepool || n.Name == c.Nodepool {
			nodepool = n
			break
		}
	}
	if nodepool == nil {
		return errors.New("Nodepool not found") // nolint:golint
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.AntiAffinityGroups)) {
		nodepool.AntiAffinityGroupIDs = make([]string, 0)
		for _, v := range c.AntiAffinityGroups {
			antiAffinityGroup, err := cs.FindAntiAffinityGroup(ctx, c.Zone, v)
			if err != nil {
				return fmt.Errorf("error retrieving Anti-Affinity Group: %s", err)
			}
			nodepool.AntiAffinityGroupIDs = append(nodepool.AntiAffinityGroupIDs, antiAffinityGroup.ID)
		}
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.DeployTarget)) {
		deployTarget, err := cs.FindDeployTarget(ctx, c.Zone, c.DeployTarget)
		if err != nil {
			return fmt.Errorf("error retrieving Deploy Target: %s", err)
		}
		nodepool.DeployTargetID = deployTarget.ID
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Description)) {
		nodepool.Description = c.Description
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.DiskSize)) {
		nodepool.DiskSize = c.DiskSize
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.InstancePrefix)) {
		nodepool.InstancePrefix = c.InstancePrefix
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.InstanceType)) {
		nodepoolInstanceType, err := cs.FindInstanceType(ctx, c.Zone, c.InstanceType)
		if err != nil {
			return fmt.Errorf("error retrieving instance type: %s", err)
		}
		nodepool.InstanceTypeID = nodepoolInstanceType.ID
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Name)) {
		nodepool.Name = c.Name
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.SecurityGroups)) {
		nodepool.SecurityGroupIDs = make([]string, 0)
		for _, v := range c.SecurityGroups {
			securityGroup, err := cs.FindSecurityGroup(ctx, c.Zone, v)
			if err != nil {
				return fmt.Errorf("error retrieving Security Group: %s", err)
			}
			nodepool.SecurityGroupIDs = append(nodepool.SecurityGroupIDs, securityGroup.ID)
		}
		updated = true
	}

	decorateAsyncOperation(fmt.Sprintf("Updating Nodepool %q...", c.Nodepool), func() {
		if updated {
			if err = cluster.UpdateNodepool(ctx, nodepool); err != nil {
				return
			}
		}

		for _, f := range c.ResetFields {
			switch f {
			case "anti-affinity-groups":
				err = cluster.ResetNodepoolField(ctx, nodepool, &nodepool.AntiAffinityGroupIDs)
			case "deploy-target":
				err = cluster.ResetNodepoolField(ctx, nodepool, &nodepool.DeployTargetID)
			case "description":
				err = cluster.ResetNodepoolField(ctx, nodepool, &nodepool.Description)
			case "security-groups":
				err = cluster.ResetNodepoolField(ctx, nodepool, &nodepool.SecurityGroupIDs)
			}
			if err != nil {
				return
			}
		}
	})
	if err != nil {
		return err
	}

	if !gQuiet {
		return output(showSKSNodepool(c.Zone, cluster.ID, nodepool.ID))
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(sksNodepoolCmd, &sksNodepoolUpdateCmd{}))
}
