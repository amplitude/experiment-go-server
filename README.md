# experiment-go-server

Amplitude Experiment SDK for Go.

## Documentation

Visit our [developer docs site for the full SDK documentation](https://docs.developers.amplitude.com/experiment/sdks/go-sdk/).

## Command-line Interface (`xpmt`)

The `xpmt` command-line interface tool allows you to make Experiment SDK calls from the command line. This tool is meant to be used for debugging and testing, not for use in production environments.


### Build

```
make xpmt
```

### Run

> **Warning** All examples below assume the `EXPERIMENT_KEY` environment variable has been set. Alternatively, use the `-k`
flag to set the key in the command.

#### Subcommands
  * `fetch`: fetch variants for a user from the server
  * `rules`: fetch flag configs (rules) from experiment
  * `evaluate`: fetch flag configs from experiment and evaluate the user locally

### Fetch

Fetch variants via remote evaluation for the given input user.

```
Usage of fetch:
  -d string
        Device id to fetch variants for.
  -debug
        Log additional debug output to std out.
  -i string
        User id to fetch variants for.
  -k string
        Api key for authorization, or use EXPERIMENT_KEY env var.
  -staging
        Use skylab staging environment.
  -u string
        The full user object to fetch variants for.
  -url string
        The server url to use to fetch variants from.
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
./xpmt fetch -u '{"user_id":"user@company.com","user_properties":{"premium":true}}'
```

> Note: must use single quotes around JSON object string


### Evaluate

Fetch flag configurations and locally evaluate the user.

```
Usage of evaluate:
  -d string
        Device id to fetch variants for.
  -debug
        Log additional debug output to std out.
  -i string
        User id to fetch variants for.
  -k string
        Server api key for authorization, or use EXPERIMENT_KEY env var.
  -staging
        Use skylab staging environment.
  -u string
        The full user object to fetch variants for.
  -url string
        The server url to use poll for flag configs from.

```

#### Examples

Fetch variants for a user given the user ID
```
./xpmt evaluate -i user@company.com
```

Fetch variants for a user given the device ID
```
./xpmt evaluate -d Xg0nG1v3iToYA
```

Fetch variants for a user given an experiment user JSON object
```
./xpmt evaluate -u '{"user_id":"user@company.com","user_properties":{"premium":true}}'
```

> Note: must use single quotes around JSON object string

### Running unit tests suite
To set up for running test on local, create a `.env` file in `pkg/experiment/local` with following
contents, and replace `{API_KEY}` and `{SECRET_KEY}` (or `{EU_API_KEY}` and `{EU_SECRET_KEY}` for EU data center) for the project in test:

```
API_KEY={API_KEY}
SECRET_KEY={SECRET_KEY}
```
