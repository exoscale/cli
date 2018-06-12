package egoscale

func (*ListEvents) name() string {
	return "listEvents"
}

func (*ListEvents) description() string {
	return "A command to list events."
}

func (*ListEvents) response() interface{} {
	return new(ListEventsResponse)
}

func (*ListEventTypes) name() string {
	return "listEventTypes"
}

func (*ListEventTypes) description() string {
	return "List Event Types"
}

func (*ListEventTypes) response() interface{} {
	return new(ListEventTypesResponse)
}
