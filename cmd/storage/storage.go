package storage

import (
	"fmt"
	"os"
	"reflect"

	exocmd "github.com/exoscale/cli/cmd"

	"github.com/exoscale/cli/cmd/storage/cors"
	"github.com/exoscale/cli/cmd/storage/headers"
	"github.com/exoscale/cli/cmd/storage/metadata"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/storage/sos"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/smithy-go/middleware"
	v3 "github.com/exoscale/egoscale/v3"

	"github.com/exoscale/cli/cmd/storage/bucket"
	"github.com/spf13/cobra"
)

var (
	Cmd = &cobra.Command{
		Use:              "storage",
		Short:            "Object Storage management",
		Long:             storageCmdLongHelp(),
		TraverseChildren: true,
		SuggestFor:       []string{"sos"},
	}
)

// v3.UserAgent is deprecated, so we use a bit of reflection to
// retrieve the user agent of the default client
func getUserAgent(c *v3.Client) string {
	r := reflect.ValueOf(*c)
	userAgent := r.FieldByName("userAgent")
	return userAgent.String()
}

func init() {
	Cmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		// We have to wait until the actual command execution to assign a value to this variable
		// because some of the global variables used are not initialized before Cobra executes
		// the command.

		sos.CommonConfigOptFns = []func(*awsconfig.LoadOptions) error{
			// Custom HTTP client User-Agent
			awsconfig.WithAPIOptions([]func(*middleware.Stack) error{
				awsmiddleware.AddUserAgentKeyValue("Exoscale-CLI",
					fmt.Sprintf("%s (%s) %s", globalstate.GitVersion, globalstate.GitCommit, getUserAgent(globalstate.EgoscaleV3Client))),
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
	exocmd.RootCmd.AddCommand(Cmd)
	Cmd.AddCommand(bucket.Cmd)
	Cmd.AddCommand(cors.Cmd)
	Cmd.AddCommand(headers.Cmd)
	Cmd.AddCommand(metadata.Cmd)
}

var storageCmdLongHelp = func() string {
	long := "Manage Exoscale Object Storage"
	return long
}
