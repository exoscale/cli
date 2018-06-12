package egoscale

func (*CreateInstanceGroup) name() string {
	return "createInstanceGroup"
}

func (*CreateInstanceGroup) description() string {
	return "Creates a vm group"
}

func (*CreateInstanceGroup) response() interface{} {
	return new(InstanceGroup)
}

func (*UpdateInstanceGroup) name() string {
	return "updateInstanceGroup"
}

func (*UpdateInstanceGroup) description() string {
	return "Updates a vm group"
}

func (*UpdateInstanceGroup) response() interface{} {
	return new(InstanceGroup)
}

func (*DeleteInstanceGroup) name() string {
	return "deleteInstanceGroup"
}

func (*DeleteInstanceGroup) description() string {
	return "Deletes a vm group"
}

func (*DeleteInstanceGroup) response() interface{} {
	return new(booleanResponse)
}

func (*ListInstanceGroups) name() string {
	return "listInstanceGroups"
}

func (*ListInstanceGroups) description() string {
	return "Lists vm groups"
}

func (*ListInstanceGroups) response() interface{} {
	return new(ListInstanceGroupsResponse)
}
