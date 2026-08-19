# Pet Foster Operations Platform

This repository keeps the original Vue 3 + Element Plus admin frontend and replaces its Spring/MyBatis backend with a Go service. The API contract remains under `/api`, so the existing pages do not need a second client implementation.

## Run locally

```text
cp .env.example .env
go run ./cmd/pet-server
cd frontend-admin
npm ci
npm run dev
```

The Go service listens on `:8080`; Vite serves the admin application on `:5173` and proxies `/api` to the Go service. The default local accounts are `admin/admin123` and `testuser/user123`.

## Production checks

```text
go test ./... -count=1
go test -race ./... -count=1
go vet ./...
go build ./...
cd frontend-admin && npm ci && npm run build
```

The backend uses SQLite migrations with foreign keys and WAL mode. Its public workflows cover authenticated users, pet ownership, room capacity, foster-order state transitions, selected care services, daily care records, operation logs, and restart recovery. Mutating order operations run inside transactions and preserve the room/order relationship.

## Repository layout

- `cmd/pet-server`: HTTP entrypoint and graceful shutdown
- `internal/pet`: pet-foster domain, SQLite persistence, service layer, and API compatibility handlers
- `internal/domain`, `internal/service`, `internal/storage`, `internal/httpapi`, `internal/worker`: reusable Go platform infrastructure retained from the source modernization baseline
- `frontend-admin`: original Vue frontend, unchanged apart from the development proxy target
- `migrations`, `testdata`: platform migration and fixture assets
