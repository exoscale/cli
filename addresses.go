package egoscale

import (
	"context"
	"fmt"
)

// Delete removes the resource
func (ipaddress *IPAddress) Delete(ctx context.Context, client *Client) error {
	if ipaddress.ID == "" {
		return fmt.Errorf("an IPAddress may only be deleted using ID")
	}

	return client.BooleanRequestWithContext(ctx, &DisassociateIPAddress{
		ID: ipaddress.ID,
	})
}

// ResourceType returns the type of the resource
func (*IPAddress) ResourceType() string {
	return "PublicIpAddress"
}

func (*AssociateIPAddress) name() string {
	return "associateIpAddress"
}

func (*AssociateIPAddress) description() string {
	return "Acquires and associates a public IP to an account."
}

func (*AssociateIPAddress) asyncResponse() interface{} {
	return new(IPAddress)
}

func (*DisassociateIPAddress) name() string {
	return "disassociateIpAddress"
}

func (*DisassociateIPAddress) description() string {
	return "Disassociates an ip address from the account."
}

func (*DisassociateIPAddress) asyncResponse() interface{} {
	return new(booleanResponse)
}

func (*UpdateIPAddress) name() string {
	return "updateIpAddress"
}

func (*UpdateIPAddress) description() string {
	return "Updates an ip address"
}

func (*UpdateIPAddress) asyncResponse() interface{} {
	return new(IPAddress)
}

func (*ListPublicIPAddresses) name() string {
	return "listPublicIpAddresses"
}

func (*ListPublicIPAddresses) description() string {
	return "Lists all public ip addresses"
}

func (*ListPublicIPAddresses) response() interface{} {
	return new(ListPublicIPAddressesResponse)
}

// ListRequest builds the ListAdresses request
func (ipaddress *IPAddress) ListRequest() (ListCommand, error) {
	req := &ListPublicIPAddresses{
		Account:             ipaddress.Account,
		AssociatedNetworkID: ipaddress.AssociatedNetworkID,
		DomainID:            ipaddress.DomainID,
		ForDisplay:          &ipaddress.ForDisplay,
		ForVirtualNetwork:   &ipaddress.ForVirtualNetwork,
		ID:                  ipaddress.ID,
		IPAddress:           ipaddress.IPAddress,
		IsElastic:           &ipaddress.IsElastic,
		IsSourceNat:         &ipaddress.IsSourceNat,
		PhysicalNetworkID:   ipaddress.PhysicalNetworkID,
		VlanID:              ipaddress.VlanID,
		ZoneID:              ipaddress.ZoneID,
	}

	return req, nil
}

// SetPage sets the current page
func (ls *ListPublicIPAddresses) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListPublicIPAddresses) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

func (*ListPublicIPAddresses) each(resp interface{}, callback IterateItemFunc) {
	ips, ok := resp.(*ListPublicIPAddressesResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type. ListPublicIPAddressesResponse expected, got %T", resp))
		return
	}

	for i := range ips.PublicIPAddress {
		if !callback(&ips.PublicIPAddress[i], nil) {
			break
		}
	}
}
