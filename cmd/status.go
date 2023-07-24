package cmd

import (
	"bytes"
	"os"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/status"
	"github.com/exoscale/cli/table"
)

// https://www.statuspal.io/api-docs#tag/Status/operation/getStatusPageStatus
const (
	statusPageSubdomain = "exoscalestatus"
)

func init() {

	// Global flags have no effect here, hide them
	statusCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		cmd.Flags().MarkHidden("quiet")
		cmd.Flags().MarkHidden("output-format")
		cmd.Flags().MarkHidden("output-template")
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

	// Get the impacted services by zone
	incidents, maintenances, err := status.GetIncidents()
	if len(incidents) == 0 {
		buf = bytes.NewBuffer([]byte("n/a"))
	}

	it := table.NewEmbeddedTable(buf)
	it.Table.AppendBulk(incidents)
	it.Render()
	t.Append([]string{"Incidents", buf.String()})
	buf.Reset()

	if len(maintenances) == 0 {
		buf = bytes.NewBuffer([]byte("n/a"))
	}

	mt := table.NewEmbeddedTable(buf)
	mt.Table.AppendBulk(maintenances)
	mt.Render()
	t.Append([]string{"Maintenances", buf.String()})

	t.Render()

	return nil
}
