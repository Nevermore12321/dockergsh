package api_server

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/external/libcontainer/user"
	"strconv"
)

func lookupGidByName(nameOrGid string) (int, error) {
	groups, err := user.ParseGroupFilter(r, func(g *user.Group) bool {
		return g.Name == nameOrGid || strconv.Itoa(g.Gid) == nameOrGid
	})
	if err != nil {
		return -1, err
	}
	if groups != nil && len(groups) > 0 {
		return groups[0].Gid, nil
	}
	return -1, fmt.Errorf("group %s not found", nameOrGid)
}

// 将 addr 权限修改为 nameOrGid 组
func changeGroup(addr, nameOrGid string) error {
	gid, err := lookupGidByName(nameOrGid)
}
