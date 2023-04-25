package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/exoscale/cli/pkg/globalstate"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/kballard/go-shellquote"
	"github.com/spf13/cobra"
)

type instanceSSHCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	sshInfo struct {
		ipAddress string
		keyFile   string
	} `cli-cmd:"-"`
	_ bool `cli-cmd:"ssh"`

	Instance string `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`

	IPv6        bool   `cli-flag:"ipv6" cli-short:"6" cli-help:"connect to the instance via its IPv6 address"`
	Login       string `cli-short:"l" cli-help:"SSH username to use for logging in (default: instance template default username)"`
	PrintCmd    bool   `cli-flag:"print-command" cli-usage:"print the SSH command that would be executed instead of executing it"`
	PrintConfig bool   `cli-flag:"print-ssh-config" cli-usage:"print the corresponding SSH information in a format compatible with ssh_config(5)"`
	SSHOpts     string `cli-flag:"ssh-options" cli-short:"o" cli-usage:"additional options to pass to the ssh(1) command"`
	Zone        string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceSSHCmd) buildSSHCommand() []string {
	cmd := []string{"ssh"}

	if _, err := os.Stat(c.sshInfo.keyFile); err == nil {
		cmd = append(cmd, "-i", c.sshInfo.keyFile)
	}

	if c.IPv6 {
		cmd = append(cmd, "-6")
	}

	if c.Login != "" {
		cmd = append(cmd, "-l", c.Login)
	}

	if c.SSHOpts != "" {
		opts, err := shellquote.Split(c.SSHOpts)
		if err == nil {
			cmd = append(cmd, opts...)
		}
	}

	cmd = append(cmd, c.sshInfo.ipAddress)

	return cmd
}

func (c *instanceSSHCmd) cmdAliases() []string { return nil }

func (c *instanceSSHCmd) cmdShort() string { return "Log into a Compute instance via SSH" }

func (c *instanceSSHCmd) cmdLong() string {
	return `This command connects to a Compute instance via SSH (requires the ssh(1) command).

To pass custom SSH options:

    exo compute instance ssh -o "-p 2222 -A" my-instance
`
}

func (c *instanceSSHCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSSHCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	instance, err := globalstate.GlobalEgoscaleClient.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if c.Login == "" {
		instanceTemplate, err := globalstate.GlobalEgoscaleClient.GetTemplate(ctx, c.Zone, *instance.TemplateID)
		if err != nil {
			return fmt.Errorf("error retrieving instance template: %w", err)
		}
		if instanceTemplate.DefaultUser != nil {
			c.Login = *instanceTemplate.DefaultUser
		}
	}

	c.sshInfo.keyFile = getInstanceSSHKeyPath(*instance.ID)

	c.sshInfo.ipAddress = instance.PublicIPAddress.String()
	if c.IPv6 {
		if instance.IPv6Address == nil {
			return fmt.Errorf("instance %q has no IPv6 address", c.Instance)
		}
		c.sshInfo.ipAddress = instance.IPv6Address.String()
	}

	sshCmd := c.buildSSHCommand()

	switch {
	case c.PrintConfig:
		out := bytes.NewBuffer(nil)
		_, _ = fmt.Fprintf(out, "Host %s\n", c.sshInfo.ipAddress)

		if c.Login != "" {
			_, _ = fmt.Fprintf(out, "User %s\n", c.Login)
		}

		if _, err := os.Stat(c.sshInfo.keyFile); err == nil {
			_, _ = fmt.Fprintf(out, "IdentityFile %q\n", c.sshInfo.keyFile)
		}

		fmt.Print(out.String())
		return nil

	case c.PrintCmd:
		fmt.Println(strings.Join(sshCmd, " "))
		return nil

	default:
		cmd := exec.Command("ssh", sshCmd[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout

		return cmd.Run()
	}
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceCmd, &instanceSSHCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
