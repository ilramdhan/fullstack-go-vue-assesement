# Backend — Payment Dashboard API

Go service for the Durianpay take-home assignment. Layered architecture (entity / repository / usecase / handler) with OpenAPI as the single source of truth.

> Full setup, including Docker, lives in the **root README**. This file documents backend-specific workflows.

## Stack

- Go 1.22, [chi/v5](https://github.com/go-chi/chi) router
- SQLite (mattn/go-sqlite3) with WAL journaling for safe multi-reader / single-writer concurrency
- JWT auth (golang-jwt/v5)
- OpenAPI v3 → server stubs + types via [oapi-codegen v2](https://github.com/oapi-codegen/oapi-codegen); the spec is embedded into the binary, no runtime filesystem dependency
- Request validation middleware against the embedded OpenAPI spec
- Swagger UI served at `/swagger`

## Layout

```
backend/
├── main.go                       # composition root
├── internal/
│   ├── api/                      # adapter implementing generated ServerInterface
│   ├── config/                   # env-driven config
│   ├── entity/                   # domain types + errors
│   ├── middleware/               # JWT auth + role guard
│   ├── migration/                # versioned schema migrations + seed
│   ├── module/
│   │   ├── auth/                 # handler / usecase / repository
│   │   └── payment/              # handler / usecase / repository
│   ├── openapigen/               # GENERATED — do not edit
│   ├── service/http/             # chi server + middleware wiring + Swagger UI
│   └── transport/                # error rendering helpers
├── script/
│   ├── gen-secret/               # helper to print a JWT secret
│   └── seed/                     # standalone seed runner
├── data/                         # sqlite db (gitignored)
├── env.sample                    # template for .env
├── Dockerfile                    # multi-stage CGO + alpine runtime
└── Makefile
```

## Local development

Prerequisites: **Go 1.21+**, a C toolchain (gcc/clang) for the SQLite cgo driver.

```bash
cp env.sample .env
make tool-openapi      # one-time: install oapi-codegen
make openapi-gen       # regenerate internal/openapigen/openapi.gen.go
make dep               # go mod tidy
make run               # listens on :8080
```

Quick smoke check:

```bash
curl -i http://localhost:8080/healthz
curl -i -X POST http://localhost:8080/dashboard/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"cs@test.com","password":"password"}'

# Interactive API docs
open http://localhost:8080/swagger
```

Seeded users (auto-created on first boot):

| Email | Password | Role |
|---|---|---|
| `cs@test.com` | `password` | `cs` |
| `operation@test.com` | `password` | `operation` |

The first boot also seeds 50 demo payments (38 completed / 7 processing / 5 failed). Re-running `make seed` is a no-op if data already exists; `make seed-reset` wipes the local sqlite file and re-seeds.

## OpenAPI workflow

`../openapi.yaml` is the source of truth.

1. Edit `../openapi.yaml`.
2. `make openapi-gen` regenerates `internal/openapigen/openapi.gen.go`.
3. Implement the new method on `internal/api/api_handler.go` — the compile-time assertion `var _ openapigen.ServerInterface = (*APIHandler)(nil)` fails the build if you add a path without a handler.

The generated Go file embeds the spec, so `GetSwagger()` is in-process and the validator middleware doesn't read from disk. `/openapi.json` and `/swagger` reuse that embedded spec.

## Tests

```bash
make test          # plain run
make test-race     # with -race
make test-cover    # writes coverage.html
```

Strategy is documented in the root README (Section 5). Highlights:
- repository tests run against sqlite `:memory:` with the real migration runner
- usecase tests stub the repo via interface — no third-party mock framework
- HTTP integration tests run the full chi pipeline against an in-memory DB and assert every status code the OpenAPI contract promises

Coverage on `auth + payment + middleware`: **84.8% statements**, all green under `-race`.

## Environment variables

| Var | Default | Purpose |
|---|---|---|
| `HTTP_ADDR` | `:8080` | Listen address |
| `DATABASE_PATH` | `data/dashboard.db` | SQLite file location |
| `JWT_SECRET` | dev placeholder | HS256 signing key |
| `JWT_EXPIRED` | `24h` | Token TTL (Go duration) |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:5173,http://localhost:4173,http://localhost:3000` | CSV list |

## API endpoints

| Method | Path | Auth | Notes |
|---|---|---|---|
| `GET`  | `/healthz` | public | Liveness probe; returns `{"status":"ok"}`. |
| `GET`  | `/swagger` | public | Swagger UI rendered against the embedded spec. |
| `GET`  | `/openapi.json` | public | The embedded spec, marshaled to JSON. |
| `POST` | `/dashboard/v1/auth/login` | public | Email + password → JWT + user (email, role). Wrong password returns 401, malformed email returns 400 (validator). |
| `GET`  | `/dashboard/v1/payments` | bearer | List with `status`, `id`, `sort`, `limit`, `offset`. Sort whitelist: `created_at`, `-created_at`, `amount`, `-amount`. |
| `GET`  | `/dashboard/v1/payments/summary` | bearer | Counts by status. |
| `PUT`  | `/dashboard/v1/payments/{id}/review` | bearer + role `operation` | `{decision: approve\|reject}`. 403 for `cs` role, 404 if id missing, 409 if not in `processing`. |

## Auth flow

1. Client `POST /auth/login` with email + password.
2. Server verifies bcrypt hash, signs an HS256 JWT carrying `sub` (user id), `email`, `role`, `iat`, `exp`.
3. Client sends `Authorization: Bearer <token>` on every subsequent call.
4. `middleware.Auth` parses the token, verifies the signature, and injects `(userID, email, role)` into the request context. `middleware.RequireRole("operation")` (or the inline check inside `APIHandler.ReviewPayment`) gates routes that need a specific role.

The OpenAPI request validator runs *before* the auth middleware on the dashboard route group, so request shape errors return 400 even on protected routes; an `AuthenticationFunc` no-op satisfies the validator's contract that something handles `bearerAuth` (the real check is the chi middleware that runs next).
