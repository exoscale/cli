package cmd

import (
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
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) < 1 {
			return listSecurityGroups()
		}
		return firewallDetails(args[0])
	},
}

func listSecurityGroups() error {
	sgs, err := cs.List(&egoscale.SecurityGroup{})
	if err != nil {
		return err
	}

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Name", "Description", "ID"})

	for _, key := range sgs {
		k := key.(*egoscale.SecurityGroup)
		table.Append([]string{k.Name, k.Description, k.ID})
	}
	table.Render()
	return nil
}

func firewallDetails(name string) error {
	securGrp, err := getSecuGrpWithNameOrID(cs, name)
	if err != nil {
		return err
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
	return nil
}

func init() {
	firewallCmd.AddCommand(firewallListCmd)
}
