# project-seed

Full-stack SaaS template: Go backend + Next.js 15 + Cloudflare Pages + Fly.io. Clone, rename, ship.

## What's included

| Layer | Tech | Where |
|-------|------|-------|
| Backend | Go 1.22 · chi · JWT · pgx/v5 · goose | `backend/` |
| Frontend | Next.js 15 · Tailwind · App Router | `frontend/` |
| Sidecar | Playwright: PDF gen + scraping | `services/pdf-sidecar/` |
| DB | PostgreSQL + goose migrations | `backend/migrations/` |
| Storage | MinIO (dev) → Cloudflare R2 (prod) | |
| Auth | JWT (15min) + refresh tokens (30d) | |
| LLM | Anthropic (Haiku/Sonnet) + Ollama fallback | `backend/internal/llm/` |
| CI | Build + test on PRs | `.github/workflows/ci.yml` |
| Deploy | Fly.io + Cloudflare Pages on push to main | `.github/workflows/deploy.yml` |

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
open http://localhost:3000   # frontend
open http://localhost:8080/health  # backend
```

## GitHub Secrets (for CI/CD)

Set these in your repo → Settings → Secrets and variables → Actions:

| Secret | Where to get it |
|--------|----------------|
| `FLY_API_TOKEN` | `fly auth token` |
| `CLOUDFLARE_API_TOKEN` | Cloudflare dashboard → API Tokens → Pages:Edit |
| `CLOUDFLARE_ACCOUNT_ID` | Right sidebar in Cloudflare dashboard |
| `NEXT_PUBLIC_API_URL` | `https://myapp-api.fly.dev` (after first Fly deploy) |

## Deployment

After setting secrets, push to `main` — the deploy pipeline runs automatically:
1. Tests gate (Go tests)
2. Backend → Fly.io (`fly.toml`)
3. Sidecar → Fly.io (`fly.sidecar.toml`)
4. Frontend → Cloudflare Pages

Steps 2–4 run in parallel after step 1 passes.

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
  docker-compose.yml Dev environment (postgres + minio + ollama profile)
  fly.toml           Backend deploy config
  fly.sidecar.toml   Sidecar deploy config
  Makefile           make up / down / test / migrate / rename
  .env.example       Copy to .env and fill in
```
