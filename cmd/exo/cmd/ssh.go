package cmd

import (
	"fmt"
	"log"
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
}

func sshCmdRun(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		sshCmd.Usage()
		return
	}

	isInfos, err := cmd.Flags().GetBool("infos")
	if err != nil {
		log.Fatal(err)
	}

	infos, err := getSSHInfos(args[0])
	if err != nil {
		log.Fatal(err)
	}
	if isInfos {
		printSSHInfos(infos)
		return
	}
	connectSSH(infos)
}

type sshInfos struct {
	sshKeys  string
	userName string
	ip       *net.IP
	vmName   string
	vmID     string
}

func getSSHInfos(name string) (*sshInfos, error) {
	vm, err := getVMWithNameOrID(cs, name)
	if err != nil {
		return nil, err
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
		sshKeys:  path.Join(configFolder, "instances", vm.ID, "id_rsa"),
		userName: tempUser,
		ip:       vm.IP(),
		vmName:   vm.Name,
		vmID:     vm.ID,
	}, nil

}

func printSSHInfos(infos *sshInfos) {
	println("Virtual machine name", infos.vmName, "with ID", infos.vmID)
	println(" - sshkey path:", infos.sshKeys)
	println(" - username@IPadress:", infos.userName+"@"+infos.ip.String())
}

func connectSSH(cred *sshInfos) {

	args := []string{
		"-i",
		cred.sshKeys,
		cred.userName + "@" + cred.ip.String(),
	}

	cmd := exec.Command("ssh", args...)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

}

func init() {
	sshCmd.Run = sshCmdRun
	sshCmd.Flags().BoolP("infos", "i", false, "infos show ssh connection informations")
	RootCmd.AddCommand(sshCmd)
}
