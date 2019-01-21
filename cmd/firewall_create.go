package cmd

import (
	"fmt"
	"os"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var firewallCreateCmd = &cobra.Command{
	Use:   "create <name>+",
	Short: "Create security group",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}
		desc, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}

		syncTasks := []syncTask{}
		for _, arg := range args {
			syncTasks = append(syncTasks, syncTask{
				egoscale.CreateSecurityGroup{Name: arg, Description: desc},
				fmt.Sprintf("Create security group %q", arg),
			})
		}

		taskResponses := syncTasksAsync(syncTasks)
		errors := filterErrors(taskResponses)
		if len(errors) > 0 {
			return errors[0]
		}

		description := (desc != "")

		table := table.NewTable(os.Stdout)
		if !description {
			table.SetHeader([]string{"Name", "ID"})
		} else {
			table.SetHeader([]string{"Name", "Description", "ID"})
		}

		for _, resp := range taskResponses {
			r := resp.resp.(*egoscale.SecurityGroup)

			if description {
				table.Append([]string{r.Name, r.ID.String()})
				continue
			}
			table.Append([]string{r.Name, r.Description, r.ID.String()})
		}
		table.Render()

		return nil
	},
}

func init() {
	firewallCreateCmd.Flags().StringP("description", "d", "", "Security group description")
	firewallCmd.AddCommand(firewallCreateCmd)
}
