---
description: "Frontend architecture guidance for context ownership, application composition, and the repository's TypeScript, React, and Redux Toolkit stack."
applyTo: "src/frontend/**, tests/frontend/**"
---

# Frontend Instructions

The architectural principles in `../architecture/architecture.instructions.md` are language-neutral. This file applies them to the current TypeScript, React, and Redux Toolkit frontend.

## 1. Feature and Context Ownership

Each frontend bounded context owns its user-facing features and context-specific API and state integration:

```text
src/frontend/src/contexts/<context>/
├── api/
├── state/
├── components/
├── pages/
└── index.ts
```

Add folders when the context needs them; do not create empty structural placeholders.

## 2. Context Isolation and Public APIs

- A context must not import another context's internal components, state, or API implementation.
- Expose only the functionality needed by the application shell or other approved consumers through the context's `index.ts`.
- The app shell may compose context public APIs; it must not reach into context internals.
- Share UI or client utilities only when multiple contexts genuinely need the same behavior. Keep context-specific behavior in its owning context.

## 3. Application Shell Responsibilities

- Keep `app/` focused on application-wide composition, such as the root component, routing, and store configuration.
- Context pages and feature components own context-specific presentation and interactions.
- Keep API transport, feature state, and presentation concerns in their appropriate layers. Translate server responses at the API boundary when the server contract differs from the frontend's internal model.
- Do not move feature UI into the app shell merely to simplify initial wiring; the shell should compose a feature page through its public API.

## 4. TypeScript, React, and Redux Toolkit

- Use TypeScript types for API contracts, state, component props, and public context APIs.
- Use React components for presentation and interaction; keep domain and feature behavior out of unrelated application-wide components.
- Use Redux Toolkit for shared client state and asynchronous workflows when state must be coordinated across the application. Do not put every transient or server-owned value into global state without a need.
- Keep Redux slices, async actions, selectors, and related state types within their owning context.
- Keep API request functions in the owning context's `api/` area and surface user-visible failures through the established UI state and error handling.

## 5. Tests

- Keep focused frontend unit/component tests close to the context code where the test tooling supports it.
- Put cross-context or browser-level end-to-end tests under `tests/frontend/`.
- Test observable behavior and public context APIs rather than coupling tests to another context's internals.
