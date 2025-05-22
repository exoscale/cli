package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type sksNodepoolUpdateCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	Cluster  string `cli-arg:"#" cli-usage:"CLUSTER-NAME|ID"`
	Nodepool string `cli-arg:"#" cli-usage:"NODEPOOL-NAME|ID"`

	AntiAffinityGroups []string    `cli-flag:"anti-affinity-group" cli-usage:"Nodepool Anti-Affinity Group NAME|ID (can be specified multiple times)"`
	DeployTarget       string      `cli-usage:"Nodepool Deploy Target NAME|ID"`
	Description        string      `cli-usage:"Nodepool description"`
	DiskSize           int64       `cli-usage:"Nodepool Compute instances disk size"`
	InstancePrefix     string      `cli-usage:"string to prefix Nodepool member names with"`
	InstanceType       string      `cli-usage:"Nodepool Compute instances type"`
	Labels             []string    `cli-flag:"label" cli-usage:"Nodepool label (format: KEY=VALUE, can be repeated multiple times)"`
	Name               string      `cli-usage:"Nodepool name"`
	PrivateNetworks    []string    `cli-flag:"private-network" cli-usage:"Nodepool Private Network NAME|ID (can be specified multiple times)"`
	SecurityGroups     []string    `cli-flag:"security-group" cli-usage:"Nodepool Security Group NAME|ID (can be specified multiple times)"`
	Taints             []string    `cli-flag:"taint" cli-usage:"Kubernetes taint to apply to Nodepool Nodes (format: KEY=VALUE:EFFECT, can be specified multiple times)"`
	Zone               v3.ZoneName `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksNodepoolUpdateCmd) CmdAliases() []string { return nil }

func (c *sksNodepoolUpdateCmd) CmdShort() string { return "Update an SKS cluster Nodepool" }

func (c *sksNodepoolUpdateCmd) CmdLong() string {
	return fmt.Sprintf(`This command updates an SKS Nodepool.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sksNodepoolShowOutput{}), ", "),
	)
}

func (c *sksNodepoolUpdateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksNodepoolUpdateCmd) CmdRun(cmd *cobra.Command, _ []string) error { //nolint:gocyclo
	var updated bool

	ctx := GContext
	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}
	resp, err := client.ListSKSClusters(ctx)
	if err != nil {
		return err
	}

	cluster, err := resp.FindSKSCluster(c.Cluster)
	if err != nil {
		return err
	}

	var nodepool *v3.SKSNodepool
	for _, n := range cluster.Nodepools {
		if n.ID.String() == c.Nodepool || n.Name == c.Nodepool {
			nodepool = &n
			break
		}
	}
	if nodepool == nil {
		return errors.New("nodepool not found")
	}

	updateReq := v3.UpdateSKSNodepoolRequest{}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.Description)) {
		updateReq.Description = c.Description
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.DiskSize)) {
		updateReq.DiskSize = c.DiskSize
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.InstancePrefix)) {
		updateReq.InstancePrefix = c.InstancePrefix
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.Name)) {
		updateReq.Name = c.Name
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.AntiAffinityGroups)) ||
		cmd.Flags().Changed(MustCLICommandFlagName(c, &c.DeployTarget)) ||
		cmd.Flags().Changed(MustCLICommandFlagName(c, &c.InstanceType)) ||
		cmd.Flags().Changed(MustCLICommandFlagName(c, &c.PrivateNetworks)) ||
		cmd.Flags().Changed(MustCLICommandFlagName(c, &c.SecurityGroups)) {

		if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.AntiAffinityGroups)) {
			aags, err := lookupAntiAffinityGroups(ctx, client, c.AntiAffinityGroups)
			if err != nil {
				return err
			}
			updateReq.AntiAffinityGroups = aags
			updated = true
		}

		if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.DeployTarget)) {
			dt, err := lookupDeployTarget(ctx, client, c.DeployTarget)
			if err != nil {
				return err
			}
			if dt != nil {
				updateReq.DeployTarget = dt
				updated = true
			}
		}

		if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.InstanceType)) {
			it, err := lookupInstanceType(ctx, client, c.InstanceType)
			if err != nil {
				return err
			}
			if it != nil {
				updateReq.InstanceType = it
				updated = true
			}
		}

		if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.PrivateNetworks)) {
			pns, err := lookupPrivateNetworks(ctx, client, c.PrivateNetworks)
			if err != nil {
				return err
			}
			updateReq.PrivateNetworks = pns
			updated = true
		}

		if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.SecurityGroups)) {
			sgs, err := lookupSecurityGroups(ctx, client, c.SecurityGroups)
			if err != nil {
				return err
			}
			updateReq.SecurityGroups = sgs
			updated = true
		}
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.Labels)) {
		if updateReq.Labels == nil {
			updateReq.Labels = v3.SKSNodepoolLabels{}
		}
		if len(c.Labels) > 0 {
			labels, err := utils.SliceToMap(c.Labels)
			if err != nil {
				return fmt.Errorf("label: %w", err)
			}
			for k, v := range labels {
				(updateReq.Labels)[k] = v
			}
		}
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.Taints)) {
		if nodepool.Taints == nil {
			updateReq.Taints = v3.SKSNodepoolTaints{}
		}
		for _, t := range c.Taints {
			key, taint, err := parseSKSNodepoolTaint(t)
			if err != nil {
				return fmt.Errorf("invalid taint value %q: %w", t, err)
			}
			(updateReq.Taints)[key] = *taint
		}

		updated = true
	}

	if updated {
		op, err := client.UpdateSKSNodepool(ctx, cluster.ID, nodepool.ID, updateReq)
		decorateAsyncOperation(fmt.Sprintf("Updating Nodepool %q...", c.Nodepool), func() {
			_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		})
		if err != nil {
			return err
		}
	}

	if !globalstate.Quiet {
		return (&sksNodepoolShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Cluster:            cluster.ID.String(),
			Nodepool:           nodepool.ID.String(),
			Zone:               v3.ZoneName(c.Zone),
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(RegisterCLICommand(sksNodepoolCmd, &sksNodepoolUpdateCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
