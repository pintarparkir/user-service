# user-service

[![Security](https://sonarcloud.io/api/project_badges/measure?project=pintarparkir_user-service&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=pintarparkir_user-service)
[![Reliability](https://sonarcloud.io/api/project_badges/measure?project=pintarparkir_user-service&metric=reliability_rating)](https://sonarcloud.io/summary/new_code?id=pintarparkir_user-service)
[![Maintainability](https://sonarcloud.io/api/project_badges/measure?project=pintarparkir_user-service&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=pintarparkir_user-service)
[![Duplications](https://sonarcloud.io/api/project_badges/measure?project=pintarparkir_user-service&metric=duplicated_lines_density)](https://sonarcloud.io/summary/new_code?id=pintarparkir_user-service)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=pintarparkir_user-service&metric=coverage)](https://sonarcloud.io/summary/new_code?id=pintarparkir_user-service)

> **Purpose:** Driver profile management — owns user identity, vehicle registry, and MSISDN source for notifications.
> **Author:** Farid Triwicaksono

## Architecture Overview

![Architecture](docs/PintarParkir.architecture.svg)

## E2E Flow

![Flow Diagram](docs/flow.diagram.svg)

## Sequence Diagrams

- [End-to-End Flow](docs/sequence-diagrams/99-end-to-end-flow.md)

## Tech Stack

- Go 1.25 + Gin (HTTP) + gRPC
- PostgreSQL (pgcrypto for PII encryption)
- Redis (caching + distributed locks)
- RabbitMQ (async event-driven via outbox pattern)
- Cloud Run (GCP) with auto-scaling
- OpenTelemetry (traces + metrics)

**Service-specific:** pgcrypto PII encryption, JWT RS256 verification, lazy driver registration, gRPC server (h2c multiplexed)

## API

See [OpenAPI Specification](docs/api-specifications/openapi-spec.yaml) and [AsyncAPI Specification](docs/api-specifications/asyncapi-spec.yaml).

## Running Locally

```bash
cp configs/.env.example configs/.env
make run
```

## Testing

```bash
make test          # unit tests
make test-coverage # with coverage report
```

## Deployment

CD via GitHub Actions → GCP Cloud Run (asia-southeast1).
Triggers on push to `main`.

Cloud Run URL: `https://user-service-725nddkmwq-as.a.run.app`
