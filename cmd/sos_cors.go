package cmd

import (
	"fmt"
	"strings"

	minio "github.com/minio/minio-go"
	"github.com/spf13/cobra"
)

// sosCORSCmd represents the sos cors command
var sosCORSCmd = &cobra.Command{
	Use:   "cors <bucket name>",
	Short: "Bucket(s) CORS management",
}

func init() {
	sosCmd.AddCommand(sosCORSCmd)
}

// sosShowCORSCmd represents the sos cors show
var sosShowCORSCmd = &cobra.Command{
	Use:     "show <bucket name>",
	Short:   "show bucket CORSs",
	Aliases: gShowAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}

		minioClient, err := newMinioClient(sosZone)
		if err != nil {
			return err
		}

		location, err := minioClient.GetBucketLocation(args[0])
		if err != nil {
			return err
		}

		minioClient, err = newMinioClient(location)
		if err != nil {
			return err
		}

		bucketInfo, err := minioClient.GetBucketCORS(args[0])
		if err != nil {
			return err
		}

		return showBucketCors(*bucketInfo)
	},
}

func showBucketCors(bucketInfo minio.BucketInfo) error {
	for _, rule := range bucketInfo.CORS {
		fmt.Printf("Origin: %s\n", strings.Join(rule.AllowedOrigin, ","))
		fmt.Printf("Method: %s\n", strings.Join(rule.AllowedMethod, ","))
		if len(rule.AllowedHeader) > 0 {
			fmt.Printf("Header: %s\n", strings.Join(rule.AllowedHeader, ","))
		}
		if len(rule.ExposeHeader) > 0 {
			fmt.Printf("Expose Header: %s\n", strings.Join(rule.ExposeHeader, ","))
		}
		fmt.Println()
	}

	return nil
}

func init() {
	sosCORSCmd.AddCommand(sosShowCORSCmd)
}
