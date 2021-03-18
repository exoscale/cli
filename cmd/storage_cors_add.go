package cmd

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	storageCORSAddCmdFlagAllowedOrigin = "allowed-origin"
	storageCORSAddCmdFlagAllowedMethod = "allowed-method"
	storageCORSAddCmdFlagAllowedHeader = "allowed-header"
)

// storageCORSRuleFromCmdFlags returns a non-nil pointer to a storageCORSRule struct if at least
// one of the CORS-related command flags is set.
func storageCORSRuleFromCmdFlags(flags *pflag.FlagSet) *storageCORSRule {
	var cors *storageCORSRule

	flags.VisitAll(func(flag *pflag.Flag) {
		switch flag.Name {
		case storageCORSAddCmdFlagAllowedOrigin:
			if v, _ := flags.GetStringSlice(storageCORSAddCmdFlagAllowedOrigin); len(v) > 0 {
				if cors == nil {
					cors = &storageCORSRule{}
				}

				cors.AllowedOrigins = v
			}

		case storageCORSAddCmdFlagAllowedMethod:
			if v, _ := flags.GetStringSlice(storageCORSAddCmdFlagAllowedMethod); len(v) > 0 {
				if cors == nil {
					cors = &storageCORSRule{}
				}

				cors.AllowedMethods = v
			}

		case storageCORSAddCmdFlagAllowedHeader:
			if v, _ := flags.GetStringSlice(storageCORSAddCmdFlagAllowedHeader); len(v) > 0 {
				if cors == nil {
					cors = &storageCORSRule{}
				}

				cors.AllowedHeaders = v
			}

		default:
			return
		}
	})

	return cors
}

var storageCORSAddCmd = &cobra.Command{
	Use:   "add <bucket>",
	Short: "Add a CORS configuration rule to a bucket",
	Long: `This command adds a new rule to the current bucket CORS
configuration.

Example:

    exo storage cors add my-bucket \
        --allowed-origin "https://my-website.net" \
        --allowed-method "*"
`,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		return cmdCheckRequiredFlags(cmd, []string{
			storageCORSAddCmdFlagAllowedOrigin,
			storageCORSAddCmdFlagAllowedMethod,
		})
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		bucket := args[0]

		certsFile, err := cmd.Flags().GetString("certs-file")
		if err != nil {
			return err
		}

		storage, err := newStorageClient(
			storageClientOptWithCertsFile(certsFile),
			storageClientOptZoneFromBucket(bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %v", err)
		}

		cors := storageCORSRuleFromCmdFlags(cmd.Flags())
		if err := storage.addBucketCORSRule(bucket, cors); err != nil {
			return fmt.Errorf("unable to add rule to the bucket CORS configuration: %s", err)
		}

		if !gQuiet {
			return output(storage.showBucket(bucket))
		}

		return nil
	},
}

func init() {
	storageCORSAddCmd.Flags().StringSlice(storageCORSAddCmdFlagAllowedOrigin, nil,
		"allowed origin (can be repeated multiple times)")
	storageCORSAddCmd.Flags().StringSlice(storageCORSAddCmdFlagAllowedMethod, nil,
		"allowed method (can be repeated multiple times)")
	storageCORSAddCmd.Flags().StringSlice(storageCORSAddCmdFlagAllowedHeader, nil,
		"allowed header (can be repeated multiple times)")
	storageCORSCmd.AddCommand(storageCORSAddCmd)
}

func (c *storageClient) addBucketCORSRule(bucket string, cors *storageCORSRule) error {
	curCORS, err := c.GetBucketCors(gContext, &s3.GetBucketCorsInput{Bucket: aws.String(bucket)})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NoSuchCORSConfiguration" {
				curCORS = &s3.GetBucketCorsOutput{}
			}
		}

		if cors == nil {
			return fmt.Errorf("unable to retrieve bucket CORS configuration: %s", err)
		}
	}

	_, err = c.PutBucketCors(gContext, &s3.PutBucketCorsInput{
		Bucket: &bucket,
		CORSConfiguration: &s3types.CORSConfiguration{
			CORSRules: append(curCORS.CORSRules, cors.toS3()),
		},
	})

	return err
}
