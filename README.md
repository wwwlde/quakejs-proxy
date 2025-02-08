# QuakeJS-Proxy

Play on QuakeJS servers with a native ioquake3 client.

## Description

QuakeJS-Proxy is a Golang proxy server that relays UDP packets from an [ioquake3](https://ioquake3.org) client to a [QuakeJS](https://github.com/inolen/quakejs) WebSocket server. It allows you to connect to QuakeJS servers using a native ioquake3 client, enabling the use of custom keybindings, configurations, and more.

## Features

- **UDP to WebSocket Proxy**: Transforms UDP packets from the ioquake3 client into WebSocket messages for QuakeJS servers.
- **Customizable Logging**: Options to log packet exchanges and new connections for debugging.
- **Hexdump Support**: Print hex dumps of packets for advanced debugging.

## Installation

Clone the repository and build the project:

```shell
$ go build cmd/main.go
```

The project has been tested with Go 1.15 and should work with all later versions.

## Usage

Run the proxy with the following command:

```shell
$ ./quakejs-proxy --ws <quakejs-server-uri>
```

### Required Parameters

- `--ws` (or `-w`): The URI of the QuakeJS server (e.g., `quakejs.com:27960`).

### Optional Parameters

- `--listen` (or `-l`): Specify the IP address to listen on. Defaults to all available interfaces.
- `--log-exchanges`: Log every packet exchange through the proxy (useful for debugging).
- `--hexdump`: Print a hex dump of every packet (useful for debugging).
- `--log-new-conn`: Enable or disable logging of new connections (default: `true`).

### Example

```shell
$ ./quakejs-proxy --ws quakejs.com:27960 --listen 0.0.0.0:27960 --log-exchanges --hexdump
```

This will:
- Connect to the QuakeJS server at `quakejs.com:27960`.
- Listen for UDP connections on `0.0.0.0:27960`.
- Log all packet exchanges and print hex dumps of packets.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## License

This project is open-source and available under the [MIT License](LICENSE).
