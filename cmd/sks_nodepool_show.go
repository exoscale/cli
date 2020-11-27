package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	apiv2 "github.com/exoscale/egoscale/api/v2"
	"github.com/spf13/cobra"
)

type sksNodepoolShowOutput struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	CreationDate   string   `json:"creation_date"`
	InstancePoolID string   `json:"instance_pool_id"`
	InstanceType   string   `json:"instance_type"`
	Template       string   `json:"template"`
	DiskSize       int64    `json:"disk_size"`
	SecurityGroups []string `json:"security_groups"`
	Version        string   `json:"version"`
	Size           int64    `json:"size"`
	State          string   `json:"state"`
}

func (o *sksNodepoolShowOutput) toJSON()      { outputJSON(o) }
func (o *sksNodepoolShowOutput) toText()      { outputText(o) }
func (o *sksNodepoolShowOutput) toTable()     { outputTable(o) }
func (o *sksNodepoolShowOutput) Type() string { return "SKS Nodepool" }

var sksNodepoolShowCmd = &cobra.Command{
	Use:   "show <cluster name | ID> <Nodepool name | ID>",
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
		z, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		zone, err := getZoneByNameOrID(z)
		if err != nil {
			return fmt.Errorf("error retrieving zone: %s", err)
		}

		return output(showSKSNodepool(zone, args[0], args[1]))
	},
}

func showSKSNodepool(zone *egoscale.Zone, c, np string) (outputter, error) {
	var nodepool *egoscale.SKSNodepool

	ctx := apiv2.WithEndpoint(gContext, apiv2.NewReqEndpoint(gCurrentAccount.Environment, zone.Name))
	cluster, err := lookupSKSCluster(ctx, zone.Name, c)
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
		ID:             nodepool.ID,
		Name:           nodepool.Name,
		Description:    nodepool.Description,
		CreationDate:   nodepool.CreatedAt.String(),
		InstancePoolID: nodepool.InstancePoolID,
		DiskSize:       nodepool.DiskSize,
		Version:        nodepool.Version,
		State:          nodepool.State,
		Size:           nodepool.Size,
		SecurityGroups: make([]string, 0),
	}

	serviceOffering, err := getServiceOfferingByNameOrID(nodepool.InstanceTypeID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving service offering: %s", err)
	}
	out.InstanceType = serviceOffering.Name

	template, err := getTemplateByNameOrID(zone.ID,
		egoscale.MustParseUUID(nodepool.TemplateID).String(),
		"featured")
	if err != nil {
		return nil, fmt.Errorf("error retrieving template: %s", err)
	}
	out.Template = template.Name

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
