package lifecycle

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "lifecycle",
	Short: "Object Storage Bucket lifecycle management",
	Long:  "Object Storage Bucket lifecycle management",
}
