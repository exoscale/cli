package cmd

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/spf13/cobra"
)

type storageCORSRule struct {
	AllowedOrigins []string `json:"allowed_origins,omitempty"`
	AllowedMethods []string `json:"allowed_methods,omitempty"`
	AllowedHeaders []string `json:"allowed_headers,omitempty"`
}

// storageCORSRulesFromS3 converts a list of S3 CORS rules to a list of
// storageCORSRule.
func storageCORSRulesFromS3(v *s3.GetBucketCorsOutput) []storageCORSRule {
	rules := make([]storageCORSRule, 0)

	for _, rule := range v.CORSRules {
		rules = append(rules, storageCORSRule{
			AllowedOrigins: rule.AllowedOrigins,
			AllowedMethods: rule.AllowedMethods,
			AllowedHeaders: rule.AllowedHeaders,
		})
	}

	return rules
}

// toS3 converts a storageCORSRule object to the S3 CORS rule format.
func (r *storageCORSRule) toS3() s3types.CORSRule {
	return s3types.CORSRule{
		AllowedOrigins: r.AllowedOrigins,
		AllowedMethods: r.AllowedMethods,
		AllowedHeaders: r.AllowedHeaders,
	}
}

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
