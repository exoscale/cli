package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// vmCmd represents the vm command
var vmCmd = &cobra.Command{
	Use:   "vm",
	Short: "Virtual machines management",
}

func getVMWithNameOrID(name string) (*egoscale.VirtualMachine, error) {
	vm := &egoscale.VirtualMachine{}

	id, err := egoscale.ParseUUID(name)
	if err != nil {
		vm.Name = name
	} else {
		vm.ID = id
	}

	if err := cs.GetWithContext(gContext, vm); err != nil {
		return nil, err
	}

	return vm, nil
}

func getSecurityGroup(vm *egoscale.VirtualMachine) []string {
	sgs := []string{}
	for _, sgN := range vm.SecurityGroup {
		sgs = append(sgs, sgN.Name)
	}
	return sgs
}

func getKeyPairPath(vmID string) string {
	return path.Join(gConfigFolder, "instances", vmID)
}

func saveKeyPair(keyPairs *egoscale.SSHKeyPair, vmID egoscale.UUID) {
	folder := getKeyPairPath(vmID.String())

	if _, err := os.Stat(folder); os.IsNotExist(err) {
		if err := os.MkdirAll(folder, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}

	filePath := path.Join(folder, "id_rsa")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := ioutil.WriteFile(filePath, []byte(keyPairs.PrivateKey), 0600); err != nil {
			log.Fatalf("SSH private key could not be written: %s", err)
		}
	}
}

func deleteKeyPair(vmID egoscale.UUID) {
	folder := getKeyPairPath(vmID.String())

	if _, err := os.Stat(folder); !os.IsNotExist(err) {
		if err := os.RemoveAll(folder); err != nil {
			log.Fatalf("SSH private key could not be deleted: %s", err)
		}
	}
}

func init() {
	RootCmd.AddCommand(vmCmd)
}
