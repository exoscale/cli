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

var runstatusIncidentUpdateCmd = &cobra.Command{
	Use:   "update [PAGE] INCIDENT-NAME|ID",
	Short: "update an existing incident",
	Long: `Update an incident.
This is also used to close an incident,
passing a resolved status and flagging the services state as operational`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		pageName := gCurrentAccount.DefaultRunstatusPage
		incidentName := args[0]

		if gCurrentAccount.DefaultRunstatusPage == "" && len(args) == 1 {
			fmt.Fprintf(os.Stderr, `Error: No default runstat.us page is set:
  Please specify a page in parameter or add it to %q

  `, gConfigFilePath)
			return cmd.Usage()
		}

		if len(args) > 1 {
			pageName = args[0]
			incidentName = args[1]
		}

		runstatusPage, err := csRunstatus.GetRunstatusPage(gContext, egoscale.RunstatusPage{Subdomain: pageName})
		if err != nil {
			return err
		}

		incident, err := getRunstatusIncidentByNameOrID(*runstatusPage, incidentName)
		if err != nil {
			return err
		}

		reader := bufio.NewReader(os.Stdin)

		description, err := cmd.Flags().GetString(runstatusFlagDescription)
		if err != nil {
			return nil
		}
		if description == "" {
			description, err = readInput(reader, "Description of the incident", "none")
			if err != nil {
				return err
			}
		}

		state, err := cmd.Flags().GetString(runstatusFlagState)
		if err != nil {
			return nil
		}

		if state == "" {
			prompt := promptui.Select{
				Label: "Services State",
				Items: []string{"major_outage", "partial_outage", "degraded_performance", "operational"},
			}

			_, state, err = prompt.Run()

			if err != nil {
				return fmt.Errorf("prompt failed %v", err)
			}
		}

		status, err := cmd.Flags().GetString(runstatusFlagStatus)
		if err != nil {
			return nil
		}
		if status == "" {
			prompt := promptui.Select{
				Label: "Event Status",
				Items: []string{"investigating", "identified", "monitoring", "resolved"},
			}

			_, status, err = prompt.Run()

			if err != nil {
				return fmt.Errorf("prompt failed %v", err)
			}
		}

		fmt.Println()

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
		fmt.Fprintf(w, "Description:\t%s\n", description) // nolint: errcheck
		fmt.Fprintf(w, "State:\t%s\n", state)             // nolint: errcheck
		fmt.Fprintf(w, "Status:\t%s\n", status)           // nolint: errcheck

		if err := w.Flush(); err != nil {
			return err
		}

		if !askQuestion("Are you sure you want to update this incident?") {
			return nil
		}

		if err := csRunstatus.UpdateRunstatusIncident(gContext, *incident, egoscale.RunstatusEvent{
			Text:   description,
			Status: status,
			State:  state,
		}); err != nil {
			return err
		}
		fmt.Printf("Incident %q successfully updated\n", incidentName)

		return nil
	},
}

func init() {
	runstatusIncidentCmd.AddCommand(runstatusIncidentUpdateCmd)

	// required
	runstatusIncidentUpdateCmd.Flags().StringP(runstatusFlagDescription, "d", "", "incident description")
	runstatusIncidentUpdateCmd.Flags().StringP(runstatusFlagStatus, "s", "", "incident status (investigating|identified|monitoring|resolved)")
	runstatusIncidentUpdateCmd.Flags().StringP(runstatusFlagState, "t", "", "incident state (major_outage|partial_outage|degraded_performance|operational)")
}
