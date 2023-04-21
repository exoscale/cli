package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/exoscale/cli/table"
	"github.com/spf13/cobra"
)

func init() {
	storageBucketObjectOwnershipCmd.Flags().StringP(zoneFlagLong, zoneFlagShort, "", zoneFlagMsg)
	storageBucketCmd.AddCommand(storageBucketObjectOwnershipCmd)
}

var storageBucketObjectOwnershipCmd = &cobra.Command{
	// TODO
	Use:     "object-ownership {status,object-writer,bucket-owner-enforced,bucket-owner-preferred} sos://BUCKET",
	Aliases: []string{"oo"},
	Short:   "Manage the Object Ownership setting of a Storage Bucket",
	Long:    storageBucketObjectOwnershipCmdLongHelp(),

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[0] = strings.TrimPrefix(args[0], storageBucketPrefix)

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{zoneFlagLong})
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		ownershipCommand := args[0]
		bucket := args[1]

		fmt.Println(ownershipCommand)

		zone, err := cmd.Flags().GetString(zoneFlagLong)
		if err != nil {
			return err
		}

		storage, err := newStorageClient(
			storageClientOptWithZone(zone),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		return output(storage.getBucketObjectOwnership(cmd.Context(), bucket))
	},
}

var storageBucketObjectOwnershipCmdLongHelp = func() string {
	// TODO
	return "Manage the Object Ownership setting of a Storage Bucket"
}

type storageBucketObjectOwnershipOutput struct {
	Bucket          string `json:"bucket"`
	ObjectOwnership string `json:"objectOwnership"`
}

func (o *storageBucketObjectOwnershipOutput) toJSON() { outputJSON(o) }
func (o *storageBucketObjectOwnershipOutput) toText() { outputText(o) }
func (o *storageBucketObjectOwnershipOutput) toTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()
	t.SetHeader([]string{"Bucket Object Ownership"})

	t.Append([]string{"Bucket", o.Bucket})
	t.Append([]string{"Zone", o.ObjectOwnership})
}

func (c storageClient) getBucketObjectOwnership(ctx context.Context, bucket string) (outputter, error) {
	params := s3.GetBucketOwnershipControlsInput{
		Bucket: aws.String(bucket),
	}

	resp, err := c.GetBucketOwnershipControls(gContext, &params)
	if err != nil {
		// TODO wrap
		return nil, err
	}

	out := storageBucketObjectOwnershipOutput{
		Bucket:          bucket,
		ObjectOwnership: string(resp.OwnershipControls.Rules[0].ObjectOwnership),
	}

	return &out, nil
}
