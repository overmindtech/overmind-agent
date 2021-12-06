# overmind Agent

Agent code for overmind.

## Sources

### `package`

Gets details abaout packages e.g.

```json
{
    "type": "package",
    "uniqueAttribute": "name",
    "attributes": {
        "attrStruct": {
            "architecture": "amd64",
            "description": "zlib is a library implementing the deflate compression method found\nin gzip and PKZIP.  This package includes the shared library.\n",
            "maintainer": "Ubuntu Developers <ubuntu-devel-discuss@lists.ubuntu.com>",
            "name": "zlib1g",
            "priority": "required",
            "section": "libs",
            "size": 163,
            "status": "installed",
            "summary": "compression library - runtime",
            "url": {
                "Host": "zlib.net",
                "Path": "/",
                "Scheme": "http"
            },
            "version": "1:1.2.11.dfsg-2ubuntu1.2"
        }
    },
    "context": "ubuntu2004.localdomain"
}
```

#### Search Format

The search method accepts a file name and will search for the owner of that package. On apt-based systems this uses `dpkg-query --search` and on rpm-based systems it uses `rpm -q --whatprovides`

### `group`

Details about a gruop e.g.

```json
{
    "type": "group",
    "uniqueAttribute": "name",
    "attributes": {
        "attrStruct": {
            "gid": "0",
            "members": [
                "root"
            ],
            "name": "wheel"
        }
    },
    "context": "ubuntu2004.localdomain",
    "linkedItemRequests": [
        {
            "type": "user",
            "method": 2,
            "query": "root",
            "context": "ubuntu2004.localdomain"
        }
    ]
}
```

#### Search Format

The query for Search supports either a GID, or a name of the group.

### `host`

Returns details of host entries e.g.

```json
{
    "type": "host",
    "uniqueAttribute": "name",
    "attributes": {
        "attrStruct": {
            "address": "127.0.0.1",
            "name": "localhost"
        }
    },
    "context": "ubuntu2004.localdomain",
    "linkedItemRequests": [
        {
            "type": "ip",
            "query": "127.0.0.1",
            "context": "global"
        }
    ]
}
```

#### Search Format

Query accepts an IP or a hostname

### `mount`

Reurns details of a mount point e.g.

```json
{
    "type": "mount",
    "uniqueAttribute": "path",
    "attributes": {
        "attrStruct": {
            "device": "/dev/disk0s2",
            "fstype": "hfs",
            "options": [
                "local",
                "journaled"
            ],
            "path": "/"
        }
    },
    "context": "ubuntu2004.localdomain",
    "linkedItemRequests": [
        {
            "type": "file",
            "query": "/",
            "context": "ubuntu2004.localdomain"
        }
    ]
}
```

#### Search Format

Query accepts either mount path or device name.

### `user`

Returns details of OS users e.g.

```json
{
    "type": "user",
    "uniqueAttribute": "username",
    "attributes": {
        "attrStruct": {
            "comment": "Unprivileged User",
            "gid": "65535",
            "home": "/var/empty",
            "password": "*",
            "shell": "/usr/bin/false",
            "uid": "65535",
            "username": "nobody"
        }
    },
    "context": "ubuntu2004.localdomain",
    "linkedItemRequests": [
        {
            "type": "group",
            "method": 2,
            "query": "65535",
            "context": "ubuntu2004.localdomain"
        },
        {
            "type": "file",
            "query": "/var/empty",
            "context": "ubuntu2004.localdomain"
        },
        {
            "type": "file",
            "query": "/usr/bin/false",
            "context": "ubuntu2004.localdomain"
        }
    ]
}
```

#### Search Format

Query accepts either a name or a UID

### `port`

Returns details of listening ports e.g.

```json
{
    "type": "port",
    "uniqueAttribute": "port",
    "attributes": {
        "attrStruct": {
            "localIP": "127.0.0.1",
            "pid": 15366,
            "port": 34957,
            "state": "LISTEN"
        }
    },
    "context": "ubuntu2004.localdomain",
    "linkedItemRequests": [
        {
            "type": "ip",
            "query": "127.0.0.1",
            "context": "global"
        },
        {
            "type": "process",
            "query": "15366",
            "context": "ubuntu2004.localdomain"
        }
    ]
}
```

### `disk`

Details of disks e.g.

```json
{
    "type": "disk",
    "uniqueAttribute": "device",
    "attributes": {
        "attrStruct": {
            "device": "/dev/sda3",
            "free": 120277426176,
            "fstype": "ext4",
            "inodesFree": 8039680,
            "inodesTotal": 8232960,
            "inodesUsed": 193280,
            "inodesUsedPercent": 2.347636815920398,
            "ioTime": 44064,
            "iopsInProgress": 0,
            "label": "",
            "mergedReadCount": 14706,
            "mergedWriteCount": 55535,
            "mountpoint": "/",
            "name": "sda3",
            "opts": "rw,relatime",
            "path": "/",
            "readBytes": 1456608256,
            "readCount": 60049,
            "readTime": 33581,
            "serialNumber": "",
            "total": 132224544768,
            "used": 5186424832,
            "usedPercent": 4.133800126754675,
            "weightedIO": 552,
            "writeBytes": 1509060608,
            "writeCount": 12975,
            "writeTime": 5953
        }
    },
    "context": "ubuntu2004.localdomain",
    "linkedItemRequests": [
        {
            "type": "file",
            "query": "/",
            "context": "ubuntu2004.localdomain"
        }
    ]
}
```

### `process`

Gets details of processes e.g.

```json
{
    "type": "process",
    "uniqueAttribute": "pid",
    "attributes": {
        "attrStruct": {
            "cmdline": "/tmp/go-build1837870653/b271/psutil.test -test.paniconexit0 -test.timeout=10m0s -test.v=true -test.count=1",
            "cpuPercent": 0.7253417588413082,
            "createTime": "2021-12-06 14:35:47 +0000 UTC",
            "cwd": "/vagrant/sources/psutil",
            "exe": "/tmp/go-build1837870653/b271/psutil.test",
            "groups": [
                4,
                24,
                30,
                46,
                111,
                117,
                118,
                1000
            ],
            "ioReadBytes": 4096,
            "ioReadCount": 145,
            "ioWriteBytes": 0,
            "ioWriteCount": 16,
            "isRunning": true,
            "memoryPercent": 1.0320771932601929,
            "name": "psutil.test",
            "nice": 20,
            "numConnections": 0,
            "pageFaultsMajor": 0,
            "parent": 14785,
            "pid": 15422,
            "status": "Running",
            "terminal": "/pts/0",
            "username": "vagrant"
        }
    },
    "context": "ubuntu2004.localdomain",
    "linkedItemRequests": [
        {
            "type": "file",
            "query": "/vagrant/sources/psutil",
            "context": "ubuntu2004.localdomain"
        },
        {
            "type": "file",
            "query": "/tmp/go-build1837870653/b271/psutil.test",
            "context": "ubuntu2004.localdomain"
        },
        {
            "type": "package",
            "method": 2,
            "query": "/tmp/go-build1837870653/b271/psutil.test",
            "context": "ubuntu2004.localdomain"
        },
        {
            "type": "group",
            "query": "4",
            "context": "ubuntu2004.localdomain"
        },
        {
            "type": "group",
            "query": "24",
            "context": "ubuntu2004.localdomain"
        },
        {
            "type": "group",
            "query": "30",
            "context": "ubuntu2004.localdomain"
        },
        {
            "type": "group",
            "query": "46",
            "context": "ubuntu2004.localdomain"
        },
        {
            "type": "group",
            "query": "111",
            "context": "ubuntu2004.localdomain"
        },
        {
            "type": "group",
            "query": "117",
            "context": "ubuntu2004.localdomain"
        },
        {
            "type": "group",
            "query": "118",
            "context": "ubuntu2004.localdomain"
        },
        {
            "type": "group",
            "query": "1000",
            "context": "ubuntu2004.localdomain"
        },
        {
            "type": "process",
            "query": "14785",
            "context": "ubuntu2004.localdomain"
        },
        {
            "type": "user",
            "query": "vagrant",
            "context": "ubuntu2004.localdomain"
        }
    ]
}
```

### `service`

Returns service details e.g.

```json
{
    "type": "service",
    "uniqueAttribute": "Name",
    "attributes": {
        "attrStruct": {
            "ActiveState": "active",
            "Description": "D-Bus System Message Bus",
            "ExecCondition": [],
            "ExecMainPID": 529,
            "ExecReload": {
                "args": [
                    "--print-reply",
                    "--system",
                    "--type=method_call",
                    "--dest=org.freedesktop.DBus",
                    "/",
                    "org.freedesktop.DBus.ReloadConfig"
                ],
                "binary": "/usr/bin/dbus-send",
                "fullCMD": "/usr/bin/dbus-send --print-reply --system --type=method_call --dest=org.freedesktop.DBus / org.freedesktop.DBus.ReloadConfig"
            },
            "ExecStart": {
                "args": [
                    "--system",
                    "--address=systemd:",
                    "--nofork",
                    "--nopidfile",
                    "--systemd-activation",
                    "--syslog-only"
                ],
                "binary": "/usr/bin/dbus-daemon",
                "fullCMD": "/usr/bin/dbus-daemon --system --address=systemd: --nofork --nopidfile --systemd-activation --syslog-only"
            },
            "ExecStop": {
                "args": [],
                "binary": "/bin/true",
                "fullCMD": "/bin/true"
            },
            "FragmentPath": {},
            "GuessMainPID": true,
            "LoadState": "loaded",
            "MemoryCurrent": 1896448,
            "Name": "dbus.service",
            "NotifyAccess": "none",
            "OOMPolicy": "stop",
            "Path": "/org/freedesktop/systemd1/unit/dbus_2eservice",
            "Restart": "no",
            "SubState": "running",
            "Type": "simple"
        }
    },
    "context": "ubuntu2004.localdomain",
    "linkedItemRequests": [
        {
            "type": "file",
            "query": "/bin/true",
            "context": "ubuntu2004.localdomain"
        },
        {
            "type": "file",
            "query": "/usr/bin/dbus-send",
            "context": "ubuntu2004.localdomain"
        },
        {
            "type": "file",
            "query": "/usr/bin/dbus-daemon",
            "context": "ubuntu2004.localdomain"
        },
        {
            "type": "process",
            "query": "529",
            "context": "ubuntu2004.localdomain"
        }
    ]
}
```

#### Search Format

Query searches by a glob pattern

### `file`

Returns details about files. Does not support `Find()` as this would involve searching the whole disk for millions of files and doesn't make sense.

```json
{
    "type": "file",
    "uniqueAttribute": "path",
    "attributes": {
        "attrStruct": {
            "group": "vagrant",
            "mode": "-rw-------",
            "owner": "vagrant",
            "path": "/tmp/tempTestFile2124936914",
            "size": 0,
            "type": "file"
        }
    },
    "context": "ubuntu2004.localdomain",
    "linkedItemRequests": [
        {
            "type": "group",
            "query": "vagrant",
            "context": "ubuntu2004.localdomain"
        },
        {
            "type": "user",
            "query": "vagrant",
            "context": "ubuntu2004.localdomain"
        },
        {
            "type": "package",
            "method": 2,
            "query": "/tmp/tempTestFile2124936914",
            "context": "ubuntu2004.localdomain"
        }
    ]
}
```

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
