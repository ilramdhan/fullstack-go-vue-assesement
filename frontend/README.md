# Frontend — Payment Dashboard

Vue 3 + TypeScript + Vite SPA for the Durianpay take-home assignment.

> Full setup, including Docker, lives in the **root README**. This file documents frontend-specific workflow and rationale.

## Stack

| Concern | Choice | Why |
|---|---|---|
| Framework | Vue 3 (`<script setup>`) + TypeScript | Required by the assignment; SFC + TS gives compile-time safety on top of refactor-friendly templates. |
| Build | Vite 8 | Standard for new Vue 3 projects; fast HMR, esbuild for dev. |
| Routing | Vue Router 4 | The de-facto Vue 3 router. |
| Client state | Pinia 2 | Modern Vue 3 state library. We use it sparingly — only for client state that needs to outlive a route (auth token, role, email). |
| Server state | TanStack Vue Query 5 | Owns everything that comes from the API: payments list, summary, mutations. Provides cache, retry, optimistic updates, refetch-on-focus out of the box. |
| HTTP | axios + `@hey-api/openapi-ts` SDK | Axios for the request/response interceptors (Bearer + 401 redirect); the SDK is regenerated from `openapi.yaml` so types and method signatures match the backend exactly. |
| Forms | vee-validate 4 + zod 3 | Schema-driven validation, inline errors, typed values. |
| UI primitives | shadcn-vue components written into `src/components/ui` | We own the source — no opaque component library. Tailored to the small surface this app needs (button/input/label/card/badge/skeleton/dialog). |
| Styling | Tailwind 3 + CSS variables for theme tokens | Ergonomic, theme tokens make dark-mode trivial. |
| Icons | `lucide-vue-next` | Tree-shaken, consistent with the shadcn aesthetic. |
| Toasts | `vue-sonner` | Lightweight, accessible. |
| Tests | Vitest + Testing Library + MSW + happy-dom | Mirrors how the app actually runs in the browser. |

### Why split Pinia and Vue Query?

The two libraries solve different problems and mixing them tends to produce bespoke caching code. We let Vue Query do all server-state caching/invalidation/optimism, while Pinia stores the small slice of client state (auth) that any composable might need. The result: the auth store fits in 60 lines, the review mutation is ~30 lines including optimistic update + rollback, and tests don't have to mock a global store.

## Project layout

```
frontend/
├── index.html
├── vite.config.ts                # alias @ = src/, dev proxy /api -> :8080, vitest config
├── tailwind.config.js
├── tsconfig.app.json             # strict TS, paths: { "@/*": ["./src/*"] }
├── eslint.config.js              # vue + ts + prettier flat config
├── nginx.conf                    # production reverse-proxy config (used by the docker image)
├── openapi-ts.config.ts          # codegen target for src/api/generated
├── Dockerfile                    # multi-stage build -> nginx
├── src/
│   ├── api/
│   │   ├── http.ts               # axios instance + interceptors, wires the hey-api SDK
│   │   ├── query-client.ts       # vue-query QueryClient defaults
│   │   └── generated/            # GIT-IGNORED: regenerated on predev/prebuild
│   ├── assets/index.css          # tailwind + theme tokens
│   ├── components/
│   │   ├── AppLayout.vue         # sticky topbar + container
│   │   └── ui/                   # button, input, label, card, badge, skeleton, dialog
│   ├── features/
│   │   ├── auth/                 # LoginPage + use-login composable
│   │   └── payments/             # DashboardPage, SummaryWidget, PaymentsTable, queries.ts
│   ├── lib/
│   │   ├── utils.ts              # cn() helper (clsx + tailwind-merge)
│   │   └── format.ts             # IDR currency, Intl date, shortId
│   ├── router/index.ts           # routes + guard (requiresAuth, redirect on /login when authed)
│   ├── stores/auth.ts            # Pinia: token, email, role, persisted to localStorage
│   ├── test/                     # vitest setup, MSW handlers, fixtures, renderWithStack
│   ├── App.vue
│   └── main.ts
└── package.json
```

## Scripts

```bash
npm install                # installs deps; gen:api runs implicitly via predev/prebuild
npm run dev                # vite, port 5173, /api -> :8080 proxy
npm run build              # vue-tsc check + production build into dist/
npm run preview            # serves dist/ on :4173
npm run typecheck          # vue-tsc --noEmit
npm run lint               # eslint . --ext .ts,.vue
npm run lint:fix           # autofix
npm run format             # prettier write
npm run test               # vitest run
npm run test:watch         # vitest in watch mode
npm run test:cover         # vitest + v8 coverage report
npm run gen:api            # regenerate src/api/generated from ../openapi.yaml
```

`predev` and `prebuild` both call `gen:api`, so the SDK is always in sync with the spec without you having to remember.

## Environment

`cp .env.example .env` and adjust if needed:

| Var | Default | Notes |
|---|---|---|
| `VITE_API_BASE_URL` | `http://localhost:8080` | Backend origin. In Docker, the build arg is `/api` so the SPA talks to nginx and nginx proxies to the backend service. |

## OpenAPI codegen pipeline

`openapi-ts.config.ts` points at `../openapi.yaml`. `npm run gen:api` writes `src/api/generated/{client/, sdk.gen.ts, types.gen.ts}`. The `client.gen.ts` exposes a `client` you can swap with your own axios instance — that's what `api/http.ts` does:

```ts
client.setConfig({ axios: http, baseURL })
```

After that line every generated function (`loginUser`, `listPayments`, `reviewPayment`, …) inherits the Bearer-token interceptor and the 401 redirect.

## Testing

See the **Testing strategy** section in the root README for the philosophy. Key files:

- `src/test/msw/handlers.ts` — stateful request handlers (a mutation actually changes the in-memory list, so optimistic updates round-trip).
- `src/test/utils.ts` — `renderWithStack(component)` returns a render with Pinia + Vue Query + memory router pre-mounted, matching production composition.
- `src/test/setup.ts` — boots the MSW server, resets mock state and `localStorage` between tests.

Coverage targets:
- `features/auth`: **91.1% statements / 93.0% lines**
- `features/payments`: **80.4% / 85.3%**
- `stores/auth`: **96.7% / 100%**
- `lib/format`: **100% / 100%**

## Production build (without Docker)

```bash
npm run build
npm run preview                  # static server on :4173
```

For a real deploy you'd typically serve `dist/` from any static host (nginx, Cloudflare Pages, Netlify) and route `/api/*` to the Go service. The repo's `nginx.conf` is exactly that: a reference SPA fallback + reverse proxy config that the Docker image uses.
