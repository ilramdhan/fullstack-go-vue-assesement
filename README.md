# Payment Dashboard

Internal dashboard for monitoring and reviewing incoming payments. Built for the
Durianpay Full Stack Engineer take-home assignment.

- **Backend** — Go 1.22, chi router, SQLite (WAL), JWT auth, OpenAPI-first.
- **Frontend** — Vue 3 + TypeScript + Vite, Pinia, TanStack Vue Query, shadcn-vue + Tailwind 3.
- **Contract** — `openapi.yaml` at the repo root drives both the Go server stubs and the typed FE client.

```
.
├── openapi.yaml                # single source of truth for the HTTP API
├── backend/                    # Go service
├── frontend/                   # Vue app
├── docker-compose.yml          # one-command bootstrap (BE + FE)
├── Makefile                    # stack-level targets (up/down/test/...)
└── README.md
```

## 1. Prerequisites

| Tool | Version | Used for |
|---|---|---|
| Docker + Docker Compose | any recent | the recommended path; runs both services |
| Go | 1.21+ | backend (only if running locally without Docker) |
| Node.js | 20+ | frontend (only if running locally without Docker) |
| GNU Make | any | shortcut targets |

The reviewer environment described in the assignment (Go 1.21+, Node 20+, Docker, Docker Compose, Make, macOS) is fully covered.

## 2. Quick start (Docker)

```bash
git clone <this-repo> payment-dashboard
cd payment-dashboard

make up                           # builds + starts BE and FE
```

Then open:

| URL | What |
|---|---|
| <http://localhost:8088> | Frontend dashboard |
| <http://localhost:8080/healthz> | Backend liveness probe |
| <http://localhost:8080/swagger> | Swagger UI for the OpenAPI contract |
| <http://localhost:8080/openapi.json> | Raw OpenAPI 3.0.3 spec |

When you're done:

```bash
make down        # stops the stack, keeps the sqlite volume
make clean       # stops + drops the volume (database is reset on next up)
make logs        # tail logs from both services
```

### Demo accounts

The first time the backend boots it migrates the schema and seeds two users plus 50 demo payments (38 completed / 7 processing / 5 failed).

| Email | Password | Role | Capabilities |
|---|---|---|---|
| `cs@test.com` | `password` | `cs` | Read-only |
| `operation@test.com` | `password` | `operation` | Read + approve/reject `processing` payments |

## 3. Manual setup (without Docker)

Two terminals, run from the repo root.

**Backend** (terminal 1):

```bash
cd backend
cp env.sample .env             # adjust JWT_SECRET if you like
make tool-openapi              # one-time: install oapi-codegen
make openapi-gen               # regenerate openapigen package from ../openapi.yaml
make dep                       # go mod tidy
make run                       # listens on :8080
```

**Frontend** (terminal 2):

```bash
cd frontend
cp .env.example .env           # contains VITE_API_BASE_URL=http://localhost:8080
npm install
npm run dev                    # listens on :5173, proxies /api -> :8080
```

Open <http://localhost:5173>.

For a production build of the frontend: `npm run build && npm run preview` (port 4173).

## 4. API contract

The contract lives in [`openapi.yaml`](./openapi.yaml). Both the Go server stubs (`backend/internal/openapigen`) and the typed TS client (`frontend/src/api/generated`) are regenerated from it.

| Method | Path | Auth | Notes |
|---|---|---|---|
| `POST` | `/dashboard/v1/auth/login` | public | email + password → JWT + user (email, role) |
| `GET`  | `/dashboard/v1/payments` | bearer | filter by `status`, `id`, sort, paginate |
| `GET`  | `/dashboard/v1/payments/summary` | bearer | aggregated counts by status |
| `PUT`  | `/dashboard/v1/payments/{id}/review` | bearer + role `operation` | approve/reject a `processing` payment |
| `GET`  | `/healthz` | public | liveness probe |
| `GET`  | `/swagger` | public | Swagger UI for interactive exploration |

The OpenAPI request validator is mounted as middleware, so any request that violates the contract (bad enum, missing field, wrong type) is rejected with `400` before reaching a handler.

## 5. Testing strategy

Both layers have unit + integration coverage and lint passes are required for the `make test` target to succeed.

### Backend (`backend/Makefile`)

```bash
make -C backend test          # plain run
make -C backend test-race     # race detector
make -C backend test-cover    # writes coverage.html
```

- **Repository tests** — sqlite `:memory:` with the real migration runner; covers every WHERE/ORDER BY branch and a `SQL-injection-shaped sort value` is asserted to fall back to the safe default.
- **Usecase tests** — interface-driven stub repositories; covers JWT issue + verify roundtrip, expired tokens, signature rejection, the review state-machine, validation failures.
- **Middleware tests** — `httptest` request recorder; covers Bearer parser edge cases (missing/malformed/empty) and role-guard allowed/denied paths.
- **HTTP integration tests** — full chi pipeline (`OpenAPI validator → conditionalAuth → handler`) wired against an in-memory db with seeded fixtures; one happy path per endpoint plus every status code we promise: `200/400/401/403/404/409`.

Statements coverage on `auth + payment + middleware` packages: **84.8%**, all green under `-race`.

### Frontend (`frontend/package.json`)

```bash
cd frontend
npm test                      # vitest run
npm run test:cover            # vitest run --coverage
```

- **MSW at the network layer** — `msw/node` handlers stand in for the real backend, so the same axios + `@tanstack/vue-query` + hey-api SDK pipeline runs end-to-end. Handlers are stateful (`approve` mutates the in-memory list) so optimistic updates and cache invalidation are exercised.
- **Component tests** with `@testing-library/vue` and a `renderWithStack` helper that mounts Pinia, Vue Query, and Vue Router exactly the way `main.ts` does in production.
- **Pure logic tests** for `formatCurrency`/`formatDate`, the auth store (set / clear / persist / rehydrate / corrupt-storage tolerance), and the router guards.

Statements coverage: 78.2% overall, **91.1% on `features/auth`**, **80.4% on `features/payments`**, 100% on `lib/format`.

## 6. Environment variables

Backend (`backend/env.sample`):

| Var | Default | Purpose |
|---|---|---|
| `HTTP_ADDR` | `:8080` | Listen address |
| `DATABASE_PATH` | `data/dashboard.db` | SQLite file (Docker: `/app/data/dashboard.db` on a named volume) |
| `JWT_SECRET` | dev placeholder | HS256 signing key — replace before any real deploy |
| `JWT_EXPIRED` | `24h` | Token TTL (Go duration) |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:5173,http://localhost:4173,http://localhost:3000` | CSV of allowed origins |

Frontend (`frontend/.env.example`):

| Var | Default | Purpose |
|---|---|---|
| `VITE_API_BASE_URL` | `http://localhost:8080` | Where the SPA looks for the backend |

In Docker the frontend image is built with `VITE_API_BASE_URL=/api` so the SPA talks to nginx, which proxies `/api/*` to the backend service over the Docker network.

## 7. Architecture overview

```
Browser ──► nginx (FE container) ──► Go server (BE container) ──► SQLite (volume)
              │  /            (SPA + assets, gzip, SPA fallback)
              └─ /api/*       (proxy_pass http://backend:8080/)
```

Backend is layered (entity → repository → usecase → handler → API adapter), wired manually in `main.go` with no DI framework. The OpenAPI request validator runs as the outer middleware on the protected route group, then a JWT auth middleware injects the verified identity into the request context, then the route-specific role check (for `PUT /payments/{id}/review`) runs inside the handler.

Frontend splits state cleanly: **Pinia** owns client state (auth token + role + email, persisted to localStorage), **Vue Query** owns server state (payments list/summary/mutations) with optimistic updates and rollback. UI composes shadcn-vue primitives styled with Tailwind.

See `backend/README.md` and `frontend/README.md` for per-package detail.

## 8. Troubleshooting

- **Port already in use** — another process is on `:8080` or `:8088`. `lsof -nP -i :8080` to find it; either stop the other process or override the port mapping in `docker-compose.yml`.
- **Backend container marked unhealthy** — `make logs` to see the error. The volume mount path `./backend/data` is created with the container's user; if you've previously created `backend/data` as root locally, `sudo chown -R $(id -u):$(id -g) backend/data`.
- **`make up` fails on an arm64 Mac** — the backend builds CGO with the SQLite driver, which needs `gcc/musl-dev` in the builder stage; the Dockerfile already installs them. If the build fails on a fresh Mac, run `docker buildx build --platform linux/amd64` or update Docker Desktop.
- **CORS error in the browser** — make sure `CORS_ALLOWED_ORIGINS` on the backend matches the origin you're loading the frontend from. Default covers `:5173`, `:4173`, and `:3000`.
- **"Session expired" toast on every page** — your JWT_SECRET changed since the token was issued. Sign out and back in, or clear `localStorage` for the dashboard origin.
