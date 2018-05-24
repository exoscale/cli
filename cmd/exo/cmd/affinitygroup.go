package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// affinitygroupCmd represents the affinitygroup command
var affinitygroupCmd = &cobra.Command{
	Use:   "affinitygroup",
	Short: "Affinity groups management",
}

func getAffinityGroupIDByName(cs *egoscale.Client, affinityGroup string) (string, error) {
	affReq := &egoscale.ListAffinityGroups{}

	affResp, err := cs.Request(affReq)
	if err != nil {
		return "", err
	}

	affs := affResp.(*egoscale.ListAffinityGroupsResponse)

	for _, aff := range affs.AffinityGroup {
		if strings.ToLower(affinityGroup) == strings.ToLower(aff.Name) {
			return aff.ID, nil
		}
		if affinityGroup == aff.ID {
			return aff.ID, nil
		}
	}
	return "", fmt.Errorf("Affinity group not found")
}

func init() {
	rootCmd.AddCommand(affinitygroupCmd)
}
