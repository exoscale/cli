package storage

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/flags"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/pkg/storage/sos/object"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

var storageListCmd = &cobra.Command{
	Use:   "list [sos://BUCKET[/[PREFIX/]]",
	Short: "List buckets and objects",
	Long: fmt.Sprintf(`This command lists buckets and their objects.

If no argument is passed, this commands lists existing buckets. If a prefix is
specified (e.g. "sos://my-bucket/.../") the command lists the objects stored
in the bucket under the corresponding prefix.

Supported output template annotations:

  * When listing buckets: %s
  * When listing objects: %s`,
		strings.Join(output.TemplateAnnotations(&sos.ListBucketsItemOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&object.ListObjectsItemOutput{}), ", ")),
	Aliases: exocmd.GListAlias,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)
		}

		if err := flags.ValidateTimestampFlags(cmd); err != nil {
			return err
		}

		return flags.ValidateVersionFlags(cmd)
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			bucket string
			prefix string
		)

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return utils.PrintOutput(listStorageBuckets(zone))
		}

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		stream, err := cmd.Flags().GetBool("stream")
		if err != nil {
			return err
		}

		parts := strings.SplitN(args[0], "/", 2)
		bucket = parts[0]
		if len(parts) > 1 {
			prefix = parts[1]
		}

		filters, err := flags.TranslateTimeFilterFlagsToFilterFuncs(cmd)
		if err != nil {
			return err
		}

		storage, err := sos.NewStorageClient(
			exocmd.GContext,
			sos.ClientOptZoneFromBucket(exocmd.GContext, bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		listVersions, err := cmd.Flags().GetBool(flags.Versions)
		if err != nil {
			return err
		}

		versionFilters, err := flags.TranslateVersionFilterFlagsToFilterFuncs(cmd)
		if err != nil {
			return err
		}

		if listVersions || len(versionFilters) > 0 {
			list := storage.ListVersionedObjectsFunc(bucket, prefix, recursive, stream)
			return utils.PrintOutput(storage.ListObjectsVersions(exocmd.GContext, list, recursive, stream, filters, versionFilters))
		}

		list := storage.ListObjectsFunc(bucket, prefix, recursive, stream)
		return utils.PrintOutput(storage.ListObjects(exocmd.GContext, list, recursive, stream, filters))
	},
}

func init() {
	storageListCmd.Flags().BoolP("recursive", "r", false,
		"list bucket recursively")
	storageListCmd.Flags().BoolP("stream", "s", false,
		"stream listed files instead of waiting for complete listing (useful for large buckets)")
	flags.AddVersionsFlags(storageListCmd)
	flags.AddTimeFilterFlags(storageListCmd)
	flags.AddZoneListFlag(storageListCmd)
	storageCmd.AddCommand(storageListCmd)
}

func listStorageBuckets(zone string) (output.Outputter, error) {

	out := make(sos.ListBucketsOutput, 0)

	// ListSosBucketsUsageWithResponse is a global command, use default zone
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, account.CurrentAccount.DefaultZone))

	res, err := globalstate.EgoscaleClient.ListSosBucketsUsageWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	for _, b := range *res.JSON200.SosBucketsUsage {
		created := *b.CreatedAt
		if err != nil {
			return nil, err
		}

		// Filter the zone if set as flag
		if zone == "" || zone == string(*b.ZoneName) {
			out = append(out, sos.ListBucketsItemOutput{
				Name:    *b.Name,
				Zone:    string(*b.ZoneName),
				Size:    *b.Size,
				Created: created.Format(object.TimestampFormat),
			})
		}
	}

	return &out, nil
}
