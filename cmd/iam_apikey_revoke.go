package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// apiKeyRevokeCmd represents an API key revocation command
var apiKeyRevokeCmd = &cobra.Command{
	Use:     "revoke <key>+",
	Short:   "Revoke API keys",
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
				fmt.Sprintf("Revoking API key %q", cmd.Key),
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
	apiKeyRevokeCmd.Flags().BoolP("force", "f", false, "Attempt to revoke API keys without prompting for confirmation")
	iamAPIKeyCmd.AddCommand(apiKeyRevokeCmd)
}
