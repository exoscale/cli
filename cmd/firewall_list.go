package cmd

import (
	"log"
	"os"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/cmd/exo/table"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var firewallListCmd = &cobra.Command{
	Use:     "list [security group name | id]",
	Short:   "List security groups or show a security group rules details",
	Aliases: gListAlias,
}

func firewallListRun(cmd *cobra.Command, args []string) {

	if len(args) < 1 {
		listSecurityGroups()
		return
	}
	firewallDetails(args[0])
}

func listSecurityGroups() {
	sgs, err := cs.List(&egoscale.SecurityGroup{})
	if err != nil {
		log.Fatal(err)
	}

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Name", "Description", "ID"})

	for _, key := range sgs {
		k := key.(*egoscale.SecurityGroup)
		table.Append([]string{k.Name, k.Description, k.ID})
	}
	table.Render()

}

func firewallDetails(name string) {
	securGrp, err := getSecuGrpWithNameOrID(cs, name)
	if err != nil {
		log.Fatal(err)
	}

	table := table.NewTable(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader([]string{"Type", "Source", "Protocol", "Port", "Description", "ID"})

	heading := "INGRESS"
	for _, in := range securGrp.IngressRule {
		table.Append(formatRules(heading, &in))
		heading = ""
	}

	heading = "EGRESS"
	for _, out := range securGrp.EgressRule {
		table.Append(formatRules(heading, (*egoscale.IngressRule)(&out)))
		heading = ""
	}

	table.Render()

}

func init() {
	firewallListCmd.Run = firewallListRun
	firewallCmd.AddCommand(firewallListCmd)
}
