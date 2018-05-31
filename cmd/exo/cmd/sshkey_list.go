package cmd

import (
	"log"
	"os"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/cmd/exo/table"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available ssh keyPair",
	RunE: func(cmd *cobra.Command, args []string) error {
		return listSSHKey()
	},
}

func listSSHKey() error {
	sshKey := &egoscale.SSHKeyPair{}
	sshKeys, err := cs.List(sshKey)
	if err != nil {
		log.Fatal(err)
	}

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Name", "Fingerprint"})

	for _, key := range sshKeys {
		k := key.(*egoscale.SSHKeyPair)
		table.Append([]string{k.Name, k.Fingerprint})
	}
	table.Render()

	return nil
}

func init() {
	sshkeyCmd.AddCommand(listCmd)
}
