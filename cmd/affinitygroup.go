package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// affinitygroupCmd represents the affinitygroup command
var affinitygroupCmd = &cobra.Command{
	Use:   "affinitygroup",
	Short: "Affinity groups management",
}

func getAffinityGroupByName(name string) (*egoscale.AffinityGroup, error) {
	aff := &egoscale.AffinityGroup{}

	id, err := egoscale.ParseUUID(name)
	if err == nil {
		aff.ID = id
	} else {
		aff.Name = name
	}

	resp, err := cs.GetWithContext(gContext, aff)
	if err != nil {
		if e, ok := err.(*egoscale.ErrorResponse); ok && e.ErrorCode == egoscale.ParamError {
			return nil, fmt.Errorf("missing Affinity Group %q", name)
		}

		return nil, err
	}

	return resp.(*egoscale.AffinityGroup), nil
}

func getAffinityGroups(vm *egoscale.VirtualMachine) []string {
	ags := make([]string, len(vm.AffinityGroup))
	for i, agN := range vm.AffinityGroup {
		ags[i] = agN.Name
	}
	return ags
}

func init() {
	RootCmd.AddCommand(affinitygroupCmd)
}
