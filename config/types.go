package config

import "fmt"

type PostgreSQLCreds struct {
	Host   string
	Port   uint16
	User   string
	Pass   string
	DbName string
}

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

func (n Network) Validate() error {
	if len(n.DataNodesREST) == 0 {
		return fmt.Errorf("no data nodes rest endpoints")
	}

	if len(n.BootstrapPeers) == 0 {
		return fmt.Errorf("no bootstrap peers")
	}

	if len(n.RPCPeers) == 0 {
		return fmt.Errorf("no rpc peers")
	}

	if len(n.Seeds) == 0 {
		return fmt.Errorf("no seeds")
	}

	if len(n.GenesisURL) == 0 {
		return fmt.Errorf("no genesis url")
	}

	if len(n.ArtifactsRepository) == 0 {
		return fmt.Errorf("empty artifacts repository")
	}

	return nil
}
