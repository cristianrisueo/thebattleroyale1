# The Battle Royale

A learning project: a Battle Royale simulator built as Go microservices, alongside Sam Newman's _Building Microservices_. Students fight live battles on a compressed clock while bettors wager on the outcome. The domain is kept simple on purpose. The point is the architecture: Kafka, gRPC, sagas, idempotency, caching, resilience and observability.

**Status:** work in progress. Currently building the `catalog` service.

## Repository layout

| Path                       | Contents                                                                                                      |
| -------------------------- | ------------------------------------------------------------------------------------------------------------- |
| `proto/`                   | Protobuf contracts, one versioned package per domain (`catalog/v1`). Source of truth for the gRPC APIs.       |
| `gen/`                     | Go code generated from `proto/` with `make proto-gen`. Never edited by hand.                                  |
| `services/<name>/`         | One directory per service. `db/migrations` holds schema migrations and `db/seed` reference data.              |
| `deploy/`                  | Local infrastructure: Docker Compose (Kafka, Redpanda Console, Postgres, Redis) and the Postgres init script. |
| `buf.yaml`, `buf.gen.yaml` | buf configuration: lint and breaking-change rules, and code generation.                                       |
| `Makefile`                 | Entry point for every task. Run `make help` to list them.                                                     |

## Getting started

Requirements: Go 1.27 and Docker.

```sh
make up                  # start the local infrastructure
make migrate s=catalog   # apply catalog migrations and seed data
```

| Service          | Address               |
| ---------------- | --------------------- |
| Kafka            | `localhost:9092`      |
| Redpanda Console | http://localhost:8080 |
| Postgres         | `localhost:5432`      |
| Redis            | `localhost:6379`      |
