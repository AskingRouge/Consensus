# Consensus

A real-time, multiplayer tool for a research team to collaboratively score and discuss LLM outputs against a shared evaluation rubric. Reviewers join the same session, see the same batch of model outputs, submit scores live, and see each other's scores and presence update in real time. Items where reviewers disagree get flagged for discussion.

## Status

Built in phases; this checklist tracks progress.

- [x] **Phase 1** — Go backend skeleton, MongoDB models, REST CRUD for sessions/items (no auth, no WebSockets yet)
- [x] **Phase 2** — Auth + allowlist middleware *(username/password for now — see note below)*
- [x] **Phase 3** — WebSocket hub: connection management, broadcasting presence and (from Phase 5) score updates
- [x] **Phase 4** — React frontend: login/signup, session list, session view, item cards, live presence
- [x] **Phase 5** — Rubric scoring UI + disagreement flagging logic (backend + frontend)
- [x] **Phase 6** — Comment threads, live via WebSocket
- [ ] Phase 7 — Export functionality (JSON/CSV)
- [ ] Phase 8 — Terraform infrastructure + Azure deployment *(deliberately deferred — see note below)*
- [ ] Phase 9 — Final README + architecture write-up

> **Azure/Terraform note:** deployment is being held off for now by request. Everything is built so the eventual move is a config change, not a rewrite — `MONGODB_URI` is the only thing that changes to go from the local `docker-compose` Mongo to MongoDB Atlas, and the backend already builds into a portable Docker image. The `infra/` Terraform and Azure-specific docs will be added in Phase 8 when you're ready.

> **Auth note:** Phase 2 is built as **username/password** rather than Google OAuth, by request, to unblock development without a Google Cloud Console detour. It's a real implementation (bcrypt-hashed passwords, database-backed sessions, an email allowlist that fails closed) — not a stub — but it's deliberately temporary: no password-reset flow exists, since that's meaningful additional surface area (email infra or a manual admin process) not worth building for a scheme that's getting replaced. Every place that cares "who is this request from" reads it through `auth.UserEmail(r)`, so swapping in real Google OAuth later touches `internal/auth` and the login screen only — not the session/item/score/comment handlers.

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
│   │   ├── auth/            # username/password auth + allowlist (Phase 2; OAuth later)
│   │   ├── scoring/           # rubric validation + disagreement detection (pure, heavily tested)
│   │   ├── ws/                  # WebSocket hub: connections, presence, broadcast
│   │   └── export/                # JSON/CSV export (Phase 7)
│   └── Dockerfile
├── frontend/           # React app (Vite)
│   └── src/
│       ├── components/  # AuthForm, SessionList, SessionView, ItemCard, PresenceIndicator, RubricScorer, CommentThread
│       ├── hooks/         # useWebSocket — reconnect-with-backoff, presence/score/comment state
│       ├── api.js          # REST client (fetch + cookies)
│       └── App.jsx          # routing + auth gate
├── infra/              # Terraform (Phase 8, deferred)
└── docker-compose.yml  # local MongoDB for dev
```

## Prerequisites

- **Go 1.22+** — [go.dev/dl](https://go.dev/dl/)
- **Docker** (for local MongoDB) — or a MongoDB Atlas connection string, see below
- **Node 20.19+ or 22.12+** — the frontend's toolchain (Vite 8) requires it; older Node 20.x releases fail to install

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

Set `ALLOWED_EMAILS` to a comma-separated list of the emails allowed to sign up (e.g. `you@unc.edu,teammate@unc.edu`) — signup is rejected for anyone else. `SESSION_SECRET` can be left blank for local dev (the server generates a random one and logs a warning); set a real one — `openssl rand -base64 32` — before deploying anywhere, or every login breaks on restart.

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

The frontend (step 6 below) is the normal way to use this, but the REST API works standalone too — useful for scripting or debugging. Pass `-c cookies.txt` on the request that logs you in and `-b cookies.txt` on every request after, so your session cookie carries over between calls.

```bash
# health check (no auth needed)
curl http://localhost:8080/healthz

# sign up (email must be in ALLOWED_EMAILS; password 8+ chars) — logs you
# in immediately, saving the session cookie to cookies.txt
curl -c cookies.txt -X POST http://localhost:8080/api/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "you@unc.edu", "password": "a-real-password"}'

# on later visits, log in instead of signing up again
curl -c cookies.txt -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "you@unc.edu", "password": "a-real-password"}'

# who am I?
curl -b cookies.txt http://localhost:8080/api/auth/me

# create a session
curl -b cookies.txt -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{"name": "Week 3 Course-Scope Eval"}'

# list sessions
curl -b cookies.txt http://localhost:8080/api/sessions

# upload a batch of items (JSON or CSV — matches the research harness's
# question / retrieved_context / generated_answer output shape)
curl -b cookies.txt -X POST "http://localhost:8080/api/sessions/<SESSION_ID>/items/upload" \
  -F "file=@items.json"

# list items in a session
curl -b cookies.txt http://localhost:8080/api/sessions/<SESSION_ID>/items

# log out
curl -b cookies.txt -c cookies.txt -X POST http://localhost:8080/api/auth/logout
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

### 5. Try the WebSocket (live presence)

Each review session has a WebSocket at `ws://localhost:8080/ws/sessions/<SESSION_ID>` — it authenticates with the exact same session cookie as the REST API, so there's no separate WebSocket login step. Every connected client immediately gets, and receives again on any change, a full presence snapshot:

```json
{"type": "presence", "payload": [
  {"reviewer_email": "you@unc.edu", "item_id": "6a9f..."},
  {"reviewer_email": "teammate@unc.edu"}
]}
```

A client tells the server which item it's looking at by sending:

```json
{"type": "view_item", "item_id": "6a9f..."}
```

(`item_id: ""` means "not viewing a specific item" — e.g. back on the session's item list.) The frontend (step 6) drives this automatically, but you can also try it directly with any WebSocket client — e.g. [websocat](https://github.com/vi/websocat): `websocat "ws://localhost:8080/ws/sessions/<SESSION_ID>" -H "Cookie: consensus_session=<value from cookies.txt>"`.

Score updates broadcast on this same connection as `score_update` messages (`{"score": {...}, "flagged": bool, "disagreeing_dimensions": [...]}`); new comments broadcast as `comment` messages carrying the comment itself.

### 6. Run the frontend

With the backend still running (steps 1-3 above):

```bash
cd frontend
npm install   # first time only
npm run dev
```

Open **http://localhost:5173**. Sign up with an allowlisted email, create a session, upload a JSON/CSV batch (a file picker button, no curl needed now), and click into an item — you'll see the "Live" connection badge and everyone's presence avatars update in real time as reviewers move between items. Sharing a session's URL works as an invite link: an unauthenticated visitor is sent to log in first and is then dropped back on that exact session, not the session list.

Click a rubric score and submit — it's saved immediately and pushed live to every other connected reviewer. Once two reviewers have scored the same item and their ratings differ by at least the session's disagreement threshold (2, by default, on a scale dimension; any split verdict on a pass/fail dimension), the item is flagged "Needs discussion" for everyone, with a filter in the item list to jump straight to flagged items.

If you want to see two reviewers disagree, open the same session URL in two browsers (or a normal + private window) signed in as two different allowlisted accounts, and score the same item differently.

Every item also has a discussion thread — post a comment and it appears instantly for every other connected reviewer, oldest first, with your own comments visually distinguished from others'.

### 7. Run the tests

```bash
cd backend
go test ./...

cd frontend
npm run build   # also runs Vite's own type/bundle checks
npm run lint
```

## Endpoints so far

| Method | Path | Auth required | Description |
|---|---|---|---|
| GET | `/healthz` | no | Liveness check |
| POST | `/api/auth/signup` | no | Create an account (email must be allowlisted); logs in |
| POST | `/api/auth/login` | no | Log in with email/password |
| POST | `/api/auth/logout` | no | Clear the current session |
| GET | `/api/auth/me` | yes | Current logged-in user's email |
| POST | `/api/sessions` | yes | Create a session (default rubric if none supplied) |
| GET | `/api/sessions` | yes | List all sessions |
| GET | `/api/sessions/{id}` | yes | Get one session |
| GET | `/api/sessions/{id}/items` | yes | List items in a session |
| POST | `/api/sessions/{id}/items/upload` | yes | Upload a JSON or CSV batch of items |
| POST | `/api/sessions/{id}/items/{itemID}/scores` | yes | Submit/update your score for one item |
| GET | `/api/sessions/{id}/scores` | yes | List every score in the session |
| POST | `/api/sessions/{id}/items/{itemID}/comments` | yes | Post a comment on one item |
| GET | `/api/sessions/{id}/comments` | yes | List every comment in the session |
| GET (WS) | `/ws/sessions/{id}` | yes | Live presence, score, and comment updates |

More to come: export.
