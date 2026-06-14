---
name: servitor-app
description: >-
  Design and build applications that run on Servitor (Software As A Repository).
  Use when creating a servitor-backend, designing data models for file-based storage,
  or scaffolding a new servitor app. Applies to any project with a .servitor.conf.
---

# Servitor App Design

Build personal web apps where frontend, backend, and data live in one git repo, served by a single binary.

## Your Role: App Architect

✅ Design the data model FIRST — before any code
✅ Map entity relationships and dependencies explicitly
✅ Reflect dependencies in CRUD operations (validate refs on create, cascade/reject on delete)
✅ Keep backend stateless and fast (spawned per request)
✅ Store all persistent data as files in `data/`
✅ Use the CGI contract (raw HTTP on stdin/stdout)

❌ Do NOT write backend code before the data model is defined
❌ Do NOT leave orphaned references after deletes
❌ Do NOT ignore entity dependencies in API design
❌ Do NOT use databases — data is file-based only
❌ Do NOT assume persistent state between requests

---

## Design Process

### 1. Data Model First

Before writing any code, define:

1. **Entities** — the objects the app manages
2. **Attributes** — fields on each entity
3. **Relationships** — dependencies between entities (one-to-one, one-to-many, many-to-many)
4. **File layout** — how entities map to files/directories in `data/`

The data model dictates API routes, CRUD operations, frontend views, and sync behaviour.

### 2. Dependency-Aware Backend

The backend must understand the dependency graph between objects:

- **Create**: Validate that all referenced objects exist before writing
- **Read**: Resolve and include related objects as needed
- **Update**: Consider cascading effects on dependents
- **Delete**: Handle dependents explicitly — reject if dependents exist, cascade-delete, or nullify references. Never leave orphans.

### 3. API Design

Map entities to routes under the configured `api_prefix` (default `/api`):

```
GET    /api/{entity}       — list
GET    /api/{entity}/{id}  — read
POST   /api/{entity}       — create
PUT    /api/{entity}/{id}  — update
DELETE /api/{entity}/{id}  — delete
```

Return proper status codes: 201 (created), 404 (not found), 409 (conflict), 422 (validation failure).

---

## Project Structure

```
my-app/
├── .servitor.conf          # YAML configuration
├── frontend/               # Pre-built static files (HTML, CSS, JS)
│   └── index.html
├── backend/                # Backend binary source
│   ├── main.go
│   └── servitor-backend    # Compiled binary
├── data/                   # Persistent file-based data (auto-synced to git)
│   └── {entities}/         # One directory per entity type
└── .git/
```

## Backend Contract

The backend binary receives a raw HTTP request on stdin and writes a raw HTTP response to stdout. It is spawned per request and must exit after responding.

```
stdin:  POST /api/notes HTTP/1.1\r\nContent-Type: application/json\r\n\r\n{"title":"hi"}
stdout: HTTP/1.1 201 Created\r\nContent-Type: application/json\r\n\r\n{"id":"abc","title":"hi"}
```

## Data Layout Guidelines

- One directory per entity type: `data/users/`, `data/projects/`
- One JSON file per record: `data/users/{id}.json`
- Or a single JSON file per collection for small datasets: `data/tags.json`
- Structure files to minimise concurrent writes to the same file
- Keep individual files small — git diffs stay readable

## Constraints

- Localhost only (127.0.0.1), single-user, no auth
- No external dependencies beyond `servitor` binary + `git`
- Backend is cold-started per request — avoid heavy init
- Backend timeout is configurable (default 3s) — stay well under it
- Git sync is eventually-consistent — design for it
- Conflict resolution is force-push (local wins)
