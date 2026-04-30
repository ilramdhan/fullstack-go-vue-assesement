# Backend — Payment Dashboard API

Go service for the Durianpay take-home assignment. Layered architecture (entity / repository / usecase / handler) with OpenAPI as the single source of truth.

> Full setup, including Docker, lives in the **root README**. This file documents backend-specific workflows.

## Stack

- Go 1.22, [chi/v5](https://github.com/go-chi/chi) router
- SQLite (mattn/go-sqlite3) with WAL journaling for safe multi-reader/single-writer
- JWT auth (golang-jwt/v5)
- OpenAPI v3 → server stubs + types via [oapi-codegen v2](https://github.com/oapi-codegen/oapi-codegen)
- Request validation middleware against the live OpenAPI spec

## Layout

```
backend/
├── main.go                       # composition root
├── internal/
│   ├── api/                      # adapter implementing generated ServerInterface
│   ├── config/                   # env-driven config
│   ├── entity/                   # domain types + errors
│   ├── module/
│   │   └── auth/                 # auth feature: handler / usecase / repository
│   ├── openapigen/               # GENERATED — do not edit
│   ├── service/http/             # chi server + middleware wiring
│   └── transport/                # error rendering helpers
├── script/gen-secret/            # helper to print a JWT secret
├── data/                         # sqlite db lives here (gitignored)
├── env.sample                    # template for .env
└── Makefile
```

## Local development

Prerequisites: **Go 1.21+**, a C toolchain (gcc/clang) for the SQLite cgo driver.

```bash
cp env.sample .env
make tool-openapi      # one-time
make openapi-gen       # regenerates internal/openapigen/openapi.gen.go
make dep               # go mod tidy
make run               # listens on :8080
```

Quick smoke check:

```bash
curl -i http://localhost:8080/healthz
curl -i -X POST http://localhost:8080/dashboard/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"cs@test.com","password":"password"}'
```

Seeded users (auto-created on first boot):

| Email | Password | Role |
|---|---|---|
| `cs@test.com` | `password` | `cs` |
| `operation@test.com` | `password` | `operation` |

## OpenAPI workflow

`../openapi.yaml` is the source of truth.

1. Edit `../openapi.yaml`.
2. `make openapi-gen` regenerates `internal/openapigen/openapi.gen.go`.
3. Implement the new method on `internal/api/api_handler.go` — the compile-time assertion `var _ openapigen.ServerInterface = (*APIHandler)(nil)` will fail otherwise.

## Environment variables

| Var | Default | Purpose |
|---|---|---|
| `HTTP_ADDR` | `:8080` | Listen address |
| `DATABASE_PATH` | `data/dashboard.db` | SQLite file location |
| `JWT_SECRET` | dev placeholder | HS256 signing key |
| `JWT_EXPIRED` | `24h` | Token TTL (Go duration) |
| `OPENAPIYAML_LOCATION` | `../openapi.yaml` | Spec path for validator middleware |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:5173,http://localhost:4173,http://localhost:3000` | CSV list |

## API endpoints (current)

- `GET  /healthz` — liveness probe.
- `POST /dashboard/v1/auth/login` — email + password → JWT + role.
- `GET  /dashboard/v1/payments` — list payments (auth required). _Implemented in Phase 3._
- `GET  /dashboard/v1/payments/summary` — aggregate totals. _Implemented in Phase 3._
- `PUT  /dashboard/v1/payments/{id}/review` — approve/reject (role `operation`). _Implemented in Phase 3._
