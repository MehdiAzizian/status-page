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
