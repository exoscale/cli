package cmd

import (
	"os"
	"strings"

	minio "github.com/minio/minio-go/v6"
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
	client, err := minio.NewV4(endpoint, gCurrentAccount.Key, gCurrentAccount.APISecret(), true)
	if err != nil {
		return nil, err
	}

	if _, ok := os.LookupEnv("EXOSCALE_TRACE"); ok {
		client.TraceOn(os.Stderr)
	}

	client.SetAppInfo("Exoscale-CLI", gVersion)

	return client, nil
}

func init() {
	RootCmd.AddCommand(sosCmd)
}
