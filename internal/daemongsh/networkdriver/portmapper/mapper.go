package portmapper

import "github.com/Nevermore12321/dockergsh/pkg/iptables"

/* 端口映射 */

var (
	chain *iptables.Chain // 端口映射的 itables chain
)

func SetIptablesChain(c *iptables.Chain) {
	chain = c
}
