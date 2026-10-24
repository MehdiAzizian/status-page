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

### Step 6a/6b: read endpoints

**What:** `GET /components` (all) and `GET /components/{id}` (one), as JSON.
**Why:** the real feature: show the status board from the DB.

- `type Component struct` = one DB row as a Go box. `json:"id"` = name in JSON.
- `db.Query` (many rows) · `db.QueryRow` (one row) · `Scan` = row → struct.
- `{id}` in route + `r.PathValue("id")` · `strconv.Atoi` = "2" → 2.
- `$1` placeholder = safe from SQL injection.
- Tested: `/components/2` → 200 · `/99` → 404 · `/abc` → 400.
- Mistake: `undefined: json` → forgot `"encoding/json"` in imports.

### Step 6c: write endpoint

**What:** `POST /components/{id}/status` with body `{"status":"down"}`.
**Why:** support staff change status. GET reads, POST changes.

```bash
curl -i -X POST localhost:8081/components/3/status -d '{"status":"down"}'   # 204
```

- `json.Decode(r.Body)` = read client JSON · validation: only operational/degraded/down → else 400.
- `db.Exec` UPDATE · `RowsAffected()==0` → 404 · success → 204 No Content.
- Restarted app → still "down" → **stateless app**, data lives in DB.
- Delete `pg` container → data gone (no volume yet).

### Step 7: Dockerfile

**What:** multi-stage Dockerfile → image `status-page:dev` (26.7 MB).
**Why:** app + everything it needs in one box. Runs the same everywhere.

```bash
docker build -t status-page:dev .
docker images status-page
docker run --rm -p 8081:8081 -e DATABASE_URL=...@localhost:5432/...            # 503
docker run --rm -p 8081:8081 -e DATABASE_URL=...@host.docker.internal:5432/... # 200
```

- Stage 1 `golang:1.25` builds · Stage 2 `distroless` runs only `/app` as `nonroot`.
- `COPY go.mod go.sum` before code = layer caching.
- Inside a container, `localhost` = the container itself → DB not found → 503.
  `host.docker.internal` = the Mac → reaches `pg` via its `-p 5432:5432` → 200.

### Step 8: Docker Compose

**What:** `compose.yaml` with 2 services (`db`, `app`) + named volume `pgdata` + `db/init.sql`.
**Why:** one command rebuilds everything from files in git. No manual steps.

```bash
docker compose up -d --build     # build app image, create network + volume + 2 containers
docker compose down              # delete containers + network, KEEP volume
docker compose down -v           # also delete volume → data gone
```

- Compose created NEW containers from images, did not touch old `pg`.
- `init.sql` mounted into `/docker-entrypoint-initdb.d/` → runs on first start (empty data dir) only.
- App reaches DB at `db:5432` (service name = DNS name on Compose network).
- `healthcheck` (`pg_isready`) + `depends_on: service_healthy` → app starts after DB is ready.
- `restart: unless-stopped` → survives Docker/Mac restart.
- Test: POST down → `down` → `up` → still down (volume kept). `down -v` → `up` → operational.

### Step 9: tests

**What:** `validStatus()` function + table-driven test in `main_test.go`.
**Why:** CI will run tests on every push. Fail = no deploy.

```bash
go test ./... -v
```

- `validStatus` checks the client's input in POST (not the DB).
- Broke it on purpose (removed "down") → `validStatus("down") = false, want true` → FAIL. Put back → PASS.

**Phase 0 done.**

---

## Phase 1: AWS + Terraform

### Step 1: cost investigation (real zombie)

**Problem:** AWS billed ~$0.78/month since May 2026 (free year ended). $3.90 past due.
**Found:** Cost Explorer → Group by Usage type → `EBS:VolumeUsage.gp3` ($0.64) + tax ($0.14).
**Cause:** old `testInstance` (t2.micro, us-east-1) **stopped** since 2025 → its 8 GiB EBS disk still billed.
**Fix:** terminated the instance (root disk had Delete on termination = Yes) → Volumes = 0.
**Lesson:** stop ≠ delete. Disks outlive servers. Always check Volumes after cleanup.

### Step 2: account safety (new account)

- New Free-plan account: $100 credits, 6 months (+$100 via activities).
- Account is a member of an AWS Organization → **no own root** (centralized root access). MFA on my login (TOTP).
- Budget `monthly-5` (+ `monthly-10`). Budgets only email, they don't stop spending.
- Region: **eu-north-1 (Stockholm)**: EU, cheap, close enough.

### Step 3: connect Mac to AWS

```bash
brew tap hashicorp/tap && brew install hashicorp/tap/terraform
brew install awscli                    # arm64 version
aws configure                          # key of IAM user "terraform", eu-north-1, json
aws sts get-caller-identity            # → arn:aws:iam::905223168457:user/terraform
```

- IAM user `terraform`: no console, AdministratorAccess, CLI access key.
- Problems fixed: macOS upgrade → Command Line Tools outdated (`xcode-select --install`);
  old aws-cli was x86_64 → `bad CPU type` → reinstalled arm64 via brew.
- Old key from old account was in `~/.aws/credentials` → overwritten.

### Step 4: first Terraform resource

`infra/main.tf`: provider aws (eu-north-1) + `aws_s3_bucket.test` (`mahdi-tf-test-2026`).

```bash
terraform init      # download AWS provider
terraform plan      # Plan: 1 to add
terraform apply     # created → seen in S3 console
terraform plan      # No changes (desired = actual)
terraform destroy   # 1 destroyed → gone
terraform apply     # back again
```

- `terraform.tfstate` appeared = Terraform's memory. Not in git.
- `destroy` removes only what this state created. Delete one thing → remove block + apply.

### Step 5: network (VPC)

**What:** removed test bucket; added VPC `10.0.0.0/16`, public subnet `10.0.1.0/24`, Internet Gateway,
route table (`0.0.0.0/0 → igw`), association. Plan: 5 to add, 1 to destroy. Cost $0.
**Why:** the server needs a network and a path to the internet.

- Path: internet → IGW → route table → public subnet → (server next).
- Subnet shows 251 IPs: AWS reserves 5.
- Default VPC also exists in the account (AWS made it). Ours is `status-page`.
