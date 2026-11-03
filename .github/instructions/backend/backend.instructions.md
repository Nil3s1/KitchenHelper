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
- Export the context's application entry point and the command inputs and result types required to call it from other permitted backend packages. Keep implementation handlers, repository ports, persistence types, and transport DTOs unexported unless an actual external consumer requires them. Exported JSON fields alone do not make an unexported type a usable public API.
- Keep process startup and application-wide composition in `cmd/<entrypoint>/`; keep context-local dependency wiring and its small public application facade in `module.go`.
- A context must remain callable without HTTP or MCP. Use a public application API such as `NewModule(db) *Module` with context-aware operations; do not make an `http.Handler` the context's only public entry point.
- Exposing an application API does not permit direct context-to-context imports. Go's `internal/` visibility and the architecture's context-isolation rules still apply.

## 2. File Naming Conventions

Use these prefixes for files in a flat context package:

- `dom_*.go`: Domain entities, value objects, invariants, and domain-owned contracts.
- `app_cmd_<action>_<entity>.go`: A write use case and its input/handler.
- `app_command.go`: Shared command inputs/mapping and application-owned command ports, including a context-local Unit of Work when needed.
- `app_query.go`: Related query ports and shared read-model DTOs. Do not hide shared declarations in one individual query's file; split ports and models into focused files only when their size warrants it.
- `app_qry_<action>_<entity>.go`: One read use case and its concrete application handler, with a named query input only when needed.
- `inf_repo_pg_*_command.go`: PostgreSQL command-side persistence adapters and database-specific models.
- `inf_repo_pg_*_query.go`: PostgreSQL query-side persistence adapters and query projections.
- `inf_repo_sqlite_*.go`: SQLite adapters, only when that database is implemented.
- `inf_handler_http_*.go`: HTTP transport adapters and transport DTOs.
- `inf_handler_mcp_*.go`: MCP transport adapters, when required.
- `module.go`: Context-local dependency wiring and public application methods. Keep transport setup in the owning adapter file.

Do not create placeholder files or adapters for technologies that are not in use.

## 3. Domain and Mapping Boundaries

- Domain types contain business concepts and invariants; they do not depend on database drivers, HTTP frameworks, serialization formats, or infrastructure.
- In Go domain files (`dom_*.go`), keep domain structs free of transport and persistence struct tags.
- Do not use domain entities directly as HTTP request/response DTOs or database models.
- Map transport input to application input and application results to transport output at explicit boundaries.
- Command use cases should create or load domain aggregates when needed to enforce business invariants, then persist them through command-side ports.
- Query use cases should retrieve purpose-built query DTOs/projections directly from query-side ports. Do not load and map domain aggregates for a read that does not invoke domain behavior.
- Every distinct command and query use case has a concrete application handler, including simple reads that initially only delegate to a query port. This is a repository architecture convention, not a Go language requirement. Public application methods delegate to these handlers, never directly to repository ports.
- Use query inputs according to need: `Handle(ctx, id)` for a single lookup and `Handle(ctx)` for a parameterless list are appropriate. Introduce a named query struct for related filters, sorting, pagination, or multiple parameters when it improves clarity; do not require empty query structs.
- Transport adapters call the application API, not repositories or implementation handlers. Keep SQL and concrete persistence dependencies out of transport adapters and application handlers.
- Keep application operations independent of request/response transport types. HTTP, MCP, and direct Go callers use the same operations so that domain validation, transactions, and any application policies cannot be bypassed by choosing another transport.
- Expose transport integration separately, for example `Module.HTTPHandler() http.Handler` implemented in the HTTP adapter file. Creating a module or obtaining a handler must not start a server. Do not add a transport adapter until it is needed.
- Use an ordinary DTO slice for a list unless a separate result type adds metadata or meaningful behavior.
- Define public query DTOs with the fields and shape application callers need. A DTO may also serve as a transport response when the contracts match; otherwise map it at the transport boundary.
- Map domain aggregates to persistence models and back for command-side operations that need domain behavior. Query-side persistence should project directly into its DTO and must not perform an unnecessary persistence-to-domain-to-DTO round trip.
- Keep domain types free of serialization and database tags. Put transport serialization details on transport-facing DTOs and database-specific representations on persistence models; a query DTO intentionally used as a transport response is transport-facing.

## 4. Application Ports and Adapters

- Application use cases coordinate domain behavior and express required external capabilities through interfaces (ports).
- Define ports within their owning context at the layer that consumes them. Application repository and transaction contracts belong in application files, not domain files merely because they return aggregates. Keep shared ports in clearly named common files; a genuinely consumer-specific port may live with its use case.
- Infrastructure adapters implement those interfaces. Application code must not depend on concrete database, network, or transport implementations.
- Keep interfaces focused on implemented consumer needs. Related operations may share a command or query repository interface; use narrower interfaces when consumers have meaningfully different needs. Neither a complete context-wide interface nor an interface per use case is mandatory. One adapter can satisfy multiple interfaces without duplicate implementations.
- The command repository includes all operations needed by write workflows, including loading aggregates, existence checks, and concurrency checks as well as persistence. Command/query separation does not forbid reads required to perform a command correctly.
- Query ports return DTOs/projections for lookup, listing, filtering, and other implemented reads. Reads that actually need domain behavior should use an application use case with an appropriate aggregate-loading port instead.
- Add methods when required by implemented behavior, not speculative CRUD completeness. Reuse an existing port when it remains cohesive; do not add redundant wrapper interfaces or duplicate adapters.
- Inject repository ports into application handlers. Transport adapters consume the public application API. Application behavior must not depend on concrete infrastructure; constructor wiring may choose concrete adapters.
- Keep the public facade small and explicit: it delegates both commands and queries to their respective handlers. Constructors wire repositories into handlers; the facade does not retain repository ports for direct use by its operations. Keep implementation handlers private and their application behavior out of the facade.
- Concrete structs with constructors and `Handle` methods are sufficient. Do not add a generic mediator, base handler, handler interface per use case, or separate repository implementation per query without a concrete requirement.
- Database ownership stays with the caller that opens it. Module construction wires the existing connection without opening, closing, or taking ownership of it; make the selected persistence technology explicit rather than claiming the constructor is database-agnostic.
- Keep command-side and query-side repositories logically distinct even when both use the same database, schema, table, or persistence technology. Their implementations may share a connection or low-level helpers. A shared physical database does not require a separate read database.
- Separate read storage or projections are optional scaling/consistency choices, not prerequisites for command/query separation.
- Keep transport-specific validation and decoding in the transport adapter; keep business invariants in the domain and enforce them for every use case.
- Keep one focused handler per distinct implemented use case. A delegation-only handler is allowed to preserve the uniform command/query structure; do not invent business logic to justify it or combine unrelated use cases into a large handler.

## 5. Transactions and Unit of Work

- Use a transaction only when a use case needs atomic multi-operation behavior or transactional consistency/concurrency protection.
- Keep a Unit of Work contract application-owned and context-local because its repositories and transaction semantics are context-specific.
- When a Unit of Work is needed, expose the command port needed by the transactional workflow within its scope; a cohesive shared command repository is appropriate when it fits. Never expose `sql.Tx` or another driver transaction type. Do not combine command and query repositories merely because they share a database.
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
- Test direct application API calls independently of transport setup, including domain validation, command results, query DTOs, context/parameter propagation, and error propagation.
- Include tests from an external Go test package (`<context>_test`) to verify that the constructor, commands, results, and transport attachment are actually accessible to callers without relying on unexported identifiers.
- Test query handlers through their query ports, including DTO results, context/parameter propagation, list ordering and empty collections, and error propagation.
- Test transport routes through the same module API and handlers backed by port test doubles: request context and parameters, response DTO/JSON shape, list ordering and empty arrays, and error mapping.
- Test shared packages independently of any bounded context; test context-specific mapping in the owning adapter.
- Put tests requiring PostgreSQL, multiple contexts, or the complete running application under `tests/backend/` or the corresponding central `tests/` area.
- Use test-first development where practical, while prioritizing meaningful behavioral coverage over prescribed test-writing order.
