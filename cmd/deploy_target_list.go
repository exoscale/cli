package cmd

import (
	"fmt"
	"os"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type deployTargetListItemOutput struct {
	Zone string `json:"zone"`
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type deployTargetListOutput []deployTargetListItemOutput

func (o *deployTargetListOutput) toJSON()  { outputJSON(o) }
func (o *deployTargetListOutput) toText()  { outputText(o) }
func (o *deployTargetListOutput) toTable() { outputTable(o) }

var deployTargetListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Deploy Targets",
	Long: fmt.Sprintf(`This command lists existing Deploy Targets.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&vmListOutput{}), ", ")),
	RunE: func(cmd *cobra.Command, args []string) error {
		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}
		zone = strings.ToLower(zone)

		return output(listDeployTargets(zone), nil)
	},
}

func listDeployTargets(zone string) outputter {
	var listZones []string

	if zone != "" {
		listZones = []string{zone}
	} else {
		listZones = allZones
	}

	out := make(deployTargetListOutput, 0)
	res := make(chan deployTargetListItemOutput)
	defer close(res)

	go func() {
		for dt := range res {
			out = append(out, dt)
		}
	}()
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))
	err := forEachZone(listZones, func(zone string) error {
		list, err := cs.ListDeployTargets(ctx, zone)
		if err != nil {
			return fmt.Errorf("unable to list Deploy Targets in zone %s: %v", zone, err)
		}

		for _, dt := range list {
			res <- deployTargetListItemOutput{
				Zone: zone,
				ID:   dt.ID,
				Name: dt.Name,
				Type: dt.Type,
			}
		}

		return nil
	})
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr,
			"warning: errors during listing, results might be incomplete.\n%s\n", err) // nolint:golint
	}

	return &out
}

func init() {
	deployTargetListCmd.Flags().StringP("zone", "z", "", "zone to filter results to")
	deployTargetCmd.AddCommand(deployTargetListCmd)
}
