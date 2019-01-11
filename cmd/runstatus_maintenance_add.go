package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/exoscale/egoscale"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var runstatusMaintenanceAddCmd = &cobra.Command{
	Use:   "add [page name]",
	Short: "Add a maintenance to a runstat.us page",
	RunE: func(cmd *cobra.Command, args []string) error {
		if gCurrentAccount.DefaultRunstatusPage == "" && len(args) == 0 {
			fmt.Fprintf(os.Stderr, `Error: No default runstat.us page is set:
  Please specify a page in parameter or add it to %q

  `, gConfigFilePath)
			return cmd.Usage()
		}

		if len(args) == 0 {
			args = append(args, gCurrentAccount.DefaultRunstatusPage)
		}

		runstatusPage, err := csRunstatus.GetRunstatusPage(gContext, egoscale.RunstatusPage{Subdomain: args[0]})
		if err != nil {
			return err
		}

		reader := bufio.NewReader(os.Stdin)

		title, err := cmd.Flags().GetString(runstatusFlagTitle)
		if err != nil {
			return err
		}
		if title == "" {
			title, err = readInput(reader, "Title of the maintenance", "none")
			if err != nil {
				return err
			}
		}

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
			return err
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

		var startDate time.Time
		sDate, err := cmd.Flags().GetString("start-date")
		if err != nil {
			return err
		}
		if sDate == "" {
			startDate, err = pickupDatePrompt(reader, time.Now(), "Choose start date:")
		} else {
			startDate, err = time.Parse("2006-01-02T15:04:05-0700", sDate)
		}
		if err != nil {
			return err
		}

		var endDate time.Time
		eDate, err := cmd.Flags().GetString("end-date")
		if err != nil {
			return err
		}
		if eDate == "" {
			endDate, err = pickupDatePrompt(reader, startDate, "Choose end date:")
		} else {
			endDate, err = time.Parse("2006-01-02T15:04:05-0700", eDate)
		}
		if err != nil {
			return err
		}

		fmt.Println()

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
		fmt.Fprintf(w, "Title:\t%s\n", title)                             // nolint: errcheck
		fmt.Fprintf(w, "Description:\t%s\n", description)                 // nolint: errcheck
		fmt.Fprintf(w, "Status:\t%s\n", status)                           // nolint: errcheck
		fmt.Fprintf(w, "Service(s):\t%s\n", strings.Join(services, ", ")) // nolint: errcheck
		fmt.Fprintf(w, "Start date:\t%s\n", startDate)                    // nolint: errcheck
		fmt.Fprintf(w, "End date:\t%s\n", endDate)                        // nolint: errcheck

		if err := w.Flush(); err != nil {
			return err
		}

		if !askQuestion("sure you want to add this maintenance") {
			return nil
		}

		maintenance, err := csRunstatus.CreateRunstatusMaintenance(gContext, egoscale.RunstatusMaintenance{
			Title:       title,
			Description: description,
			Status:      status,
			Services:    services,
			StartDate:   &startDate,
			EndDate:     &endDate,
			PageURL:     runstatusPage.URL,
		})
		if err != nil {
			return err
		}

		fmt.Printf("Maintenance %q successfully created\n", maintenance.Title)
		return nil
	},
}

func pickupDatePrompt(reader *bufio.Reader, now time.Time, text string) (time.Time, error) {
	var errTime time.Time
	fmt.Println(text)
	day, err := readNumberInput(reader, "Day", fmt.Sprintf("%d", now.Day()))
	if err != nil {
		return errTime, err
	}
	month, err := readNumberInput(reader, "Month", fmt.Sprintf("%d", now.Month()))
	if err != nil {
		return errTime, err
	}
	year, err := readNumberInput(reader, "Year", fmt.Sprintf("%d", now.Year()))
	if err != nil {
		return errTime, err
	}
	hour, err := readNumberInput(reader, "Hour", fmt.Sprintf("%d", now.Hour()))
	if err != nil {
		return errTime, err
	}
	minute, err := readNumberInput(reader, "Minute", fmt.Sprintf("%d", now.Minute()))
	if err != nil {
		return errTime, err
	}

	local, err := time.LoadLocation("")
	if err != nil {
		return errTime, err
	}

	date := time.Date(year, time.Month(month), day, hour, minute, 0, 0, local)

	return date, nil
}

func readNumberInput(reader *bufio.Reader, text, value string) (int, error) {
	for {
		val, err := readInput(reader, text, value)
		if err != nil {
			return 0, err
		}
		number, err := strconv.ParseInt(val, 10, 64)
		if err == nil {
			return int(number), nil
		}
		fmt.Printf("%q: not a number\n", val)
	}
}

func init() {
	runstatusMaintenanceCmd.AddCommand(runstatusMaintenanceAddCmd)
	//required
	runstatusMaintenanceAddCmd.Flags().StringSliceP(runstatusFlagServices, "", []string{}, "The list of services affected by the maintenance. e.g: <service1,service2,...>")
	runstatusMaintenanceAddCmd.Flags().StringP(runstatusFlagTitle, "t", "", "Title for the maintenance")
	runstatusMaintenanceAddCmd.Flags().StringP(runstatusFlagDescription, "d", "", "Description for the maintenance")
	runstatusMaintenanceAddCmd.Flags().StringP(runstatusFlagStatus, "", "", "<scheduled | in-progress | completed>")
	runstatusMaintenanceAddCmd.Flags().StringP("start-date", "s", "", "The planned start date for the maintenance, in UTC format e.g. 2016-05-31T21:11:32.378Z")
	runstatusMaintenanceAddCmd.Flags().StringP("end-date", "e", "", "End date, in UTC format e.g. 2016-05-31T21:11:32.378Z")
}
