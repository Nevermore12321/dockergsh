package aufs

import "github.com/Nevermore12321/dockergsh/internal/daemongsh/graphdriver"

// 初始化，注册 aufs driver
func init() {
	graphdriver.Register("aufs", Init)
}

// Driver todo
type Driver struct {
}

func (Driver) String() string {
	return "aufs"
}

func (a *Driver) Create(id, parent string) error {
	//TODO implement me
	return nil
}

func (a *Driver) Remove(id string) error {
	//TODO implement me
	return nil
}

func (a *Driver) Exists(id string) bool {
	//TODO implement me
	return false
}

func (a *Driver) Cleanup() error {
	//TODO implement me
	return nil
}

func Init(root string, options []string) (graphdriver.Driver, error) {
	return &Driver{}, nil
}
