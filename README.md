# project-seed

Full-stack SaaS template: Go backend + Next.js 15 + Cloudflare Workers + Fly.io. Clone, rename, ship.

This repo is a project factory, not just a pile of boilerplate. New projects should be generated from a business idea, converted into a capability manifest, approved by the owner, then scaffolded and verified before Codex starts implementation.

## What's included

| Layer | Tech | Where |
|-------|------|-------|
| Backend | Go 1.24 · chi · JWT · pgx/v5 · goose | `backend/` |
| Frontend | Next.js 15 · Tailwind · App Router | `frontend/` |
| Sidecar | Playwright: PDF gen + scraping | `services/pdf-sidecar/` |
| DB | PostgreSQL + goose migrations | `backend/migrations/` |
| Storage | MinIO (dev) → Cloudflare R2 (prod) | |
| Auth | JWT (15min) + refresh tokens (30d) | |
| LLM | Anthropic (Haiku/Sonnet) + Ollama fallback | `backend/internal/llm/` |
| CI | Build + test on PRs | `.github/workflows/ci.yml` |
| Deploy | Fly.io + Cloudflare Workers on push to main | `.github/workflows/deploy.yml` |

## Factory workflow

1. CEO intake produces the business case and v1 scope.
2. Engineering review converts the idea into `project.yaml` capability flags.
3. Owner approves the product and architecture lock before code changes.
4. Template is cloned, renamed, stripped to the selected capabilities, then verified.
5. Codex executes from `_docs/CODEX.md`, appending to the log after every task.

The docs are part of the product:

| File | Role |
|------|------|
| `project.yaml` | Machine-readable capability manifest and confidence gates |
| `_docs/DESIGN.md` | Product source of truth; do not change without owner review |
| `_docs/CODEX.md` | Codex execution plan and verification contract |
| `_docs/IMPLEMENTATION_LOG.md` | Append-only task history |
| `_docs/EXECUTIVE_SUMMARY.md` | Phase handoff summary |

## Quick start

```bash
# 1. Create new repo from this template
gh repo create myapp --template butlerwang/project-seed --private --clone

# 2. Rename all occurrences of "project-seed"
cd myapp
make rename NAME=myapp

# 3. Copy env and configure
cp .env.example .env
# Edit .env — at minimum set JWT_SECRET and ANTHROPIC_API_KEY

# 4. Start everything
make up

# 5. Open
open http://localhost:23000   # frontend
open http://localhost:28080/health  # backend
```

## GitHub Secrets (for CI/CD)

Set these in your repo → Settings → Secrets and variables → Actions:

| Secret | Where to get it |
|--------|----------------|
| `FLY_API_TOKEN` | `fly auth token` |
| `CLOUDFLARE_API_TOKEN` | Cloudflare dashboard → API Tokens → Workers:Edit |
| `CLOUDFLARE_ACCOUNT_ID` | Right sidebar in Cloudflare dashboard |
| `NEXT_PUBLIC_API_URL` | `https://myapp-api.fly.dev` (after first Fly deploy) |

## Deployment

After setting secrets, push to `main` — the deploy pipeline runs automatically:
1. Tests gate (Go tests)
2. Backend → Fly.io (`fly.toml`)
3. Sidecar → Fly.io (`fly.sidecar.toml`)
4. Frontend → Cloudflare Workers

Steps 2–4 run in parallel after step 1 passes.

## Optional mobile client

Do not include mobile by default. Add it only when the product needs app-store distribution or native device capabilities.

| Need | Default |
|------|---------|
| Mobile-friendly dashboard/forms | Responsive web or PWA |
| iOS/Android app with shared product surface | Expo React Native |
| Push, camera, location, offline drafts | Expo React Native development build |
| Mobile-first custom UI or heavy animation | Flutter |
| Deep OS integration, widgets, watch, AR | Native Swift/Kotlin review |

When mobile is enabled, keep the backend as the system of record and add `apps/mobile`, `packages/api-client`, and `packages/design-tokens`. Do not share browser auth code or Next.js components with the mobile app.

## Confidence policy

The factory is not allowed to claim confidence from preference alone. It earns confidence by passing the verification gate: Go tests, frontend type-check/build, OpenNext build, Docker builds, compose smoke tests, and deploy dry-run where credentials allow it. Any skipped check must be written into the handoff as residual risk.

## Estimated cost

| Stage | Monthly |
|-------|---------|
| Early (< 50 users) | ~$0 (all free tiers) |
| Growing (50–500 users) | ~$6–25/mo |
| Scale (500+ users) | Neon Launch $19 + Fly.io ~$10–20 |

## File map

```
project-seed/
  backend/           Go API (auth, LLM, health)
  frontend/          Next.js 15 (landing, login, register, dashboard, admin)
  services/
    pdf-sidecar/     Playwright: /health /generate-pdf /scrape
  migrations/        goose SQL migrations
  _docs/             Design + implementation doc templates (fill in per project)
  project.yaml       Capability manifest and factory verification gates
  docker-compose.yml Dev environment (postgres + minio + ollama profile)
  fly.toml           Backend deploy config
  fly.sidecar.toml   Sidecar deploy config
  Makefile           make up / down / test / migrate / rename
  .env.example       Copy to .env and fill in
```
