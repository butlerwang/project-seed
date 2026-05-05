# [Project Name] — Design Document

> Source of truth. Do not change without owner review.

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
