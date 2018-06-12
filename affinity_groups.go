package egoscale

import (
	"context"
	"fmt"
	"net/url"
)

// ListRequest builds the ListAffinityGroups request
func (ag *AffinityGroup) ListRequest() (ListCommand, error) {
	return &ListAffinityGroups{
		ID:   ag.ID,
		Name: ag.Name,
	}, nil
}

// Delete removes the given Affinity Group
func (ag *AffinityGroup) Delete(ctx context.Context, client *Client) error {
	if ag.ID == "" && ag.Name == "" {
		return fmt.Errorf("an Affinity Group may only be deleted using ID or Name")
	}

	req := &DeleteAffinityGroup{
		Account:  ag.Account,
		DomainID: ag.DomainID,
	}

	if ag.ID != "" {
		req.ID = ag.ID
	} else {
		req.Name = ag.Name
	}

	return client.BooleanRequestWithContext(ctx, req)
}

func (*CreateAffinityGroup) name() string {
	return "createAffinityGroup"
}

func (*CreateAffinityGroup) description() string {
	return "Creates an affinity/anti-affinity group"
}

func (*CreateAffinityGroup) asyncResponse() interface{} {
	return new(AffinityGroup)
}

func (*UpdateVMAffinityGroup) name() string {
	return "updateVMAffinityGroup"
}

func (*UpdateVMAffinityGroup) description() string {
	return "Updates the affinity/anti-affinity group associations of a virtual machine. The VM has to be stopped and restarted for the new properties to take effect."
}

func (*UpdateVMAffinityGroup) asyncResponse() interface{} {
	return new(VirtualMachine)
}

func (req *UpdateVMAffinityGroup) onBeforeSend(params *url.Values) error {
	// Either AffinityGroupIDs or AffinityGroupNames must be set
	if len(req.AffinityGroupIDs) == 0 && len(req.AffinityGroupNames) == 0 {
		params.Set("affinitygroupids", "")
	}
	return nil
}

func (*DeleteAffinityGroup) name() string {
	return "deleteAffinityGroup"
}

func (*DeleteAffinityGroup) description() string {
	return "Deletes affinity group"
}

func (*DeleteAffinityGroup) asyncResponse() interface{} {
	return new(booleanResponse)
}

func (*ListAffinityGroups) name() string {
	return "listAffinityGroups"
}

func (*ListAffinityGroups) description() string {
	return "Lists affinity groups"
}

func (*ListAffinityGroups) response() interface{} {
	return new(ListAffinityGroupsResponse)
}

func (*ListAffinityGroupTypes) name() string {
	return "listAffinityGroupTypes"
}

func (*ListAffinityGroupTypes) description() string {
	return "Lists affinity group types available"
}

// SetPage sets the current page
func (ls *ListAffinityGroups) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListAffinityGroups) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

func (*ListAffinityGroups) each(resp interface{}, callback IterateItemFunc) {
	vms, ok := resp.(*ListAffinityGroupsResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type. ListAffinityGroupsResponse expected, got %T", resp))
		return
	}

	for i := range vms.AffinityGroup {
		if !callback(&vms.AffinityGroup[i], nil) {
			break
		}
	}
}

func (*ListAffinityGroupTypes) response() interface{} {
	return new(ListAffinityGroupTypesResponse)
}
