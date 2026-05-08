# CLAUDE.md — project-seed

This is a GitHub template. After cloning, run `make rename NAME=yourapp` then update this file.

## Project Overview

<!-- FILL IN: What does this product do? Who is it for? -->

This template implements the standard project factory strategy: business intake -> capability manifest -> owner approval -> scaffold/strip -> verification -> Codex execution. Do not treat the template defaults as product requirements; the generated `project.yaml` decides which capabilities survive into a real project.

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

- Backend: Go 1.24 + chi + pgx/v5 + goose (module: `github.com/butlerwang/PROJECTNAME/backend`)
- Frontend: Next.js 15 App Router + Tailwind (Cloudflare Workers via OpenNext)
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
- Capability manifest: read `project.yaml` before designing or implementing features
- Architecture lock: `_docs/DESIGN.md` is source of truth and requires owner review before changes
- Confidence gate: do not call work complete until tests/builds/Docker/smoke checks are run or skipped with a recorded reason
- Mobile default: no mobile client unless `project.yaml` enables it; use Expo React Native first for normal iOS/Android app needs

## GitHub Secrets Required

```
FLY_API_TOKEN
CLOUDFLARE_API_TOKEN
CLOUDFLARE_ACCOUNT_ID
NEXT_PUBLIC_API_URL
```

## Doc Files

- `project.yaml` — machine-readable capability manifest and verification gates
- `_docs/DESIGN.md` — product spec (source of truth)
- `_docs/CODEX.md` — implementation plan for Codex
- `_docs/IMPLEMENTATION_LOG.md` — append-only status log
- `_docs/EXECUTIVE_SUMMARY.md` — phase handoff notes
