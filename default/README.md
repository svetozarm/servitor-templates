# Todo App (Servitor Default Template)

A simple todo app demonstrating the servitor pattern.

## Structure

```
├── .servitor.conf          # Server + backend + sync config
├── frontend/index.html     # Single-page frontend
├── backend/
│   ├── openapi.yaml        # API spec (source of truth)
│   ├── main.go             # CGI backend
│   └── servitor-backend    # Compiled binary
└── data/todos/             # One JSON file per todo
```

## Usage

```bash
cd backend && go build -o servitor-backend . && cd ..
## Alternatively, you can use the build.sh script to build
servitor
```

Open http://127.0.0.1:8080 — add, toggle, and delete todos.

Data persists in `data/todos/{id}.json` and auto-syncs to git.

## API

See `backend/openapi.yaml` for the full spec. Summary:

| Method | Path              | Description  | Status |
|--------|-------------------|--------------|--------|
| GET    | /api/todos        | List todos   | 200    |
| POST   | /api/todos        | Create todo  | 201    |
| GET    | /api/todos/{id}   | Get todo     | 200    |
| PUT    | /api/todos/{id}   | Update todo  | 200    |
| DELETE | /api/todos/{id}   | Delete todo  | 204    |
