# Snapshot Testing Tool

The Snapshot Testing Tool is a Golang-based utility designed to facilitate snapshot testing for non-validator nodes. It enables users to start a local non-validator node using data from a remote snapshot (including network history) and run it for a specified duration.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Flags](#flags)
- [Examples](#examples)
- [Contributing](#contributing)

## Installation

To use the Snapshot Testing Tool, ensure you have Go installed on your system. Then, clone this repository and navigate to its directory:

```bash
git clone https://github.com/vegaprotocol/snapshot-testing.git
cd snapshot-testing
```

## Usage

To run the tool, execute the following command:

```bash
go run main.go run [flags]
```

## Flags

The tool supports the following flags:

- `--duration`: Defines the duration for which the node must be running.
- `--environment`: Specifies the network the started node should connect to. Available options include:
  - `mainnet`
  - `validator-testnet`
  - `fairground`
  - `mainnet-mirror`
  - `stagnet1`
  - `devnet1`
- `--work-dir`: Local folder where all temporary files, configs, binaries, and logs are stored.
- `--config-path`: Path to the config.toml file. It can be URL or local file-path. See config.toml in this repository for the example config
- `--external-address`: IP or the DNS name of your node if the node is running behind NAT w/o symmetric routing

## Examples

Here are a few examples of how to use the tool with different configurations:

1. Start a node on the mainnet environment and run it for 24 hours:
   ```bash
   go run main.go run --environment=mainnet --duration=24h --work-dir=/path/to/work/dir
   ```

2. Run a node on the validator-testnet environment for 12 hours:
   ```bash
   go run main.go run --environment=validator-testnet --duration=12h --work-dir=/path/to/work/dir
   ```

## Result structure

The result structure returns some information about the run. The result is printed in the STDOUT and created in the `path/to/work/dir/results.json` file. The file is optional and may not be created if any result has not been produced.

Possible values:

- `catchup-duration` - tells how much time the network needed from start to catch-up
- `last-known-node-height` - the last known height for started node before snapshot-testing has stopped
- `network-stopped-producing blocks` - the date when the network stopped producing blocks for some reasons(usually crash of the network)
- `node-last-healthy` - the date when the node was healthy for the last time
- `node-last-lag` - the date when the snapshot-testing noted node was more than 500 blocks behind rest of the network for the last time
- `node-startup` - the date when the vega process started
- `reason` - the reason of the failure
- `snapshot-max` - the block for the latest available snapshot for the node
- `snapshot-min` - the first available snapshot for the node
- `status` - the status of the snapshot testing pipeline
- `test-startup` - the date when the snapshot-testing started
- `visor-extra-log-lines` - the log from the vegavisor stdout when the snapshot-testing node started
- `should-skip-failure` - the flag tells if we can safety skip reporting error(e.g. When all the nodes were unhealthy)

Example result:

```json
{
    "catchup-duration": "1m38.724939934s",
    "last-known-node-height": 15725,
    "network-stopped-producing blocks": "0001-01-01 00:00:00 +0000 UTC",
    "node-catch-up": "2024-06-12 19:03:00.515063819 +0000 UTC m=+107.143023398",
    "node-last-healthy": "2024-06-12 19:18:16.733774049 +0000 UTC m=+1023.361733653",
    "node-last-lag": "2024-06-12 19:02:55.251644873 +0000 UTC m=+101.879604442",
    "node-startup": "2024-06-12 19:02:28.948219743 +0000 UTC m=+75.576179386",
    "reason": "",
    "should-skip-failure": false,
    "snapshot-max": 15600,
    "snapshot-min": 13800,
    "status": "HEALTHY",
    "test-startup": "2024-06-12 19:01:21.790123892 +0000 UTC m=+8.418083464",
    "visor-extra-log-lines": ""
}
```

## Contributing

Contributions to the Snapshot Testing Tool are welcome! If you encounter any issues or have suggestions for improvements, please feel free to open an issue or submit a pull request on [GitHub](https://github.com/vegaprotocol/snapshot-testing/).
