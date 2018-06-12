package egoscale

func (*ListHosts) name() string {
	return "listHosts"
}

func (*ListHosts) description() string {
	return "List hosts."
}

func (*ListHosts) response() interface{} {
	return new(ListHostsResponse)
}
