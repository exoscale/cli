package cmd

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/exoscale/egoscale"
	"github.com/kballard/go-shellquote"
	"github.com/spf13/cobra"
)

var sshCmd = &cobra.Command{
	Use:               "ssh INSTANCE-NAME|ID",
	Short:             "SSH into a Compute instance",
	ValidArgsFunction: completeVMNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		ipv6, err := cmd.Flags().GetBool("ipv6")
		if err != nil {
			return err
		}

		printInfo, err := cmd.Flags().GetBool("info")
		if err != nil {
			return err
		}

		printCmd, err := cmd.Flags().GetBool("print")
		if err != nil {
			return err
		}

		sshInfo, err := getSSHInfo(args[0], ipv6)
		if err != nil {
			return err
		}

		sshOpts, err := cmd.Flags().GetString("ssh-options")
		if err != nil {
			return err
		}
		sshInfo.opts = sshOpts

		if printInfo {
			printSSHInfo(sshInfo)
			return nil
		}

		sshCmd := buildSSHCommand(sshInfo)

		if printCmd {
			fmt.Println(strings.Join(sshCmd, " "))
			return nil
		}

		return connectSSH(sshCmd[1:])
	},
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		fmt.Fprintln(os.Stderr,
			`**********************************************************************
The "exo ssh" command is deprecated and will be removed in a future
version, please use "exo compute instance ssh" replacement command.
**********************************************************************`)
		time.Sleep(3 * time.Second)
	},
	Hidden: true,
}

type sshInfo struct {
	sshKeys  string
	username string
	opts     string
	ip       net.IP
	vmName   string
	vmID     *egoscale.UUID
}

func getSSHInfo(name string, isIpv6 bool) (*sshInfo, error) {
	var info sshInfo

	vm, err := getVirtualMachineByNameOrID(name)
	if err != nil {
		return nil, err
	}
	info.vmID = vm.ID
	info.vmName = vm.Name

	info.sshKeys = getKeyPairPath(vm.ID.String())

	nic := vm.DefaultNic()
	if nic == nil {
		return nil, fmt.Errorf("this instance %q has no default NIC", vm.ID)
	}

	info.ip = *vm.IP()
	if isIpv6 {
		if nic.IP6Address == nil {
			return nil, fmt.Errorf("missing IPv6 address on the instance %q", vm.ID)
		}
		info.ip = nic.IP6Address
	}

	if info.ip == nil {
		return nil, fmt.Errorf("no valid IP address found")
	}

	query := &egoscale.Template{
		ID:         vm.TemplateID,
		IsFeatured: true,
		ZoneID:     vm.ZoneID,
	}

	resp, err := cs.GetWithContext(gContext, query)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Compute instance template: %v", err)
	}

	template := resp.(*egoscale.Template)
	username, ok := template.Details["username"]
	if ok {
		info.username = username
	}

	return &info, nil
}

func buildSSHCommand(info *sshInfo) []string {
	cmd := []string{"ssh"}

	if _, err := os.Stat(info.sshKeys); err == nil {
		cmd = append(cmd, "-i", info.sshKeys)
	}

	if info.opts != "" {
		opts, err := shellquote.Split(info.opts)
		if err == nil {
			cmd = append(cmd, opts...)
		}
	}

	if info.username != "" {
		cmd = append(cmd, "-l", info.username)
	}

	cmd = append(cmd, info.ip.String())

	return cmd
}

func printSSHInfo(info *sshInfo) {
	fmt.Println("Host", info.vmName)
	fmt.Println("\tHostName", info.ip.String())

	if info.username != "" {
		fmt.Println("\tUser", info.username)
	}

	if _, err := os.Stat(info.sshKeys); err == nil {
		fmt.Println("\tIdentityFile", info.sshKeys)
	}
}

func connectSSH(args []string) error {
	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

func getInstanceSSHKeyPath(instanceID string) string {
	return path.Join(gConfigFolder, "instances", instanceID, "id_rsa")
}

func init() {
	sshCmd.Flags().BoolP("info", "i", false, "Print SSH config information")
	sshCmd.Flags().StringP("ssh-options", "o", "",
		"Additional options to pass to the `ssh` command (e.g. -o \"-l my-user -p 2222\"`)")
	sshCmd.Flags().BoolP("print", "p", false, "Print SSH command")
	sshCmd.Flags().BoolP("ipv6", "6", false, "Use IPv6 for SSH")
	RootCmd.AddCommand(sshCmd)
}
