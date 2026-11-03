---
description: "Repository-wide architecture, bounded-context boundaries, local orchestration, dependency policy, and test placement."
applyTo: "**"
---

# Repository Architecture Instructions

These principles describe responsibilities and dependency boundaries, not a required programming language or framework. Apply the language-neutral guidance first; implementation details belong in the relevant technology-scoped instructions.

## 1. Repository Layout

Keep application code and container configuration in the repository structure below. Adapt individual files and test folders to the chosen stack without changing ownership boundaries.

```text
/
├── docker-compose.yml
├── src/
│   ├── database/
│   │   ├── migrations/
│   │   └── ...
│   ├── backend/
│   │   ├── cmd/
│   │   ├── internal/
│   │   │   ├── shared/
│   │   │   ├── platform/          # Optional shared technical capabilities
│   │   │   └── contexts/<context>/
│   │   └── ...
│   └── frontend/
│       └── src/
│           ├── app/
│           ├── shared/
│           └── contexts/<context>/
└── tests/
    ├── database/
    ├── backend/contexts/<context>/
    └── frontend/contexts/<context>/
```

Place unit tests near the code they exercise when supported by the language and tooling. Keep cross-component integration, end-to-end, and database migration tests under the central `tests/` tree.

## 2. Technology Stack

The current implementation uses:

- **Database:** PostgreSQL, with a schema per bounded context.
- **Database migrations:** Go migration runner using `golang-migrate`.
- **Backend:** Go, primarily using the standard library and minimizing external dependencies.
- **Frontend:** TypeScript, React, and Redux Toolkit.

Do not introduce a new language, framework, or database technology without a deliberate architecture change.

### Dependency Approval

External third-party dependencies may be added only after explicit user approval. Before changing a dependency manifest or running a package installation command, explain why the dependency is needed and ask for approval. Prefer standard-library or already-approved alternatives where they meet the requirement.

## 3. Local Container Orchestration

The root `docker-compose.yml` owns the local application stack. Keep service startup dependencies explicit:

- PostgreSQL has a health check.
- The migration service starts only after PostgreSQL is healthy.
- The backend starts only after migrations complete successfully.
- The frontend is built and served as a separate service.

Keep service-specific build and runtime configuration in that service's Dockerfile. Do not rely on startup ordering alone as a substitute for health or completion conditions.

## 4. Bounded Contexts and Dependency Direction

- Each bounded context owns its domain rules, application use cases, persistence mapping, and external adapters.
- A context must not import another context's implementation or reach into its internal data structures.
- Cross-context collaboration uses an explicit contract at the boundary, such as a stable shared identifier, an application-facing interface, or an integration message. Choose the simplest mechanism that fits the interaction.
- Shared abstractions must not create reverse dependencies from shared code into a context.
- Domain rules do not depend on transport, persistence, framework, or user-interface details. Application use cases coordinate domain behavior through ports; infrastructure adapters implement those ports.
- Keep presentation and transport concerns at their respective boundaries. Translate external input into application-level input and map results into the external response format.

The backend-specific rules for the Shared Kernel and reusable technical capabilities are defined in `../backend/backend.instructions.md`.

## 5. Transactions and Consistency

Use a transaction when a use case needs multiple changes to succeed or fail atomically, or when its consistency and concurrency requirements need transactional protection. Do not require a transaction wrapper for every command by default.

Keep transaction boundaries behind an application-facing port or context-owned Unit of Work. Application and domain code must not depend on database-driver transaction types. A context that does not need multi-operation atomicity should use the simplest suitable persistence contract.

## 6. Testing

- Test domain invariants and application behavior at the owning context boundary.
- Put unit tests near the code under test where the language and test tooling make that practical.
- Put database, cross-context integration, and end-to-end tests under the central `tests/` tree.
- Use integration tests for database behavior that cannot be established by mapping or unit tests alone.
- Prefer behavior-focused tests; strict test-first ordering is encouraged where practical, not a substitute for meaningful coverage.

Technology-scoped instruction files may specify the exact test locations and commands for their stack.
