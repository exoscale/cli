package cmd

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/cmd/exo/table"
	"github.com/spf13/cobra"
)

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload [name] [path]",
	Short: "Upload ssh keyPair from given path",
}

func runUploadCmd(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		uploadCmd.Usage()
		return
	}
	uploadSSHKey(args[0], args[1])
}

func uploadSSHKey(name, publicKeyPath string) {
	pbk, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := cs.Request(&egoscale.RegisterSSHKeyPair{Name: name, PublicKey: string(pbk)})
	if err != nil {
		log.Fatal(err)
	}

	keyPair := resp.(*egoscale.RegisterSSHKeyPairResponse).KeyPair
	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Name", "Fingerprint"})
	table.Append([]string{keyPair.Name, keyPair.Fingerprint})
	table.Render()
}

func init() {
	uploadCmd.Run = runUploadCmd
	sshkeyCmd.AddCommand(uploadCmd)
}
