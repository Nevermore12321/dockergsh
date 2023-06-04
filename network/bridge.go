package network

type BridgeNetworkDriver struct {
}

func (bd *BridgeNetworkDriver) Name() string {
	return "bridge"
}
