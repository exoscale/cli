package cmd

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/egoscale"
)

type sshkeyUploadOutput struct {
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
}

func (o *sshkeyUploadOutput) Type() string { return "SSH Key" }
func (o *sshkeyUploadOutput) ToJSON()      { output.JSON(o) }
func (o *sshkeyUploadOutput) ToText()      { output.Text(o) }
func (o *sshkeyUploadOutput) ToTable()     { output.Table(o) }

func init() {
	sshkeyCmd.AddCommand(&cobra.Command{
		Use:   "upload NAME PUBLIC-KEY-FILE",
		Short: "Upload SSH key",
		Long: fmt.Sprintf(`This command uploads a locally existing SSH key.

Supported output template annotations: %s`,
			strings.Join(output.TemplateAnnotations(&sshkeyUploadOutput{}), ", ")),
		Aliases: gUploadAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return cmd.Usage()
			}

			return printOutput(uploadSSHKey(args[0], args[1]))
		},
	})
}

func uploadSSHKey(name, publicKeyPath string) (output.Outputter, error) {
	pbk, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}

	resp, err := globalstate.EgoscaleClient.RequestWithContext(gContext, &egoscale.RegisterSSHKeyPair{
		Name:      name,
		PublicKey: string(pbk),
	})
	if err != nil {
		return nil, err
	}

	keyPair := resp.(*egoscale.SSHKeyPair)

	if !globalstate.Quiet {
		return &sshkeyUploadOutput{
			Name:        keyPair.Name,
			Fingerprint: keyPair.Fingerprint,
		}, nil
	}

	return nil, nil
}
