package ssh_key

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type computeSSHKeyRegisterCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"register"`

	Name          string `cli-arg:"#"`
	PublicKeyFile string `cli-arg:"#"`
}

func (c *computeSSHKeyRegisterCmd) CmdAliases() []string { return exocmd.GCreateAlias }

func (c *computeSSHKeyRegisterCmd) CmdShort() string {
	return "Register an SSH key"
}

func (c *computeSSHKeyRegisterCmd) CmdLong() string {
	return fmt.Sprintf(`This command registers a new SSH key.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&computeSSHKeyShowOutput{}), ", "))
}

func (c *computeSSHKeyRegisterCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *computeSSHKeyRegisterCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	// Template registration can take a _long time_, raising
	// the Exoscale API client timeout as a precaution.
	client := globalstate.EgoscaleV3Client.WithHttpClient(&http.Client{Timeout: 30 * time.Minute})

	ctx := exocmd.GContext

	publicKey, err := os.ReadFile(c.PublicKeyFile)
	if err != nil {
		return err
	}

	registerKeyRequest := v3.RegisterSSHKeyRequest{
		Name:      c.Name,
		PublicKey: string(publicKey),
	}

	err = utils.DecorateAsyncOperations(fmt.Sprintf("Registering SSH key %q...", c.Name), func() error {
		op, err := client.RegisterSSHKey(ctx, registerKeyRequest)
		if err != nil {
			return fmt.Errorf("exoscale: error while registering SSH key: %w", err)
		}

		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		if err != nil {
			return fmt.Errorf("exoscale: error while waiting for SSH key registration: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&computeSSHKeyShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Key:                c.Name,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(computeSSHKeyCmd, &computeSSHKeyRegisterCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
