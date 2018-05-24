package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/cmd/exo/table"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var sshCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create ssh keyPair",
}

func runListCmd(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		sshCreateCmd.Usage()
		return
	}
	keyPair, err := createSSHKey(args[0])
	if err != nil {
		log.Fatal(err)
	}
	displayResult(keyPair)
}

func createSSHKey(name string) (*egoscale.SSHKeyPair, error) {
	resp, err := cs.Request(&egoscale.CreateSSHKeyPair{Name: name})
	if err != nil {
		return nil, err
	}

	sshKEYPair, ok := resp.(*egoscale.CreateSSHKeyPairResponse)
	if !ok {
		return nil, fmt.Errorf("Expected %q, got %t", "egoscale.CreateSSHKeyPairResponse", resp)
	}
	return &sshKEYPair.KeyPair, nil
}

func displayResult(sshKEYPair *egoscale.SSHKeyPair) {
	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Name", "Fingerprint"})
	table.Append([]string{sshKEYPair.Name, sshKEYPair.Fingerprint})
	table.Render()

	fmt.Println(sshKEYPair.PrivateKey)
}

func init() {
	sshCreateCmd.Run = runListCmd
	sshkeyCmd.AddCommand(sshCreateCmd)
}
