package cmd

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// sshCmd represents the ssh command
var sshCmd = &cobra.Command{
	Use:   "ssh <vm name | id>",
	Short: "SSH into a virtual machine instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		ipv6, err := cmd.Flags().GetBool("ipv6")
		if err != nil {
			return err
		}

		isInfos, err := cmd.Flags().GetBool("infos")
		if err != nil {
			return err
		}

		isConnectionSTR, err := cmd.Flags().GetBool("print")
		if err != nil {
			return err
		}

		infos, err := getSSHInfos(args[0], ipv6)
		if err != nil {
			return err
		}

		if isConnectionSTR {
			return printSSHConnectSTR(infos)
		}

		if isInfos {
			printSSHInfos(infos)
			return nil
		}
		return connectSSH(infos)
	},
}

type sshInfos struct {
	sshKeys  string
	userName string
	ip       *net.IP
	vmName   string
	vmID     string
}

func getSSHInfos(name string, isIpv6 bool) (*sshInfos, error) {
	vm, err := getVMWithNameOrID(cs, name)
	if err != nil {
		return nil, err
	}

	sshKeyPath := path.Join(gConfigFolder, "instances", vm.ID, "id_rsa")

	if _, err := os.Stat(sshKeyPath); os.IsNotExist(err) {
		sshKeyPath = "Default ssh keypair not found"
	}

	nic := vm.DefaultNic()
	if nic == nil {
		return nil, fmt.Errorf("No default NIC on this instance")
	}

	vmIP := vm.IP()

	if isIpv6 {
		if nic.IP6Address != nil {
			vmIP = &nic.IP6Address
		} else {
			return nil, fmt.Errorf("IPv6 not found on this virtual machine ID %q", vm.ID)
		}
	}

	template := &egoscale.Template{ID: vm.TemplateID, IsFeatured: true, ZoneID: "1"}

	if err := cs.Get(template); err != nil {
		return nil, err
	}

	tempUser, ok := template.Details["username"]
	if !ok {
		return nil, fmt.Errorf("User name not found in template id %q", template.ID)
	}

	return &sshInfos{
		sshKeys:  sshKeyPath,
		userName: tempUser,
		ip:       vmIP,
		vmName:   vm.Name,
		vmID:     vm.ID,
	}, nil

}

func printSSHConnectSTR(infos *sshInfos) error {

	if _, err := os.Stat(infos.sshKeys); os.IsNotExist(err) {
		return fmt.Errorf("Default ssh keypair not found")
	}

	fmt.Printf("ssh -i %s %s@%s\n", infos.sshKeys, infos.userName, infos.ip)

	return nil
}

func printSSHInfos(infos *sshInfos) {
	println("Virtual machine name", infos.vmName, "with ID", infos.vmID)
	println(" - sshkey path:", infos.sshKeys)
	println(" - username@IPadress:", infos.userName+"@"+infos.ip.String())
}

func connectSSH(cred *sshInfos) error {

	args := []string{
		"-i",
		cred.sshKeys,
		cred.userName + "@" + cred.ip.String(),
	}

	cmd := exec.Command("ssh", args...)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

func init() {
	sshCmd.Flags().BoolP("infos", "i", false, "infos show ssh connection informations")
	sshCmd.Flags().BoolP("print", "p", false, "Print SSH connection command")
	sshCmd.Flags().BoolP("ipv6", "6", false, "Using ipv6 for SSH")
	RootCmd.AddCommand(sshCmd)
}
