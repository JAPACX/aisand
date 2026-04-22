package lima

// VM represents a Lima virtual machine.
type VM struct {
	Name   string  `json:"name"`
	Status string  `json:"status"`
	CPUs   int     `json:"cpus"`
	Memory int64   `json:"memory"` // bytes
	Disk   int64   `json:"disk"`   // bytes
	Mounts []Mount `json:"mounts"`
}

// Mount represents a host directory mounted inside a Lima VM.
type Mount struct {
	Location string `json:"location"`
	Writable bool   `json:"writable"`
}
