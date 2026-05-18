package dbaas

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
)

type dbaasReadReplicaShowOutput struct {
	Name          string `json:"name"`
	Zone          string `json:"zone"`
	Type          string `json:"type"`
	Plan          string `json:"plan"`
	State         string `json:"state"`
	SourceService string `json:"source_service"`
	SourceZone    string `json:"source_zone,omitempty"`
	Status        string `json:"status"`
	IsActive      bool   `json:"is_active"`
	IsEnabled     bool   `json:"is_enabled"`
	DiskSize      int64  `json:"disk_size"`
	NodeCount     int64  `json:"node_count"`
	NodeCPUCount  int64  `json:"node_cpu_count"`
	NodeMemory    int64  `json:"node_memory"`
}

func (o *dbaasReadReplicaShowOutput) ToJSON()  { output.JSON(o) }
func (o *dbaasReadReplicaShowOutput) ToText()  { output.Text(o) }
func (o *dbaasReadReplicaShowOutput) ToTable() { output.Table(o) }

type dbaasReadReplicaShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Name string `cli-arg:"#" cli-usage:"REPLICA-NAME"`
	Zone string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasReadReplicaShowCmd) CmdAliases() []string { return exocmd.GShowAlias }

func (c *dbaasReadReplicaShowCmd) CmdShort() string {
	return "Show Database Service read replica details"
}

func (c *dbaasReadReplicaShowCmd) CmdLong() string {
	return "Show details of a specific Database Service read replica."
}

func (c *dbaasReadReplicaShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasReadReplicaShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	return c.OutputFunc(c.showReadReplica(exocmd.GContext))
}

func (c *dbaasReadReplicaShowCmd) showReadReplica(ctx context.Context) (output.Outputter, error) {
	commonSvc, err := dbaasGetV3(ctx, c.Name, c.Zone)
	if err != nil {
		return nil, err
	}

	if !dbaasReadReplicaSupportedServiceType(string(commonSvc.Type)) {
		return nil, fmt.Errorf("read replicas are not supported for Database Service type %q", commonSvc.Type)
	}

	replicaIntegration := dbaasGetReadReplicaIntegrationForReplica(commonSvc)
	if replicaIntegration == nil {
		return nil, fmt.Errorf("%q is not a read replica", c.Name)
	}

	sourceZone := ""
	sourceService, err := dbaasFindServiceByNameAllZones(ctx, replicaIntegration.Source)
	if err == nil {
		sourceZone = sourceService.Zone
	}

	out := &dbaasReadReplicaShowOutput{
		Name:          string(commonSvc.Name),
		Zone:          c.Zone,
		Type:          string(commonSvc.Type),
		Plan:          commonSvc.Plan,
		State:         string(commonSvc.State),
		SourceService: replicaIntegration.Source,
		SourceZone:    sourceZone,
		Status:        replicaIntegration.Status,
		IsActive:      utils.DefaultBool(replicaIntegration.ISActive, false),
		IsEnabled:     utils.DefaultBool(replicaIntegration.ISEnabled, false),
		DiskSize:      commonSvc.DiskSize,
		NodeCount:     commonSvc.NodeCount,
		NodeCPUCount:  commonSvc.NodeCPUCount,
		NodeMemory:    commonSvc.NodeMemory,
	}

	return out, nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasReadReplicaCmd, &dbaasReadReplicaShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
