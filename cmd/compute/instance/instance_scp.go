package instance

import (
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

type instanceSCPCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	scpInfo struct {
		ipAddress string
		keyFile   string
	} `cli-cmd:"-"`
	_ bool `cli-cmd:"scp"`

	Instance string `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`
	Source   string `cli-arg:"#"`
	Target   string `cli-arg:"#"`

	IPv6      bool   `cli-flag:"ipv6" cli-short:"6" cli-help:"connect to the instance via its IPv6 address"`
	Login     string `cli-short:"l" cli-help:"SCP username to use for logging in (default: instance template default username)"`
	PrintCmd  bool   `cli-flag:"print-command" cli-usage:"print the SCP command that would be executed instead of executing it"`
	Recursive bool   `cli-short:"r" cli-usage:"recursively copy entire directories"`
	ReplStr   string `cli-flag:"replace-str" cli-short:"i" cli-usage:"string to replace with the actual Compute instance information (i.e. username@IP-ADDRESS)"`
	SCPOpts   string `cli-flag:"scp-options" cli-short:"o" cli-usage:"additional options to pass to the scp(1) command"`
	Zone      string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceSCPCmd) buildSCPCommand() []string {
	cmd := []string{"scp"}

	if _, err := os.Stat(c.scpInfo.keyFile); err == nil {
		cmd = append(cmd, "-i", c.scpInfo.keyFile)
	}

	if c.Recursive {
		cmd = append(cmd, "-r")
	}

	if c.SCPOpts != "" {
		opts, err := shellquote.Split(c.SCPOpts)
		if err == nil {
			cmd = append(cmd, opts...)
		}
	}

	// Parse arguments to find the replacement string to interpolate with <[username@]target instance IP address>
	for _, arg := range []string{c.Source, c.Target} {
		if strings.Contains(arg, c.ReplStr) {
			remote := c.scpInfo.ipAddress
			if c.Login != "" {
				remote = fmt.Sprintf("%s@%s", c.Login, remote)
			}

			arg = strings.Replace(arg, c.ReplStr, remote, 1)
		}
		cmd = append(cmd, arg)
	}

	return cmd
}

func (c *instanceSCPCmd) CmdAliases() []string { return nil }

func (c *instanceSCPCmd) CmdShort() string { return "SCP files to/from a Compute instance" }

func (c *instanceSCPCmd) CmdLong() string {
	return `This command executes the scp(1) command to send or receive files to/from the
specified Compute instance. TARGET (or SOURCE, depending on the direction
of the transfer) must contain the "{}" marker which will be interpolated at
run time with the actual Compute instance IP address, similar to xargs(1).
This marker can be replaced by another string via the --replace-str|-i flag.

Example:

    exo compute instance scp my-instance hello-world.txt {}:
    exo compute instance scp -i%% my-instance %%:/etc/motd .
`
}

func (c *instanceSCPCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSCPCmd) CmdRun(_ *cobra.Command, _ []string) error {
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

	if c.Login == "" {
		instanceTemplate, err := client.GetTemplate(ctx, instance.Template.ID)
		if err != nil {
			return fmt.Errorf("error retrieving instance template: %w", err)
		}
		if instanceTemplate.DefaultUser != "" {
			c.Login = instanceTemplate.DefaultUser
		}
	}

	c.scpInfo.keyFile = ssh.GetInstanceSSHKeyPath(instance.ID.String())

	c.scpInfo.ipAddress = instance.PublicIP.String()
	if c.IPv6 {
		if instance.Ipv6Address == "" {
			return fmt.Errorf("instance %q has no IPv6 address", c.Instance)
		}
		c.scpInfo.ipAddress = instance.Ipv6Address
	}

	scpCmd := c.buildSCPCommand()

	if c.PrintCmd {
		fmt.Println(strings.Join(scpCmd, " "))
		return nil
	}

	cmd := exec.Command("scp", scpCmd[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceCmd, &instanceSCPCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),

		ReplStr: "{}",
	}))
}
