package cmd

import (
	"bufio"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/exoscale/egoscale"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var runstatusMaintenanceUpdateCmd = &cobra.Command{
	Use:   "update [PAGE] MAINTENANCE-NAME",
	Short: "update a maintenance",
	Long: `Update a maintenance.
This is also used to close an maintenance`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		pageName := gCurrentAccount.DefaultRunstatusPage
		maintenanceName := args[0]

		if gCurrentAccount.DefaultRunstatusPage == "" && len(args) == 1 {
			fmt.Fprintf(os.Stderr, `Error: No default runstat.us page is set:
  Please specify a page in parameter or add it to %q

  `, gConfigFilePath)
			return cmd.Usage()
		}

		if len(args) > 1 {
			pageName = args[0]
			maintenanceName = args[1]
		}

		runstatusPage, err := csRunstatus.GetRunstatusPage(gContext, egoscale.RunstatusPage{Subdomain: pageName})
		if err != nil {
			return err
		}

		maintenance, err := getRunstatusMaintenanceByNameOrID(*runstatusPage, maintenanceName)
		if err != nil {
			return err
		}

		reader := bufio.NewReader(os.Stdin)

		description, err := cmd.Flags().GetString(runstatusFlagDescription)
		if err != nil {
			return err
		}
		if description == "" {
			description, err = readInput(reader, "Description for the maintenance", "none")
			if err != nil {
				return err
			}
		}

		status, err := cmd.Flags().GetString(runstatusFlagStatus)
		if err != nil {
			return nil
		}
		if status == "" {
			prompt := promptui.Select{
				Label: "Status",
				Items: []string{"scheduled", "in-progress", "completed"},
			}

			_, status, err = prompt.Run()

			if err != nil {
				return fmt.Errorf("prompt failed %v", err)
			}
		}

		fmt.Println()

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
		fmt.Fprintf(w, "Description:\t%s\n", description) // nolint: errcheck
		fmt.Fprintf(w, "Status:\t%s\n", status)           // nolint: errcheck

		if err := w.Flush(); err != nil {
			return err
		}

		if !askQuestion("Are you sure you want to update this maintenance?") {
			return nil
		}

		if err := csRunstatus.UpdateRunstatusMaintenance(gContext, *maintenance, egoscale.RunstatusEvent{
			Text:   description,
			Status: status,
		}); err != nil {
			return err
		}

		fmt.Printf("Maintenance %q successfully updated\n", maintenanceName)
		return nil
	},
}

func init() {
	runstatusMaintenanceCmd.AddCommand(runstatusMaintenanceUpdateCmd)
	runstatusMaintenanceUpdateCmd.Flags().StringP(runstatusFlagDescription, "d", "", "maintenance description")
	runstatusMaintenanceUpdateCmd.Flags().StringP(runstatusFlagStatus, "s", "", "maintenance status (scheduled|in-progress|completed)")
}
