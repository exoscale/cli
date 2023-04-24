package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type sshkeyListItemOutput struct {
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
}

type sshkeyListOutput []sshkeyListItemOutput

func (o *sshkeyListOutput) toJSON()  { output.JSON(o) }
func (o *sshkeyListOutput) toText()  { output.Text(o) }
func (o *sshkeyListOutput) toTable() { output.Table(o) }

func init() {
	sshkeyCmd.AddCommand(&cobra.Command{
		Use:   "list [filter ...]",
		Short: "List SSH Keys",
		Long: fmt.Sprintf(`This command lists existing SSH Keys.
Optional patterns can be provided to filter results by name or fingerprint.

Supported output template annotations: %s`,
			strings.Join(output.output.OutputterTemplateAnnotations(&sshkeyListOutput{}), ", ")),
		Aliases: gListAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			return printOutput(listSSHKey(args))
		},
	})
}

func listSSHKey(filters []string) (output.Outputter, error) {
	sshKeys, err := getSSHKeys(cs)
	if err != nil {
		return nil, err
	}

	out := sshkeyListOutput{}

	for _, k := range sshKeys {
		keep := true
		if len(filters) > 0 {
			keep = false
			s := strings.ToLower(fmt.Sprintf("%s#%s", k.Name, k.Fingerprint))

			for _, filter := range filters {
				substr := strings.ToLower(filter)
				if strings.Contains(s, substr) {
					keep = true
					break
				}
			}
		}

		if !keep {
			continue
		}

		out = append(out, sshkeyListItemOutput{
			Name:        k.Name,
			Fingerprint: k.Fingerprint,
		})
	}

	return &out, nil
}

func getSSHKeys(cs *egoscale.Client) ([]egoscale.SSHKeyPair, error) {
	sshKeys, err := cs.ListWithContext(gContext, &egoscale.SSHKeyPair{})
	if err != nil {
		return nil, err
	}

	res := make([]egoscale.SSHKeyPair, len(sshKeys))

	for i, key := range sshKeys {
		k := key.(*egoscale.SSHKeyPair)
		res[i] = *k
	}

	return res, nil
}
