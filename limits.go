package egoscale

// name returns the CloudStack API command name
func (*ListResourceLimits) name() string {
	return "listResourceLimits"
}

func (*ListResourceLimits) response() interface{} {
	return new(ListResourceLimitsResponse)
}
