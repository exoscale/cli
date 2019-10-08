package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type apiKeyRevokeItemOutput egoscale.RevokeAPIKeyResponse

func (o *apiKeyRevokeItemOutput) toJSON()  { outputJSON(o) }
func (o *apiKeyRevokeItemOutput) toText()  { outputText(o) }
func (o *apiKeyRevokeItemOutput) toTable() { outputTable(o) }

// apiKeyCreateCmd represents the create command
var apiKeyRevokeCmd = &cobra.Command{
	Use:     "revoke <APIKey key>+",
	Short:   "Revoke an APIKeys",
	Aliases: gRevokeAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		tasks := make([]task, 0, len(args))
		for _, arg := range args {
			if !force {
				if !askQuestion(fmt.Sprintf("sure you want to revoke %q", arg)) {
					return nil
				}
			}

			cmd := &egoscale.RevokeAPIKey{Key: arg}
			tasks = append(tasks, task{
				cmd,
				fmt.Sprintf("Revoke APIKey %q", cmd.Key),
			})
		}

		resps := asyncTasks(tasks)
		errs := filterErrors(resps)
		if len(errs) > 0 {
			return errs[0]
		}
		return nil
	},
}

func init() {
	apiKeyRevokeCmd.Flags().BoolP("force", "f", false, "Attempt to revoke APIKey without prompting for confirmation")
	iamAPIKeyCmd.AddCommand(apiKeyRevokeCmd)
}
