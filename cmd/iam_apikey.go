package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// apiKeycmd represent the API key command
var apiKeyCmd = &cobra.Command{
	Use:   "apikey",
	Short: "API Keys management",
}

func getAPIKeyByKey(key string) (*egoscale.APIKey, error) {
	resp, err := cs.RequestWithContext(gContext, egoscale.GetAPIKey{
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

	resp, err := cs.RequestWithContext(gContext, egoscale.ListAPIKeys{})
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

	return &apiKeys[0], nil
}

func init() {
	iamCmd.AddCommand(apiKeyCmd)
}
