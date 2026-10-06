# Shatranj backend

Go HTTP/WebSocket API for Shatranj. PostgreSQL is the durable store; Redis is
optional and currently used as an optional startup-checked infrastructure
dependency.

## Setup

Install Go 1.23+ and PostgreSQL. Copy/configure environment values before
starting the API:

| Variable | Required | Default / purpose |
| --- | --- | --- |
| `DATABASE_URL` | For feature APIs/workers | PostgreSQL connection string. Without it, only health and an in-memory worker scaffold are available. |
| `SHATRANJ_JWT_SECRET` | With `DATABASE_URL` | Random signing secret of at least 32 bytes. |
| `SHATRANJ_HTTP_ADDR` | No | Listen address; defaults to `:8080`. |
| `REDIS_URL` | No | `redis://` or `rediss://` URL. If set, connection failures stop startup. |
| `SHATRANJ_CORS_ORIGINS` | No | Comma-separated allowed browser origins. |
| `SHATRANJ_SHUTDOWN_TIMEOUT` | No | Positive Go duration; defaults to `10s`. |
| `SHATRANJ_PREMIUM_PRICE_PAISE` | No | Monthly premium price in paise; defaults to `49900` (INR 499). |
| `RAZORPAY_KEY_ID`, `RAZORPAY_KEY_SECRET`, `RAZORPAY_WEBHOOK_SECRET` | For Razorpay | Configure all three together to enable payment orders and signed webhooks. |
| `STOCKFISH_PATH` | No | Stockfish executable; review requests remain unavailable without it. |

When configured, PostgreSQL and Redis are pinged during process startup. A
connection or configuration failure is returned and logged rather than
silently ignored.

## Migrations

Apply every SQL file in `internal/database/migrations` in ascending filename
order, once per database. Migration numbering intentionally skips unused
feature stages; do not invent missing migrations. For PowerShell:

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
Get-ChildItem .\cmd,.\internal -Recurse -Filter *.go |
  ForEach-Object { gofmt -w $_.FullName }
go test ./...
go build ./...
go vet ./...
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
- **Games, matchmaking, ratings, review:** `/games`, `/matchmaking`, `/ratings`
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
