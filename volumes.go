package egoscale

import (
	"fmt"
)

// ResourceType returns the type of the resource
func (*Volume) ResourceType() string {
	return "Volume"
}

// ListRequest builds the ListVolumes request
func (vol *Volume) ListRequest() (ListCommand, error) {
	req := &ListVolumes{
		Account:          vol.Account,
		DomainID:         vol.DomainID,
		Name:             vol.Name,
		Type:             vol.Type,
		VirtualMachineID: vol.VirtualMachineID,
		ZoneID:           vol.ZoneID,
	}

	return req, nil
}

// name returns the CloudStack API command name
func (*ResizeVolume) name() string {
	return "resizeVolume"
}

func (*ResizeVolume) asyncResponse() interface{} {
	return new(Volume)
}

// name returns the CloudStack API command name
func (*ListVolumes) name() string {
	return "listVolumes"
}

func (*ListVolumes) response() interface{} {
	return new(ListVolumesResponse)
}

// SetPage sets the current page
func (ls *ListVolumes) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListVolumes) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

func (*ListVolumes) each(resp interface{}, callback IterateItemFunc) {
	volumes, ok := resp.(*ListVolumesResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type. ListVolumesResponse expected, got %T", resp))
		return
	}

	for i := range volumes.Volume {
		if !callback(&volumes.Volume[i], nil) {
			break
		}
	}
}
