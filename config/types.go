package config

type ContainerConfig struct {
	Name        string
	Image       string
	Environment map[string]string
	Command     []string
	Ports       map[uint16]uint16
}

type EndpointWithREST struct {
	CoreREST string // Usually used for health-check
	Endpoint string
}

type Network struct {
	ArtifactsRepository   string
	GenesisURL            string
	BinaryVersionOverride string // This is used when We deploy a patch to the mainnet

	DataNodesREST  []string
	RPCPeers       []EndpointWithREST
	Seeds          []string
	BootstrapPeers []string
}
