# Cheatsheet

Short hints to remember. ⭐ = asked in interviews.

## Big ideas

- ⭐ **App reports, platform acts.** App says "I'm sick" (`/healthz` → 503).
  Kubernetes / Docker / load balancer decides what to do.
- ⭐ **Config in env vars, not in code.** Same code runs on laptop, Docker, AWS.
- ⭐ **Fail fast.** Missing config → crash at start with a clear message.
- ⭐ **Pin versions.** `postgres:16`, never `latest`. Same input → same result.
- ⭐ **Lock files = receipts.** `go.sum`, `package-lock.json`,
  `.terraform.lock.hcl`. Exact versions + fingerprints.
- ⭐ **Containers are disposable.** Delete container → data gone, unless volume.
- ⭐ **Restart policy.** Container without one stays dead after a reboot.

## HTTP status codes ⭐

| Code | Meaning | Remember |
|---|---|---|
| 200 | OK | all good |
| 404 | Not Found | wrong path |
| 405 | Method Not Allowed | right path, wrong GET/POST |
| 500 | Internal Server Error | app bug |
| 503 | Service Unavailable | app alive, dependency down |

## Errors → meaning ⭐

| Error | Means |
|---|---|
| `connection refused` | nothing is listening on that port |
| `go.mod file not found` | wrong folder → `pwd` |
| `not a git repository` | wrong folder → `pwd` |
| `missing go.sum entry` | run `go mod tidy` |
| `/healthz` 503 | check the DB: `docker ps -a` |

## Git

```bash
git status                 # what changed?
git add . && git commit -m "msg" && git push
git log --oneline          # history
git remote -v              # where is GitHub?
```

## Go

```bash
go run .                   # compile + run
go mod tidy                # fix go.mod / go.sum
```
`main.go` = front door. `go.mod` = shopping list. `go.sum` = receipt.

## Docker ⭐

```bash
docker run -d --name X -e KEY=val -p HOST:CONTAINER image:tag
docker ps                  # running containers
docker ps -a               # ALL containers, also stopped ones
docker start X / docker stop X
docker exec -it X bash     # go inside a container
docker logs X              # what did it print?
```

## Postgres

```bash
docker exec -it pg psql -U postgres
```
`\q` quit · `\dt` list tables · `SELECT * FROM components;`
URL format: `postgres://user:password@host:port/dbname`

## Troubleshooting loop ⭐

Symptom → Evidence → Hypothesis → Test → Fix → Verify

## nano

`^O` save · `^X` exit · `^W` search · `^K` cut line · `^U` paste

---

# Tips (one-liners, keep growing)

- Our app is **one service**. `main.go` is just where it starts.
- `/healthz` = the app checking **itself**: "can I reach my DB?"
- **K8s is the doctor.** It calls `/healthz` every few seconds.
  App only reports. K8s takes action (stop traffic, restart).
- Every language does `/healthz`. It's just a URL: 200 or 503.
- `go.mod` = shopping list. `go.sum` = receipt.
- `connection refused` = nobody is listening on that port.
- 404 = wrong path. 405 = right path, wrong method.
- Git is on your laptop. GitHub is a copy on the internet. `push` = send.
- Commit = a photo of your code, with an ID (SHA).
- `docker ps` = alive only. `docker ps -a` = alive + dead.
- Mac reboot → containers without restart policy stay dead.
- Starting ≠ ready. A DB needs a few seconds after `start`.
  → That's why deploys **retry** health checks, not check once.
- Wrong folder is the #1 beginner error. `pwd` first.
- **Why a DB?** Status page must remember "API = down". App restarts → memory
  is wiped. DB keeps data. Almost every real app = app + DB.
- **Why /healthz? Who uses it:**
  deploy pipeline (bad → rollback) · Kubernetes (bad → restart / no traffic) ·
  load balancer (bad → skip this server) · monitoring (bad → alert).
- **App = server.** It waits on a port forever for requests.
  `listening on :8081` = "I'm open, knock on door 8081".
- **Log line** = time + message. First thing you read when something breaks.
- **2 terminals** = one runs the server (busy, waiting), one is the client (curl).
- `Ctrl+C` = stop the program running in this terminal.
- `curl` = a browser for the terminal. Sends a request, prints the answer.
- **Close terminal → its programs die + its `export`s are gone.**
  `docker run -d` survives: it lives in Docker, not in the terminal.
- `echo $VAR` = read an env var. Empty line = not set.
- `ping` = is the **computer** alive? `curl` = is the **app** alive?
- `$` = "value of". `echo DATABASE_URL` prints the word. `echo $DATABASE_URL` prints the value.
- `docker run` = buy a **new** car. `docker start` = turn the key in **your** car.
  `start` keeps your data. New `run` = empty DB.
- **Restart policy** (`--restart unless-stopped`) = container wakes up when Docker restarts.
- `go mod init` = birth certificate. Once per project. `go run .` = every time.
- `DATABASE_URL` = DB address + password. No URL → app can't start.
- **Start dependencies first.** DB → then app.
- Config goes in the **environment**, not the code. Code only **reads** it (`os.Getenv`).
- `export` = sticky note on the terminal. Programs started from it can read it.
- ⭐ **Same idea everywhere, different sticky note:**
  Mac `export` · Docker `-e` · Compose `environment:` · K8s ConfigMap/Secret · GitHub Actions Secrets.
- Why not in code? 🔐 password would be public · 🌍 one code, different URL per place.
- ⭐ **12-factor app** = config lives in the environment. Same code everywhere.
- `.env` file = sticky notes saved in a file. Always in `.gitignore`.
- ⭐ **Config vs Secret.** Port = config (ConfigMap). Password = secret (Secret).
- `struct` = a box with labeled slots. `Component` = one row of the table, in Go.
- `undefined: X` = you use X but forgot to `import` it.
- Go compiles at start → code change = Ctrl+C + `go run .` again.
- JSON: `[ ]` = list · `{ }` = one object. DB row → Go struct → JSON → client.
- ⭐ **SQL injection**: never glue user input into SQL. Use placeholders (`$1`).
- 400 = **your** request is bad (abc is not a number). 404 = thing doesn't exist.
- ⭐ **4xx = client's fault. 5xx = server's fault** → alerts fire on 5xx, not 4xx.
- GET = **read**. POST = **change**. `curl -X POST -d '{...}'` = send a body.
- Encode = Go → JSON (send). Decode = JSON → Go (receive).
- ⭐ **Validate input. Never trust the client.**
- 204 = "done, nothing to send back".
