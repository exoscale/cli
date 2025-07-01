package instance

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kballard/go-shellquote"
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/ssh"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceSSHCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	sshInfo struct {
		ipAddress string
		keyFile   string
	} `cli-cmd:"-"`
	_ bool `cli-cmd:"ssh"`

	Instance        string `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`
	CommandArgument string `cli-arg:"?" cli-usage:"COMMAND ARGUMENT"`

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

func (c *instanceSSHCmd) CmdAliases() []string { return nil }

func (c *instanceSSHCmd) CmdShort() string { return "Log into a Compute instance via SSH" }

func (c *instanceSSHCmd) CmdLong() string {
	return `This command connects to a Compute instance via SSH (requires the ssh(1) command).

To pass custom SSH options:

    exo compute instance ssh -o "-p 2222 -A" my-instance
`
}

func (c *instanceSSHCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSSHCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	instances, err := client.ListInstances(ctx)
	if err != nil {
		return err
	}
	instance, err := instances.FindListInstancesResponseInstances(c.Instance)
	if err != nil {
		return err
	}

	// No ssh possible for Private Instances
	if instance.PublicIP.String() == "" || instance.PublicIP.String() == "none" {
		return fmt.Errorf("instance %q is a Private Instance (`exo compute instance ssh` is not supported)", c.Instance)
	}

	if c.Login == "" {
		instanceTemplate, err := client.GetTemplate(ctx, instance.Template.ID)
		if err != nil {
			return fmt.Errorf("error retrieving instance template: %w", err)
		}
		if instanceTemplate.DefaultUser != "" {
			c.Login = instanceTemplate.DefaultUser
		}
	}

	c.sshInfo.keyFile = ssh.GetInstanceSSHKeyPath(instance.ID.String())

	c.sshInfo.ipAddress = instance.PublicIP.String()
	if c.IPv6 {
		if instance.Ipv6Address == "" {
			return fmt.Errorf("instance %q has no IPv6 address", c.Instance)
		}
		c.sshInfo.ipAddress = instance.Ipv6Address
	}

	sshCmd := c.buildSSHCommand()
	if c.CommandArgument != "" {
		sshCmd = append(sshCmd, c.CommandArgument)
	}

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
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceCmd, &instanceSSHCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
