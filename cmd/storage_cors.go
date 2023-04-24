package cmd

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/spf13/cobra"
)

var storageCORSCmd = &cobra.Command{
	Use:   "cors",
	Short: "Manage buckets CORS configuration",
	Long: `These commands allow you to manage the CORS configuration of a bucket.

For more information on CORS, please refer to the Exoscale Storage
documentation:
https://community.exoscale.com/documentation/storage/cors/

Notes:

  * It is not possible to edit a CORS configuration rule once it's been
    created, nor to delete rules individually -- the whole configuration must
    be reset using the "exo storage cors reset" command.
`,
}

func init() {
	storageCmd.AddCommand(storageCORSCmd)
}

// CORSRulesFromS3 converts a list of S3 CORS rules to a list of
// CORSRule.
func CORSRulesFromS3(v *s3.GetBucketCorsOutput) []CORSRule {
	rules := make([]CORSRule, 0)

	for _, rule := range v.CORSRules {
		rules = append(rules, CORSRule{
			AllowedOrigins: rule.AllowedOrigins,
			AllowedMethods: rule.AllowedMethods,
			AllowedHeaders: rule.AllowedHeaders,
		})
	}

	return rules
}

// toS3 converts a sos.CORSRule object to the S3 CORS rule format.
func (r *sos.CORSRule) toS3() s3types.CORSRule {
	return s3types.CORSRule{
		AllowedOrigins: r.AllowedOrigins,
		AllowedMethods: r.AllowedMethods,
		AllowedHeaders: r.AllowedHeaders,
	}
}
