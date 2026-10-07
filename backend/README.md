# Shatranj backend

Go HTTP/WebSocket API for Shatranj. The canonical backend architecture lives
under `internal/...`, with `cmd/server` for the API process and `cmd/worker`
for background workers. PostgreSQL is the durable store. Redis is optional and
powers distributed rate limiting and presence when configured, falling back to
in-memory stores when it is not.

## Setup

Install Go 1.23+ and PostgreSQL. Copy/configure environment values before
starting the API:

The repository includes `.env.example` as a reference. The Go application
currently reads environment variables from the process and does not load a
`.env` file automatically. For a local PowerShell setup, copy it to
`.env.ps1`, review the placeholder values, then load it with
`. .\.env.ps1` before running the server or worker. Do not commit files
containing real credentials.

| Variable | Required | Default / purpose |
| --- | --- | --- |
| `DATABASE_URL` | For feature APIs/workers | PostgreSQL connection string. Without it, only health and an in-memory worker scaffold are available. |
| `SHATRANJ_JWT_SECRET` | With `DATABASE_URL` | Random signing secret of at least 32 bytes. |
| `SHATRANJ_HTTP_ADDR` | No | Listen address; defaults to `:8080`. |
| `REDIS_URL` | No | `redis://` or `rediss://` URL. When set it backs distributed rate limiting and presence. If it is empty, malformed, or unreachable the process logs a warning and uses in-memory fallbacks instead of blocking startup. |
| `SHATRANJ_CORS_ORIGINS` | No | Comma-separated allowed browser origins. |
| `SHATRANJ_LOG_LEVEL` | No | One of `debug`, `info`, `warn`, `error`; defaults to `info`. |
| `SHATRANJ_LOG_FORMAT` | No | `text` (default) or `json` for structured logs. |
| `SHATRANJ_SHUTDOWN_TIMEOUT` | No | Positive Go duration; defaults to `10s`. |
| `SHATRANJ_PREMIUM_PRICE_PAISE` | No | Monthly premium price in paise; defaults to `49900` (INR 499). |
| `RAZORPAY_KEY_ID`, `RAZORPAY_KEY_SECRET`, `RAZORPAY_WEBHOOK_SECRET` | For Razorpay | Configure all three together to enable payment orders and signed webhooks. |
| `STOCKFISH_PATH` | No | Stockfish executable; review requests remain unavailable without it. |

PostgreSQL is pinged during startup; a connection or configuration failure stops
the process so a misconfigured database is never silently ignored. Redis is
treated as optional: when `REDIS_URL` is set, `cmd/server` and `cmd/worker`
attempt to connect and, on failure, log a warning and continue with in-memory
fallbacks. Leaving `REDIS_URL` empty disables Redis outright.

## Redis

Redis is an optional accelerator, never a source of truth. `internal/cache`
selects a backend at startup and exposes it through a single `cache.Client`:

- **Redis-backed when configured:** atomic per-window request counters for
  auth rate limiting (`register`, `login`, and the `otp/request` guard) and
  TTL-based user presence records (`presence:user:<id>`, refreshed on connect
  and cleared on disconnect).
- **In-memory fallback otherwise:** a process-local rate limiter and presence
  map, identical in behavior within a single instance but not shared across
  processes.
- **Still PostgreSQL-owned:** matchmaking uses a transactional
  `pg_advisory_xact_lock` queue, and the worker claims jobs with
  `FOR UPDATE SKIP LOCKED`; neither depends on Redis.
- **Future work:** a Redis-backed background job queue and wiring the
  `internal/cache` lock abstraction (`cache.WithLock`) into a call site. The
  abstraction is implemented and tested but kept unwired because the database
  claim and advisory locks already provide the required guarantees.

## Migrations

Apply every SQL file in `internal/database/migrations` in ascending filename
order, once per database. `000_extensions.sql` enables `pgcrypto`, which is
required by migrations that use `gen_random_uuid()`. Migration numbering
intentionally skips unused feature stages; do not invent missing migrations.
For PowerShell:

```powershell
Get-ChildItem .\internal\database\migrations\*.sql |
  Sort-Object Name |
  ForEach-Object { psql $env:DATABASE_URL -v ON_ERROR_STOP=1 -f $_.FullName }
```

To bootstrap an administrator, promote a trusted account using a secured
database session after migrations:

```sql
UPDATE users SET role = 'admin' WHERE email = 'admin@example.com';
```

## Run and verify

```powershell
go mod download
go run .\cmd\server
```

In another terminal, start background workers:

```powershell
go run .\cmd\worker
```

With PostgreSQL, the worker processes queued game-analysis jobs and marks
analysis unavailable when Stockfish is not configured. Without PostgreSQL, it
runs the cleanly stoppable in-memory queue scaffold. Both processes stop on
Ctrl+C or termination signals.

Run the full checks:

```powershell
go fmt ./...
go list ./...
go test ./...
go vet ./...
go build ./...
```

`GET /healthz` is available without database configuration. Feature APIs return
a structured service-unavailable response until PostgreSQL and JWT signing are
configured.

## API areas

All feature APIs are versioned under `/api/v1`. Protected routes require a
Bearer access token. `/api/v1/admin/*` uses token authentication and a
database-backed administrator role check. Public routes include health,
registration/login, published news, subscription plan information, and the
Razorpay webhook (which requires a valid provider signature).

- **Auth, users, profiles:** `/auth`, `/users`, `/profiles`
- **Games, matchmaking, ratings, leaderboard, review:** `/games`,
  `/matchmaking`, `/ratings`, `/leaderboard`, `/games/{gameID}/review`
- **Puzzles and lessons:** `/puzzles`, `/lessons`
- **Community:** `/tournaments`, `/clubs`, `/friends`, `/chat`
- **Product features:** `/achievements`, `/notifications`, `/subscriptions`,
  `/payments`, `/coaches`, `/news`
- **Administration:** `/admin` dashboard, user and coach moderation, news
  publishing, and payment review

Game and chat WebSockets are mounted at
`/api/v1/games/{gameID}/ws` and `/api/v1/chat/rooms/{roomID}/ws`; both require
authentication and authorize room membership before upgrading.

HTTP success bodies are JSON values written through the shared JSON helper.
Errors use the common JSON error shape with a stable code and safe message.

## Current limitations

- Redis is optional and currently powers only rate limiting and presence. Auth
  rate limits and WebSocket presence use Redis when configured and per-process
  memory otherwise. Matchmaking is PostgreSQL-backed (advisory locks, not
  Redis), the background worker is database- or in-memory-based, and
  Redis-backed job queues and wired locks remain future work.
- OTP login has request/response structure but no email provider or durable OTP
  store yet.
- Object storage is represented by `internal/storage` with a development-safe
  in-memory implementation. Cloud provider integration is intentionally not
  wired yet.
- Stockfish analysis requires `STOCKFISH_PATH`; without it, review requests are
  recorded as unavailable rather than crashing workers.
- Razorpay payment orders and webhooks require all Razorpay environment values.
