package egoscale

// ResourceType returns the type of the resource
func (*Snapshot) ResourceType() string {
	return "Snapshot"
}

func (*CreateSnapshot) name() string {
	return "createSnapshot"
}

func (*CreateSnapshot) description() string {
	return "Creates an instant snapshot of a volume."
}

func (*CreateSnapshot) asyncResponse() interface{} {
	return new(Snapshot)
}

func (*ListSnapshots) name() string {
	return "listSnapshots"
}

func (*ListSnapshots) description() string {
	return "Lists all available snapshots for the account."
}

func (*ListSnapshots) response() interface{} {
	return new(ListSnapshotsResponse)
}

func (*DeleteSnapshot) name() string {
	return "deleteSnapshot"
}

func (*DeleteSnapshot) description() string {
	return "Deletes a snapshot of a disk volume."
}

func (*DeleteSnapshot) asyncResponse() interface{} {
	return new(booleanResponse)
}

func (*RevertSnapshot) name() string {
	return "revertSnapshot"
}

func (*RevertSnapshot) description() string {
	return "revert a volume snapshot."
}

func (*RevertSnapshot) asyncResponse() interface{} {
	return new(booleanResponse)
}
