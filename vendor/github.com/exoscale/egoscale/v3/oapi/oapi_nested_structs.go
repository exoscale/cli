package oapi

import (
	"time"

	openapi_types "github.com/deepmap/oapi-codegen/pkg/types"
)

type InstancesListElement struct {
	CreatedAt          *time.Time
	Id                 *openapi_types.UUID
	InstanceType       *InstanceType
	Ipv6Address        *string
	Labels             *Labels
	Manager            *Manager
	Name               *string
	PrivateNetworks    *[]PrivateNetwork
	PublicIp           *string
	PublicIpAssignment *PublicIpAssignment
	SecurityGroups     *[]SecurityGroup
	SshKey             *SshKey
	SshKeys            *[]SshKey
	State              *InstanceState
	Template           *Template
}

func FromListInstancesResponse(r *ListInstancesResponse) []InstancesListElement {
	if r.JSON200.Instances == nil {
		return nil
	}

	t := *r.JSON200.Instances

	ret := make([]InstancesListElement, 0, len(t))

	for _, v := range t {
		ret = append(ret, InstancesListElement{
			CreatedAt:          v.CreatedAt,
			Id:                 v.Id,
			InstanceType:       v.InstanceType,
			Ipv6Address:        v.Ipv6Address,
			Labels:             v.Labels,
			Manager:            v.Manager,
			Name:               v.Name,
			PrivateNetworks:    v.PrivateNetworks,
			PublicIp:           v.PublicIp,
			PublicIpAssignment: v.PublicIpAssignment,
			SecurityGroups:     v.SecurityGroups,
			SshKey:             v.SshKey,
			SshKeys:            v.SshKeys,
			State:              v.State,
			Template:           v.Template,
		})
	}

	return ret
}

type DBaaSIntegrationSettings struct {
	AdditionalProperties *bool
	Properties           *map[string]interface{}
	Title                *string
	Type                 *string
}

func FromListDbaasIntegrationSettingsResponse(r *ListDbaasIntegrationSettingsResponse) *DBaaSIntegrationSettings {
	t := r.JSON200.Settings

	return &DBaaSIntegrationSettings{
		AdditionalProperties: t.AdditionalProperties,
		Properties:           t.Properties,
		Title:                t.Title,
		Type:                 t.Type,
	}
}
