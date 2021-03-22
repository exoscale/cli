package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var apiKeyRevokeCmd = &cobra.Command{
	Use:     "revoke KEY|NAME",
	Short:   "Revoke an API key",
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
				if !askQuestion(fmt.Sprintf("Are you sure you want to revoke API key %q?", arg)) {
					return nil
				}
			}

			apiKey, err := getAPIKeyByName(arg)
			if err != nil {
				return err
			}

			cmd := &egoscale.RevokeAPIKey{Key: apiKey.Key}
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
	apiKeyRevokeCmd.Flags().BoolP("force", "f", false, cmdFlagForceHelp)
	apiKeyCmd.AddCommand(apiKeyRevokeCmd)
}
