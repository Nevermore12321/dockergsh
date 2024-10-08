package vfs

import "github.com/Nevermore12321/dockergsh/internal/daemongsh/graphdriver"

func init() {
	graphdriver.Register("vfs", Init)
}

// todo
type Driver struct {
}

func (v *Driver) String() string {
	//TODO implement me
	panic("implement me")
}

func (v *Driver) Create(id, parent string) error {
	//TODO implement me
	panic("implement me")
}

func (v *Driver) Remove(id string) error {
	//TODO implement me
	panic("implement me")
}

func (v *Driver) Exists(id string) bool {
	//TODO implement me
	panic("implement me")
}

func (v *Driver) Cleanup() error {
	//TODO implement me
	panic("implement me")
}

func Init(root string, options []string) (graphdriver.Driver, error) {
	return &Driver{}, nil
}
