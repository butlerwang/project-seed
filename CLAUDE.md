# CLAUDE.md — project-seed

This is a GitHub template. After cloning, run `make rename NAME=yourapp` then update this file.

## Project Overview

<!-- FILL IN: What does this product do? Who is it for? -->

## Current Status

<!-- FILL IN: What phase are we in? What was last worked on? -->

## Skill Routing

Before responding, check if any skill applies:
- `/plan` — use when designing a new feature
- `/ship` — default mode: execute, minimal questions
- `/review` — code review after implementation
- `/investigate` — debugging unknown issues
- `/qa` — test coverage check

## Stack

- Backend: Go 1.22 + chi + pgx/v5 + goose (module: `github.com/butlerwang/PROJECTNAME/backend`)
- Frontend: Next.js 15 App Router + Tailwind (Cloudflare Pages via `@cloudflare/next-on-pages`)
- Sidecar: Playwright Node.js server (`services/pdf-sidecar/`)
- DB: PostgreSQL (Neon in prod, Docker in dev) — goose migrations in `backend/migrations/`
- Storage: MinIO (dev) → Cloudflare R2 (prod)
- Auth: JWT (15min access) + refresh token (30d, HTTP-only cookie) stored in `sessions` table
- LLM: Anthropic API (Haiku fast / Sonnet capable) with Ollama fallback

## Key Conventions

- Repository pattern: always inject `repository.Repository` interface, never concrete type
- Memory fallback: server starts without DATABASE_URL using `MemoryRepository`
- Admin: set `role = 'admin'` directly in DB; `RequireAdmin()` middleware checks JWT claim
- Rename: `make rename NAME=myapp` replaces all `project-seed` occurrences

## GitHub Secrets Required

```
FLY_API_TOKEN
CLOUDFLARE_API_TOKEN
CLOUDFLARE_ACCOUNT_ID
NEXT_PUBLIC_API_URL
```

## Doc Files

- `_docs/DESIGN.md` — product spec (source of truth)
- `_docs/CODEX.md` — implementation plan for Codex
- `_docs/IMPLEMENTATION_LOG.md` — append-only status log
- `_docs/EXECUTIVE_SUMMARY.md` — phase handoff notes
