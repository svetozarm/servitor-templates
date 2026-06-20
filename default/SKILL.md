---
name: servitor-app
description: Interact with a servitor Go CGI backend by piping raw HTTP requests to the binary via stdin. Use when managing data, invoking the backend API, or extending a servitor app.
---

# Servitor Default Template — Backend & Datastore

Use this skill when working with a servitor-based project that uses a Go CGI backend with flat-file JSON storage.

## Source of Truth

**`backend/openapi.yaml`** — API contract. Read this for endpoints, schemas, and status codes.

## Prerequisites

The backend binary must exist at `backend/servitor-backend`. If it's missing, build it:

```bash
./build.sh
```

## Datastore

Location: `data/todos/` (one JSON file per record, named `{id}.json`)

## API Reference

| Method | Path | Body | Response | Status |
|--------|------|------|----------|--------|
| GET | /api/todos | — | `Todo[]` | 200 |
| POST | /api/todos | `{"text": "..."}` | `Todo` | 201 |
| GET | /api/todos/{id} | — | `Todo` | 200 |
| PUT | /api/todos/{id} | `{"text?": "...", "done?": bool}` | `Todo` | 200 |
| DELETE | /api/todos/{id} | — | — | 204 |

Errors: 404 (not found), 422 (validation — `text` required on POST).

### Todo Schema

```json
{"id": "string", "text": "string", "done": boolean, "created_at": "RFC3339 datetime"}
```

## Interacting with the App

The backend is a CGI binary that reads a raw HTTP request from stdin. Invoke it directly — no HTTP server or curl needed.

```bash
# List all todos
printf 'GET /api/todos HTTP/1.1\r\nHost: localhost\r\n\r\n' | ./backend/servitor-backend

# Create
printf 'POST /api/todos HTTP/1.1\r\nHost: localhost\r\nContent-Length: 20\r\n\r\n{"text":"Buy milk"}' | ./backend/servitor-backend

# Update (partial — only send fields to change)
printf 'PUT /api/todos/{id} HTTP/1.1\r\nHost: localhost\r\nContent-Length: 13\r\n\r\n{"done":true}' | ./backend/servitor-backend

# Delete
printf 'DELETE /api/todos/{id} HTTP/1.1\r\nHost: localhost\r\n\r\n' | ./backend/servitor-backend
```

The `Content-Length` header must match the body byte length exactly. For requests with no body (GET, DELETE), omit it.
