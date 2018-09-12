package cmd

import (
	"strings"

	minio "github.com/minio/minio-go"
	"github.com/spf13/cobra"
)

const (
	sosZone = "ch-dk-2"
)

// sosCmd represents the sos command
var sosCmd = &cobra.Command{
	Use:   "sos",
	Short: "Simple Object Storage management",
}

func newMinioClient(zone string) (*minio.Client, error) {
	endpoint := strings.Replace(gCurrentAccount.SosEndpoint, "https://", "", -1)
	endpoint = strings.Replace(endpoint, "{zone}", zone, -1)
	return minio.NewV4(endpoint, gCurrentAccount.Key, gCurrentAccount.Secret, true)
}

func init() {
	RootCmd.AddCommand(sosCmd)
}
