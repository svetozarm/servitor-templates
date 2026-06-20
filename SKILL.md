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
❌ Do NOT implement handlers before the OpenAPI spec exists
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

### 4. OpenAPI Specification

Define the backend API as an OpenAPI 3.1 spec in `backend/openapi.yaml` before writing handler code. This spec is the single source of truth for routes, request/response schemas, and validation rules.

```yaml
openapi: "3.1.0"
info:
  title: My App API
  version: "1.0.0"
paths:
  /api/notes:
    get:
      summary: List notes
      responses:
        "200":
          description: Array of notes
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: "#/components/schemas/Note"
    post:
      summary: Create a note
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/NoteInput"
      responses:
        "201":
          description: Created note
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Note"
        "422":
          description: Validation error
  /api/notes/{id}:
    get:
      summary: Get a note
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: The note
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Note"
        "404":
          description: Not found
    put:
      summary: Update a note
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/NoteInput"
      responses:
        "200":
          description: Updated note
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Note"
        "404":
          description: Not found
        "422":
          description: Validation error
    delete:
      summary: Delete a note
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        "204":
          description: Deleted
        "404":
          description: Not found
        "409":
          description: Conflict — dependents exist
components:
  schemas:
    Note:
      type: object
      required: [id, title, created_at]
      properties:
        id:
          type: string
        title:
          type: string
        body:
          type: string
        created_at:
          type: string
          format: date-time
    NoteInput:
      type: object
      required: [title]
      properties:
        title:
          type: string
        body:
          type: string
```

Guidelines:
- Place the spec at `backend/openapi.yaml`
- Define all entity schemas under `components/schemas`
- Use `$ref` to avoid duplication between endpoints
- Separate input schemas (for create/update) from full entity schemas (which include generated fields like `id`, timestamps)
- Document error responses (404, 409, 422) on every endpoint that can produce them
- Write the spec AFTER the data model but BEFORE implementing handlers

---

## Project Structure

```
my-app/
├── .servitor.conf          # YAML configuration
├── frontend/               # Pre-built static files (HTML, CSS, JS)
│   └── index.html
├── backend/                # Backend binary source
│   ├── openapi.yaml        # API spec (source of truth for routes & schemas)
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

## Multi-App Mode

Servitor supports hosting multiple apps from a single repo via cascading configs.

### Top-Level Config

```yaml
server:
  port: 8080
  host: 127.0.0.1

apps:
  - name: notes
    path: ./apps/notes
  - name: wiki
    path: ./apps/wiki
```

Each app directory contains its own `.servitor.conf` with `frontend`, `backend`, and `sync` settings. Apps are served at `/<name>/` (e.g., `/notes/`, `/wiki/`).

### Multi-App Structure

```
my-repo/
├── .servitor.conf          # Top-level: server + apps list
├── apps/
│   ├── notes/
│   │   ├── .servitor.conf  # frontend/backend/sync config
│   │   ├── frontend/
│   │   ├── backend/
│   │   └── data/
│   └── wiki/
│       ├── .servitor.conf
│       ├── frontend/
│       ├── backend/
│       └── data/
└── .git/
```

### Per-App Config

```yaml
frontend:
  path: ./frontend

backend:
  path: ./backend/servitor-backend
  api_prefix: /api
  timeout: 3s

sync:
  enabled: true
  inactivity_delay: 30s
  max_interval: 5m
```

All paths are resolved relative to the app's directory.

### Routing

- Static files: `/<name>/` → app's `frontend/`
- API calls: `/<name>/<api_prefix>/` → app's backend binary
- Root `/` → HTML index page with links to all apps

### Frontend Requirement

Frontends must use **relative** API paths (e.g. `fetch("api/types")` not `fetch("/api/types")`). The browser resolves relative URLs against the page's base URL (`/<name>/`), producing the correct prefixed path.

Each app gets its own sync engine and CGI proxy. The single-app config format continues to work unchanged.