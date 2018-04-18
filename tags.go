package egoscale

// APIName returns the CloudStack API command name
func (*CreateTags) APIName() string {
	return "createTags"
}

func (*CreateTags) asyncResponse() interface{} {
	return new(booleanAsyncResponse)
}

// APIName returns the CloudStack API command name
func (*DeleteTags) APIName() string {
	return "deleteTags"
}

func (*DeleteTags) asyncResponse() interface{} {
	return new(booleanAsyncResponse)
}

// APIName returns the CloudStack API command name
func (*ListTags) APIName() string {
	return "listTags"
}

func (*ListTags) response() interface{} {
	return new(ListTagsResponse)
}
