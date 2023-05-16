package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
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
		sshKey *egoscale.SSHKey
		err    error
	)

	// Template registration can take a _long time_, raising
	// the Exoscale API client timeout as a precaution.
	globalstate.EgoscaleClient.Client.SetTimeout(30 * time.Minute)

	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(account.CurrentAccount.Environment, account.CurrentAccount.DefaultZone),
	)

	publicKey, err := os.ReadFile(c.PublicKeyFile)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Registering SSH key %q...", c.Name), func() {
		sshKey, err = globalstate.EgoscaleClient.RegisterSSHKey(ctx, account.CurrentAccount.DefaultZone, c.Name, string(publicKey))
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return c.outputFunc(&computeSSHKeyShowOutput{
			Fingerprint: *sshKey.Fingerprint,
			Name:        *sshKey.Name,
		}, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(computeSSHKeyCmd, &computeSSHKeyRegisterCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
