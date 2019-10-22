package egoscale

// InstancePoolState reprsents a state of an instance pool
type InstancePoolState string

const (
	// InstancePoolCreating creating state
	InstancePoolCreating InstancePoolState = "creating"
	// InstancePoolRunning running state
	InstancePoolRunning InstancePoolState = "running"
	// InstancePoolDestroying destroying state
	InstancePoolDestroying InstancePoolState = "destroying"
	// InstancePoolScalingUp scaling up state
	InstancePoolScalingUp InstancePoolState = "scaling-up"
	// InstancePoolScalingDown scaling down state
	InstancePoolScalingDown InstancePoolState = "scaling-down"
)

// InstancePool represents an instrance pool
type InstancePool struct {
	ID                *UUID             `json:"id"`
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	ServiceOfferingID *UUID             `json:"serviceofferingid"`
	TemplateID        *UUID             `json:"templateid"`
	ZoneID            *UUID             `json:"zoneid"`
	SecurityGroupIDs  []UUID            `json:"securitygroupids"`
	NetworkIDs        []UUID            `json:"networkids"`
	KeyPair           string            `json:"keypair"`
	UserData          string            `json:"userdata"`
	Size              int               `json:"size"`
	RootDiskSize      int               `json:"rootdisksize"`
	State             InstancePoolState `json:"state"`
	VirtualMachines   []VirtualMachine  `json:"virtualmachines"`
}

// CreateInstancePool create an instance pool
type CreateInstancePool struct {
	Name              string `json:"name"`
	Description       string `json:"description,omitempty"`
	ServiceOfferingID *UUID  `json:"serviceofferingid"`
	TemplateID        *UUID  `json:"templateid"`
	ZoneID            *UUID  `json:"zoneid"`
	SecurityGroupIDs  []UUID `json:"securitygroupids,omitempty"`
	NetworkIDs        []UUID `json:"networkids,omitempty"`
	KeyPair           string `json:"keypair,omitempty"`
	UserData          string `json:"userdata,omitempty"`
	Size              int    `json:"size"`
	RootDiskSize      int    `json:"rootdisksize"`
	_                 bool   `name:"createInstancePool" description:"Creates an Instance Pool with the provided parameters"`
}

// CreateInstancePoolResponse instance pool create response
type CreateInstancePoolResponse struct {
	ID                *UUID             `json:"id"`
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	ServiceOfferingID *UUID             `json:"serviceofferingid"`
	TemplateID        *UUID             `json:"templateid"`
	ZoneID            *UUID             `json:"zoneid"`
	SecurityGroupIDs  []UUID            `json:"securitygroupids"`
	NetworkIDs        []UUID            `json:"networkids"`
	KeyPair           string            `json:"keypair"`
	UserData          string            `json:"userdata"`
	Size              int64             `json:"size"`
	RootDiskSize      int               `json:"rootdisksize"`
	State             InstancePoolState `json:"state"`
}

// Response returns the struct to unmarshal
func (CreateInstancePool) Response() interface{} {
	return new(CreateInstancePoolResponse)
}

// UpdateInstancePool update an instance pool
type UpdateInstancePool struct {
	ID          *UUID  `json:"id"`
	ZoneID      *UUID  `json:"zoneid"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	TemplateID  *UUID  `json:"templateid,omitempty"`
	UserData    string `json:"userdata,omitempty"`
	_           bool   `name:"updateInstancePool" description:""`
}

// UpdateInstancePoolResponse update instance pool response
type UpdateInstancePoolResponse struct {
	Success bool `json:"success"`
}

// Response returns the struct to unmarshal
func (UpdateInstancePool) Response() interface{} {
	return new(UpdateInstancePoolResponse)
}

// ScaleInstancePool scale an instance pool
type ScaleInstancePool struct {
	ID     *UUID `json:"id"`
	ZoneID *UUID `json:"zoneid"`
	Size   int   `json:"size"`
	_      bool  `name:"scaleInstancePool" description:""`
}

// ScaleInstancePoolResponse scale instance pool response
type ScaleInstancePoolResponse struct {
	Success bool `json:"success"`
}

// Response returns the struct to unmarshal
func (ScaleInstancePool) Response() interface{} {
	return new(ScaleInstancePoolResponse)
}

// DestroyInstancePool destroy an instance pool
type DestroyInstancePool struct {
	ID     *UUID `json:"id"`
	ZoneID *UUID `json:"zoneid"`
	_      bool  `name:"destroyInstancePool" description:""`
}

// DestroyInstancePoolResponse destroy instance pool response
type DestroyInstancePoolResponse struct {
	Success bool `json:"success"`
}

// Response returns the struct to unmarshal
func (DestroyInstancePool) Response() interface{} {
	return new(DestroyInstancePoolResponse)
}

// GetInstancePool get an instance pool
type GetInstancePool struct {
	ID     *UUID `json:"id"`
	ZoneID *UUID `json:"zoneid"`
	_      bool  `name:"getInstancePool" description:""`
}

// GetInstancePoolResponse get instance pool response
type GetInstancePoolResponse struct {
	Count         int
	InstancePools []InstancePool `json:"instancepool"`
}

// Response returns the struct to unmarshal
func (GetInstancePool) Response() interface{} {
	return new(GetInstancePoolResponse)
}

// ListInstancePools list instance pool
type ListInstancePools struct {
	ZoneID *UUID `json:"zoneid"`
	_      bool  `name:"listInstancePools" description:""`
}

// ListInstancePoolsResponse list instance pool response
type ListInstancePoolsResponse struct {
	Count         int
	InstancePools []InstancePool `json:"instancepool"`
}

// Response returns the struct to unmarshal
func (ListInstancePools) Response() interface{} {
	return new(ListInstancePoolsResponse)
}
