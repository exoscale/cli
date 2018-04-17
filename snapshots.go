package egoscale

// ResourceType returns the type of the resource
func (*Snapshot) ResourceType() string {
	return "Snapshot"
}

// APIName returns the CloudStack API command name
func (*CreateSnapshot) APIName() string {
	return "createSnapshot"
}

func (*CreateSnapshot) asyncResponse() interface{} {
	return new(CreateSnapshotResponse)
}

// CreateSnapshotResponse represents a freshly created snapshot
type CreateSnapshotResponse struct {
	Snapshot Snapshot `json:"snapshot"`
}

// APIName returns the CloudStack API command name
func (*ListSnapshots) APIName() string {
	return "listSnapshots"
}

func (*ListSnapshots) response() interface{} {
	return new(ListSnapshotsResponse)
}

// ListSnapshotsResponse represents a list of volume snapshots
type ListSnapshotsResponse struct {
	Count    int        `json:"count"`
	Snapshot []Snapshot `json:"snapshot"`
}

// APIName returns the CloudStack API command name
func (*DeleteSnapshot) APIName() string {
	return "deleteSnapshot"
}

func (*DeleteSnapshot) asyncResponse() interface{} {
	return new(booleanAsyncResponse)
}

// APIName returns the CloudStack API command name
func (*RevertSnapshot) APIName() string {
	return "revertSnapshot"
}

func (*RevertSnapshot) asyncResponse() interface{} {
	return new(booleanAsyncResponse)
}
