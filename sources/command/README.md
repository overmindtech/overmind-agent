# Command source

This source provides the ability to run commands on systems as part of the operation of a secondary source. The items returned by this source are marked as `hidden` meaning they won't be saved in the database or shown in a GUI, they only exist top help inform the creation of other, higher level items.

## `Get`

Running a Get request against this source will simplex execute the command in the query and return the item, for example the following request:

```json
{
    "context": "some-server.company.com",
    "linkdepth": 0,
    "method": 0,
    "query": "hostname",
    "type": "command"
}
```

Would return an item with the following attributes:

```json
{
    "exitCode": 0,
    "name": "hostname",
    "stderr": '',
    "stdout": "some-server"
}
```

## `Search`

The Search method is used to run commands with more complex requirements. The query for the Search method should be an instance of the `CommandParams` struct as JSON e.g.

```json
{
    // Command specifies the command to run, including all arguments
    "command": "cat /etc/hosts",

    // ExpectedExit is the expected exit code (usually 0)
    "expected_exit": 0,

    // Timeout before cancelling the command. This can be provided in any
    // format that can be parsed using `time.ParseDuration` such as "300ms",
    // or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m",
    // "h".
    "timeout": "10s",

    // Dir specifies the working directory of the command.
    "dir": "/tmp",

    // Env specifies environment variables that should be set when running the
    // command
    "env": {
        "ENV_VAR": "foo",
    },

    // STDIN specifies the binary data that should be piped to the command as
    // STDIN. This can be used for example to simulate user intaction for
    // programs that read from STDIN. This will be encoded using base64 to a
    // string in JSON
    "stdin": "eWVzCnllcwpubwo=",
}
```