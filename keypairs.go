package egoscale

import (
	"context"
	"fmt"
)

// Delete removes the given SSH key, by Name
func (ssh *SSHKeyPair) Delete(ctx context.Context, client *Client) error {
	if ssh.Name == "" {
		return fmt.Errorf("an SSH Key Pair may only be deleted using Name")
	}

	return client.BooleanRequestWithContext(ctx, &DeleteSSHKeyPair{
		Name:     ssh.Name,
		Account:  ssh.Account,
		DomainID: ssh.DomainID,
	})
}

// ListRequest builds the ListSSHKeyPairs request
func (ssh *SSHKeyPair) ListRequest() (ListCommand, error) {
	req := &ListSSHKeyPairs{
		Account:     ssh.Account,
		DomainID:    ssh.DomainID,
		Fingerprint: ssh.Fingerprint,
		Name:        ssh.Name,
	}

	return req, nil
}

func (*CreateSSHKeyPair) name() string {
	return "createSSHKeyPair"
}

func (*CreateSSHKeyPair) description() string {
	return "Create a new keypair and returns the private key"
}

func (*CreateSSHKeyPair) response() interface{} {
	return new(SSHKeyPair)
}

func (*DeleteSSHKeyPair) name() string {
	return "deleteSSHKeyPair"
}

func (*DeleteSSHKeyPair) description() string {
	return "Deletes a keypair by name"
}

func (*DeleteSSHKeyPair) response() interface{} {
	return new(booleanResponse)
}

func (*RegisterSSHKeyPair) name() string {
	return "registerSSHKeyPair"
}

func (*RegisterSSHKeyPair) description() string {
	return "Register a public key in a keypair under a certain name"
}

func (*RegisterSSHKeyPair) response() interface{} {
	return new(SSHKeyPair)
}

func (*ListSSHKeyPairs) name() string {
	return "listSSHKeyPairs"
}

func (*ListSSHKeyPairs) description() string {
	return "List registered keypairs"
}

func (*ListSSHKeyPairs) response() interface{} {
	return new(ListSSHKeyPairsResponse)
}

func (*ListSSHKeyPairs) each(resp interface{}, callback IterateItemFunc) {
	sshs, ok := resp.(*ListSSHKeyPairsResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type. ListSSHKeyPairsResponse expected, got %T", resp))
		return
	}

	for i := range sshs.SSHKeyPair {
		if !callback(&sshs.SSHKeyPair[i], nil) {
			break
		}
	}
}

// SetPage sets the current page
func (ls *ListSSHKeyPairs) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListSSHKeyPairs) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

func (*ResetSSHKeyForVirtualMachine) name() string {
	return "resetSSHKeyForVirtualMachine"
}

func (*ResetSSHKeyForVirtualMachine) description() string {
	return `Resets the SSH Key for virtual machine. The virtual machine must be in a "Stopped" state.`
}

func (*ResetSSHKeyForVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}
