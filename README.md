# Deviant Agent

Agent code for Deviant.

## Config

All configuration options can be provided via the command line or as environment variables:

| Environment Variable | CLI Flag | Automatic | Description |
|----------------------|----------|-----------|-------------|
| `CONFIG`| `--config` | ✅ | Config file location. Can be used instead of the CLI or environment variables if needed |
| `LOG`| `--log` | ✅ | Set the log level. Valid values: panic, fatal, error, warn, info, debug, trace |
| `NATS_SERVERS`| `--nats-servers` | ✅ | A list of NATS servers to connect to |
| `NATS_NAME_PREFIX`| `--nats-name-prefix` | ✅ | A name label prefix. Sources should append a dot and their hostname .{hostname} to this, then set this is the NATS connection name which will be sent to the server on CONNECT to identify the client |
| `NATS_CA_FILE`| `--nats-ca-file` | ✅ | Path to the CA file that NATS should use when connecting over TLS |
| `NATS_JWT_FILE`| `--nats-jwt-file` | ✅ | Path to the file containing the user JWT |
| `NATS_NKEY_FILE`| `--nats-nkey-file` | ✅ | Path to the file containing the NKey seed |
| `MAX-PARALLEL`| `--max-parallel` | ✅ | Max number of requests to run in parallel |
| `YOUR_CUSTOM_FLAG`| `--your-custom-flag` |   | Configuration that you add should be documented here |

Config can also be provided as a config f

## Developing

## Creating New Commands

This project uses the [Cobra](https://github.com/spf13/cobra/blob/master/cobra/README.md) framework as a CLI. New commands can be created by running things like this:

```shell
cobra add serve
cobra add config
cobra add create -p 'configCmd'
```

### Running with NATS

#### Docker Compose

This method only incudes the bare minimum for NATS to work, just NATS itself. useful for integration testing when the larger kubernetes cluster is not required.

Start the required services in the foreground and stop when cancelled:

```
docker compose up && docker compose down
```

Or start the service in the background:

```
docker compose up -d
```

Run the Agent:

```
go run main.go
```

### Local Debugging

This will compile and run the code on your local laptop and can be run using the "local discover" debugging routine. This is likely not to be super helpful as it will only discover local resources. Note that in order for debugging to work you will need to add a host entry that resolves `nats.debug` to a URL that runs a NATS server

### Remote Debugging

Vagrant is used for remote debugging against VMs. Firstly install the [Remote - SSH](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-ssh) plugin. Then ssh to either the "centos" or "ubuntu" vms. This will install VSCode on the servers and allow you to run debugging locally on them.

TODO: Make the process of configuring SSH easier

### Updating Go Dependencies

```shell
go get -v -u -t && go mod tidy && go mod vendor
```

### TODOS

This project uses the following tags for TODOs:

* **`PANIC`**: Places in the code that could in theory cause an unrecovered panic that probably should be handled better
* **`PERF`**: Opportunities for performance gains
* **`TODO`**: This is for general things that should be fixed that don't fit into the other categories

The Todo Tree plugin is recommended (in `.vscode/extensions.json`) for managing these.

### Running Locally

The source CLI can be interacted with locally by running:

```shell
go run main.go --help
```

### Testing

Tests in this package can be run using:

```shell
go test ./...
```
