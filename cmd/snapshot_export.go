package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type snapshotExportOutput struct {
	URL      string `json:"url"`
	Checksum string `json:"checksum"`
}

func (o *snapshotExportOutput) toJSON()  { outputJSON(o) }
func (o *snapshotExportOutput) toText()  { outputText(o) }
func (o *snapshotExportOutput) toTable() { outputTable(o) }

func init() {
	snapshotCmd.AddCommand(&cobra.Command{
		Use:   "export <snapshot ID>",
		Short: "export snapshot",
		Long: fmt.Sprintf(`This command exports a volume snapshot.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&snapshotExportOutput{}), ", ")),
		Aliases: gListAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}
			return output(exportSnapshot(args[0]))
		},
	})
}

func exportSnapshot(snapshotID string) (outputter, error) {
	id, err := egoscale.ParseUUID(snapshotID)
	if err != nil {
		return nil, err
	}

	res, err := asyncRequest(&egoscale.ExportSnapshot{ID: id}, fmt.Sprintf("exporting snapshot %q", id))
	if err != nil {
		return nil, err
	}

	snapshot := res.(*egoscale.ExportSnapshotResponse)

	return &snapshotExportOutput{
		URL:      snapshot.PresignedURL,
		Checksum: snapshot.MD5sum,
	}, nil
}
