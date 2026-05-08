# [Project Name] — Codex Implementation Plan

## Rules for Codex

1. Read `project.yaml` first — it is the machine-readable capability manifest
2. Follow `_docs/DESIGN.md` exactly — it is the product source of truth
3. Do not change `_docs/DESIGN.md` or architecture scope without owner review
4. After each task: append to `_docs/IMPLEMENTATION_LOG.md`
5. After each phase: update `_docs/EXECUTIVE_SUMMARY.md`
6. Verify after every task: `go build ./...` + `go test ./...` + `npx tsc --noEmit`
7. For deploy or infrastructure changes, also run the relevant Docker and Cloudflare/Fly dry-runs
8. Never skip verification silently; record skipped checks and why
9. Commit after each task with conventional commit format

## Capability Rules

- Strip optional services when `project.yaml` disables them.
- Do not add a Playwright sidecar unless the product needs PDF, browser automation, or scraping.
- Do not activate Stripe behavior unless `payments.enabled` is true and webhook secrets are configured.
- Do not scaffold a mobile app unless `mobile_client.enabled` is true.
- If mobile is enabled, default to Expo React Native and generate a shared TypeScript API client. Use Flutter or native only when the design document says why Expo is not enough.

## Phase 1 — [Name]

### Task 1.1 — [Name]

**Files:** 
- Create: `path/to/file.go`

**Implementation:**
<!-- Steps -->

**Verify:** `go build ./...`

---
<!-- Add more tasks following this pattern -->
