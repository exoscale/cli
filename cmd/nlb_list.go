package cmd

import (
	"fmt"
	"os"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type nlbListItemOutput struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Zone      string `json:"zone"`
	IPAddress string `json:"ip_address"`
}

type nlbListOutput []nlbListItemOutput

func (o *nlbListOutput) toJSON()  { outputJSON(o) }
func (o *nlbListOutput) toText()  { outputText(o) }
func (o *nlbListOutput) toTable() { outputTable(o) }

var nlbListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Network Load Balancers",
	Long: fmt.Sprintf(`This command lists Network Load Balancers.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&nlbListItemOutput{}), ", ")),
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}
		zone = strings.ToLower(zone)

		return output(listNLBs(zone), nil)
	},
}

func listNLBs(zone string) outputter {
	var nlbZones []string

	if zone != "" {
		nlbZones = []string{zone}
	} else {
		nlbZones = allZones
	}

	out := make(nlbListOutput, 0)
	res := make(chan nlbListItemOutput)
	defer close(res)

	go func() {
		for nlb := range res {
			out = append(out, nlb)
		}
	}()
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))
	err := forEachZone(nlbZones, func(zone string) error {
		list, err := cs.ListNetworkLoadBalancers(ctx, zone)
		if err != nil {
			return fmt.Errorf("unable to list Network Load Balancers in zone %s: %v", zone, err)
		}

		for _, nlb := range list {
			res <- nlbListItemOutput{
				ID:        nlb.ID,
				Name:      nlb.Name,
				Zone:      zone,
				IPAddress: nlb.IPAddress.String(),
			}
		}

		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"warning: errors during listing, results might be incomplete.\n%s\n", err) // nolint:golint
	}

	return &out
}

func init() {
	nlbListCmd.Flags().StringP("zone", "z", "", "Zone to filter results to")
	nlbCmd.AddCommand(nlbListCmd)
}
