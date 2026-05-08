# Project Seed Factory Strategy Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Convert `project-seed` from a static template into a capability-gated project factory with explicit mobile rules and verification gates.

**Architecture:** Keep the tested code scaffold in `project-seed`, and make `project.yaml` the machine-readable source for generated docs and capability choices. Human docs explain the workflow; ADRs capture why; the `../new-project` skill applies the workflow to future repos.

**Tech Stack:** Go 1.24, Next.js 15, Cloudflare Workers via OpenNext, Fly.io, Neon PostgreSQL, Expo React Native when mobile is enabled.

---

### Task 1: Capture The Accepted Architecture

**Files:**
- Create: `project.yaml`
- Create: `docs/adr/README.md`
- Create: `docs/adr/template.md`
- Create: `docs/adr/0001-use-capability-gated-project-factory.md`

- [x] Add a root manifest that records stack defaults, capability gates, mobile decision rules, verification requirements, and known loopholes closed.
- [x] Add an ADR index and template.
- [x] Record the accepted decision: template plus orchestration skill, not template-only or skill-only.

### Task 2: Update Template Handoff Docs

**Files:**
- Modify: `README.md`
- Modify: `CLAUDE.md`
- Modify: `_docs/DESIGN.md`
- Modify: `_docs/CODEX.md`

- [x] Document the factory workflow: intake, manifest, owner approval, scaffold, verification, Codex execution.
- [x] Add mobile-client decision rules and repo layout.
- [x] Make Codex read `project.yaml` before implementation.
- [x] Require skipped verification checks to be recorded as residual risk.

### Task 3: Close Rename Loophole

**Files:**
- Modify: `Makefile`

- [x] Include `*.jsonc` in `make rename` so `frontend/wrangler.jsonc` gets renamed.

### Task 4: Align The `new-project` Skill

**Files:**
- Modify: `../new-project/SKILL.md`
- Modify: `../new-project/README.md`
- Modify: `../new-project/RETROFIT_SKILL.md`
- Modify: `../new-project/TODOS.md`

- [x] Add mobile-client inference and confirmation.
- [x] Generate the expanded `project.yaml` schema.
- [x] Require a post-scaffold verification gate before claiming confidence.
- [x] Keep Cloudflare Workers/OpenNext as the current default.

### Task 5: Verify

**Commands:**
- [x] `GOCACHE=$(pwd)/backend/.gocache go test ./...` in `backend/`
- [x] `npx tsc --noEmit` in `frontend/`
- [x] `npm run build` in `frontend/`
- [x] `npm run cf:build` in `frontend/`
- [x] `npx wrangler deploy --dry-run` in `frontend/` — passed; Wrangler emitted a sandbox log-file EPERM warning but exited 0 and produced the dry-run bundle summary.
- [x] `docker compose build backend`
- [x] `docker compose build frontend`
- [x] `docker compose up -d` and smoke checks for backend `/health`, frontend `/`, and sidecar `/health`
- [x] `npm audit --audit-level=moderate` — reports a moderate PostCSS advisory through Next/OpenNext; no acceptable package-only fix because `npm audit fix --force` downgrades Next to 9.3.3.

**Expected:** All checks pass, or skipped checks are recorded with exact reasons.
