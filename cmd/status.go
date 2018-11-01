package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/exoscale/cli/table"

	"github.com/spf13/cobra"
)

const (
	statusURL         = "https://exoscalestatus.com/api.json"
	statusContentPage = "application/json"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Exoscale status",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := http.Get(statusURL)
		if err != nil {
			return err
		}
		defer r.Body.Close()

		contentType := r.Header.Get("content-type")
		if contentType != statusContentPage {
			return fmt.Errorf("status page content type expected %q, but got %q", statusContentPage, contentType)
		}

		response := &ExoscaleStatus{}
		if err := json.NewDecoder(r.Body).Decode(response); err != nil {
			return err
		}

		tableWriter := table.NewTable(os.Stdout)

		tableWriter.SetHeader([]string{"Detailed Status", "State"})

		tableWriter.Append([]string{"Compute", response.Status.Compute.State})
		tableWriter.Append([]string{"Compute API", response.Status.ComputeAPI.State})
		tableWriter.Append([]string{"DNS", response.Status.DNS.State})
		tableWriter.Append([]string{"Object Storage", response.Status.ObjectStorage.State})

		tableWriter.Render()

		tableWriter = table.NewTable(os.Stdout)

		tableWriter.SetHeader([]string{"Current Events", "Upcoming Maintenances"})

		tableWriter.Render()

		return nil
	},
}

//ExoscaleStatus represent exoscale statsus
type ExoscaleStatus struct {
	URL                  string        `json:"url"`
	Incidents            []interface{} `json:"incidents"`
	UpcomingMaintenances []interface{} `json:"upcoming_maintenances"`
	Status               struct {
		Compute struct {
			State string `json:"state"`
		} `json:"Compute"`
		ComputeAPI struct {
			State string `json:"state"`
		} `json:"Compute API"`
		DNS struct {
			State string `json:"state"`
		} `json:"DNS"`
		ObjectStorage struct {
			State string `json:"state"`
		} `json:"Object Storage"`
	} `json:"status"`
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
