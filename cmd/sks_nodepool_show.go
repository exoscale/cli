package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type sksNodepoolShowOutput struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	Description        string            `json:"description"`
	CreationDate       string            `json:"creation_date"`
	InstancePoolID     string            `json:"instance_pool_id"`
	InstancePrefix     string            `json:"instance_prefix"`
	InstanceType       string            `json:"instance_type"`
	Template           string            `json:"template"`
	DiskSize           int64             `json:"disk_size"`
	AntiAffinityGroups []string          `json:"anti_affinity_groups"`
	SecurityGroups     []string          `json:"security_groups"`
	PrivateNetworks    []string          `json:"private_networks"`
	Version            string            `json:"version"`
	Size               int64             `json:"size"`
	State              string            `json:"state"`
	Taints             []string          `json:"taints"`
	Labels             map[string]string `json:"labels"`
	AddOns             []string          `json:"addons"`
}

func (o *sksNodepoolShowOutput) Type() string { return "SKS Nodepool" }
func (o *sksNodepoolShowOutput) toJSON()      { output.JSON(o) }
func (o *sksNodepoolShowOutput) toText()      { output.Text(o) }
func (o *sksNodepoolShowOutput) toTable()     { output.Table(o) }

type sksNodepoolShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Cluster  string `cli-arg:"#" cli-usage:"CLUSTER-NAME|ID"`
	Nodepool string `cli-arg:"#" cli-usage:"NODEPOOL-NAME|ID"`

	Zone string `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksNodepoolShowCmd) cmdAliases() []string { return gShowAlias }

func (c *sksNodepoolShowCmd) cmdShort() string { return "Show an SKS cluster Nodepool details" }

func (c *sksNodepoolShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows an SKS cluster Nodepool details.

Supported output template annotations: %s`,
		strings.Join(output.OutputterTemplateAnnotations(&sksNodepoolShowOutput{}), ", "))
}

func (c *sksNodepoolShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksNodepoolShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	var nodepool *egoscale.SKSNodepool

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	cluster, err := cs.FindSKSCluster(ctx, c.Zone, c.Cluster)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	for _, n := range cluster.Nodepools {
		if *n.ID == c.Nodepool || *n.Name == c.Nodepool {
			nodepool = n
			break
		}
	}
	if nodepool == nil {
		return errors.New("Nodepool not found") // nolint:golint
	}

	out := sksNodepoolShowOutput{
		AddOns: func() (v []string) {
			if nodepool.AddOns != nil {
				v = *nodepool.AddOns
			}
			return
		}(),
		AntiAffinityGroups: make([]string, 0),
		CreationDate:       nodepool.CreatedAt.String(),
		Description:        utils.DefaultString(nodepool.Description, ""),
		DiskSize:           *nodepool.DiskSize,
		ID:                 *nodepool.ID,
		InstancePoolID:     *nodepool.InstancePoolID,
		InstancePrefix:     utils.DefaultString(nodepool.InstancePrefix, ""),
		Labels: func() (v map[string]string) {
			if nodepool.Labels != nil {
				v = *nodepool.Labels
			}
			return
		}(),
		Name:            *nodepool.Name,
		SecurityGroups:  make([]string, 0),
		PrivateNetworks: make([]string, 0),
		Size:            *nodepool.Size,
		State:           *nodepool.State,
		Taints: func() (v []string) {
			if nodepool.Taints != nil {
				v = make([]string, 0)
				for k, t := range *nodepool.Taints {
					v = append(v, fmt.Sprintf("%s=%s:%s", k, t.Value, t.Effect))
				}
			}
			return
		}(),
		Version: *nodepool.Version,
	}

	if nodepool.AntiAffinityGroupIDs != nil {
		for _, id := range *nodepool.AntiAffinityGroupIDs {
			antiAffinityGroup, err := cs.GetAntiAffinityGroup(ctx, c.Zone, id)
			if err != nil {
				return fmt.Errorf("error retrieving Anti-Affinity Group: %w", err)
			}
			out.AntiAffinityGroups = append(out.AntiAffinityGroups, *antiAffinityGroup.Name)
		}
	}

	if nodepool.PrivateNetworkIDs != nil {
		for _, id := range *nodepool.PrivateNetworkIDs {
			privateNetwork, err := cs.GetPrivateNetwork(ctx, c.Zone, id)
			if err != nil {
				return fmt.Errorf("error retrieving Private Network: %w", err)
			}
			out.PrivateNetworks = append(out.PrivateNetworks, *privateNetwork.Name)
		}
	}

	if nodepool.SecurityGroupIDs != nil {
		for _, id := range *nodepool.SecurityGroupIDs {
			securityGroup, err := cs.GetSecurityGroup(ctx, c.Zone, id)
			if err != nil {
				return fmt.Errorf("error retrieving Security Group: %w", err)
			}
			out.SecurityGroups = append(out.SecurityGroups, *securityGroup.Name)
		}
	}

	serviceOffering, err := cs.GetInstanceType(ctx, c.Zone, *nodepool.InstanceTypeID)
	if err != nil {
		return fmt.Errorf("error retrieving service offering: %w", err)
	}
	out.InstanceType = *serviceOffering.Size

	template, err := cs.GetTemplate(ctx, c.Zone, *nodepool.TemplateID)
	if err != nil {
		return fmt.Errorf("error retrieving template: %w", err)
	}
	out.Template = *template.Name

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(sksNodepoolCmd, &sksNodepoolShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
