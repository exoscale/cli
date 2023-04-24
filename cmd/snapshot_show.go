package cmd

import (
	"fmt"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type snapshotShowOutput struct {
	ID           string `json:"id"`
	Date         string `json:"date"`
	InstanceID   string `json:"instance_id"`
	InstanceName string `json:"instance_name"`
	State        string `json:"state"`
	Size         string `json:"size"`
	TemplateID   string `json:"template_id"`
	TemplateName string `json:"template_name"`
}

func (o *snapshotShowOutput) Type() string { return "Snapshot" }
func (o *snapshotShowOutput) toJSON()      { output.JSON(o) }
func (o *snapshotShowOutput) toText()      { output.Text(o) }
func (o *snapshotShowOutput) toTable()     { output.Table(o) }

func init() {
	snapshotCmd.AddCommand(&cobra.Command{
		Use:   "show ID",
		Short: "Show a snapshot details",
		Long: fmt.Sprintf(`This command shows a snapshot details.

Supported output template annotations: %s`,
			strings.Join(output.output.OutputterTemplateAnnotations(&snapshotShowOutput{}), ", ")),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}

			snapshot, err := getSnapshotByNameOrID(args[0])
			if err != nil {
				return err
			}

			return printOutput(showSnapshot(snapshot))
		},
	})
}

func showSnapshot(snapshot *egoscale.Snapshot) (output.Outputter, error) {
	resp, err := cs.GetWithContext(gContext, &egoscale.Volume{ID: snapshot.VolumeID})
	if err != nil {
		return nil, err
	}
	volume := resp.(*egoscale.Volume)

	return &snapshotShowOutput{
		ID:           snapshot.ID.String(),
		InstanceID:   volume.VirtualMachineID.String(),
		InstanceName: volume.VMName,
		Date:         snapshot.Created,
		State:        snapshot.State,
		Size:         humanize.IBytes(uint64(snapshot.Size)),
		TemplateID:   volume.TemplateID.String(),
		TemplateName: volume.TemplateName,
	}, nil
}
