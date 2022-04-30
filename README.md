# experiment-go-server

Amplitude Experiment SDK for Go.

## Documentation

Visit our [developer docs site for the full SDK documentation](https://docs.developers.amplitude.com/experiment/sdks/go-sdk/).

## Command-line Interface (`xpmt`)

The `xpmt` command-line interface tool allows you to make Experiment SDK calls from the command line. This tool is meant to be used for debugging and testing, not for use in production environments.


### Build

**Makefile currently only builds for macos (amd64 & x64), add a line to the `Makefile` to support your OS and Architecture.**

```
make xpmt
```

### Run

> **NOTE:** All examples below assume the `EXPERIMENT_KEY` environment variable has been set. Alternatively, use the `-k` 
flag to set the key in the command.

#### Subcommands
  * `fetch`: fetch variants for a user from the server

### Fetch

```
Usage of fetch:
  -d string
        Device id to fetch variants for.
  -i string
        User id to fetch variants for.
  -k string
        Api key for authorization, or use EXPERIMENT_KEY env var.
  -u string
        The full user object to fetch variants for.
```

#### Examples

Fetch variants for a user given the user ID
```
./xpmt fetch -i user@company.com
```

Fetch variants for a user given the device ID
```
./xpmt fetch -d Xg0nG1v3iToYA
```

Fetch variants for a user given an experiment user JSON object
```
./xpmt fetch -u '{"user_id":"user@company.com","country":"France"}'
```

> Note: must use single quotes around JSON object string


### Evaluate

Fetch flag configurations and remotely evaluate the user.

```
Usage of fetch:
  -d string
        Device id to fetch variants for.
  -i string
        User id to fetch variants for.
  -k string
        Api key for authorization, or use EXPERIMENT_KEY env var.
  -u string
        The full user object to fetch variants for.
```

#### Examples

Fetch variants for a user given the user ID
```
./xpmt fetch -i user@company.com
```

Fetch variants for a user given the device ID
```
./xpmt fetch -d Xg0nG1v3iToYA
```

Fetch variants for a user given an experiment user JSON object
```
./xpmt fetch -u '{"user_id":"user@company.com","country":"France"}'
```

> Note: must use single quotes around JSON object string