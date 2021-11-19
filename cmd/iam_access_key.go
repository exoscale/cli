package cmd

import (
	"fmt"
	"strings"

	egoscale "github.com/exoscale/egoscale/v2"
	"github.com/spf13/cobra"
)

var iamAccessKeyCmd = &cobra.Command{
	Use:     "access-key",
	Aliases: []string{"key"},
	Short:   "IAM access keys management",
}

// parseIAMAccessKeyResource parses a string-encoded IAM access key resource formatted such as
// DOMAIN/TYPE:NAME and deserializes it into an egoscale.IAMAccessKeyResource struct.
func parseIAMAccessKeyResource(v string) (*egoscale.IAMAccessKeyResource, error) {
	var iamAccessKeyResource egoscale.IAMAccessKeyResource

	parts := strings.SplitN(v, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid format")
	}
	iamAccessKeyResource.ResourceName = parts[1]

	parts = strings.SplitN(parts[0], "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid format")
	}
	iamAccessKeyResource.Domain = parts[0]
	iamAccessKeyResource.ResourceType = parts[1]

	if iamAccessKeyResource.Domain == "" ||
		iamAccessKeyResource.ResourceType == "" ||
		iamAccessKeyResource.ResourceName == "" {
		return nil, fmt.Errorf("invalid format")
	}

	return &iamAccessKeyResource, nil
}

func init() {
	iamCmd.AddCommand(iamAccessKeyCmd)
}
