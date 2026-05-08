# ADR-0001: Use Capability-Gated Project Factory

**Date**: 2026-05-07
**Status**: accepted
**Deciders**: Butler Wang, Codex

## Context

The user repeatedly builds projects with the same shape: CEO review, engineering review, Go backend, Next.js frontend, optional sidecars, low-cost deployment, and a strict documentation handoff system. A raw template saves boilerplate time, but it can also copy unneeded services, stale deployment defaults, insecure placeholder config, or mobile/browser clients into projects that do not need them. The factory needs to preserve speed while forcing explicit product and architecture decisions before implementation.

## Decision

Use `project-seed` as a capability-gated project factory. Every new project must produce `project.yaml`, `_docs/DESIGN.md`, `_docs/CODEX.md`, `_docs/IMPLEMENTATION_LOG.md`, and `_docs/EXECUTIVE_SUMMARY.md`; optional capabilities are included only when the manifest enables them. The frontend default is Cloudflare Workers via OpenNext, and mobile is disabled by default with Expo React Native as the first app-store option when needed.

## Alternatives Considered

### Static GitHub Template Only

- **Pros**: Simple, fast, easy to clone.
- **Cons**: Copies too much, cannot explain why capabilities are present, and drifts when platform defaults change.
- **Why not**: It saves setup time but does not protect against architecture mistakes.

### Claude Skill Only

- **Pros**: Flexible and conversational.
- **Cons**: Recreates boilerplate from prompts and depends too much on model memory.
- **Why not**: The stable code should live in a tested template; the skill should orchestrate, not invent.

### Mobile In Every Project

- **Pros**: Future-ready for app-store distribution.
- **Cons**: Adds dependency, auth, CI/CD, and UX burden for projects that only need responsive web.
- **Why not**: Mobile should be a product requirement, not a default artifact.

## Consequences

### Positive

- New projects start faster without hiding important architecture decisions.
- Optional services have explicit reasons and can be stripped safely.
- Handoff docs stay aligned with the scaffold instead of becoming narrative drift.
- Verification gates make confidence evidence-based.

### Negative

- Scaffolding is more procedural than a simple clone.
- The skill and template must evolve together.
- Some decisions still need human review, especially auth, mobile, tenancy, and data model.

### Risks

- Cloudflare, Next.js, Fly.io, and package ecosystems can change. Mitigation: verify current docs before changing deployment defaults and run deploy dry-runs.
- Inferred capability flags can be wrong. Mitigation: require owner approval before architecture lock.
- A future mobile-first project may outgrow Expo. Mitigation: decision matrix allows Flutter or native when the product justifies it.

