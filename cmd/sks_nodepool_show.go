package cmd

import (
	"errors"
	"fmt"
	"strings"

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
}

func (o *sksNodepoolShowOutput) Type() string { return "SKS Nodepool" }
func (o *sksNodepoolShowOutput) toJSON()      { outputJSON(o) }
func (o *sksNodepoolShowOutput) toText()      { outputText(o) }
func (o *sksNodepoolShowOutput) toTable()     { outputTable(o) }

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
		strings.Join(outputterTemplateAnnotations(&sksNodepoolShowOutput{}), ", "))
}

func (c *sksNodepoolShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksNodepoolShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	return output(showSKSNodepool(c.Zone, c.Cluster, c.Nodepool))
}

func showSKSNodepool(zone, c, np string) (outputter, error) {
	var nodepool *egoscale.SKSNodepool

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

	cluster, err := cs.FindSKSCluster(ctx, zone, c)
	if err != nil {
		return nil, err
	}

	for _, n := range cluster.Nodepools {
		if *n.ID == np || *n.Name == np {
			nodepool = n
			break
		}
	}
	if nodepool == nil {
		return nil, errors.New("Nodepool not found") // nolint:golint
	}

	out := sksNodepoolShowOutput{
		AntiAffinityGroups: make([]string, 0),
		CreationDate:       nodepool.CreatedAt.String(),
		Description:        defaultString(nodepool.Description, ""),
		DiskSize:           *nodepool.DiskSize,
		ID:                 *nodepool.ID,
		InstancePoolID:     *nodepool.InstancePoolID,
		InstancePrefix:     defaultString(nodepool.InstancePrefix, ""),
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
			antiAffinityGroup, err := cs.GetAntiAffinityGroup(ctx, zone, id)
			if err != nil {
				return nil, fmt.Errorf("error retrieving Anti-Affinity Group: %v", err)
			}
			out.AntiAffinityGroups = append(out.AntiAffinityGroups, *antiAffinityGroup.Name)
		}
	}

	if nodepool.PrivateNetworkIDs != nil {
		for _, id := range *nodepool.PrivateNetworkIDs {
			privateNetwork, err := cs.GetPrivateNetwork(ctx, zone, id)
			if err != nil {
				return nil, fmt.Errorf("error retrieving Private Network: %v", err)
			}
			out.PrivateNetworks = append(out.PrivateNetworks, *privateNetwork.Name)
		}
	}

	if nodepool.SecurityGroupIDs != nil {
		for _, id := range *nodepool.SecurityGroupIDs {
			securityGroup, err := cs.GetSecurityGroup(ctx, zone, id)
			if err != nil {
				return nil, fmt.Errorf("error retrieving Security Group: %v", err)
			}
			out.SecurityGroups = append(out.SecurityGroups, *securityGroup.Name)
		}
	}

	serviceOffering, err := cs.GetInstanceType(ctx, zone, *nodepool.InstanceTypeID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving service offering: %s", err)
	}
	out.InstanceType = *serviceOffering.Size

	template, err := cs.GetTemplate(ctx, zone, *nodepool.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving template: %s", err)
	}
	out.Template = *template.Name

	return &out, nil
}

func init() {
	cobra.CheckErr(registerCLICommand(sksNodepoolCmd, &sksNodepoolShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
