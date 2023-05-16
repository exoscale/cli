package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/egoscale"
)

const (
	runstatusFlagTitle       = "title"
	runstatusFlagDescription = "description"
	runstatusFlagStatus      = "status"
	runstatusFlagState       = "state"
	runstatusFlagServices    = "services"
)

var runstatusIncidentAddCmd = &cobra.Command{
	Use:   "add PAGE",
	Short: "Add an incident to a runstat.us page",
	RunE: func(cmd *cobra.Command, args []string) error {
		if account.CurrentAccount.DefaultRunstatusPage == "" && len(args) == 0 {
			fmt.Fprintf(os.Stderr, `Error: No default runstat.us page is set:
  Please specify a page in parameter or add it to %q

  `, gConfigFilePath)
			return cmd.Usage()
		}

		if len(args) == 0 {
			args = append(args, account.CurrentAccount.DefaultRunstatusPage)
		}

		runstatusPage, err := csRunstatus.GetRunstatusPage(gContext, egoscale.RunstatusPage{Subdomain: args[0]})
		if err != nil {
			return err
		}

		reader := bufio.NewReader(os.Stdin)

		title, err := cmd.Flags().GetString(runstatusFlagTitle)
		if err != nil {
			return nil
		}
		if title == "" {
			title, err = readInput(reader, "Title of the incident", "none")
			if err != nil {
				return err
			}
		}

		description, err := cmd.Flags().GetString(runstatusFlagDescription)
		if err != nil {
			return nil
		}
		if description == "" {
			description, err = readInput(reader, "Description for the initial creation event", "none")
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
				Items: []string{"investigating", "identified", "monitoring"},
			}

			_, status, err = prompt.Run()

			if err != nil {
				return fmt.Errorf("prompt failed %v", err)
			}
		}

		services, err := cmd.Flags().GetStringSlice(runstatusFlagServices)
		if err != nil {
			return err
		}
		if len(services) == 0 {
			services, err = promptGetService(*runstatusPage)
			if err != nil {
				return err
			}
		}

		fmt.Println()

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
		fmt.Fprintf(w, "Title:\t%s\n", title)                             // nolint: errcheck
		fmt.Fprintf(w, "Description:\t%s\n", description)                 // nolint: errcheck
		fmt.Fprintf(w, "State:\t%s\n", state)                             // nolint: errcheck
		fmt.Fprintf(w, "Status:\t%s\n", status)                           // nolint: errcheck
		fmt.Fprintf(w, "Service(s):\t%s\n", strings.Join(services, ", ")) // nolint: errcheck

		if err := w.Flush(); err != nil {
			return err
		}

		if !askQuestion("Are you sure you want to add this incident?") {
			return nil
		}

		incident, err := csRunstatus.CreateRunstatusIncident(gContext, egoscale.RunstatusIncident{
			PageURL:    runstatusPage.URL,
			Services:   services,
			State:      state,
			Status:     status,
			StatusText: description,
			Title:      title,
		})
		if err != nil {
			return err
		}

		fmt.Printf("Incident %q successfully created\n", incident.Title)
		return nil
	},
}

func promptGetService(page egoscale.RunstatusPage) ([]string, error) {
	servs, err := csRunstatus.ListRunstatusServices(gContext, page)
	if err != nil {
		return nil, err
	}
	var services []string
	tmp := make([]string, 0, len(servs))
	for _, serv := range servs {
		tmp = append(tmp, serv.Name)
	}

	if len(servs) > 0 {
		tmp = append([]string{"[DONE] (Validate choice)"}, tmp...)
		var index int
		for {
			prompt := promptui.Select{
				Label: "Select some service",
				Items: tmp,
			}

			var result string
			index, result, err = prompt.Run()

			if index != 0 {
				services = append(services, result)
				tmp = append(tmp[:index], tmp[index+1:]...)
				continue
			}
			break
		}

		if err != nil {
			return nil, fmt.Errorf("prompt failed %v", err)
		}

		return services, nil
	}

	return []string{}, nil
}

func init() {
	runstatusIncidentCmd.AddCommand(runstatusIncidentAddCmd)

	// required
	runstatusIncidentAddCmd.Flags().StringSliceP(runstatusFlagServices, "", []string{}, "List of strings with the services impacted. e.g: service1,service2,...")
	runstatusIncidentAddCmd.Flags().StringP(runstatusFlagTitle, "t", "", "incident title")
	runstatusIncidentAddCmd.Flags().StringP(runstatusFlagDescription, "d", "", "incident initial event description")
	runstatusIncidentAddCmd.Flags().StringP(runstatusFlagStatus, "", "", "incident status (investigating|identified|monitoring)")
	runstatusIncidentAddCmd.Flags().StringP(runstatusFlagState, "s", "", "incident state (major_outage|partial_outage|degraded_performance|operational)")
}
