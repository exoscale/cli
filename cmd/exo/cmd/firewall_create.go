package cmd

import (
	"log"
	"os"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/cmd/exo/table"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var firewallCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create security group",
}

func firewallCreateRun(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		firewallCreateCmd.Usage()
		return
	}
	desc, err := cmd.Flags().GetString("description")
	if err != nil {
		log.Fatal(err)
	}
	firewallCreate(args[0], desc)
}

func firewallCreate(name, desc string) {
	req := &egoscale.CreateSecurityGroup{Name: name}

	if desc != "" {
		req.Description = desc
	}

	resp, err := cs.Request(req)
	if err != nil {
		log.Fatal(err)
	}

	sgResp := resp.(*egoscale.SecurityGroup)

	table := table.NewTable(os.Stdout)
	if desc == "" {
		table.SetHeader([]string{"Name", "ID"})
		table.Append([]string{sgResp.Name, sgResp.ID})
	} else {
		table.SetHeader([]string{"Name", "Description", "ID"})
		table.Append([]string{sgResp.Name, sgResp.Description, sgResp.ID})
	}
	table.Render()
}

func init() {
	firewallCreateCmd.Run = firewallCreateRun
	firewallCreateCmd.Flags().StringP("description", "d", "", "Security group description")
	firewallCmd.AddCommand(firewallCreateCmd)
}
