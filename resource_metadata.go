package egoscale

func (*ListResourceDetails) name() string {
	return "listResourceDetails"
}

func (*ListResourceDetails) description() string {
	return "List resource detail(s)"
}

func (*ListResourceDetails) response() interface{} {
	return new(ListResourceDetailsResponse)
}
