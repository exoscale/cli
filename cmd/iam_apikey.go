package cmd

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/egoscale"
)

// apiKeycmd represent the API key command
var apiKeyCmd = &cobra.Command{
	Use:   "apikey",
	Short: "API Keys management",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		fmt.Fprintln(os.Stderr,
			`**********************************************************************
The "exo iam apikey" commands are deprecated and will be removed in a future
version, please use "exo iam access-key" replacement commands.
**********************************************************************`)
		time.Sleep(3 * time.Second)
	},
	Hidden: true,
}

func getAPIKeyByKey(key string) (*egoscale.APIKey, error) {
	resp, err := globalstate.EgoscaleClient.RequestWithContext(gContext, egoscale.GetAPIKey{
		Key: key,
	})
	if err != nil {
		return nil, err
	}

	return resp.(*egoscale.APIKey), nil
}

func getAPIKeyByName(name string) (*egoscale.APIKey, error) {
	apiKeys := []egoscale.APIKey{}

	if apiKey, err := getAPIKeyByKey(name); err == nil {
		return apiKey, err
	}

	resp, err := globalstate.EgoscaleClient.RequestWithContext(gContext, egoscale.ListAPIKeys{})
	if err != nil {
		return nil, err
	}
	r := resp.(*egoscale.ListAPIKeysResponse)

	for _, i := range r.APIKeys {
		if i.Name == name {
			apiKeys = append(apiKeys, i)
		}
	}

	switch count := len(apiKeys); {
	case count == 0:
		return nil, fmt.Errorf("not found: %q", name)
	case count > 1:
		return nil, fmt.Errorf(`more than one element found: %d`, count)
	}

	sort.Strings(apiKeys[0].Operations)

	return &apiKeys[0], nil
}

func init() {
	iamCmd.AddCommand(apiKeyCmd)
}
