package cmd

import (
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var affinitygroupDeleteCmd = &cobra.Command{
	Use:     "delete <name | id>",
	Short:   "Delete affinity group",
	Aliases: gDeleteAlias,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Usage()
			return
		}

		deleteAffinityGroup(args[0])
	},
}

func deleteAffinityGroup(name string) error {
	id, err := getAffinityGroupIDByName(cs, name)
	if err != nil {
		return err
	}

	_, err = cs.Request(&egoscale.DeleteAffinityGroup{ID: id})
	if err != nil {
		return err
	}

	println(id)

	return nil
}

func init() {
	affinitygroupCmd.AddCommand(affinitygroupDeleteCmd)
}
