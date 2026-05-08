# [Project Name] — Design Document

> Source of truth. Do not change without owner review.

## Factory Strategy

This project starts from `project-seed`, but the template is not the source of product truth. The accepted strategy is:

1. Capture the business idea and v1 scope.
2. Generate `project.yaml` as the capability manifest.
3. Review and approve `_docs/DESIGN.md` before implementation.
4. Scaffold only the required capabilities.
5. Run verification before handoff.
6. Record residual risks instead of claiming false confidence.

## Problem

<!-- What problem does this solve? For whom? -->

## Solution

<!-- One paragraph: what we're building and why it works. -->

## Users

| Persona | Description | Key need |
|---------|-------------|----------|
| | | |

## Core Features (v1)

- [ ] Feature 1
- [ ] Feature 2

## Out of Scope (v1)

- 

## Data Model

```sql
-- Key tables (abbreviated)
```

## Capability Manifest

Keep this section aligned with `project.yaml`.

| Capability | Enabled | Reason |
|------------|---------|--------|
| Auth | yes | Default for external SaaS |
| Admin | yes | Customer support and user maintenance |
| Storage | TBD | Enable only for files, photos, documents, attachments |
| LLM | TBD | Enable only for generation, analysis, summaries, recommendations, or chat |
| Sidecar | TBD | Enable only for PDF, browser automation, or scraping |
| Jobs | TBD | Enable only for async/background work |
| Payments | TBD | Enable only when charging at launch |
| Browser extension | TBD | Enable only when workflow must happen inside the browser |
| Mobile client | TBD | See mobile decision matrix below |

## Mobile Client Decision

Do not scaffold mobile by default. Choose the smallest surface that matches the product:

| Need | Choice |
|------|--------|
| Mobile-friendly dashboard, forms, or read-only workflows | Responsive web or PWA |
| App-store app with same core surface as web | Expo React Native |
| Push notifications, camera upload, location, offline drafts | Expo React Native development build |
| Mobile-first product with heavy custom UI or animation | Flutter |
| Widgets, watch, AR, background services, deep OS integration | Native Swift/Kotlin review |

If mobile is enabled, use this layout:

```text
apps/
  web/       Next.js
  mobile/    Expo React Native
packages/
  api-client/      generated TypeScript client
  design-tokens/   colors, spacing, typography only
backend/
```

Mobile must not reuse browser-only auth, localStorage token logic, Next.js components, or web dashboard layouts. Share API schema, generated client, design tokens, product language, and analytics event names.

## API Routes

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /api/v1/auth/register | public | Register |
| POST | /api/v1/auth/login | public | Login |
| POST | /api/v1/auth/refresh | cookie | Refresh JWT |
| GET | /api/v1/auth/me | JWT | Current user |
| GET | /api/v1/admin/users | JWT+admin | List users |

## Pages

| Route | Auth | Description |
|-------|------|-------------|
| / | public | Landing |
| /login | public | Login |
| /register | public | Register |
| /dashboard | required | Main app |
| /admin | admin | Admin portal |

## Monetization

<!-- How does this make money? -->

## Tech Decisions

<!-- Any non-default choices and why. -->

## Verification Gate

Before claiming this project is ready for Codex execution or handoff, run:

- Backend: `go test ./...`
- Frontend: `npx tsc --noEmit`
- Frontend: `npm run build`
- Frontend deploy package: `npm run cf:build`
- Cloudflare deploy dry-run: `npx wrangler deploy --dry-run` when credentials/tooling allow it
- Docker: `docker compose build backend` and `docker compose build frontend`
- Smoke: compose up, then backend `/health`, frontend `/`, sidecar `/health` if enabled

If a check is skipped, write the reason and residual risk into `_docs/EXECUTIVE_SUMMARY.md`.
