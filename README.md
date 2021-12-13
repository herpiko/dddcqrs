# DDD CQRS learning ground

- `cmd` - All the executable binaries
	- `http` - HTTP service that face up against front proxy / users
	- `service` - gRPC service acts as bridging to the mesh
	- `article-command` - Command module that write to databases
	- `article-query` - Query module that read from Elastic
	- `eventstore` - Event store module to catch all events that happened in the mesh
- `conn` - Connection singleton for nats, psql, elastic, etc
- `delivery` - Delivery hook for each domain
- `domain` - Domain homebase, including repositories
- `internal` - utils, helpers
- `migrations` - PostgreSQL migrations for domain
- `proto` - Holds the protobuf specs. The generated files will be placed to project root dir so they can be imported from anywhere and we can avoid import cycle
- `tests` - End to end tests

All configuration are done by environment variables. See `.env`

The grey blocks bellow are not implemented yet.

![cqrs drawio](https://user-images.githubusercontent.com/2534060/145761150-a34d6617-8ad1-4f1d-8e3d-87f62dcbab1a.png)

## Development

You need `make`. Install `build-essential` if you are on Ubuntu or Command Line Tools through XCode if you are on macOS.

### Preparation


#### Linux

`protoc` and `max_map_count`. Requires `sudo`.
```
make prep-linux
```

#### macOS (x86_64)

`protoc`. Requires `sudo`.
```
make prep-macos
```

### Infrastructure
```
make dev
```

### Proto generator
```
make gen
```

### Run locally

In Docker,
```
make dockerbuild
make dockerrun
```

Or natively: open 5 separate terminals then run this for each:
- `go run cmd/http/main`
- `go run cmd/service/main`
- `go run cmd/article-command/main`
- `go run cmd/article-query/main`
- `go run cmd/event-store/main`


### Test

In Docker,
```
make dev
make dockerbuild
make dockertest
```

Or natively: start the modules in separate terminals except `http`,
- `go run cmd/service/main`
- `go run cmd/article-command/main`
- `go run cmd/article-query/main`
- `go run cmd/event-store/main`

Then,
```
make test
```

### Images

Using `herpiko/dddcqrs` namespace,

```
make dockerbuild
make dockerpush
```
