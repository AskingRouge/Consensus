# Consensus

A real-time, multiplayer tool for a research team to collaboratively score and discuss LLM outputs against a shared evaluation rubric. Reviewers join the same session, see the same batch of model outputs, submit scores live, and see each other's scores and presence update in real time. Items where reviewers disagree get flagged for discussion.

## Status

Built in phases; this checklist tracks progress.

- [x] **Phase 1** — Go backend skeleton, MongoDB models, REST CRUD for sessions/items (no auth, no WebSockets yet)
- [ ] Phase 2 — Google OAuth + allowlist middleware
- [ ] Phase 3 — WebSocket hub: connection management, broadcasting score updates and presence
- [ ] Phase 4 — React frontend: session list, session view, item cards, live score display
- [ ] Phase 5 — Rubric scoring UI + disagreement flagging logic (backend + frontend)
- [ ] Phase 6 — Comment threads, live via WebSocket
- [ ] Phase 7 — Export functionality (JSON/CSV)
- [ ] Phase 8 — Terraform infrastructure + Azure deployment *(deliberately deferred — see note below)*
- [ ] Phase 9 — Final README + architecture write-up

> **Azure/Terraform note:** deployment is being held off for now by request. Everything is built so the eventual move is a config change, not a rewrite — `MONGODB_URI` is the only thing that changes to go from the local `docker-compose` Mongo to MongoDB Atlas, and the backend already builds into a portable Docker image. The `infra/` Terraform and Azure-specific docs will be added in Phase 8 when you're ready.

## Stack

| Piece | Choice | Why |
|---|---|---|
| Backend | Go | Goroutines make many concurrent WebSocket connections (one per active reviewer) cheap and simple to reason about. |
| Realtime | WebSockets | Live score updates, presence, and disagreement flags need push, not polling — this is the core mechanic, not a bolt-on. |
| Database | MongoDB | Rubrics and item shapes vary by research batch; a document store fits that better than a fixed relational schema. |
| Frontend | React | A multi-user dashboard with frequently-updating state (scores, presence, flags) fits component-based UI with real state management. |
| Deployment | Azure Container Apps + MongoDB Atlas | *(deferred, see above)* |
| IaC | Terraform | *(deferred, see above)* |

A full architecture write-up (WebSocket broadcast model, why Go/Mongo) lands in Phase 9.

## Project layout

```
consensus/
├── backend/            # Go API server
│   ├── main.go
│   ├── internal/
│   │   ├── config/     # env var loading
│   │   ├── models/     # session, item, score, comment structs
│   │   ├── db/          # MongoDB connection + indexes
│   │   ├── parse/        # JSON/CSV item-batch parsing
│   │   ├── handlers/      # REST endpoints
│   │   ├── auth/            # Google OAuth + allowlist (Phase 2)
│   │   ├── ws/                # WebSocket hub (Phase 3)
│   │   └── export/              # JSON/CSV export (Phase 7)
│   └── Dockerfile
├── frontend/           # React app (Phase 4+)
├── infra/              # Terraform (Phase 8, deferred)
└── docker-compose.yml  # local MongoDB for dev
```

## Prerequisites

- **Go 1.22+** — [go.dev/dl](https://go.dev/dl/)
- **Docker** (for local MongoDB) — or a MongoDB Atlas connection string, see below
- **Node 20+** — needed starting Phase 4

## Local setup (what's built so far)

### 1. Start MongoDB

Easiest option — a local container via docker-compose (from the repo root):

```bash
docker compose up -d
```

This matches what MongoDB Atlas provides, so moving to Atlas later is just swapping `MONGODB_URI` — no code changes. If you'd rather set up Atlas's free tier now instead of running Mongo locally:

1. Go to [mongodb.com/cloud/atlas/register](https://www.mongodb.com/cloud/atlas/register) and create a free account.
2. Create a free (M0) cluster.
3. Under **Database Access**, add a database user with a username/password.
4. Under **Network Access**, add your current IP (or `0.0.0.0/0` for local dev only).
5. Click **Connect > Drivers**, copy the `mongodb+srv://...` connection string, and use it as `MONGODB_URI` below.

### 2. Configure the backend

```bash
cd backend
cp .env.example .env
```

The defaults in `.env.example` already point at the local docker-compose Mongo, so no edits are needed for local dev. If you're using Atlas, replace `MONGODB_URI` with your `mongodb+srv://...` string.

### 3. Run it

```bash
cd backend
go mod tidy   # first time only, resolves dependencies
go run .
```

You should see:

```
consensus backend listening on :8080 (frontend origin: http://localhost:5173)
```

### 4. Try the API

There's no frontend yet and no auth yet — requests authenticate via a dev-only `X-Dev-User-Email` header (replaced by real Google OAuth in Phase 2).

```bash
# health check
curl http://localhost:8080/healthz

# create a session
curl -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -H "X-Dev-User-Email: you@unc.edu" \
  -d '{"name": "Week 3 Course-Scope Eval"}'

# list sessions
curl http://localhost:8080/api/sessions

# upload a batch of items (JSON or CSV — matches the research harness's
# question / retrieved_context / generated_answer output shape)
curl -X POST "http://localhost:8080/api/sessions/<SESSION_ID>/items/upload" \
  -H "X-Dev-User-Email: you@unc.edu" \
  -F "file=@items.json"

# list items in a session
curl http://localhost:8080/api/sessions/<SESSION_ID>/items
```

Example `items.json`:

```json
[
  {
    "question": "What is the withdrawal deadline?",
    "retrieved_context": "syllabus.pdf p2",
    "generated_answer": "The withdrawal deadline is March 3rd."
  }
]
```

A CSV with the header `question,retrieved_context,generated_answer` works the same way — pass `?format=csv` if the filename doesn't end in `.csv`.

### 5. Run the tests

```bash
cd backend
go test ./...
```

## Endpoints so far

| Method | Path | Description |
|---|---|---|
| GET | `/healthz` | Liveness check |
| POST | `/api/sessions` | Create a session (default rubric if none supplied) |
| GET | `/api/sessions` | List all sessions |
| GET | `/api/sessions/{id}` | Get one session |
| GET | `/api/sessions/{id}/items` | List items in a session |
| POST | `/api/sessions/{id}/items/upload` | Upload a JSON or CSV batch of items |

More to come: score submission, comments, WebSocket endpoint, export.
