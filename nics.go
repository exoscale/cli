package egoscale

import (
	"errors"
)

// ListRequest build a ListNics request from the given Nic
func (nic *Nic) ListRequest() (ListCommand, error) {
	if nic.VirtualMachineID == "" {
		return nil, errors.New("command ListNics requires the VirtualMachineID field to be set")
	}

	req := &ListNics{
		VirtualMachineID: nic.VirtualMachineID,
		NicID:            nic.ID,
		NetworkID:        nic.NetworkID,
	}

	return req, nil
}

func (*ListNics) name() string {
	return "listNics"
}

func (*ListNics) description() string {
	return "list the vm nics  IP to NIC"
}

func (*ListNics) response() interface{} {
	return new(ListNicsResponse)
}

// SetPage sets the current page
func (ls *ListNics) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListNics) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

func (*ListNics) each(resp interface{}, callback IterateItemFunc) {
	nics := resp.(*ListNicsResponse)
	for i := range nics.Nic {
		if !callback(&(nics.Nic[i]), nil) {
			break
		}
	}
}

func (*AddIPToNic) name() string {
	return "addIpToNic"
}

func (*AddIPToNic) description() string {
	return "Assigns secondary IP to NIC"
}

func (*AddIPToNic) asyncResponse() interface{} {
	return new(NicSecondaryIP)
}

func (*RemoveIPFromNic) name() string {
	return "removeIpFromNic"
}

func (*RemoveIPFromNic) description() string {
	return "Removes secondary IP from the NIC."
}

func (*RemoveIPFromNic) asyncResponse() interface{} {
	return new(booleanResponse)
}

func (*ActivateIP6) name() string {
	return "activateIp6"
}

func (*ActivateIP6) description() string {
	return "Activate the IPv6 on the VM's nic"
}

func (*ActivateIP6) asyncResponse() interface{} {
	return new(Nic)
}
