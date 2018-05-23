package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/cmd/exo/table"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var privnetCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "A brief description of your command",
}

func isEmptyArgs(args ...string) bool {
	for _, arg := range args {
		if arg == "" {
			return true
		}
	}
	return false
}

func scanText(name string) string {
	fmt.Printf("%s: ", name)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func privnetCreateRun(cmd *cobra.Command, args []string) {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		log.Fatal(err)
	}
	desc, err := cmd.Flags().GetString("description")
	if err != nil {
		log.Fatal(err)
	}
	zone, err := cmd.Flags().GetString("zone")
	if err != nil {
		log.Fatal(err)
	}

	if name == "" && zone == "" {
		if name == "" {
			name = scanText("Name")
		}
		if desc == "" {
			desc = scanText("Description")
		}
		if zone == "" {
			zone = scanText("Zone")
		}
	}

	if isEmptyArgs(name, zone) {
		privnetCreateCmd.Usage()
		return
	}

	privnetCreate(name, desc, zone)
}

func privnetCreate(name, desc, zone string) {
	var err error
	zone, err = getZoneIDByName(cs, zone)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := cs.Request(&egoscale.ListNetworkOfferings{ZoneID: zone, Name: "PrivNet"})
	if err != nil {
		log.Fatal(err)
	}

	s := resp.(*egoscale.ListNetworkOfferingsResponse)

	offNetID := ""

	if len(s.NetworkOffering) > 0 {
		offNetID = s.NetworkOffering[0].ID
	}

	creatResp, err := cs.Request(&egoscale.CreateNetwork{DisplayText: desc, Name: name, NetworkOfferingID: offNetID, ZoneID: zone})
	if err != nil {
		log.Fatal(err)
	}

	newNet := creatResp.(*egoscale.CreateNetworkResponse).Network

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Name", "Description", "ID"})
	table.Append([]string{newNet.Name, newNet.DisplayText, newNet.ID})
	table.Render()
}

func init() {
	privnetCreateCmd.Run = privnetCreateRun
	privnetCreateCmd.Flags().StringP("name", "n", "", "Private network name")
	privnetCreateCmd.Flags().StringP("description", "d", "", "Private network description")
	privnetCreateCmd.Flags().StringP("zone", "z", "", "Assign private network to a zone")
	privnetCmd.AddCommand(privnetCreateCmd)
}
