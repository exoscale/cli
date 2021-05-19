package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	exov2 "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type sksNodepoolShowOutput struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	CreationDate       string   `json:"creation_date"`
	InstancePoolID     string   `json:"instance_pool_id"`
	InstancePrefix     string   `json:"instance_prefix"`
	InstanceType       string   `json:"instance_type"`
	Template           string   `json:"template"`
	DiskSize           int64    `json:"disk_size"`
	AntiAffinityGroups []string `json:"anti_affinity_groups"`
	SecurityGroups     []string `json:"security_groups"`
	Version            string   `json:"version"`
	Size               int64    `json:"size"`
	State              string   `json:"state"`
}

func (o *sksNodepoolShowOutput) toJSON()      { outputJSON(o) }
func (o *sksNodepoolShowOutput) toText()      { outputText(o) }
func (o *sksNodepoolShowOutput) toTable()     { outputTable(o) }
func (o *sksNodepoolShowOutput) Type() string { return "SKS Nodepool" }

var sksNodepoolShowCmd = &cobra.Command{
	Use:   "show CLUSTER-NAME|ID NODEPOOL-NAME|ID",
	Short: "Show a SKS cluster Nodepool details",
	Long: fmt.Sprintf(`This command shows a SKS cluster Nodepool details.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&sksNodepoolShowOutput{}), ", ")),
	Aliases: gShowAlias,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		return output(showSKSNodepool(zone, args[0], args[1]))
	},
}

func showSKSNodepool(zone, c, np string) (outputter, error) {
	var nodepool *exov2.SKSNodepool

	zoneV1, err := getZoneByNameOrID(zone)
	if err != nil {
		return nil, err
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))
	cluster, err := lookupSKSCluster(ctx, zone, c)
	if err != nil {
		return nil, err
	}

	for _, n := range cluster.Nodepools {
		if n.ID == np || n.Name == np {
			nodepool = n
			break
		}
	}
	if nodepool == nil {
		return nil, errors.New("Nodepool not found") // nolint:golint
	}

	out := sksNodepoolShowOutput{
		ID:                 nodepool.ID,
		Name:               nodepool.Name,
		Description:        nodepool.Description,
		CreationDate:       nodepool.CreatedAt.String(),
		InstancePoolID:     nodepool.InstancePoolID,
		InstancePrefix:     nodepool.InstancePrefix,
		DiskSize:           nodepool.DiskSize,
		AntiAffinityGroups: make([]string, 0),
		SecurityGroups:     make([]string, 0),
		Version:            nodepool.Version,
		Size:               nodepool.Size,
		State:              nodepool.State,
	}

	serviceOffering, err := getServiceOfferingByNameOrID(nodepool.InstanceTypeID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving service offering: %s", err)
	}
	out.InstanceType = serviceOffering.Name

	template, err := getTemplateByNameOrID(zoneV1.ID,
		egoscale.MustParseUUID(nodepool.TemplateID).String(),
		"featured")
	if err != nil {
		return nil, fmt.Errorf("error retrieving template: %s", err)
	}
	out.Template = template.Name

	if len(nodepool.AntiAffinityGroupIDs) > 0 {
		allAntiAffinityGroups, err := cs.ListWithContext(gContext, &egoscale.AffinityGroup{})
		if err != nil {
			return nil, fmt.Errorf("error listing Anti-Affinity Groups: %s", err)
		}

		for _, s := range allAntiAffinityGroups {
			sg := s.(*egoscale.AffinityGroup)

			for _, id := range nodepool.AntiAffinityGroupIDs {
				if sg.ID.String() == id {
					out.AntiAffinityGroups = append(out.AntiAffinityGroups, sg.Name)
				}
			}
		}
	}

	if len(nodepool.SecurityGroupIDs) > 0 {
		allSecurityGroups, err := cs.ListWithContext(gContext, &egoscale.SecurityGroup{})
		if err != nil {
			return nil, fmt.Errorf("error listing Security Groups: %s", err)
		}

		for _, s := range allSecurityGroups {
			sg := s.(*egoscale.SecurityGroup)

			for _, id := range nodepool.SecurityGroupIDs {
				if sg.ID.String() == id {
					out.SecurityGroups = append(out.SecurityGroups, sg.Name)
				}
			}
		}
	}

	return &out, nil
}

func init() {
	sksNodepoolShowCmd.Flags().StringP("zone", "z", "", "SKS cluster zone")
	sksNodepoolCmd.AddCommand(sksNodepoolShowCmd)
}
