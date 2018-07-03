package cmd

import (
	"fmt"
	"os"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/cmd/exo/table"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var sshCreateCmd = &cobra.Command{
	Use:     "create <name>",
	Short:   "Create ssh keyPair",
	Aliases: gCreateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}
		keyPair, err := createSSHKey(args[0])
		if err != nil {
			return err
		}
		displayResult(keyPair)
		return nil
	},
}

func createSSHKey(name string) (*egoscale.SSHKeyPair, error) {
	resp, err := cs.Request(&egoscale.CreateSSHKeyPair{Name: name})
	if err != nil {
		return nil, err
	}

	sshKEYPair, ok := resp.(*egoscale.SSHKeyPair)
	if !ok {
		return nil, fmt.Errorf("Expected %q, got %t", "egoscale.CreateSSHKeyPairResponse", resp)
	}
	return sshKEYPair, nil
}

func displayResult(sshKEYPair *egoscale.SSHKeyPair) {
	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Name", "Fingerprint"})
	table.Append([]string{sshKEYPair.Name, sshKEYPair.Fingerprint})
	table.Render()

	println(sshKEYPair.PrivateKey)
}

func init() {
	sshkeyCmd.AddCommand(sshCreateCmd)
}
