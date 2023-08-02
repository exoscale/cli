package cmd

import (
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/status"
	"github.com/exoscale/cli/table"
)

// REF: https://www.statuspal.io/api-docs#tag/Status/operation/getStatusPageStatus
const (
	statusPageSubdomain = "exoscalestatus"
)

func init() {

	// Global flags have no effect here, hide them
	statusCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		for _, flags := range []string{"quiet", "output-format", "output-template"} {

			err := cmd.Flags().MarkHidden(flags)
			if err != nil {
				fmt.Print(err)
			}
		}
		cmd.Parent().HelpFunc()(cmd, args)
	})
	RootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Exoscale status",
	RunE: func(cmd *cobra.Command, args []string) error {
		return statusShow()
	},
}

func statusShow() error {

	status, err := status.GetStatusPage(statusPageSubdomain)
	if err != nil {
		return err
	}

	// First show the global status per zone
	global, err := status.GetStatusByZone()
	if err != nil {
		return err
	}
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Exoscale Status"})

	buf := bytes.NewBuffer(nil)
	st := table.NewEmbeddedTable(buf)
	st.Table.AppendBulk(global)
	st.Render()
	t.Append([]string{"Services", buf.String()})
	buf.Reset()

	// Get the impacted services by zone (incidents and maintenances)
	incidents, maintenances, err := status.GetIncidents()
	if err != nil {
		return err
	}

	// Show incidents currently taking place
	if len(incidents) > 0 {
		it := table.NewEmbeddedTable(buf)
		it.Table.AppendBulk(incidents)
		it.Render()
	} else {
		buf = bytes.NewBuffer([]byte("n/a"))
	}
	t.Append([]string{"Incidents", buf.String()})
	buf.Reset()

	// Show maintenances currently taking place
	if len(maintenances) > 0 {
		mt := table.NewEmbeddedTable(buf)
		mt.Table.AppendBulk(maintenances)
		mt.Render()
	} else {
		buf = bytes.NewBuffer([]byte("n/a"))
	}
	t.Append([]string{"Maintenances", buf.String()})

	t.Render()
	return nil
}
