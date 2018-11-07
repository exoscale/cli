package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

const (
	statusURL         = "https://exoscalestatus.com"
	jsonStatusURL     = statusURL + "/api.json"
	statusContentPage = "application/json"
	twitterURL        = "https://twitter.com/exoscalestatus"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Exoscale status",
	RunE: func(cmd *cobra.Command, args []string) error {
		status, err := fetchRunStatus(jsonStatusURL)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.FilterHTML)

		fmt.Printf("Exoscale Status\n\t%s\n\n", statusURL)

		for k, service := range status.Status {
			fmt.Fprintf(w, "%s\t%s\n", k, service.State) // nolint: errcheck
		}
		fmt.Fprintln(w) // nolint: errcheck

		if len(status.Incidents) > 0 {
			suffix := ""
			if len(status.Incidents) > 1 {
				suffix = "s"
			}
			fmt.Printf("%d ongoing Incident%s (last: %s)\n",
				len(status.Incidents),
				suffix,
				status.Incidents[0].Title);
			fmt.Printf("Full updates available at %s\n", twitterURL)
		}

		w.Flush()
		return nil
	},
}

func fetchRunStatus(url string) (*RunStatus, error) {
	// XXX need gContext
	r, err := http.Get(jsonStatusURL)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	contentType := r.Header.Get("content-type")
	if contentType != statusContentPage {
		return nil, fmt.Errorf("status page content type expected %q, but got %q", statusContentPage, contentType)
	}

	response := &RunStatus{}
	if err := json.NewDecoder(r.Body).Decode(response); err != nil {
		return nil, err
	}

	return response, nil
}

// ServiceStatus represents the state of a service
type ServiceStatus struct {
	State string `json:"state"`
}

// RunStatus represents a runstatus struct
type RunStatus struct {
	URL       string `json:"url"`
	Incidents []struct {
		Message string    `json:"message"`
		Status  string    `json:"status"`
		Updated time.Time `json:"updated"`
		Title   string    `json:"title"`
		Created time.Time `json:"created"`
	} `json:"incidents"`
	UpcomingMaintenances []struct {
		Description string    `json:"description"`
		Title       string    `json:"title"`
		Date        time.Time `json:"date"`
	} `json:"upcoming_maintenances"`
	Status map[string]ServiceStatus `json:"status"`
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
