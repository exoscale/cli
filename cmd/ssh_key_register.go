package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type computeSSHKeyRegisterCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"register"`

	Name          string `cli-arg:"#"`
	PublicKeyFile string `cli-arg:"#"`
}

func (c *computeSSHKeyRegisterCmd) cmdAliases() []string { return gCreateAlias }

func (c *computeSSHKeyRegisterCmd) cmdShort() string {
	return "Register an SSH key"
}

func (c *computeSSHKeyRegisterCmd) cmdLong() string {
	return fmt.Sprintf(`This command registers a new SSH key.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&computeSSHKeyShowOutput{}), ", "))
}

func (c *computeSSHKeyRegisterCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *computeSSHKeyRegisterCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	var (
		sshKey *v3.SSHKey
		err    error
	)

	// Template registration can take a _long time_, raising
	// the Exoscale API client timeout as a precaution.
	client := globalstate.EgoscaleV3Client.WithHttpClient(&http.Client{Timeout: 30 * time.Minute})

	ctx := gContext

	publicKey, err := os.ReadFile(c.PublicKeyFile)
	if err != nil {
		return err
	}

	registerKeyRequest := v3.RegisterSSHKeyRequest{
		Name: c.Name,
		PublicKey: string(publicKey),
	}

	decorateAsyncOperations(fmt.Sprintf("Registering SSH key %q...", c.Name), func() error {
		op, err := client.RegisterSSHKey(ctx, registerKeyRequest)

		if err != nil {
			return err
		}

		 _, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		 if err != nil {
			return err
		}

		return nil
	})

	if !globalstate.Quiet {

		sshKey, err = client.GetSSHKey(ctx, c.Name)
		if err != nil {
			return err
		}

		return c.outputFunc(&computeSSHKeyShowOutput{
			Fingerprint: sshKey.Fingerprint,
			Name:        sshKey.Name,
		}, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(computeSSHKeyCmd, &computeSSHKeyRegisterCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
