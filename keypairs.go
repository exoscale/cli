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

// name returns the CloudStack API command name
func (*CreateSSHKeyPair) name() string {
	return "createSSHKeyPair"
}

func (*CreateSSHKeyPair) response() interface{} {
	return new(SSHKeyPair)
}

// name returns the CloudStack API command name
func (*DeleteSSHKeyPair) name() string {
	return "deleteSSHKeyPair"
}

func (*DeleteSSHKeyPair) response() interface{} {
	return new(booleanResponse)
}

// name returns the CloudStack API command name
func (*RegisterSSHKeyPair) name() string {
	return "registerSSHKeyPair"
}

func (*RegisterSSHKeyPair) response() interface{} {
	return new(SSHKeyPair)
}

// name returns the CloudStack API command name
func (*ListSSHKeyPairs) name() string {
	return "listSSHKeyPairs"
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

// name returns the CloudStack API command name
func (*ResetSSHKeyForVirtualMachine) name() string {
	return "resetSSHKeyForVirtualMachine"
}

func (*ResetSSHKeyForVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}
