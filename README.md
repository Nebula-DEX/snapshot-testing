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

## Contributing

Contributions to the Snapshot Testing Tool are welcome! If you encounter any issues or have suggestions for improvements, please feel free to open an issue or submit a pull request on [GitHub](https://github.com/vegaprotocol/snapshot-testing/).
