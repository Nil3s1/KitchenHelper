---
description: "Repository-wide technology decisions, component interaction, bounded-context boundaries, local orchestration, dependency policy, and development workflow."
applyTo: "**"
---

# Repository Architecture Instructions

This file defines the overall system, technology decisions, component responsibilities, and development workflow. Implementation conventions belong in the component instructions; business rules belong in the owning bounded context's domain instructions.

## Instruction Responsibilities

- [Backend instructions](../backend/backend.instructions.md): Go packages, ports/adapters, command/query implementations, transactions, and shared backend capabilities.
- [Frontend instructions](../frontend/frontend.instructions.md): TypeScript, React, Redux Toolkit, UI composition, API clients, and frontend state.
- [Database instructions](../database/database.instructions.md): PostgreSQL schemas, SQL migrations, migration runner, and database validation.
- [Domain instructions](../domain/): Technology-neutral business meaning and invariants, one file per bounded context. Apply them across the context's backend, frontend, persistence changes, and tests.

Keep each rule at its owning level. Reference other instruction files rather than duplicating their detailed rules.

## 1. Repository Layout

Keep application code and container configuration in the repository structure below. Create packages and folders only when needed.

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

The backend's `internal/shared/` owns context-independent backend capabilities and explicitly shared concepts. Context-specific code stays in `contexts/<context>/`; detailed shared-package rules belong in the backend instructions.

## 2. Technology Stack

The current implementation uses:

- **Database:** One PostgreSQL database, with a schema per bounded context.
- **Database migrations:** Go migration runner using `golang-migrate`.
- **Backend:** Go, primarily using the standard library and minimizing external dependencies.
- **Frontend:** TypeScript, React, and Redux Toolkit.
- **Frontend build and serving:** Vite builds the SPA; Nginx serves static assets and proxies API requests.
- **Local object storage:** MinIO provides an S3-compatible service.
- **Local orchestration:** Docker Compose builds and runs the component containers.

Do not introduce a new language, framework, or database technology without a deliberate architecture change.

### Dependency Approval

External third-party dependencies may be added only after explicit user approval. Before changing a dependency manifest to introduce a dependency or installing a new package, explain why it is needed and ask for approval. Prefer standard-library or already-approved alternatives where they meet the requirement.

## 3. Component Responsibilities and Interaction

The application is a modular monolith: bounded contexts are logical ownership boundaries within the backend, not separately deployed services.

```text
Browser -> Frontend (Nginx: SPA + /api proxy) -> Go backend
                                                   |
                                           Context use cases
                                                   |
                                           PostgreSQL adapters -> PostgreSQL

Migration runner ---------------------------------------------> PostgreSQL
MinIO: local object storage, available but not yet wired into application uploads.
```

- The frontend owns presentation and client-side interaction; the backend authoritatively enforces business rules.
- The backend composes context modules and their adapters. Application data access goes through the backend, never directly from the browser to PostgreSQL.
- PostgreSQL stores context-owned data. The migration service owns versioned schema changes, not request processing.
- MinIO is available locally, but its presence does not imply an implemented upload API. Recipe image references do not themselves implement asset storage or retrieval.
- Future contexts fit the same boundaries; do not make architecture or shared-code rules depend on the current number of contexts.

## 4. Local Container Orchestration

The root [docker-compose.yml](../../../docker-compose.yml) owns the local application stack. Keep service startup dependencies explicit:

- PostgreSQL has a health check.
- The migration service starts only after PostgreSQL is healthy.
- The backend starts only after migrations complete successfully.
- The frontend is built and served as a separate service.
- MinIO runs as a separate local storage service; add application startup dependencies only when an implemented integration requires them.

Keep service-specific build and runtime configuration in that service's Dockerfile. Do not rely on startup ordering alone as a substitute for health or completion conditions.

## 5. Bounded Contexts and Dependency Direction

- Each bounded context owns its domain rules, application use cases, persistence mapping, and external adapters.
- A context must not import another context's implementation or reach into its internal data structures.
- Logical context ownership is independent of physical deployment. Multiple contexts may use schemas in the same database; a shared database does not grant contexts ownership of one another's tables or schemas.
- Cross-context collaboration uses an explicit contract at the boundary, such as a stable shared identifier, an application-facing interface, or an integration message. Choose the simplest mechanism that fits the interaction.
- Shared abstractions must not create reverse dependencies from shared code into a context.
- Domain rules do not depend on transport, persistence, framework, or user-interface details. Application use cases coordinate domain behavior through ports; infrastructure adapters implement those ports.
- Keep presentation and transport concerns at their respective boundaries. Translate external input into application-level input and map results into the external response format.

Command/query separation is logical, not a requirement for separate databases or deployments. Keep consistency guarantees explicit at context boundaries; detailed repository and transaction contracts belong in the backend instructions.

## 6. Development Workflow and Verification

1. Read these instructions, the affected component instructions, and the owning context's domain instructions before changing behavior. For cross-component changes, load every affected instruction file.
2. Keep domain rules, API contracts, persistence mapping/migrations, and frontend behavior consistent across the affected surfaces. Do not change business meaning to accommodate an adapter.
3. Keep changes in their owning context or shared capability. Use `internal/shared/` for context-independent backend behavior without waiting for a second context; do not build unused future features.
4. Run the smallest relevant tests and builds using the existing tooling. Keep unit tests near their code; put database, cross-context integration, and end-to-end tests under `tests/`.
5. For migration changes, test apply/rollback/reapply against PostgreSQL as described in the database instructions. A skipped database test is not a successful integration check.
6. For cross-component behavior, verify the affected flow through the running local stack. Distinguish an image build or successful startup from a working end-to-end request.
7. Keep the practical startup, build, and test commands in [README.md](../../../README.md) current. Report checks that could not run and missing prerequisites explicitly.
