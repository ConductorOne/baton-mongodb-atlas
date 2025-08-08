![Baton Logo](./docs/images/baton-logo.png)

# `baton-mongodb-atlas` [![Go Reference](https://pkg.go.dev/badge/github.com/conductorone/baton-mongodb-atlas.svg)](https://pkg.go.dev/github.com/conductorone/baton-mongodb-atlas) ![main ci](https://github.com/conductorone/baton-mongodb-atlas/actions/workflows/main.yaml/badge.svg)

`baton-mongodb-atlas` is a connector for Baton built using the [Baton SDK](https://github.com/conductorone/baton-sdk). It works with MongoDB Atlas API.

Check out [Baton](https://github.com/conductorone/baton) to learn more about the project in general.

# Prerequisites

Connector requires API key that is used throughout the communication with API. To obtain this token, you have to create one in MongoDB. More in information about how to generate token [here](https://www.mongodb.com/docs/atlas/configure-api-access/)). For synchronization **Organization Read Only** permission is enough for api key. For provisioning **Organization Owner** permission is needed.

After you have obtained both private and public key, you can use it with connector. You can do this by setting `BATON_PUBLIC_KEY` and `BATON_PRIVATE_KEY` or by passing `--public-key` and `--private-key`.

# Getting Started

## brew

```
brew install conductorone/baton/baton conductorone/baton/baton-mongodb-atlas
BATON_PUBLIC_KEY=key BATON_PRIVATE_KEY=private-key baton-mongodb-atlas
baton resources
```

## docker

```
docker run --rm -v $(pwd):/out -e BATON_PUBLIC_KEY=key BATON_PRIVATE_KEY=private-key ghcr.io/conductorone/baton-mongodb-atlas:latest -f "/out/sync.c1z"
docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-mongodb-atlas/cmd/baton-mongodb-atlas@main
BATON_PUBLIC_KEY=key BATON_PRIVATE_KEY=private-key baton-mongodb-atlas
baton resources
```

# Data Model

`baton-mongodb-atlas` will fetch information about the following Baton resources:

- Users
- Database Users
- Projects
- Teams
- Organizations

# Contributing, Support and Issues

We started Baton because we were tired of taking screenshots and manually building spreadsheets. We welcome contributions, and ideas, no matter how small -- our goal is to make identity and permissions sprawl less painful for everyone. If you have questions, problems, or ideas: Please open a Github Issue!

See [CONTRIBUTING.md](https://github.com/ConductorOne/baton/blob/main/CONTRIBUTING.md) for more details.

# `baton-mongodb-atlas` Command Line Usage

```
baton-mongodb-atlas

Usage:
  baton-mongodb-atlas [flags]
  baton-mongodb-atlas [command]

Available Commands:
  capabilities       Get connector capabilities
  completion         Generate the autocompletion script for the specified shell
  config             Get the connector config schema
  help               Help about any command

Flags:
      --client-id string                                 The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string                             The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
      --create-invite-key                                Create the invitation user email ($BATON_CREATE_INVITE_KEY)
      --external-resource-c1z string                     The path to the c1z file to sync external baton resources with ($BATON_EXTERNAL_RESOURCE_C1Z)
      --external-resource-entitlement-id-filter string   The entitlement that external users, groups must have access to sync external baton resources ($BATON_EXTERNAL_RESOURCE_ENTITLEMENT_ID_FILTER)
  -f, --file string                                      The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                                             help for baton-mongodb-atlas
      --log-format string                                The output format for logs: json, console ($BATON_LOG_FORMAT) (default "console")
      --log-level string                                 The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
      --log-level-debug-expires-at string                The timestamp indicating when debug-level logging should expire ($BATON_LOG_LEVEL_DEBUG_EXPIRES_AT)
      --otel-collector-endpoint string                   The endpoint of the OpenTelemetry collector to send observability data to (used for both tracing and logging if specific endpoints are not provided) ($BATON_OTEL_COLLECTOR_ENDPOINT)
      --private-key string                               required: ($BATON_PRIVATE_KEY)
  -p, --provisioning                                     This must be set in order for provisioning actions to be enabled ($BATON_PROVISIONING)
      --public-key string                                required: ($BATON_PUBLIC_KEY)
      --skip-full-sync                                   This must be set to skip a full sync ($BATON_SKIP_FULL_SYNC)
      --sync-resources strings                           The resource IDs to sync ($BATON_SYNC_RESOURCES)
      --ticketing                                        This must be set to enable ticketing support ($BATON_TICKETING)
  -v, --version                                          version for baton-mongodb-atlas

Use "baton-mongodb-atlas [command] --help" for more information about a command.
```
