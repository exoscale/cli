package egoscale

func (*CreateTags) name() string {
	return "createTags"
}

func (*CreateTags) description() string {
	return "Creates resource tag(s)"
}

func (*CreateTags) asyncResponse() interface{} {
	return new(booleanResponse)
}

func (*DeleteTags) name() string {
	return "deleteTags"
}

func (*DeleteTags) description() string {
	return "Deleting resource tag(s)"
}

func (*DeleteTags) asyncResponse() interface{} {
	return new(booleanResponse)
}

func (*ListTags) name() string {
	return "listTags"
}

func (*ListTags) description() string {
	return "List resource tag(s)"
}

func (*ListTags) response() interface{} {
	return new(ListTagsResponse)
}
