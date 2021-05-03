package cmd

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instancePoolShowOutput struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	ServiceOffering    string   `json:"service_offering"`
	Template           string   `json:"templateid"`
	Zone               string   `json:"zoneid"`
	AntiAffinityGroups []string `json:"anti_affinity_groups" outputLabel:"Anti-Affinity Groups"`
	SecurityGroups     []string `json:"security_groups"`
	PrivateNetworks    []string `json:"private_networks"`
	ElasticIPs         []string `json:"elastic_ips" outputLabel:"Elastic IPs"`
	IPv6               bool     `json:"ipv6" outputLabel:"IPv6"`
	SSHKey             string   `json:"ssh_key"`
	Size               int64    `json:"size"`
	DiskSize           string   `json:"disk_size"`
	InstancePrefix     string   `json:"instance_prefix"`
	State              string   `json:"state"`
	Instances          []string `json:"instances"`
}

func (o *instancePoolShowOutput) toJSON()  { outputJSON(o) }
func (o *instancePoolShowOutput) toText()  { outputText(o) }
func (o *instancePoolShowOutput) toTable() { outputTable(o) }

var instancePoolShowCmd = &cobra.Command{
	Use:   "show NAME|ID",
	Short: "Show an Instance Pool details",
	Long: fmt.Sprintf(`This command shows an Instance Pool details.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instancePoolShowOutput{}), ", ")),
	Aliases: gShowAlias,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
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

		if showUserData, _ := cmd.Flags().GetBool("user-data"); showUserData {
			instancePool, err := lookupInstancePool(
				exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone)),
				zone,
				args[0],
			)
			if err != nil {
				return err
			}

			if instancePool.UserData != "" {
				userData, err := decodeUserData(instancePool.UserData)
				if err != nil {
					return fmt.Errorf("error decoding user data: %s", err)
				}

				fmt.Print(userData)
			}

			return nil
		}

		return output(showInstancePool(zone, args[0]))
	},
}

func showInstancePool(zone, i string) (outputter, error) {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

	instancePool, err := lookupInstancePool(ctx, zone, i)
	if err != nil {
		return nil, err
	}

	out := instancePoolShowOutput{
		ID:             instancePool.ID,
		Name:           instancePool.Name,
		Description:    instancePool.Description,
		Zone:           zone,
		IPv6:           instancePool.IPv6Enabled,
		SSHKey:         instancePool.SSHKey,
		Size:           instancePool.Size,
		DiskSize:       humanize.IBytes(uint64(instancePool.DiskSize << 30)),
		InstancePrefix: instancePool.InstancePrefix,
		State:          instancePool.State,
	}

	zoneV1, err := getZoneByNameOrID(zone)
	if err != nil {
		return nil, err
	}

	serviceOffering, err := getServiceOfferingByNameOrID(instancePool.InstanceTypeID)
	if err != nil {
		return nil, err
	}
	out.ServiceOffering = serviceOffering.Name

	template, err := getTemplateByNameOrID(zoneV1.ID, instancePool.TemplateID, "featured")
	if err != nil {
		return nil, err
	}
	out.Template = template.Name

	antiAffinityGroups, err := instancePool.AntiAffinityGroups(ctx)
	if err != nil {
		return nil, err
	}
	for _, antiAffinityGroup := range antiAffinityGroups {
		out.AntiAffinityGroups = append(out.AntiAffinityGroups, antiAffinityGroup.Name)
	}

	securityGroups, err := instancePool.SecurityGroups(ctx)
	if err != nil {
		return nil, err
	}
	for _, securityGroup := range securityGroups {
		out.SecurityGroups = append(out.SecurityGroups, securityGroup.Name)
	}

	privateNetworks, err := instancePool.PrivateNetworks(ctx)
	if err != nil {
		return nil, err
	}
	for _, privateNetwork := range privateNetworks {
		out.PrivateNetworks = append(out.PrivateNetworks, privateNetwork.Name)
	}

	instances, err := instancePool.Instances(ctx)
	if err != nil {
		return nil, err
	}
	for _, instance := range instances {
		out.Instances = append(out.Instances, instance.Name)
	}

	elasticIPs, err := instancePool.ElasticIPs(ctx)
	if err != nil {
		return nil, err
	}
	for _, elasticIP := range elasticIPs {
		out.ElasticIPs = append(out.ElasticIPs, elasticIP.IPAddress.String())
	}

	return &out, nil
}

func init() {
	instancePoolShowCmd.Flags().BoolP("user-data", "u", false, "show cloud-init user data configuration")
	instancePoolShowCmd.Flags().StringP("zone", "z", "", "Instance pool zone")
	instancePoolCmd.AddCommand(instancePoolShowCmd)
}
