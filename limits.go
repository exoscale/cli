package egoscale

func (*ListResourceLimits) name() string {
	return "listResourceLimits"
}

func (*ListResourceLimits) description() string {
	return "Lists esource limits."
}

func (*ListResourceLimits) response() interface{} {
	return new(ListResourceLimitsResponse)
}

func (*UpdateResourceLimit) name() string {
	return "updateResourceLimit"
}

func (*UpdateResourceLimit) description() string {
	return "Updates resource limits for an account or domain."
}

func (*UpdateResourceLimit) response() interface{} {
	return new(UpdateResourceLimitResponse)
}

func (*GetAPILimit) name() string {
	return "getAPILimit"
}

func (*GetAPILimit) description() string {
	return "Get API limit count for the caller"
}

func (*GetAPILimit) response() interface{} {
	return new(GetAPILimitResponse)
}
