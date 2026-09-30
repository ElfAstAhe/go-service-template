# forge-go 🛠️🔥

`forge-go` is a high-performance, resilient, and highly observable Enterprise application framework (production-grade backbone) written in Go. Engineered strictly in accordance with Clean Architecture, Domain-Driven Design (DDD), and SOLID principles, it provides a unified, production-ready foundation for building robust Cloud-Native microservices capable of handling heavy Highload traffic.

[![Go Reference](https://go.dev)](https://go.dev)
[![License](https://shields.io)](https://opensource.org)
[![Quality Gate](https://shields.io)](#quality-gate-requirements)

---

## 💎 Core Architectural Features

* **Advanced Observability (Production Telemetry):** End-to-end distributed tracing context propagation (`RequestID`/`TraceID`) is seamlessly baked into HTTP and gRPC layers. Native support for **Prometheus Exemplars** in latency histograms allows engineers to jump from Grafana spikes directly to exact Jaeger/Tempo trace spans in a single click.
* **Defensive Runtime Design (DoS/OOM Protection):** Adaptive resource budgeting explicitly protects pod memory from malicious payloads and decompression exploits (Zip Bombs) via strict `MaxBytesReader` constraints enforced during JSON decoding and dynamic stream decompression (Brotli/Gzip).
* **High Availability & Fault Tolerance:** Database connection pools and asynchronous messaging clients (Apache Kafka, AMQP 1.0/Azure Service Bus) implement robust **Double-Check Locking** thread-safety patterns. Connection failover handles infrastructure flips out of critical mutex paths via aggressive lazy invalidation hooks (`Invalidate`).
* **Deterministic Lifecycle Management:** The IoC/DI application orchestration lifecycle (`BaseApplication` / `container.Runner`) coordinates smooth **Graceful Shutdown** sequences with dual-barrier timeout safety nets, ensuring inflight message batches and database transactions finish before the container terminates.

---

## 🏗️ Project Layout (Standard Go Structure)

The framework strictly adheres to the **Standard Go Project Layout** combined with decoupled layer boundaries to shield the core business logic from direct vendor dependencies.

```text
.
├── api/                    # API contracts and specifications
│   ├── proto/              # Protobuf/gRPC service schemas
│   └── rest/               # Client generators, HTTP collections, and test configurations
├── bin/                    # Compiled binary build artifacts
├── cmd/                    # Application architecture entry points
│   ├── gen-tz/             # CLI utility for compiled timezone constants code generation
│   └── example-service/    # Main entry: boots DI containers and executes example runtime runner
├── configs/                # Infrastructure configurations & blueprints documentation (brokers, cache, etc)
├── deployments/            # Infrastructure blueprints (Dockerfile, compose, k8s manifests)
├── docs/                   # Automatically generated Swagger UI API documentation
├── internal/               # Private application scope (Encapsulated Business Logic)
│   ├── app/                # Bootstrapping orchestrator: wires DI layers & Graceful Shutdown
│   │   └── container/      # Service-specific dependency injection containers
│   ├── config/             # Passport bindings: unmarshals YAML/ENV/Flags via Viper
│   ├── domain/             # Enterprise Core: Domain entities, invariants, and aggregate definitions
│   │   └── mocks/          # Automated mock generation scope for business components
│   ├── facade/             # System Boundaries: BLL boundary façade routers
│   │   ├── dto/            # Data Transfer Objects mapping boundary requests
│   │   └── mapper/         # Decoupled structural data conversion transformers
│   ├── usecase/            # Application Logic: Pure single-responsibility business workflows
│   │   └── trace/          # Business logic telemetry and span context wrapping
│   ├── repository/         # DAL: Storage implementation adapters (Postgres, Redis, Cache)
│   │   ├── metrics/        # Storage adapter Prometheus metric wrappers
│   │   ├── trace/          # Storage adapter OpenTelemetry trace spans tracking
│   │   └── postgres/       # SQL storage engine execution blueprints for Postgres
│   └── transport/          # Inbound network delivery handlers (REST / gRPC)
├── migrations/             # Database structural state evolution schemas (SQL)
│   └── example-service/    # Example Service database migrations
├── pkg/                    # Reusable Public Framework Libraries (The Platform Engine)
│   ├── api/                # Generated client packages and network interfaces
│   ├── app/                # Common application engine blueprints
│   ├── auth/               # Unified authentication & security tokens management
│   ├── config/             # Common configuration engine blueprints
│   ├── container/          # Generic IoC/DI container orchestrators & helpers
│   ├── db/                 # SQL database abstraction wrappers & Transaction Managers
│   │   └── postgres/       # Native relational database adapter implementation for PostgreSQL
│   ├── domain/             # Framework-level generic domain entity models & repositories
│   ├── errs/               # Centralized, strongly-typed enterprise error taxonomy
│   ├── helper/             # High-performance system utilities and helpers
│   ├── infra/              # Core Telemetry & Caching Engine
│   │   ├── cache/          # In-memory and distributed cache wrappers
│   │   ├── metrics/        # Custom Prometheus collectors and metrics interceptors
│   │   ├── pubsub/         # Low-overhead implementation of the Pub/Sub architectural pattern
│   │   └── telemetry/      # OpenTelemetry distributed tracing ecosystem wrappers
│   ├── logger/             # Zero-allocation logging core layered on top of Uber Zap
│   ├── migration/          # Framework-level database migration engine abstractions
│   │   └── goose/          # Production-grade database migration wrapper utilizing Goose engine
│   ├── repository/         # High-performance Generic CRUD & Owned Repository primitives
│   │   ├── metrics/        # Generic metric tracking decorator components
│   │   └── trace/          # Generic distributed tracing span decorator components
│   └── transport/          # Abstract Network Protocol Implementations
│       ├── brokers/        # Vendor-agnostic Unified Event Bus Abstraction Wrapper
│       │   ├── amqp/       # AMQP protocol layer definition
│       │   │   └── azure/  # Azure Service Bus client driver implementing AMQP 1.0 (go-amqp)
│       │   └── kafka/      # Apache Kafka consumer/producer client driver (segmentio/kafka-go)
│       ├── grpc/           # Production gRPC servers, KeepAlive configs, & interceptors
│       │   └── interceptors/ # Custom Unary and Stream gRPC processing hooks
│       ├── http/           # Standard HTTP engines, Chi routers, & decompression pipelines
│       │   └── middleware/ # Structured logging, recovery, and tracing REST middleware
│       └── worker/         # Abstract asynchronous background process runners
└── utils/                  # Low-level system runtime utilities and helpers
```

---

## 🧩 Architectural Layer Boundaries

1. **Domain:** Fully decoupled. It does not import any external or framework libraries. Contains raw data structures, domain invariants, and BLL abstraction repository interfaces.
2. **Usecase:** Houses core business execution paths. Strictly follows the *Single Responsibility Principle*. Depends on Domain abstractions only.
3. **Repository (Adapter Layer):** Concrete realization of Domain data interfaces. Manages raw SQL execution, context transactions, Redis pools, or upstream HTTP microservice consumers.
4. **Transport (Delivery Layer):** Edge input boundary controllers. Translates network-specific wire packages (JSON streams, Protobuf binaries) into pure data types understandable by the Usecase layer.
5. **App:** The IoC Dependency Injection boundary. Assembles and glues all application layers together at startup.
6. **Pkg:** Shareable, highly-optimized utility and driver ecosystem which can be dropped into any cloud service seamlessly.

---

## 🚀 Implemented Capabilities in Example Service

* **Context Transaction Manager:** Thread-safe ACID transaction propagation across multi-repository Usecase execution blocks.
* **Generic CRUD Repository:** Production-grade generic abstraction layer minimizing structural boilerplates while maintaining full support for database context transactions.
* **Telemetry Decorators:** Non-invasive monitoring layer wrapping Repository instances with tracing and Prometheus collection via the *Decorator pattern*.
* **Defensive HTTP Pipeline:** Scalable Chi-router orchestration armed with Gzip/Brotli stream decoders, Structured Logging, Metrics collection, and Exemplar tracking.
* **Enterprise gRPC Runner:** Highly tuned gRPC engine fortified with KeepAlive pings, Unary/Stream tracing, and metric interceptors out of the box.
* **Application Configuration:** Fast, safe environment configurations mapping via the Viper engine.

---

## 📈 Execution Request Pipeline Sequence

Execution traffic cascades down through decoupled interfaces, enforcing clean boundary isolation:

```text
[Client Request] ──> Transport ──> Façade (DTO/Mapper) ──> Use Case ──> Repository ──> [Database/Broker]
```
The return path mirrors this flow, with mappers converting core domain models back into transport-specific DTOs safely.

---

## 📜 Quality Gate Requirements

The platform enforces zero technical debt. Code quality boundaries are governed by `golangci-lint` generation **v2.x**. To run the static analysis suite locally from the root:

```bash
make lint
```

## 📄 License

Distributed under the terms of the **Apache License 2.0**. This license grants an irrevocable, perpetual, worldwide right to use the software while explicitly defending the repository from patent traps and intellectual property disputes. See [LICENSE](./LICENSE) for details.

---
