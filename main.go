package main

import (
	"fmt"

	"github.com/vegaprotocol/snapshot-testing/config"
	"github.com/vegaprotocol/snapshot-testing/networkutils"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	network, err := networkutils.NewNetwork(logger, config.Mainnet, "./")
	if err != nil {
		panic(err)
	}

	endpoints, err := network.GetHealthyRESTEndpoints()
	if err != nil {
		panic(err)
	}

	appVersion, err := network.GetAppVersion()
	if err != nil {
		panic(err)
	}

	// vegaURL, err := network.VegaBinaryURL()
	// if err != nil {
	// 	panic(err)
	// }
	// visorURL, err := network.VisorBinaryURL()
	// if err != nil {
	// 	panic(err)
	// }

	fmt.Printf("Endpoints: %v\n", endpoints)
	fmt.Printf("App version: %v\n", appVersion)
	vegaPath, err := network.DownloadVegaBinary()
	if err != nil {
		panic(err)
	}
	visorPath, err := network.DownloadVegaVisorBinary()
	if err != nil {
		panic(err)
	}
	restartSnapshot, err := network.GetRestartSnapshot()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Vega path: %s\n", vegaPath)
	fmt.Printf("Visor path: %s\n", visorPath)

	fmt.Printf("Restart snapshot: %#v\n", restartSnapshot)
}

// cli, err := docker.NewClient()
// if err != nil {
// 	panic(err)
// }

// containerExist, err := cli.ContainerExist(context.TODO(), config.PostgresqlConfig.Name)
// if err != nil {
// 	panic(err)
// }

// if containerExist {
// 	err := cli.ContainerRemoveForce(context.TODO(), config.PostgresqlConfig.Name)
// 	if err != nil {
// 		panic(err)
// 	}
// }

// err = cli.RunContainer(context.TODO(), config.PostgresqlConfig)
// if err != nil {
// 	panic(err)
// }

// go func() {
// 	stream, err := cli.Stdout(context.TODO(), config.PostgresqlConfig.Name, true)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer stream.Close()

// 	scanner := bufio.NewScanner(stream)
// 	for scanner.Scan() {
// 		// loglineBytes := scanner.By

// 		fmt.Printf("POSTGRESQL: %s\n ", scanner.Text())
// 	}
// 	if err := scanner.Err(); err != nil {
// 		panic(err)
// 	}

// }()

// for {
// 	running, err := cli.ContainerRunning(context.TODO(), config.PostgresqlConfig.Name)
// 	if err != nil {
// 		panic(err)
// 	}

// 	if !running {
// 		break
// 	}
// }

// fmt.Println("FINISHED")
