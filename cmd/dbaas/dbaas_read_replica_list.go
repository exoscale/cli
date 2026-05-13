package dbaas

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasReadReplicaListItemOutput struct {
	ReplicaName string `json:"replica_name"`
	ReplicaZone string `json:"replica_zone"`
	Type        string `json:"type"`
	Plan        string `json:"plan"`
	State       string `json:"state"`
	Status      string `json:"status"`
	IsActive    bool   `json:"is_active"`
	IsEnabled   bool   `json:"is_enabled"`
}

type dbaasReadReplicaListOutput []dbaasReadReplicaListItemOutput

func (o *dbaasReadReplicaListOutput) ToJSON()  { output.JSON(o) }
func (o *dbaasReadReplicaListOutput) ToText()  { output.Text(o) }
func (o *dbaasReadReplicaListOutput) ToTable() { output.Table(o) }

type dbaasReadReplicaListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	ServiceName string `cli-arg:"#" cli-usage:"SERVICE-NAME"`
}

func (c *dbaasReadReplicaListCmd) CmdAliases() []string { return exocmd.GListAlias }

func (c *dbaasReadReplicaListCmd) CmdShort() string {
	return "List Database Service read replicas"
}

func (c *dbaasReadReplicaListCmd) CmdLong() string {
	return "List all read replicas of a primary Database Service across all zones."
}

func (c *dbaasReadReplicaListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasReadReplicaListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	services, err := dbaasListServicesAllZones(ctx)
	if err != nil {
		return err
	}

	primaryService, err := dbaasFindServiceByNameInServices(c.ServiceName, services)
	if err != nil {
		return err
	}

	primaryDetails, err := dbaasGetV3(ctx, c.ServiceName, primaryService.Zone)
	if err != nil {
		return err
	}

	if !dbaasReadReplicaSupportedServiceType(string(primaryDetails.Type)) {
		return fmt.Errorf("read replicas are not supported for Database Service type %q", primaryDetails.Type)
	}

	out := dbaasReadReplicaListFromServices(primaryDetails, services)

	return c.OutputFunc(&out, nil)
}

func dbaasReadReplicaListFromServices(primaryService v3.DBAASServiceCommon, services []dbaasServiceWithZone) dbaasReadReplicaListOutput {
	serviceByName := make(map[string]dbaasServiceWithZone, len(services))
	for _, service := range services {
		serviceByName[string(service.Service.Name)] = service
	}

	out := make(dbaasReadReplicaListOutput, 0)
	primaryName := string(primaryService.Name)

	for _, integration := range primaryService.Integrations {
		if integration.Type != "read_replica" || integration.Source != primaryName || integration.Dest == "" {
			continue
		}

		replicaService, ok := serviceByName[integration.Dest]
		if !ok {
			continue
		}

		out = append(out, dbaasReadReplicaListItemOutput{
			ReplicaName: string(replicaService.Service.Name),
			ReplicaZone: replicaService.Zone,
			Type:        string(replicaService.Service.Type),
			Plan:        replicaService.Service.Plan,
			State:       string(replicaService.Service.State),
			Status:      integration.Status,
			IsActive:    utils.DefaultBool(integration.ISActive, false),
			IsEnabled:   utils.DefaultBool(integration.ISEnabled, false),
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].ReplicaName == out[j].ReplicaName {
			return out[i].ReplicaZone < out[j].ReplicaZone
		}
		return out[i].ReplicaName < out[j].ReplicaName
	})

	return out
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasReadReplicaCmd, &dbaasReadReplicaListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
