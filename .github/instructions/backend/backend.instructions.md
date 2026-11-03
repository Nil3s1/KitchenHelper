---
description: "Go backend implementation rules for context packages, ports/adapters, CQRS, transactions, shared capabilities, container builds, and tests."
applyTo: "src/backend/**, tests/backend/**"
---

# Go Backend Instructions

Follow the system boundaries and technology decisions in the [architecture instructions](../architecture/architecture.instructions.md). This file defines backend-specific implementation conventions. Business invariants belong in the owning context's [domain instructions](../domain/).

## 1. Context and Package Ownership

- Each bounded context lives in `src/backend/internal/contexts/<context>/` and owns its domain behavior, use cases, persistence adapters, and transport adapters.
- Keep contexts as independent Go packages. A context must not import another context's package or internal implementation.
- Keep a context's files in a flat package while it remains understandable; split packages only when a clear package boundary improves ownership or dependency direction.
- Follow idiomatic Go naming and visibility. File prefixes below are repository conventions for readability, not substitutes for package boundaries.
- Keep process startup and application-wide composition in `cmd/<entrypoint>/`; keep context-local dependency wiring in `module.go`.

## 2. File Naming Conventions

Use these prefixes for files in a flat context package:

- `dom_*.go`: Domain entities, value objects, invariants, and domain-owned contracts.
- `app_cmd_<action>_<entity>.go`: A write use case and its input/handler.
- `app_qry_<action>_<entity>.go`: A read use case and its query DTO/handler.
- `inf_repo_pg_*_command.go`: PostgreSQL command-side persistence adapters and database-specific models.
- `inf_repo_pg_*_query.go`: PostgreSQL query-side persistence adapters and query projections.
- `inf_repo_sqlite_*.go`: SQLite adapters, only when that database is implemented.
- `inf_handler_http_*.go`: HTTP transport adapters and transport DTOs.
- `inf_handler_mcp_*.go`: MCP transport adapters, when required.
- `module.go`: Context-local dependency wiring.

Do not create placeholder files or adapters for technologies that are not in use.

## 3. Domain and Mapping Boundaries

- Domain types contain business concepts and invariants; they do not depend on database drivers, HTTP frameworks, serialization formats, or infrastructure.
- In Go domain files (`dom_*.go`), keep domain structs free of transport and persistence struct tags.
- Do not use domain entities directly as HTTP request/response DTOs or database models.
- Map transport input to application input and application results to transport output at explicit boundaries.
- Command use cases should create or load domain aggregates when needed to enforce business invariants, then persist them through command-side ports.
- Query use cases should retrieve purpose-built query DTOs/projections directly from query-side ports. Do not load and map domain aggregates for a read that does not invoke domain behavior.
- Define query DTOs with the fields and shape the read use case needs. When the query is specific to one transport and the API contract matches, the query DTO may also serve as that transport's response type; otherwise map it at the transport boundary.
- Map domain aggregates to persistence models and back for command-side operations that need domain behavior. Query-side persistence should project directly into its DTO and must not perform an unnecessary persistence-to-domain-to-DTO round trip.
- Keep domain types free of serialization and database tags. Put transport serialization details on transport-facing DTOs and database-specific representations on persistence models; a query DTO intentionally used as a transport response is transport-facing.

## 4. Application Ports and Adapters

- Application use cases coordinate domain behavior and express required external capabilities through interfaces (ports).
- Define repository contracts centrally within their owning context, not separately inside each use case. Keep the aggregate-oriented command repository contract with the domain and the query repository contract with the application read model. Other external-service ports normally belong to the application layer.
- Infrastructure adapters implement those interfaces. Application code must not depend on concrete database, network, or transport implementations.
- Define one complete command repository interface and one complete query repository interface per bounded context. Each contains all persistence methods needed across that context's respective write or read use cases. Do not fragment repositories into per-use-case or single-method interfaces, adapter types, or files.
- The command repository includes all operations needed by write workflows, including loading aggregates, existence checks, and concurrency checks as well as persistence. Command/query separation does not forbid reads required to perform a command correctly.
- The query repository includes all required read operations, such as lookup by ID, listing, filtering, and projections. It returns query DTOs/projections rather than domain aggregates unless the read actually requires domain behavior.
- Add methods when required by implemented behavior, not speculative CRUD completeness. Extend the existing repository for that side instead of introducing another small repository.
- Inject the context's complete command repository into command handlers and its complete query repository into query handlers; do not wrap them in narrower repository interfaces for individual consumers.
- Keep command-side and query-side repositories logically distinct even when both use the same database, schema, table, or persistence technology. Their implementations may share a connection or low-level helpers. A shared physical database does not require a separate read database.
- Separate read storage or projections are optional scaling/consistency choices, not prerequisites for command/query separation.
- Keep transport-specific validation and decoding in the transport adapter; keep business invariants in the domain and enforce them for every use case.
- Prefer one clear handler per distinct use case. Do not combine unrelated commands or queries into a large handler merely to reduce file count.

## 5. Transactions and Unit of Work

- Use a transaction only when a use case needs atomic multi-operation behavior or transactional consistency/concurrency protection.
- Keep a Unit of Work contract context-local because its repositories and transaction semantics are context-specific.
- When a Unit of Work is needed, expose the context's complete command repository within the transaction scope, not a new per-use-case subset; never expose `sql.Tx` or another driver transaction type. Do not combine command and query repositories merely because they share a database.
- A closure-based `Do(ctx, work)` API is a suitable Go option when it makes begin, commit, and rollback behavior explicit and reliable. It is not mandatory for use cases that need no transaction.
- Ensure failed work rolls back, successful work commits, and rollback failures remain observable in error reporting.

## 6. Shared Backend Capabilities (`internal/shared`)

Use `src/backend/internal/shared/` for context-independent backend behavior and explicitly shared concepts. This placement depends on responsibility, not the number of bounded contexts already implemented: a capability with a clear application-wide contract may belong here from its first use.

- Organize shared code into focused Go packages such as `shared/httpresponse`, `shared/auth`, or `shared/messaging` when those capabilities are needed. Do not create empty packages or a catch-all utilities package.
- Put context-independent HTTP response/error encoding, authentication middleware, logging, messaging infrastructure, or low-level database helpers here when needed. Context adapters should use the shared implementation rather than duplicate the same technical policy.
- Keep shared domain concepts (the Shared Kernel) separate from technical helpers in appropriately named packages under `shared/`. A shared identifier or principal needs an explicit common meaning; similar-looking context types are not sufficient reason to merge them.
- Shared packages must not import bounded contexts or contain their business decisions. Technical packages may use approved infrastructure dependencies; shared domain concepts must remain independent of transport, persistence, and frameworks.
- Preserve dependency direction: context adapters may use shared technical packages, but domain and application code must not import shared infrastructure to bypass their ports.
- Keep each context's domain rules, DTOs, repository contracts, and transaction contracts in that context.
- Context adapters map domain/application errors to shared transport-level error contracts; shared response helpers must not know recipe errors or other context-specific types.
- Keep shared contracts small and testable. Implement capabilities needed now with context-neutral inputs; do not speculate about future consumers or require a second context before using `shared/`.
- Generic repositories and a universal Unit of Work do not belong in `shared`.
- A common `DomainModel` is optional, not a required base class. Add only fields whose meaning, lifecycle, and rules are genuinely identical for all consumers. Keep aggregate-specific versioning in the aggregate unless the same version semantics are established across contexts.

## 7. Build and Runtime Configuration

- Use the Go standard library and approved modules; keep backend dependencies in `src/backend/go.mod`.
- Keep `src/backend/Dockerfile` a multi-stage build that compiles the required entrypoint binaries and copies them into a minimal runtime image.
- Read runtime connection settings and listen addresses from environment configuration. Validate required configuration at startup; keep secrets out of code and images.

## 8. Tests

- Keep Go unit tests next to the context package and the code they exercise.
- Test domain invariants without database or transport dependencies.
- Test application use cases through their ports, using small test doubles where useful.
- Test command-side persistence mapping separately from database integration behavior.
- Test query-side projections directly, including that query operations return the expected DTO without requiring domain aggregate mapping.
- Test shared packages independently of any bounded context; test context-specific mapping in the owning adapter.
- Put tests requiring PostgreSQL, multiple contexts, or the complete running application under `tests/backend/` or the corresponding central `tests/` area.
- Use test-first development where practical, while prioritizing meaningful behavioral coverage over prescribed test-writing order.
