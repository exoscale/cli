package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// runstatusServiceCmd represents the service command
var runstatusServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Service management",
}

func getServiceByName(page egoscale.RunstatusPage, name string) (*egoscale.RunstatusService, error) {
	services, err := csRunstatus.ListRunstatusServices(gContext, page)
	if err != nil {
		return nil, err
	}

	var result []egoscale.RunstatusService

	for i, service := range services {
		if service.Name == name {
			result = append(result, services[i])
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("%q: no service found", name)
	}
	if len(result) > 1 {
		return nil, fmt.Errorf("%q: more than one elements found", name)
	}
	return &result[0], nil
}

func init() {
	runstatusCmd.AddCommand(runstatusServiceCmd)
}
