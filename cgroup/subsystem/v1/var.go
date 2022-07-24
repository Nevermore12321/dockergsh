package v1

import "github.com/Nevermore12321/dockergsh/cgroup/subsystem"

var (
	SubsystemIns = []subsystem.Subsystem{
		&CpuSubSystem{},
		&CpuSetSubSystem{},
		&MemorySubSystem{},
	}
)
