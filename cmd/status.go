package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/exoscale/cli/table"
	"github.com/spf13/cobra"
)

const (
	statusURL         = "https://exoscalestatus.com"
	jsonStatusURL     = statusURL + "/api.json"
	statusContentPage = "application/json"
	twitterURL        = "https://twitter.com/exoscalestatus"
)

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
	RootCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Exoscale status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return statusShow()
		},
	})
}

func statusShow() error {
	status, err := fetchRunStatus(jsonStatusURL)
	if err != nil {
		return err
	}

	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Exoscale Status"})

	buf := bytes.NewBuffer(nil)
	st := table.NewEmbeddedTable(buf)
	for service, status := range status.Status {
		st.Append([]string{service, status.State})
	}
	st.Render()

	t.Append([]string{"Services", buf.String()})

	buf = bytes.NewBuffer([]byte("n/a"))
	if len(status.Incidents) > 0 {
		buf.Reset()
		it := table.NewEmbeddedTable(buf)
		for _, i := range status.Incidents {
			it.Append([]string{i.Title, i.Status, fmt.Sprint(i.Created), fmt.Sprint(i.Updated)})
		}
		it.Render()
	}
	t.Append([]string{"Incidents", buf.String()})

	buf = bytes.NewBuffer([]byte("n/a"))
	if len(status.UpcomingMaintenances) > 0 {
		buf.Reset()
		mt := table.NewEmbeddedTable(buf)
		for _, m := range status.UpcomingMaintenances {
			mt.Append([]string{m.Title, m.Description, fmt.Sprint(m.Date)})
		}
		mt.Render()
	}
	t.Append([]string{"Maintenances", buf.String()})

	t.Render()

	fmt.Println("Updates available at", twitterURL)

	return nil
}

func fetchRunStatus(url string) (*RunStatus, error) {
	// XXX need gContext
	r, err := http.Get(url)
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
