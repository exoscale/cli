package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type sshkeyCreateOutput struct {
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
	PrivateKey  string `json:"private_key"`
}

func (o *sshkeyCreateOutput) Type() string { return "SSH Key" }
func (o *sshkeyCreateOutput) ToJSON()      { output.JSON(o) }
func (o *sshkeyCreateOutput) ToText()      { output.Text(o) }
func (o *sshkeyCreateOutput) ToTable()     { output.Table(o) }

func init() {
	sshkeyCmd.AddCommand(&cobra.Command{
		Use:   "create NAME",
		Short: "Create SSH key",
		Long: fmt.Sprintf(`This command creates an SSH key.

Supported output template annotations: %s`,
			strings.Join(output.OutputterTemplateAnnotations(&sshkeyCreateOutput{}), ", ")),
		Aliases: gCreateAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}

			sshKey, err := createSSHKey(args[0])
			if err != nil {
				return err
			}

			if !globalstate.Quiet {
				return printOutput(&sshkeyCreateOutput{
					Name:        sshKey.Name,
					Fingerprint: sshKey.Fingerprint,
					PrivateKey:  sshKey.PrivateKey,
				}, err)
			}

			return nil
		},
	})
}

func createSSHKey(name string) (*egoscale.SSHKeyPair, error) {
	resp, err := globalstate.GlobalEgoscaleClient.RequestWithContext(gContext, &egoscale.CreateSSHKeyPair{
		Name: name,
	})
	if err != nil {
		return nil, err
	}

	sshKeyPair, ok := resp.(*egoscale.SSHKeyPair)
	if !ok {
		return nil, fmt.Errorf("wrong type expected %q, got %T", "egoscale.CreateSSHKeyPairResponse", resp)
	}

	return sshKeyPair, nil
}
