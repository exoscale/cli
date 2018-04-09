package egoscale

import (
	"context"
	"fmt"

	"github.com/jinzhu/copier"
)

// Get populates the given SSHKeyPair
func (ssh *SSHKeyPair) Get(ctx context.Context, client *Client) error {
	resp, err := client.RequestWithContext(ctx, &ListSSHKeyPairs{
		Account:     ssh.Account,
		DomainID:    ssh.DomainID,
		Name:        ssh.Name,
		Fingerprint: ssh.Fingerprint,
		ProjectID:   ssh.ProjectID,
	})

	if err != nil {
		return err
	}

	sshs := resp.(*ListSSHKeyPairsResponse)
	count := len(sshs.SSHKeyPair)
	if count == 0 {
		return &ErrorResponse{
			ErrorCode: ParamError,
			ErrorText: fmt.Sprintf("SSHKeyPair not found"),
		}
	} else if count > 1 {
		return fmt.Errorf("More than one SSHKeyPair was found")
	}

	return copier.Copy(ssh, sshs.SSHKeyPair[0])
}

// Delete removes the given SSH key, by Name
func (ssh *SSHKeyPair) Delete(ctx context.Context, client *Client) error {
	if ssh.Name == "" {
		return fmt.Errorf("An SSH Key Pair may only be deleted using Name")
	}

	return client.BooleanRequestWithContext(ctx, &DeleteSSHKeyPair{
		Name:      ssh.Name,
		Account:   ssh.Account,
		DomainID:  ssh.DomainID,
		ProjectID: ssh.ProjectID,
	})
}

// APIName returns the CloudStack API command name
func (*CreateSSHKeyPair) APIName() string {
	return "createSSHKeyPair"
}

func (*CreateSSHKeyPair) response() interface{} {
	return new(CreateSSHKeyPairResponse)
}

// APIName returns the CloudStack API command name
func (*DeleteSSHKeyPair) APIName() string {
	return "deleteSSHKeyPair"
}

func (*DeleteSSHKeyPair) response() interface{} {
	return new(booleanSyncResponse)
}

// APIName returns the CloudStack API command name
func (*RegisterSSHKeyPair) APIName() string {
	return "registerSSHKeyPair"
}

func (*RegisterSSHKeyPair) response() interface{} {
	return new(RegisterSSHKeyPairResponse)
}

// APIName returns the CloudStack API command name
func (*ListSSHKeyPairs) APIName() string {
	return "listSSHKeyPairs"
}

func (*ListSSHKeyPairs) response() interface{} {
	return new(ListSSHKeyPairsResponse)
}

// APIName returns the CloudStack API command name
func (*ResetSSHKeyForVirtualMachine) APIName() string {
	return "resetSSHKeyForVirtualMachine"
}

func (*ResetSSHKeyForVirtualMachine) asyncResponse() interface{} {
	return new(ResetSSHKeyForVirtualMachineResponse)
}
