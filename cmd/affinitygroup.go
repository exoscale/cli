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

func getAffinityGroupByNameOrID(v string) (*egoscale.AffinityGroup, error) {
	aff := &egoscale.AffinityGroup{}

	id, err := egoscale.ParseUUID(v)
	if err == nil {
		aff.ID = id
	} else {
		aff.Name = v
	}

	resp, err := cs.GetWithContext(gContext, aff)
	switch err {
	case nil:
		return resp.(*egoscale.AffinityGroup), nil

	case egoscale.ErrNotFound:
		return nil, fmt.Errorf("unknown Affinity Group %q", v)

	case egoscale.ErrTooManyFound:
		return nil, fmt.Errorf("multiple Affinity Groups match %q", v)

	default:
		return nil, err
	}
}

func getAffinityGroupIDs(params []string) ([]egoscale.UUID, error) {
	ids := make([]egoscale.UUID, len(params))

	for i, aff := range params {
		s, err := getAffinityGroupByNameOrID(aff)
		if err != nil {
			return nil, err
		}

		ids[i] = *s.ID
	}

	return ids, nil
}

func init() {
	RootCmd.AddCommand(affinitygroupCmd)
}
