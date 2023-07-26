package ssh

import (
	"path"

	"github.com/exoscale/cli/pkg/globalstate"
)

func GetInstanceSSHKeyPath(instanceID string) string {
	return path.Join(globalstate.ConfigFolder, "instances", instanceID, "id_rsa")
}
