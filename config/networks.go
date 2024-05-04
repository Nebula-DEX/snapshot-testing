package config

import "fmt"

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

var (
	Mainnet = Network{
		ArtifactsRepository: "vegaprotocol/vega",
		GenesisURL:          "https://raw.githubusercontent.com/vegaprotocol/networks/master/mainnet1/genesis.json",
		DataNodesREST: []string{
			"https://api0.vega.community",
			"https://api1.vega.community",
			"https://api2.vega.community",
			"https://api3.vega.community",
		},
		RPCPeers: []EndpointWithREST{
			{CoreREST: "https://api1.vega.community", Endpoint: "api1.vega.community:26657"},
			{CoreREST: "https://api2.vega.community", Endpoint: "api2.vega.community:26657"},
			{CoreREST: "https://api3.vega.community", Endpoint: "api3.vega.community:26657"},
			{CoreREST: "https://be0.vega.community", Endpoint: "be0.vega.community:26657"},
			{CoreREST: "https://be1.vega.community", Endpoint: "be1.vega.community:26657"},
			{CoreREST: "https://be3.vega.community", Endpoint: "be3.vega.community:26657"},
		},
		Seeds: []string{
			"b0db58f5651c85385f588bd5238b42bedbe57073@13.125.55.240:26656",
			"abe207dae9367995526812d42207aeab73fd6418@18.158.4.175:26656",
			"198ecd046ebb9da0fc5a3270ee9a1aeef57a76ff@144.76.105.240:26656",
			"211e435c2162aedb6d687409d5d7f67399d198a9@65.21.60.252:26656",
			"c5b11e1d819115c4f3974d14f76269e802f3417b@34.88.191.54:26656",
			"61051c21f083ee30c835a34a0c17c5d1ceef3c62@51.178.75.45:26656",
			"b0db58f5651c85385f588bd5238b42bedbe57073@18.192.52.234:26656",
			"36a2ca7bb6a50427be2181c8ebb7f62ac62ebaf5@m2.vega.community:26656",
			"9903c02a0ff881dc369fc7daccb22c1f9680d2dd@api0.vega.community:26656",
			"9903c02a0ff881dc369fc7daccb22c1f9680d2dd@api0.vega.community:26656",
			"32d7380b195c088c0605c5d24bcf15ff1dade05f@api1.vega.community:26656",
			"4f26ec99d3cf6f0e9e973c0a5f3da87d89ec6677@api2.vega.community:26656",
			"eafacd11af53cd9fb2a14eada53485779cbee4ab@api3.vega.community:26656",
			"9de3ca2bbeb62d165d39acbbcf174e7ac3a6b7c9@be3.vega.community:26656",
		},
		BootstrapPeers: []string{
			"/dns/api1.vega.community/tcp/4001/ipfs/12D3KooWDZrusS1p2XyJDbCaWkVDCk2wJaKi6tNb4bjgSHo9yi5Q",
			"/dns/api2.vega.community/tcp/4001/ipfs/12D3KooWEH9pQd6P7RgNEpwbRyavWcwrAdiy9etivXqQZzd7Jkrh",
			"/dns/api3.vega.community/tcp/4001/ipfs/12D3KooWEH9pQd6P7RgNEpwbRyavWcwrAdiy9etivXqQZzd7Jkrh",
		},
	}

	Fairground = Network{
		ArtifactsRepository: "vegaprotocol/vega",
		GenesisURL:          "https://raw.githubusercontent.com/vegaprotocol/networks-internal/main/fairground/genesis.json",
		DataNodesREST: []string{
			"https://api.n00.testnet.vega.rocks",
			"https://api.n06.testnet.vega.rocks",
			"https://api.n07.testnet.vega.rocks",
			"https://api.n08.testnet.vega.rocks",
		},
		RPCPeers: []EndpointWithREST{
			{CoreREST: "https://n00.testnet.vega.rocks/", Endpoint: "n00.testnet.vega.rocks:26657"},
			{CoreREST: "https://n01.testnet.vega.rocks/", Endpoint: "n01.testnet.vega.rocks:26657"},
			{CoreREST: "https://n02.testnet.vega.rocks/", Endpoint: "n02.testnet.vega.rocks:26657"},
			{CoreREST: "https://n03.testnet.vega.rocks/", Endpoint: "n03.testnet.vega.rocks:26657"},
			{CoreREST: "https://n04.testnet.vega.rocks/", Endpoint: "n04.testnet.vega.rocks:26657"},
			{CoreREST: "https://n05.testnet.vega.rocks/", Endpoint: "n05.testnet.vega.rocks:26657"},
			{CoreREST: "https://n06.testnet.vega.rocks/", Endpoint: "n06.testnet.vega.rocks:26657"},
			{CoreREST: "https://n07.testnet.vega.rocks/", Endpoint: "n07.testnet.vega.rocks:26657"},
			{CoreREST: "https://n08.testnet.vega.rocks/", Endpoint: "n08.testnet.vega.rocks:26657"},
		},
		Seeds: []string{
			"503a32dbd88dfddaaedb26c08bf94e3b88271527@n01.testnet.vega.rocks:26656",
			"d11e5c33795d1759db8bc50061e6a0c445aef47e@n02.testnet.vega.rocks:26656",
			"f8a64e85493e52e68f3ed6025e026fd049477e4f@n03.testnet.vega.rocks:26656",
			"0e8d71252e579115da5ab89f2ecac6cb57319b37@n04.testnet.vega.rocks:26656",
			"611e3cf6a12e58ba8a4ce577c202562214107b7d@n05.testnet.vega.rocks:26656",
		},
		BootstrapPeers: []string{
			"/dns/n00.testnet.vega.rocks/tcp/4001/ipfs/12D3KooWNiWcT93S3P3eiHqGq4a6feaD2cUfbWw9AxgdVt8RzTHJ",
			"/dns/n06.testnet.vega.rocks/tcp/4001/ipfs/12D3KooWMSaQevxg1JcaFxWTpxMjKw1J13bLVLmoxbeSJ5gpXjRh",
			"/dns/n07.testnet.vega.rocks/tcp/4001/ipfs/12D3KooWACJuzchZQH8Tz1zNmkGCatgcS2DUoiQnMFaALVMo7DpC",
			"/dns/n08.testnet.vega.rocks/tcp/4001/ipfs/12D3KooWGKPFor9TThHKDCwVWHcmgtm1A4DKF5g25cLaAZpTWUZ2",
		},
	}

	Stagnet1 = Network{
		ArtifactsRepository: "vegaprotocol/vega",
		GenesisURL:          "https://raw.githubusercontent.com/vegaprotocol/networks-internal/main/stagnet1/genesis.json",
		DataNodesREST: []string{
			"https://api.n00.stagnet1.vega.rocks",
			"https://api.n05.stagnet1.vega.rocks",
			"https://api.n06.stagnet1.vega.rocks",
		},
		RPCPeers: []EndpointWithREST{
			{CoreREST: "https://n00.stagnet1.vega.rocks", Endpoint: "n00.stagnet1.vega.rocks:26657"},
			{CoreREST: "https://n01.stagnet1.vega.rocks", Endpoint: "n01.stagnet1.vega.rocks:26657"},
			{CoreREST: "https://n02.stagnet1.vega.rocks", Endpoint: "n02.stagnet1.vega.rocks:26657"},
			{CoreREST: "https://n03.stagnet1.vega.rocks", Endpoint: "n03.stagnet1.vega.rocks:26657"},
			{CoreREST: "https://n04.stagnet1.vega.rocks", Endpoint: "n04.stagnet1.vega.rocks:26657"},
			{CoreREST: "https://n05.stagnet1.vega.rocks", Endpoint: "n05.stagnet1.vega.rocks:26657"},
			{CoreREST: "https://n06.stagnet1.vega.rocks", Endpoint: "n06.stagnet1.vega.rocks:26657"},
		},
		Seeds: []string{
			"6a473fa0c9571deb3c494c9ac64d4dda41adde3f@n01.stagnet1.vega.rocks:26656",
			"ca6e178a32324e07893049f1090077b520912803@n02.stagnet1.vega.rocks:26656",
			"49d9e6ee15e249c21d35ebe46f72f1ac631b0586@n03.stagnet1.vega.rocks:26656",
			"eea179e9eef3d760c7d6cc675d6a374347806e62@n04.stagnet1.vega.rocks:26656",
		},
		BootstrapPeers: []string{
			"/dns/n00.stagnet1.vega.rocks/tcp/4001/ipfs/12D3KooWJ5HxcmfVgstPNFquf8DTwAJg5BWmvZ1oLsqqY3g1ygDG",
			"/dns/n05.stagnet1.vega.rocks/tcp/4001/ipfs/12D3KooWHNyJBuN9GmYp23FAdMbL3nmwe5DzixFNL8d4oBTMzxag",
			"/dns/n06.stagnet1.vega.rocks/tcp/4001/ipfs/12D3KooWQpceAbYaEaas65tEt8CJofHgjRPANaojwA7oaQApHTvB",
		},
	}

	Devnet1 = Network{
		ArtifactsRepository: "vegaprotocol/vega-dev-releases",
		GenesisURL:          "https://raw.githubusercontent.com/vegaprotocol/networks-internal/main/devnet1/genesis.json",
		DataNodesREST: []string{
			"https://api.n00.devnet1.vega.rocks",
			"https://api.n06.devnet1.vega.rocks",
			"https://api.n07.devnet1.vega.rocks",
		},
		RPCPeers: []EndpointWithREST{
			{CoreREST: "https://n00.devnet1.vega.rocks", Endpoint: "n00.devnet1.vega.rocks:26657"},
			{CoreREST: "https://n01.devnet1.vega.rocks", Endpoint: "n01.devnet1.vega.rocks:26657"},
			{CoreREST: "https://n02.devnet1.vega.rocks", Endpoint: "n02.devnet1.vega.rocks:26657"},
			{CoreREST: "https://n03.devnet1.vega.rocks", Endpoint: "n03.devnet1.vega.rocks:26657"},
			{CoreREST: "https://n04.devnet1.vega.rocks", Endpoint: "n04.devnet1.vega.rocks:26657"},
			{CoreREST: "https://n05.devnet1.vega.rocks", Endpoint: "n05.devnet1.vega.rocks:26657"},
			{CoreREST: "https://n06.devnet1.vega.rocks", Endpoint: "n06.devnet1.vega.rocks:26657"},
		},
		Seeds: []string{
			"a0928bc929506560c66f5ae4fa2f73df3ed8aab8@n01.devnet1.vega.rocks:26656",
			"0c0f1575d159ed02ac05670c333593b2deb4d57e@n02.devnet1.vega.rocks:26656",
			"091cb0675d0f59305d6b72072fe423206bf17048@n03.devnet1.vega.rocks:26656",
			"e475c424a3f20313f5b0911a06b438c850b89066@n04.devnet1.vega.rocks:26656",
			"7f2b12134155929f70ef162a58a8ad5c289eacde@n05.devnet1.vega.rocks:26656",
		},
		BootstrapPeers: []string{
			"/dns/n00.devnet1.vega.rocks/tcp/4001/ipfs/12D3KooWBsVeEhCjG2djhpwexZWb76Afd7Nh6gUfpxNBr61KKojj",
			"/dns/n06.devnet1.vega.rocks/tcp/4001/ipfs/12D3KooWEbFqpQc2srFtrPcYK5t1e8mfouDutyzwW3XBEPhqYrLi",
			"/dns/n07.devnet1.vega.rocks/tcp/4001/ipfs/12D3KooWSjnLDRMwrNxWqyyzkWCkiP7JaHpKkgbNGpo8fWWfkXoy",
		},
	}
)
