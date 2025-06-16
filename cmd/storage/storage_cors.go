package storage

import (
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
