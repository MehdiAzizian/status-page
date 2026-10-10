# Learning Log

What I did, why, how. Read this when I come back.

Goal of the project: a small Go status page, taken from my laptop to AWS with
automatic deploy, rollback and Kubernetes. The app is small on purpose. The
real skills are Docker, Linux, CI/CD, AWS, Terraform, Ansible.

---

## Phase 0: build the app locally

### Step 1: project skeleton

**What:** empty Go project under git.
**Why:** everything later (CI/CD, deploy) starts from a git push.

```bash
git init -b main                                   # create local repo, branch "main"
go mod init github.com/MehdiAzizian/status-page    # create go.mod
```

- `go.mod` = the ID card of a Go project: module name + Go version + dependencies.
- `.gitignore` = files git must never track. `.env` and `*.tfstate` are in it
  because they contain passwords.

### Step 2: first commit + push to GitHub

**What:** save a snapshot, send it to GitHub.
**Why:** git is local only. GitHub is the remote copy that CI/CD will watch.

```bash
git add .                      # put files in the staging area
git commit -m "message"        # save a snapshot (gets a SHA id)
gh auth login                  # log terminal into GitHub
gh repo create status-page --public --source=. --remote=origin --push
git remote -v                  # show where "origin" points
```

- Staging area = waiting room for the next commit.
- Commit = snapshot with a unique SHA. Later, Docker images are tagged with it.
- `origin` = nickname for the GitHub URL.
- Hint: `gh auth login` shows a code in the terminal, not on the phone.
  Check the IP on the GitHub authorize page before clicking.

### Step 3: basic HTTP server

**What:** tiny Go web server with `GET /healthz`.
**Why:** every deploy tool (CI, Docker, Kubernetes) asks "is the app alive?"
by calling `/healthz`. Later it will also check the database.

```bash
go run .                              # compile + run (must be inside status-page/)
curl -i localhost:8080/healthz        # -i shows status code + headers
```

- Router (`ServeMux`) = maps URL + method to a function.
- Experiment: changed route to `POST`, curl gave **405 Method Not Allowed**.
- Interview: **404** = path does not exist. **405** = path exists, wrong method.
- Mistake I made: ran `go run .` and `git` from `P2/` instead of
  `P2/status-page/`. Fix: `pwd` first.

### Step 4: PostgreSQL in Docker

**What:** run Postgres in a container, create a table by hand.
**Why:** the app needs a database. Running DBs in Docker is daily DevOps work.

```bash
docker run -d --name pg -e POSTGRES_PASSWORD=secret -p 5432:5432 postgres:16
docker ps                               # list running containers
docker exec -it pg psql -U postgres     # open SQL shell inside the container
```

| Flag | Meaning |
|---|---|
| `-d` | run in background |
| `--name pg` | short name to use later |
| `-e KEY=value` | environment variable inside the container |
| `-p 5432:5432` | Mac port → container port |
| `postgres:16` | image:version. Always pin a version, never `latest` |

```sql
CREATE TABLE components (
  id     SERIAL PRIMARY KEY,                    -- auto 1,2,3...
  name   TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'operational'
);
INSERT INTO components (name) VALUES ('Website'), ('Login'), ('API');
SELECT * FROM components;
```

- Leave psql with `\q`.
- Data lives inside the container. Delete the container → data is gone.
  (Fix later with a volume.)

### Step 5: connect Go to Postgres + real /healthz

**What:** app reads config from env vars, connects to Postgres, `/healthz` pings the DB.
**Why:** "app running" ≠ "app working". If DB is down, health check must fail.

```bash
go get github.com/jackc/pgx/v5        # add Postgres driver
go mod tidy                           # fix go.mod/go.sum (missing go.sum entry error)
export DATABASE_URL=postgres://postgres:secret@localhost:5432/postgres
go run .                              # now on port 8081 (8080 used by jobmon)
curl -i localhost:8081/healthz        # 200 = DB ok, 503 = DB down
```

- `os.Getenv` = config from environment. No passwords in code.
- `pgxpool` = reusable DB connections.
- `db.Ping` with 2s timeout → 503 if DB unreachable.
- Real incident: Mac rebooted → `pg` container stayed dead (no restart policy)
  → `/healthz` said 503. Health check did its job.
- Right after `docker start pg` still 503: DB was booting. Few seconds later 200.
  Starting ≠ ready.

### Drill: restore after closing terminals

Closed terminals → app dead, `export` gone. Docker restarted → `pg` dead (no restart policy).
Restored by myself, in order:

```bash
docker start pg                        # existing container (NOT docker run)
export DATABASE_URL=postgres://postgres:secret@localhost:5432/postgres
go run .                               # terminal 1 = server
curl -i localhost:8081/healthz         # terminal 2 = client → 200
```
