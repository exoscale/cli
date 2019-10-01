package egoscale

type InstancePoolState string

const (
	InstancePoolCreating    InstancePoolState = "creating"
	InstancePoolRunning     InstancePoolState = "running"
	InstancePoolDestroying  InstancePoolState = "destroying"
	InstancePoolScalingUP   InstancePoolState = "scaling-up"
	InstancePoolScalingDown InstancePoolState = "scaling-down"
)

type InstancePool struct {
	ID                *UUID             `json:"id"`
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	ServiceofferingID *UUID             `json:"serviceofferingid"`
	TemplateID        *UUID             `json:"templateid"`
	ZoneID            *UUID             `json:"zoneid"`
	AffinitygroupIDs  []UUID            `json:"affinitygroupids"`
	SecuritygroupIDs  []UUID            `json:"securitygroupids"`
	NetworkIDs        []UUID            `json:"networkids"`
	Keypair           string            `json:"keypair"`
	Userdata          string            `json:"userdata"`
	Size              int               `json:"size"`
	State             InstancePoolState `json:"state"`
	Virtualmachines   []VirtualMachine  `json:"virtualmachines"`
}

type CreateInstancePool struct {
	Name              string `json:"name"`
	Description       string `json:"description,omitempty"`
	ServiceofferingID *UUID  `json:"serviceofferingid"`
	TemplateID        *UUID  `json:"templateid"`
	ZoneID            *UUID  `json:"zoneid"`
	AffinitygroupIDs  []UUID `json:"affinitygroupids,omitempty"`
	SecuritygroupIDs  []UUID `json:"securitygroupids,omitempty"`
	NetworkIDs        []UUID `json:"networkids,omitempty"`
	Keypair           string `json:"keypair,omitempty"`
	Userdata          string `json:"userdata,omitempty"`
	Size              int    `json:"size"`
	_                 bool   `name:"createInstancePool" description:"Creates an Instance Pool with the provided parameters"`
}

type CreateInstancePoolResponse struct {
	ID                *UUID             `json:"id"`
	Name              string            `json:"name"`
	ServiceofferingID *UUID             `json:"serviceofferingid"`
	TemplateID        *UUID             `json:"templateid"`
	ZoneID            *UUID             `json:"zoneid"`
	AffinitygroupIDs  []UUID            `json:"affinitygroupids"`
	SecuritygroupIDs  []UUID            `json:"securitygroupids"`
	NetworkIDs        []UUID            `json:"networkids"`
	Keypair           string            `json:"keypair"`
	Userdata          string            `json:"userdata"`
	Size              int64             `json:"size"`
	State             InstancePoolState `json:"state"`
}

// Response returns the struct to unmarshal
func (CreateInstancePool) Response() interface{} {
	return new(CreateInstancePoolResponse)
}

type UpdateInstancePool struct {
	ID          *UUID  `json:"id"`
	ZoneID      *UUID  `json:"zoneid"`
	Description string `json:"description,omitempty"`
	Userdata    string `json:"userdata,omitempty"`
	_           bool   `name:"updateInstancePool" description:""`
}

type UpdateInstancePoolResponse struct {
	Success bool `json:"success"`
}

// Response returns the struct to unmarshal
func (UpdateInstancePool) Response() interface{} {
	return new(UpdateInstancePoolResponse)
}

type ScaleInstancePool struct {
	ID     *UUID `json:"id"`
	ZoneID *UUID `json:"zoneid"`
	Size   int   `json:"size"`
	_      bool  `name:"scaleInstancePool" description:""`
}

type ScaleInstancePoolResponse struct {
	Success bool `json:"success"`
}

// Response returns the struct to unmarshal
func (ScaleInstancePool) Response() interface{} {
	return new(ScaleInstancePoolResponse)
}

type DestroyInstancePool struct {
	ID     *UUID `json:"id"`
	ZoneID *UUID `json:"zoneid"`
	_      bool  `name:"destroyInstancePool" description:""`
}

type DestroyInstancePoolResponse struct {
	Success bool `json:"success"`
}

// Response returns the struct to unmarshal
func (DestroyInstancePool) Response() interface{} {
	return new(DestroyInstancePoolResponse)
}

type GetInstancePool struct {
	ID     *UUID `json:"id"`
	ZoneID *UUID `json:"zoneid"`
	_      bool  `name:"getInstancePool" description:""`
}

type GetInstancePoolsResponse struct {
	Count                     int
	ListInstancePoolsResponse []InstancePool `json:"instancepool"`
}

// Response returns the struct to unmarshal
func (GetInstancePool) Response() interface{} {
	return new(GetInstancePoolsResponse)
}

type ListInstancePool struct {
	ZoneID *UUID `json:"zoneid"`
	_      bool  `name:"listInstancePools" description:""`
}

type ListInstancePoolsResponse struct {
	Count                     int
	ListInstancePoolsResponse []InstancePool `json:"instancepool"`
}

// Response returns the struct to unmarshal
func (ListInstancePool) Response() interface{} {
	return new(ListInstancePoolsResponse)
}
