package cmd

import (
	"fmt"
	"os"
	"strings"

	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instanceSnapshotListItemOutput struct {
	ID           string `json:"id"`
	CreationDate string `json:"creation_date"`
	Instance     string `json:"instance"`
	State        string `json:"state"`
	Zone         string `json:"zone"`
}

type instanceSnapshotListOutput []instanceSnapshotListItemOutput

func (o *instanceSnapshotListOutput) toJSON()  { outputJSON(o) }
func (o *instanceSnapshotListOutput) toText()  { outputText(o) }
func (o *instanceSnapshotListOutput) toTable() { outputTable(o) }

type instanceSnapshotListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone string `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *instanceSnapshotListCmd) cmdAliases() []string { return nil }

func (c *instanceSnapshotListCmd) cmdShort() string { return "List Compute instance snapshots" }

func (c *instanceSnapshotListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists existing Compute instance snapshots.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instanceSnapshotListOutput{}), ", "))
}

func (c *instanceSnapshotListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSnapshotListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	var zones []string

	if c.Zone != "" {
		zones = []string{c.Zone}
	} else {
		zones = allZones
	}

	out := make(instanceSnapshotListOutput, 0)
	res := make(chan instanceSnapshotListItemOutput)
	defer close(res)

	instances := make(map[string]*egoscale.Instance) // For caching

	go func() {
		for dt := range res {
			out = append(out, dt)
		}
	}()
	err := forEachZone(zones, func(zone string) error {
		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

		list, err := cs.ListSnapshots(ctx, zone)
		if err != nil {
			return fmt.Errorf("unable to list Compute instance snapshots in zone %s: %v", zone, err)
		}

		for _, s := range list {
			instance, cached := instances[*s.InstanceID]
			if !cached {
				instance, err = cs.GetInstance(ctx, zone, *s.InstanceID)
				if err != nil {
					return fmt.Errorf("unable to retrieve Compute instance %q: %s", *s.InstanceID, err)
				}
				instances[*s.InstanceID] = instance
			}

			res <- instanceSnapshotListItemOutput{
				ID:           *s.ID,
				CreationDate: s.CreatedAt.String(),
				Instance:     *instance.Name,
				State:        *s.State,
				Zone:         zone,
			}
		}

		return nil
	})
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr,
			"warning: errors during listing, results might be incomplete.\n%s\n", err) // nolint:golint
	}

	return output(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(computeInstanceSnapshotCmd, &instanceSnapshotListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
