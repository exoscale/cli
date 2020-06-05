package cmd

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

const maxUserDataLength = 32768

// vmCmd represents the vm command
var vmCmd = &cobra.Command{
	Use:   "vm",
	Short: "Virtual machines management",
}

func getVirtualMachineByNameOrID(name string) (*egoscale.VirtualMachine, error) {
	vmQuery := egoscale.VirtualMachine{}
	id, err := egoscale.ParseUUID(name)
	if err != nil {
		vmQuery.Name = name
	} else {
		vmQuery.ID = id
	}

	vms, err := cs.ListWithContext(gContext, vmQuery)
	if err != nil {
		return nil, err
	}

	var vm *egoscale.VirtualMachine
	switch len(vms) {
	case 0:
		return nil, fmt.Errorf("no VMs has been found")
	case 1:
		vm = vms[0].(*egoscale.VirtualMachine)
	default:
		names := []string{}
		for _, i := range vms {
			v := i.(*egoscale.VirtualMachine)
			if v.Name != vmQuery.Name {
				continue
			}

			vm = v
			names = append(names, fmt.Sprintf("\t%s\t%s\t%s", v.ID.String(), v.ZoneName, v.IP()))
		}

		if len(names) == 1 {
			break
		}

		fmt.Println("More than one VM has been found, use the ID instead:")
		for _, name := range names {
			fmt.Println(name)
		}
		return nil, fmt.Errorf("abort vm name %q is ambiguous", vmQuery.Name)
	}

	return vm, nil
}

func getKeyPairPath(vmID string) string {
	return path.Join(gConfigFolder, "instances", vmID, "id_rsa")
}

func saveKeyPair(keyPairs *egoscale.SSHKeyPair, vmID egoscale.UUID) {
	filePath := getKeyPairPath(vmID.String())
	folder := path.Dir(filePath)

	if _, err := os.Stat(folder); os.IsNotExist(err) {
		if err := os.MkdirAll(folder, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := ioutil.WriteFile(filePath, []byte(keyPairs.PrivateKey), 0600); err != nil {
			log.Fatalf("SSH private key could not be written: %s", err)
		}
	}
}

func deleteKeyPair(vmID egoscale.UUID) error {
	folder := getKeyPairPath(vmID.String())

	if _, err := os.Stat(folder); !os.IsNotExist(err) {
		if err := os.RemoveAll(folder); err != nil {
			return fmt.Errorf("the SSH private key could not be deleted: %s", err)
		}
	}

	return nil
}

func getUserDataFromFile(path string) (string, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	userData := base64.StdEncoding.EncodeToString(buf)

	if len(userData) >= maxUserDataLength {
		return "", fmt.Errorf("user-data maximum allowed length is %d bytes", maxUserDataLength)
	}

	return userData, nil
}

func init() {
	RootCmd.AddCommand(vmCmd)
}
