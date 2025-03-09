package main

import (
	_ "github.com/Nevermore12321/dockergsh/internal/daemongsh/execdriver/lxc"
	_ "github.com/Nevermore12321/dockergsh/internal/daemongsh/execdriver/native"
	"github.com/Nevermore12321/dockergsh/internal/reexec"
)

func main() {
	reexec.Init()
}
