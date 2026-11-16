package storage

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/pkg/storage/sos/object"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

var (
	// A version filter is a version ID (865029700534464769) or a version number (v123).
	listVersionRegex       = regexp.MustCompile(`v?\d+`)
	listVersionNumberRegex = regexp.MustCompile(`v\d+`)
)

type storageListCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Path string `cli-arg:"?" cli-usage:"sos://BUCKET[/[PREFIX/]"`

	Recursive bool `cli-short:"r" cli-usage:"list bucket recursively"`
	Stream    bool `cli-short:"s" cli-usage:"stream listed files instead of waiting for complete listing (useful for large buckets)"`

	// The --only-versions and --exclude-versions usages repeat the same text:
	// a struct tag can't be built from a constant.
	Versions        bool     `cli-usage:"list all versions of objects(if the bucket is versioned)"`
	OnlyVersions    []string `cli-usage:"limit the versions to be listed; accepts comma separated version IDs(865029700534464769) and numbers(v123); implies --versions"`
	ExcludeVersions []string `cli-usage:"exclude versions from being listed; accepts comma separated version IDs(865029700534464769) and numbers(v123); implies --versions"`

	OlderThan          time.Duration `cli-usage:"only objects older than a duration. Accepts durations in the format of Go's time.ParseDuration. examples: \"2h45m\", \"10m\", \"45s\""`
	OlderThanTimestamp time.Time     `cli-usage:"only objects older than an ISO 8601 timestamp. examples: '2023-06-07T10:00:00+02:00', use the date command $(date -d \"yesterday 10am\" --iso-8601=seconds)"`
	NewerThan          time.Duration `cli-usage:"only objects newer than a duration. Accepts durations in the format of Go's time.ParseDuration. examples: \"2h45m\", \"10m\", \"45s\""`
	NewerThanTimestamp time.Time     `cli-usage:"only objects newer than an ISO 8601 timestamp. examples: '2023-06-07T10:00:00+02:00', use the date command $(date -d \"yesterday 10am\" --iso-8601=seconds)"`

	// Empty means the bucket's zone, or all zones when listing buckets.
	Zone string `cli-short:"z" cli-usage:"Exoscale zone"`
}

func (c *storageListCmd) CmdAliases() []string { return clicmd.GListAlias }
func (c *storageListCmd) CmdShort() string     { return "List buckets and objects" }
func (c *storageListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists buckets and their objects.

If no argument is passed, this commands lists existing buckets. If a prefix is
specified (e.g. "sos://my-bucket/.../") the command lists the objects stored
in the bucket under the corresponding prefix.

Supported output template annotations:

  * When listing buckets: %s
  * When listing objects: %s`,
		strings.Join(output.TemplateAnnotations(&sos.ListBucketsItemOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&object.ListObjectsItemOutput{}), ", "))
}

func (c *storageListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) > 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	if len(args) == 1 {
		args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)
	}

	if err := clicmd.CliCommandDefaultPreRun(c, cmd, args); err != nil {
		return err
	}

	if len(c.OnlyVersions) > 0 && len(c.ExcludeVersions) > 0 {
		return fmt.Errorf("--only-versions and --exclude-versions are mutually exclusive")
	}

	// At most one of the two lists is set here.
	for _, v := range slices.Concat(c.OnlyVersions, c.ExcludeVersions) {
		if !listVersionRegex.MatchString(v) {
			return fmt.Errorf("%q is not a valid version id(865029700534464769) or version number(v123)", v)
		}

		// Version numbers are computed from the complete listing, which the
		// stream mode doesn't wait for.
		if c.Stream && listVersionNumberRegex.MatchString(v) {
			return fmt.Errorf("cannot use version number filter %q in combination with --stream flag", v)
		}
	}

	return nil
}

func (c *storageListCmd) CmdRun(_ *cobra.Command, args []string) error {
	// Not c.Path == "": "list sos://" (an empty bucket) lists objects, not buckets.
	if len(args) == 0 {
		return utils.PrintOutput(listStorageBuckets(c.Zone))
	}

	bucket, prefix, _ := strings.Cut(c.Path, "/")

	filters := listTimeFilters(c.OlderThan, c.OlderThanTimestamp, c.NewerThan, c.NewerThanTimestamp)

	zoneOpt := sos.ClientOptZoneFromBucket(exocmd.GContext, bucket)
	if c.Zone != "" {
		zoneOpt = sos.ClientOptWithZone(c.Zone)
	}

	storage, err := sos.NewStorageClient(exocmd.GContext, zoneOpt)
	if err != nil {
		return fmt.Errorf("unable to initialize storage client: %w", err)
	}

	versionFilters := listVersionFilters(c.OnlyVersions, c.ExcludeVersions)

	if c.Versions || len(versionFilters) > 0 {
		list := storage.ListVersionedObjectsFunc(bucket, prefix, c.Recursive, c.Stream)
		return utils.PrintOutput(storage.ListObjectsVersions(exocmd.GContext, list, c.Recursive, c.Stream, filters, versionFilters))
	}

	list := storage.ListObjectsFunc(bucket, prefix, c.Recursive, c.Stream)
	return utils.PrintOutput(storage.ListObjects(exocmd.GContext, list, c.Recursive, c.Stream, filters))
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &storageListCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}

func listStorageBuckets(zone string) (output.Outputter, error) {

	out := make(sos.ListBucketsOutput, 0)

	// ListSosBucketsUsageWithResponse is a global command, use default zone
	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return nil, err
	}

	bucketsList, err := client.ListSOSBucketsUsage(ctx)
	if err != nil {
		return nil, err
	}

	for _, b := range bucketsList.SOSBucketsUsage {
		created := b.CreatedAT

		// Filter the zone if set as flag
		if zone == "" || zone == string(b.ZoneName) {
			out = append(out, sos.ListBucketsItemOutput{
				Name:    b.Name,
				Zone:    string(b.ZoneName),
				Size:    b.Size,
				Created: created.Format(object.TimestampFormat),
			})
		}
	}

	return &out, nil
}
