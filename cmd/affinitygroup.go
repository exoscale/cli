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

func getAffinityGroupIDByName(cs *egoscale.Client, name string) (string, error) {
	affs, err := cs.List(&egoscale.AffinityGroup{})
	if err != nil {
		return "", err
	}

	n := strings.ToLower(name)
	for _, a := range affs {
		aff := a.(*egoscale.AffinityGroup)
		if n == strings.ToLower(aff.Name) || n == aff.ID {
			return aff.ID, nil
		}
	}
	return "", fmt.Errorf("missing Affinity Group %q", name)
}

func init() {
	RootCmd.AddCommand(affinitygroupCmd)
}
