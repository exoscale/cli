package storage

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/smithy-go/middleware"
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/storage/sos"
	v3 "github.com/exoscale/egoscale/v3"
)

var (
	storageCmd = &cobra.Command{
		Use:              "storage",
		Short:            "Object Storage management",
		Long:             storageCmdLongHelp(),
		TraverseChildren: true,
		SuggestFor:       []string{"sos"},
	}
)

func init() {
	storageCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		// We have to wait until the actual command execution to assign a value to this variable
		// because some of the global variables used are not initialized before Cobra executes
		// the command.
		sos.CommonConfigOptFns = []func(*awsconfig.LoadOptions) error{
			// Custom HTTP client User-Agent
			awsconfig.WithAPIOptions([]func(*middleware.Stack) error{
				awsmiddleware.AddUserAgentKeyValue("Exoscale-CLI",
					fmt.Sprintf("%s (%s) %s", exocmd.GVersion, exocmd.GCommit, v3.UserAgent)),
			}),

			// Conditional HTTP client request tracing
			awsconfig.WithClientLogMode(func() aws.ClientLogMode {
				if _, ok := os.LookupEnv("EXOSCALE_TRACE"); ok {
					return aws.LogRequest | aws.LogResponse
				}
				return 0
			}()),
		}

		return nil
	}
	exocmd.RootCmd.AddCommand(storageCmd)
}

var storageCmdLongHelp = func() string {
	long := "Manage Exoscale Object Storage"
	return long
}
