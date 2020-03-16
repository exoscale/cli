package cmd

import (
	"fmt"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type snapshotShowOutput struct {
	ID       string `json:"id"`
	Date     string `json:"date"`
	Instance string `json:"instance"`
	State    string `json:"state"`
	Size     string `json:"size"`
}

func (o *snapshotShowOutput) Type() string { return "Snapshot" }
func (o *snapshotShowOutput) toJSON()      { outputJSON(o) }
func (o *snapshotShowOutput) toText()      { outputText(o) }
func (o *snapshotShowOutput) toTable()     { outputTable(o) }

func init() {
	snapshotCmd.AddCommand(&cobra.Command{
		Use:   "show <ID>",
		Short: "Show a snapshot details",
		Long: fmt.Sprintf(`This command shows a snapshot details.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&snapshotShowOutput{}), ", ")),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}

			snapshot, err := getSnapshotWithNameOrID(args[0])
			if err != nil {
				return err
			}

			return output(showSnapshot(snapshot))
		},
	})
}

func showSnapshot(snapshot *egoscale.Snapshot) (outputter, error) {
	return &snapshotShowOutput{
		ID:       snapshot.ID.String(),
		Instance: snapshotVMName(*snapshot),
		Date:     snapshot.Created,
		State:    snapshot.State,
		Size:     humanize.IBytes(uint64(snapshot.Size)),
	}, nil
}
