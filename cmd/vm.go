package cmd

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/egoscale"
)

const (
	defaultServiceOffering = "medium"
	maxUserDataLength      = 32768
)

var vmCmd = &cobra.Command{
	Use:   "vm",
	Short: "Compute instances management",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		fmt.Fprintln(os.Stderr,
			`**********************************************************************
The "exo vm" commands are deprecated and will be removed in a future
version, please use "exo compute instance" replacement commands.
**********************************************************************`)
		time.Sleep(3 * time.Second)
	},
	Hidden: true,
}

func getVirtualMachineByNameOrID(name string) (*egoscale.VirtualMachine, error) {
	vmQuery := egoscale.VirtualMachine{}
	id, err := egoscale.ParseUUID(name)
	if err != nil {
		vmQuery.Name = name
	} else {
		vmQuery.ID = id
	}

	vms, err := globalstate.EgoscaleClient.ListWithContext(gContext, vmQuery)
	if err != nil {
		return nil, err
	}

	var vm *egoscale.VirtualMachine
	switch len(vms) {
	case 0:
		return nil, fmt.Errorf("Compute instance %q not found", name) // nolint
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

		return nil, errors.New("multiple Compute instances found, specify an ID instead")
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
		if err := ioutil.WriteFile(filePath, []byte(keyPairs.PrivateKey), 0o600); err != nil {
			log.Fatalf("SSH private key could not be written: %s", err)
		}
	}
}

func getUserDataFromFile(path string, compress bool) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	userData, err := encodeUserData(data, compress)
	if err != nil {
		return "", err
	}

	if len(userData) >= maxUserDataLength {
		return "", fmt.Errorf("user-data maximum allowed length is %d bytes", maxUserDataLength)
	}

	return userData, nil
}

func encodeUserData(data []byte, compress bool) (string, error) {
	if compress {
		b := new(bytes.Buffer)
		gz := gzip.NewWriter(b)

		if _, err := gz.Write(data); err != nil {
			return "", err
		}
		if err := gz.Flush(); err != nil {
			return "", err
		}
		if err := gz.Close(); err != nil {
			return "", err
		}

		data = b.Bytes()
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

func decodeUserData(data string) (string, error) {
	base64Decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	gz, err := gzip.NewReader(bytes.NewReader(base64Decoded))
	if err != nil {
		// User data are not compressed, returning as-is.
		if errors.Is(err, gzip.ErrHeader) {
			return string(base64Decoded), nil
		}

		return "", err
	}
	defer gz.Close()

	userData, err := ioutil.ReadAll(gz)
	if err != nil {
		return "", err
	}

	return string(userData), nil
}

func init() {
	RootCmd.AddCommand(vmCmd)
}
