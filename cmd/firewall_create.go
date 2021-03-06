package cmd

import (
	"fmt"
	"os"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var firewallCreateCmd = &cobra.Command{
	Use:   "create NAME",
	Short: "Create a Security Group",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}
		desc, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}

		syncTasks := []task{}
		for _, arg := range args {
			syncTasks = append(syncTasks, task{
				egoscale.CreateSecurityGroup{Name: arg, Description: desc},
				fmt.Sprintf("Creating Security Group %q", arg),
			})
		}

		taskResponses := asyncTasks(syncTasks)
		errors := filterErrors(taskResponses)
		if len(errors) > 0 {
			return errors[0]
		}

		if !gQuiet {
			table := table.NewTable(os.Stdout)
			table.SetHeader([]string{"ID", "Name", "Description"})
			for _, resp := range taskResponses {
				r := resp.resp.(*egoscale.SecurityGroup)
				table.Append([]string{r.ID.String(), r.Name, r.Description})
			}
			table.Render()
		}

		return nil
	},
}

func init() {
	firewallCreateCmd.Flags().StringP("description", "d", "", "Security Group description")
	firewallCmd.AddCommand(firewallCreateCmd)
}
