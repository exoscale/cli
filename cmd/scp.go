package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kballard/go-shellquote"
	"github.com/spf13/cobra"
)

type scpInfo struct {
	sshInfo

	recursive bool
	replStr   string
}

var scpCmd = &cobra.Command{
	Use:   "scp <instance name | ID> <source> <target>",
	Short: "SCP files to/from a Compute instance",
	Long: `This command executes the "scp" command to send or receive files to/from the
specified Compute instance. The target (or source, depending on the direction
of the transfer) must contain the "{}" marker which will be interpolated at
run time with the actual Compute instance IP address, similar to xargs(1).
This marker can be replaced by another string via the --replace-str|-i flag.

Example:

    exo scp my-instance hello-world.txt {}:
    exo scp -i%% my-instance %%:/etc/motd .
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 3 {
			return cmd.Usage()
		}

		ipv6, err := cmd.Flags().GetBool("ipv6")
		if err != nil {
			return err
		}

		printCmd, err := cmd.Flags().GetBool("print")
		if err != nil {
			return err
		}

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		replStr, err := cmd.Flags().GetString("replace-str")
		if err != nil {
			return err
		}

		sshOpts, err := cmd.Flags().GetString("ssh-options")
		if err != nil {
			return err
		}

		sshInfo, err := getSSHInfo(args[0], ipv6)
		if err != nil {
			return err
		}
		sshInfo.opts = sshOpts

		scpCmd := buildSCPCommand(&scpInfo{
			sshInfo:   *sshInfo,
			recursive: recursive,
			replStr:   replStr,
		}, args[1:])

		if printCmd {
			fmt.Println(strings.Join(scpCmd, " "))
			return nil
		}

		return runSCP(scpCmd[1:])
	},
}

func buildSCPCommand(info *scpInfo, args []string) []string {
	cmd := []string{"scp"}

	if _, err := os.Stat(info.sshKeys); err == nil {
		cmd = append(cmd, "-i", info.sshKeys)
	}

	if info.recursive {
		cmd = append(cmd, "-r")
	}

	if info.opts != "" {
		opts, err := shellquote.Split(info.opts)
		if err == nil {
			cmd = append(cmd, opts...)
		}
	}

	// Parse arguments to find the replacement string to interpolate with <[username@]target instance IP address>
	for i, a := range args {
		if strings.Contains(a, info.replStr) {
			remote := info.ip.String()
			if info.username != "" {
				remote = fmt.Sprintf("%s@%s", info.username, remote)
			}

			args[i] = strings.Replace(args[i], info.replStr, remote, 1)
			break
		}
	}
	cmd = append(cmd, args...)

	return cmd
}

func runSCP(args []string) error {
	cmd := exec.Command("scp", args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

func init() {
	scpCmd.Flags().StringP("replace-str", "i", "{}",
		"String to replace with the actual Compute instance information (i.e. username@<IP address>)")
	scpCmd.Flags().StringP("ssh-options", "o", "",
		"Additional options to pass to the `scp` command (e.g. -o \"-l my-user -p 2222\"`)")
	scpCmd.Flags().BoolP("print", "p", false, "Print SCP command")
	scpCmd.Flags().BoolP("recursive", "r", false, "Recursively copy entire directories")
	scpCmd.Flags().BoolP("ipv6", "6", false, "Use IPv6")
	RootCmd.AddCommand(scpCmd)
}
