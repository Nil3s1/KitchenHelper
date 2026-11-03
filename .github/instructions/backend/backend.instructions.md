---
description: "Go backend guidance for bounded-context packages, domain and application boundaries, adapters, transactions, tests, and a deliberately small Shared Kernel."
applyTo: "src/backend/**/*.go, tests/backend/**/*.go"
---

# Go Backend Instructions

The architectural principles in `../architecture/architecture.instructions.md` are language-neutral. This file defines how they are applied to the repository's Go backend.

## 1. Context and Package Ownership

- Each bounded context lives in `src/backend/internal/contexts/<context>/` and owns its domain behavior, use cases, persistence adapters, and transport adapters.
- Keep contexts as independent Go packages. A context must not import another context's package or internal implementation.
- Keep a context's files in a flat package while it remains understandable; split packages only when a clear package boundary improves ownership or dependency direction.
- Follow idiomatic Go naming and visibility. File prefixes below are repository conventions for readability, not substitutes for package boundaries.

## 2. File Naming Conventions

Use these prefixes for files in a flat context package:

- `dom_*.go`: Domain entities, value objects, invariants, and domain-owned contracts.
- `app_cmd_<action>_<entity>.go`: A write use case and its input/handler.
- `app_qry_<action>_<entity>.go`: A read use case and its result model/handler.
- `inf_repo_pg_*.go`: PostgreSQL persistence adapters and database-specific models.
- `inf_repo_sqlite_*.go`: SQLite adapters, only when that database is implemented.
- `inf_handler_http_*.go`: HTTP transport adapters and transport DTOs.
- `inf_handler_mcp_*.go`: MCP transport adapters, when required.
- `module.go`: Context-local dependency wiring.

Do not create placeholder files or adapters for technologies that are not in use.

## 3. Domain and Mapping Boundaries

- Domain types contain business concepts and invariants; they do not depend on database drivers, HTTP frameworks, serialization formats, or infrastructure.
- In Go domain files (`dom_*.go`), keep domain structs free of transport and persistence struct tags.
- Do not use domain entities directly as HTTP request/response DTOs or database models.
- Map transport input to application input, application input to domain behavior, and domain results to transport output at explicit boundaries.
- Keep application read models transport-neutral; map them to HTTP or other transport response DTOs in the relevant adapter.
- Map domain entities to persistence models and back inside the persistence adapter.
- Keep serialization tags and database-specific representations on adapter-owned types.

## 4. Application Ports and Adapters

- Application use cases coordinate domain behavior and express required external capabilities through interfaces (ports).
- Define each port near the layer that consumes it. A repository contract may live with the domain when it expresses an aggregate's persistence needs in domain terms; other external-service ports normally belong to the application layer.
- Infrastructure adapters implement those interfaces. Application code must not depend on concrete database, network, or transport implementations.
- Keep transport-specific validation and decoding in the transport adapter; keep business invariants in the domain and enforce them for every use case.
- Prefer one clear handler per distinct use case. Do not combine unrelated commands or queries into a large handler merely to reduce file count.

## 5. Transactions and Unit of Work

- Use a transaction only when a use case needs atomic multi-operation behavior or transactional consistency/concurrency protection.
- Keep a Unit of Work contract context-local because its repositories and transaction semantics are context-specific.
- When a Unit of Work is needed, expose repositories or a scoped operation to application code; never expose `sql.Tx` or another driver transaction type.
- A closure-based `Do(ctx, work)` API is a suitable Go option when it makes begin, commit, and rollback behavior explicit and reliable. It is not mandatory for use cases that need no transaction.
- Ensure failed work rolls back, successful work commits, and rollback failures remain observable in error reporting.

## 6. Shared Kernel (`internal/shared`)

The Shared Kernel exists to centralize genuinely recurring cross-context concepts and responsibilities. It should remove repeated implementations when contexts share the same need and rules; it is not a general-purpose utilities package or a place to centralize code merely because two implementations look alike.

- Add a shared concept only when at least two bounded contexts need the same concept with the same meaning and rules. Prefer waiting for demonstrated reuse over speculative abstraction.
- Shared code may depend on language and standard-library primitives, but must not import a bounded context or contain context-specific behavior.
- Keep each context's domain rules, DTOs, repository contracts, and transaction contracts in that context.
- Appropriate candidates may include a stable cross-context identifier, authentication principal, shared error classification, or event contract, once multiple contexts actually need and use it consistently.
- Reusable technical behavior such as authentication middleware, logging, messaging infrastructure, or transport/database helpers may also be centralized when it removes demonstrated duplication. Keep it in an appropriately named technical package (for example, `internal/platform/<capability>`), separate from domain concepts; do not force it into the Shared Kernel merely because it is reused.
- Reuse means sharing behavior and policy, not only moving duplicate lines. Keep context-specific decisions at the context boundary and make shared behavior explicit and testable.
- Generic repositories and a universal Unit of Work do not belong in `shared`.
- A common `DomainModel` is optional, not a required base class. Add only fields whose meaning, lifecycle, and rules are genuinely identical for all consumers. Keep aggregate-specific versioning in the aggregate unless the same version semantics are established across contexts.

## 7. Tests

- Keep Go unit tests next to the context package and the code they exercise.
- Test domain invariants without database or transport dependencies.
- Test application use cases through their ports, using small test doubles where useful.
- Test persistence mapping separately from database integration behavior.
- Put tests requiring PostgreSQL, multiple contexts, or the complete running application under `tests/backend/` or the corresponding central `tests/` area.
- Use test-first development where practical, while prioritizing meaningful behavioral coverage over prescribed test-writing order.
