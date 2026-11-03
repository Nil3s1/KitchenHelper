---
description: "Language-neutral business rules and invariants for the Recipes bounded context."
applyTo: "src/backend/internal/contexts/recipes/**, src/frontend/src/contexts/recipes/**, src/database/migrations/*_recipes_*.sql, tests/backend/contexts/recipes/**, tests/frontend/contexts/recipes/**, tests/database/**/*recipes*"
---

# Recipes Domain Rules

This file defines business meaning and behavior, not a required programming language, framework, persistence format, or UI design. Technical representations belong in the relevant backend, frontend, or database instructions.

## 1. Identity and Lifecycle

- The Recipes context owns the identity policy for recipes. New recipes receive a unique identifier when they are created.
- A new recipe starts at version `1`.
- Every successful update operation increments the recipe's version by exactly `1`, including a replacement whose values happen to match the current recipe.
- The recipe retains its identity and creation time across updates; its update time reflects the successful modification.

## 2. Recipe Invariants

- A recipe must have a non-empty name after trimming surrounding whitespace.
- An ingredient must have a non-empty name and unit after trimming surrounding whitespace.
- Ingredient amounts must be finite and non-negative; zero is valid.
- Ingredient names must be unique within a recipe, ignoring surrounding whitespace and case.
- Ingredients are recipe-owned values consisting of a name, amount, and unit. They do not refer to storage items through identifiers or foreign keys.
- Image values are references to image assets. The Recipes context does not own binary upload, storage, or retrieval behavior.

## 3. Recipe Information

A recipe may contain:

- `name` (required)
- `description` (optional)
- `author` (optional)
- `version` (required, starts at `1`)
- `text` (optional preparation instructions)
- `tags` (zero or more labels)
- `ingredients` (zero or more name/amount/unit values)
- `images` (zero or more image references)

The domain rules do not prescribe how these values are stored or serialized. An external representation may use different naming or formatting as long as it preserves the domain semantics.
