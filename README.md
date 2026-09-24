# The Battle Royale

A learning project: a Battle Royale simulator built as Go microservices, alongside Sam Newman's _Building Microservices_. Students fight live battles on a compressed clock while bettors wager on the outcome. The domain is kept simple on purpose. The point is the architecture: Kafka, gRPC, sagas, idempotency, caching, resilience and observability.

**Status:** work in progress. Currently building the `catalog` service.

## Repository layout

```text
.
├── .gitignore                         # Build output, coverage files, .env and editor folders kept out of git
├── CLAUDE.md                          # Working rules for Claude Code: language, comment style and repository conventions
├── Makefile                           # Entry point for every task; run `make help` to list them
├── README.md                          # This file: project overview, repository map and getting started
├── buf.gen.yaml                       # buf code generation: protoc-gen-go and protoc-gen-go-grpc into gen/
├── buf.yaml                           # buf module config: proto/ as the module, STANDARD lint and FILE breaking rules
├── go.mod                             # Go module, dependencies and tools (buf, grpcurl, protoc plugins)
├── go.sum                             # Dependency checksums
├── deploy/                            # Local infrastructure
│   ├── docker-compose.yml             # Kafka, Redpanda Console, Postgres, Redis and Jaeger with pinned versions
│   └── postgres/                      # Postgres initialisation scripts, mounted into the container
│       └── init.sql                   # Creates one user and database per service on first start
├── docs/                              # Project documentation, in Spanish
│   └── events.md                      # Event catalog: topics, keys, ordering guarantees and the fields of every event
├── gen/                               # Go code generated from proto/ with `make proto-gen`; never edited by hand
├── pkg/                               # Shared infrastructure code, never domain logic
│   ├── app/                           # Service runtime: startup and graceful shutdown of components
│   │   ├── app.go                     # Component interface and Run, which stops everything on a signal or first failure
│   │   └── grpc_server.go             # gRPC server as a Component, with health check and reflection
│   ├── cache/                         # Redis client
│   │   └── redis.go                   # Creates a traced Redis client, routes go-redis logs to slog and checks it with a Ping
│   ├── database/                      # Postgres connection pool
│   │   └── postgres.go                # Creates a traced pgx pool and checks it with a Ping
│   ├── logger/                        # Logging
│   │   └── logger.go                  # Root JSON slog logger with the service name, level from LOG_LEVEL
│   └── otel/                          # Distributed tracing
│       └── tracer.go                  # Registers the global TracerProvider exporting spans over OTLP gRPC
├── proto/                             # Protobuf contracts, source of truth for the gRPC APIs and Kafka events
│   ├── arena/                         # Arena domain contracts
│   │   └── v1/                        # Version 1 of the arena events
│   │       └── events.proto           # BattleEvent envelope and the five battle events published to arena.battles
│   ├── catalog/                       # Catalog domain contracts
│   │   └── v1/                        # Version 1 of the catalog API
│   │       └── catalog.proto          # CatalogService: get and list students, weapons and locations
│   └── common/                        # Entities shared by more than one contract
│       └── v1/                        # Version 1 of the shared entities
│           └── catalog.proto          # Student, Weapon and Location, used by both the catalog API and arena events
└── services/                          # One directory per service
    └── catalog/                       # Catalog service: students, weapons and locations reference data
        ├── local.env                  # Environment for `make run s=catalog`: database, gRPC address, OTLP, Redis, cache TTL, log level
        ├── cmd/                       # Service entry point
        │   └── main.go                # Reads config, wires tracer, Postgres, Redis and domain, and runs the gRPC server
        ├── db/                        # Database scripts applied with `make migrate s=catalog`
        │   ├── migrations/            # Schema migrations
        │   │   ├── 000001_create_catalog_tables.down.sql  # Drops the students, weapons and locations tables
        │   │   └── 000001_create_catalog_tables.up.sql    # Creates the students, weapons and locations tables
        │   └── seed/                  # Reference data, tracked in its own migrations table
        │       ├── 000001_seed_catalog.down.sql           # Deletes the seed rows
        │       └── 000001_seed_catalog.up.sql             # Inserts students, weapons and locations with fixed UUIDs
        └── internal/                  # Domain code, private to the service
            ├── cached_repository.go   # Cache-aside Repository in Redis in front of Postgres
            ├── handler.go             # gRPC handler: maps requests, responses and domain errors to gRPC codes
            ├── repository.go          # Repository interface and its Postgres implementation
            ├── service.go             # Service interface and implementation, validates ids before the repository
            └── types.go               # Domain types (Student, Weapon, Location) and domain errors
```

## Getting started

Requirements: Go 1.27 and Docker.

```sh
make up                  # start the local infrastructure
make migrate s=catalog   # apply catalog migrations and seed data
make run s=catalog       # run a service locally
```

| Service          | Address                |
| ---------------- | ---------------------- |
| Kafka            | `localhost:9092`       |
| Redpanda Console | http://localhost:8080  |
| Postgres         | `localhost:5432`       |
| Redis            | `localhost:6379`       |
| Jaeger UI        | http://localhost:16686 |
| Jaeger OTLP gRPC | `localhost:4317`       |
