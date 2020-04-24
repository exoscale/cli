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
		Use:   "export <snapshotID>",
		Short: "export snapshot",
		Long: fmt.Sprintf(`This command exports a volume snapshot.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&snapshotExportOutput{}), ", ")),
		Aliases: gListAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}
			return exportSnapshot(args[0])
		},
	})
}

func exportSnapshot(snapshotID string) error {
	id, err := egoscale.ParseUUID(snapshotID)
	if err != nil {
		return err
	}

	res, err := asyncRequest(&egoscale.ExportSnapshot{ID: id}, fmt.Sprintf("exporting snapshot of %q", id))
	if err != nil {
		return err
	}

	snapshot := res.(*egoscale.ExportSnapshotResponse)

	if !gQuiet {
		return output(&snapshotExportOutput{
			URL:      snapshot.PresignedURL,
			Checksum: snapshot.MD5sum,
		}, err)
	}

	return nil
}
