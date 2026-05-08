package dbaas

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
)

type dbaasReadReplicaListItemOutput struct {
	ReplicaName string `json:"replica-name"`
	ReplicaZone string `json:"replica-zone"`
	Type        string `json:"type"`
	Plan        string `json:"plan"`
	State       string `json:"state"`
	Status      string `json:"status"`
	IsActive    bool   `json:"is-active"`
	IsEnabled   bool   `json:"is-enabled"`
}

type dbaasReadReplicaListOutput []dbaasReadReplicaListItemOutput

func (o *dbaasReadReplicaListOutput) ToJSON()  { output.JSON(o) }
func (o *dbaasReadReplicaListOutput) ToText()  { output.Text(o) }
func (o *dbaasReadReplicaListOutput) ToTable() { output.Table(o) }

type dbaasReadReplicaListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	ServiceName string `cli-arg:"#" cli-usage:"name of the primary service"`
}

func (c *dbaasReadReplicaListCmd) CmdAliases() []string { return exocmd.GListAlias }

func (c *dbaasReadReplicaListCmd) CmdShort() string {
	return "List DBaaS read replicas"
}

func (c *dbaasReadReplicaListCmd) CmdLong() string {
	return "List all read replicas of a primary DBaaS service across all zones."
}

func (c *dbaasReadReplicaListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasReadReplicaListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	primaryService, err := dbaasFindServiceByNameAllZones(ctx, c.ServiceName)
	if err != nil {
		return err
	}

	if !dbaasReadReplicaSupportedServiceType(string(primaryService.Service.Type)) {
		return fmt.Errorf("read replicas are not supported for Database Service type %q", primaryService.Service.Type)
	}

	services, err := dbaasListServicesAllZones(ctx)
	if err != nil {
		return err
	}

	out := make(dbaasReadReplicaListOutput, 0)

	for _, service := range services {
		replicaIntegration := dbaasGetReadReplicaIntegrationForReplica(service.Service)
		if replicaIntegration == nil {
			continue
		}
		if replicaIntegration.Source != c.ServiceName {
			continue
		}

		out = append(out, dbaasReadReplicaListItemOutput{
			ReplicaName: string(service.Service.Name),
			ReplicaZone: service.Zone,
			Type:        string(service.Service.Type),
			Plan:        service.Service.Plan,
			State:       string(service.Service.State),
			Status:      replicaIntegration.Status,
			IsActive:    utils.DefaultBool(replicaIntegration.ISActive, false),
			IsEnabled:   utils.DefaultBool(replicaIntegration.ISEnabled, false),
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].ReplicaName == out[j].ReplicaName {
			return out[i].ReplicaZone < out[j].ReplicaZone
		}
		return out[i].ReplicaName < out[j].ReplicaName
	})

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasReadReplicaCmd, &dbaasReadReplicaListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
