package cmd

import (
	"fmt"
	"os"
	"strings"

	minio "github.com/minio/minio-go/v6"
	"github.com/spf13/cobra"
)

const (
	minioMaxRetry = 2
)

var sosCmdLongHelp = func() string {
	long := "Manage Exoscale Object Storage (SOS)"
	return long
}

var sosCmd = &cobra.Command{
	Use:              "sos",
	Short:            "Simple Object Storage management",
	Long:             sosCmdLongHelp(),
	TraverseChildren: true,

	Hidden: true,
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		fmt.Fprintln(os.Stderr,
			`**********************************************************************
The "exo sos" commands are deprecated and replaced by "exo storage",
they will be removed in a future version.
**********************************************************************`)
	},
}

type sosClient struct {
	*minio.Client
}


func newSOSClient() (*sosClient, error) {
	var (
		c   sosClient
		err error
	)

	z := gCurrentAccount.DefaultZone

	if err = c.setZone(z); err != nil {
		return nil, err
	}

	_, ok := os.LookupEnv("EXOSCALE_TRACE")
	if ok {
		c.TraceOn(os.Stderr)
	}

	return &c, nil
}

func (s *sosClient) setZone(zone string) error {
	// When a user wants to set the SOS zone to use for an operation, we actually have to re-create the
	// underlying Minio S3 client to specify the zone-based endpoint.

	endpoint := strings.TrimPrefix(
		strings.Replace(gCurrentAccount.SosEndpoint, "{zone}", zone, 1),
		"https://")
	minioClient, err := minio.NewV4(endpoint, gCurrentAccount.Key, gCurrentAccount.APISecret(), true)
	if err != nil {
		return err
	}

	minioClient.SetAppInfo("Exoscale-CLI", gVersion)

	if _, ok := os.LookupEnv("EXOSCALE_TRACE"); ok {
		minioClient.TraceOn(os.Stderr)
	}

	s.Client = minioClient

	return nil
}

func init() {
	minio.MaxRetry = minioMaxRetry

	RootCmd.AddCommand(sosCmd)
}
