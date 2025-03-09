package native

import "github.com/Nevermore12321/dockergsh/internal/reexec"

func init() {
	reexec.Register(DriverName, initializer)
}

func initializer() {
	// todo
}
