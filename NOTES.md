# Notes

The important points only. ⭐ = asked in interviews.
**Before each session: read only the ⭐ lines (5 min).**

---

## 1. The project

- Small Go app (status page) + PostgreSQL database.
- The app is only practice. The real skills: Docker, Linux, CI/CD, AWS, Terraform, Ansible, K8s.
- Path: laptop → Docker → AWS → automatic deploy → break it → recover → Kubernetes.

## 2. Names

| Name | What it is |
|---|---|
| database | the idea: software that stores data |
| PostgreSQL (Postgres) | one database product. `16` = its version |
| `postgres:16` | image of Postgres v16 |
| `status-page-app` | image of our Go app (built from the Dockerfile) |
| `pg` | old container from Step 4. Not used anymore |
| `db` | DB container in Compose. Used now |
| `db:5432` | address: container `db`, port 5432 (Postgres default port) |
| `pgdata` | volume where Postgres saves its files |

## 3. App

- `main.go` = where the app starts. `go.mod` = list of libraries. `go.sum` = exact versions + fingerprints.
- The app is a **server**: it waits on a port (8081) for requests.
- `/healthz` = the app checks itself: "can I reach my DB?" → 200 or 503.
- ⭐ Who uses `/healthz`: deploy pipeline (fail → rollback), Kubernetes (fail → restart),
  load balancer (fail → skip server), monitoring (fail → alert).
- ⭐ **The app reports, the platform acts.**
- ⭐ **Stateless app**: the app keeps no data. The DB keeps it.
  → you can restart or copy the app freely. Kubernetes depends on this.

## 4. HTTP

| Code | Meaning |
|---|---|
| 200 | OK |
| 204 | done, no content to return |
| 400 | bad request (e.g. `abc` is not a number) |
| 404 | not found |
| 405 | right path, wrong method |
| 500 | app error |
| 503 | app alive, dependency (DB) down |

- ⭐ **4xx = client's fault. 5xx = server's fault.** Alerts fire on 5xx.
- GET = read. POST = change.
- ⭐ **SQL injection**: never put user input directly in SQL. Use placeholders (`$1`).
- ⭐ Always validate input.

## 5. Config

- ⭐ Config goes in **environment variables**, not in code (**12-factor app**).
  Reasons: password stays out of GitHub · same code works on laptop and AWS.
- `export X=...` sets it in the current terminal only. Close terminal → gone.
- Same idea everywhere: `export` · Docker `-e` · Compose `environment:` · K8s ConfigMap/Secret · GitHub Secrets.
- ⭐ **Config vs secret**: port = config, password = secret. Store them separately.
- `.env` file = env vars saved in a file → always in `.gitignore`.

## 6. Docker

- ⭐ **Image** = template (program installed, no data). **Container** = a running copy of an image.
- **Dockerfile** = the instructions to build an image.
- ⭐ **Multi-stage build**: stage 1 builds the app with Go tools, stage 2 keeps only the binary. 900 MB → 27 MB.
- ⭐ **Layer caching**: copy `go.mod` before the code → rebuilds are fast.
- ⭐ Never run as root (`USER nonroot`).
- ⭐ Each container has its own network. `localhost` inside a container = that container, not your Mac.
- `docker run` = create a NEW container. `docker start` = start an EXISTING one (keeps its data).
- **Restart policy** (`restart: unless-stopped`) = container comes back after a Docker/Mac restart.

## 7. Docker Compose

- ⭐ One file (`compose.yaml`) describes all containers, networks, volumes. One command starts all.
- Compose creates its **own** containers from images. It does not reuse old ones.
- Compose network: each **service name is a DNS name** → app reaches DB at `db:5432`.
- `depends_on: service_healthy` → app starts only after DB is ready. (Starting ≠ ready.)
- `init.sql` runs automatically **only when the volume is empty** (first start).

## 8. Volumes ⭐

- Container deleted → its files deleted. **Volume** = storage outside the container, managed by Docker.
- **Container = Excel (the program). Volume = your .xlsx file (the data).**
- `pgdata:/var/lib/postgresql/data` → Postgres writes its files into the volume.
- **Named volume** (`pgdata`) = Docker manages it. **Bind mount** (`./db/init.sql`) = a real file from your Mac.
- `compose down` → volume kept. `compose down -v` → volume deleted, data gone.
- Rule: **anything you can't lose → volume.**
- Real world: AWS disk = **EBS** · Kubernetes = **PersistentVolume** · managed DB = **RDS**.

## 9. Tests

- Test = code that checks your code. CI runs it on every push → fail = no deploy.
- Go: files `*_test.go`, functions `TestXxx(t *testing.T)`, run `go test ./... -v`.
- ⭐ **Table-driven test**: list of input → expected output, checked in a loop.
- Small functions are easy to test → move logic out of handlers.
- `validStatus` = guard of the POST endpoint: checks the word the **client sends**, before the DB.
  It does NOT check the DB. Test calls it directly: no server, no DB.
- ⭐ Read a failure: `file:line: func(input) = got, want expected`.
- Break the code on purpose → test must go red. A test that never fails proves nothing.
- ⭐ Test good input, bad input, and edge cases (`""`).
- Good names say exactly what they check (`validStatus`, `isValidStatus`).

## 10. Debugging

- ⭐ Loop: Symptom → Evidence → Hypothesis → Test → Fix → Verify.

| Error | Meaning |
|---|---|
| `connection refused` | nothing is listening on that port |
| `go.mod file not found` / `not a git repository` | wrong folder → `pwd` |
| `missing go.sum entry` | `go mod tidy` |
| `undefined: json` | forgot an `import` |
| `/healthz` 503 | DB down → `docker ps -a` |

- `ping` = is the machine up? `curl` = is the app up?
- Logs (time + message) = first thing to read.

## 11. Commands

```bash
# git
git status · git add . · git commit -m "msg" · git push · git log --oneline

# go
go run .              # build + run
go mod tidy           # fix dependencies

# docker
docker ps -a          # all containers (also stopped)
docker logs NAME      # what the container printed

# compose
docker compose up -d --build
docker compose ps
docker compose logs app
docker compose down        # keep volumes
docker compose down -v     # delete volumes too

# test
curl -i localhost:8081/healthz
curl -X POST localhost:8081/components/3/status -d '{"status":"down"}'
echo $DATABASE_URL
```

nano: `Ctrl+O` save · `Ctrl+X` exit · `Ctrl+W` search

---

# Phase 1: AWS

## 12. AWS basics

- ⭐ **Delete/stop a server → its disk (EBS volume) can stay → you keep paying.** Same as Docker volumes.
- **Zombie resource** = forgotten leftover that still costs money (disk, snapshot, public IP).
- Find costs: Cost Explorer → Group by **Usage type**. No region prefix = us-east-1.
- `EBS:VolumeUsage.gp3` = disk storage. ~$0.08 per GB per month.
- Stopped server = no compute cost, but **disk still billed**.
- Blocking AWS at the bank ≠ cancel. Debt stays → account suspended.
- **Stop** = server paused, disk kept (still billed). **Terminate** = server deleted forever.
- ⭐ **Delete on termination** (Storage tab): Yes → disk deleted with server. No → disk stays → zombie.
- Always verify after cleanup: Instances + Volumes empty.
- ⭐ **EC2 instance** = virtual server (CPU, RAM, OS). **EBS volume** = its disk.
  Instance ≈ container (but a full computer) · EBS ≈ Docker volume.
- Attached disk can't be deleted → delete the **instance**, root disk follows. Delete the parent, children follow.
- **Free plan** (new account): $100 credits, 6 months. Credits first, no surprise bills. Out of credits/days → account stops.
- ⭐ **Root account = master key.** MFA on it, then don't use it daily.
- ⭐ **Root** = owner (signup email), can do everything. **IAM** = service to create users/roles with limited permissions.
  Root for emergencies only. Daily work → IAM user.
- ⭐ **Least privilege** = give only the permissions needed.
- ⭐ **AWS Organizations** = many accounts under one parent. **Centralized root access** = member accounts
  have no own root; parent controls it. My Free-plan account works like this.
- Every account starts with a **default security group** (firewall). Normal.
- ⭐ **Choose a region by:** law (EU data → EU), price, latency. We use **eu-north-1 (Stockholm)**.
- ⭐ **Resources are per region.** Console shows only the selected region → zombies hide in other regions.
- IAM user `terraform` = identity for programs (no console). **Access key** = its login for CLI.
  Secret shown once. Never in chat/code/git. Later replaced by **OIDC** (no stored keys).
- ⭐ **Budget alert** = email before money/credits burn silently.
- ⭐ **Budgets do NOT stop spending.** Only email. You must act (destroy).

## 13. AWS CLI

- `aws configure` → saves keys in `~/.aws/credentials` (plain text, never in git) + region in `~/.aws/config`.
- ⭐ `aws sts get-caller-identity` = "who am I?" First command when AWS access fails.
- Terraform install: `brew tap hashicorp/tap && brew install hashicorp/tap/terraform`.
- OS upgrade → "Command Line Tools too outdated" → reinstall them (`xcode-select --install`). Read the error first.
- ⭐ `bad CPU type in executable` = program built for another CPU. **x86_64/amd64** (Intel) vs **arm64** (Apple, AWS Graviton).
  Same issue exists for Docker images and EC2 types. `file $(which X)` shows the architecture.
- `which X` = where is this program? `hash -r` = forget old paths.

## 14. Terraform

- Why AWS: we need a real server on the internet. AWS rents servers (EC2) + network + firewall + disk.
- ⭐ **IaC (Infrastructure as Code)**: infra described in files, not clicks → repeatable, reviewable, no forgotten clicks.
- ⭐ **Terraform = Compose for the cloud.** `main.tf` ≈ `compose.yaml` · `apply` ≈ `up` · `destroy` ≈ `down`.
- Terraform remembers what it created → `destroy` removes all of it → no zombies.
- Terraform uses the `terraform` IAM user's key (from `aws configure`) to act in AWS.
- ⭐ **ARN** = unique ID of anything in AWS: `arn:aws:SERVICE::ACCOUNT:TYPE/NAME`.
- **Provider** = plugin that talks to a cloud's API (`hashicorp/aws`).
- `resource "TYPE" "NICKNAME" { ... }` → e.g. `aws_s3_bucket` `test`.
- ⭐ Workflow: `terraform init` (download provider) → `plan` (dry run, changes nothing) → `apply` (do it) → `destroy`.
- **S3** = cloud file storage. Bucket names are unique worldwide. Empty bucket = $0.
- ⭐ **EBS (block storage)** = disk for 1 server, you pick size, pay full size even empty.
  **S3 (object storage)** = no size, grows itself, pay only what you store. Backups, images, logs, Terraform state.
- Plan symbols: `+` create · `~` change · `-` destroy. `(known after apply)` = AWS decides it (e.g. ARN).
- ⭐ **Always read `Plan: X to add, Y to change, Z to destroy` before apply.** Surprise destroy = danger.
- ⭐ `plan` compares 3 things: `main.tf` (desired) · `terraform.tfstate` (what TF made) · real AWS (actual).
  Same → "No changes". Different → + / ~ / -.
- ⭐ **`terraform.tfstate`** = Terraform's memory (IDs of what it created). May contain secrets → never in git, never edit.
  Later: remote state in S3.
- ⭐ **Desired state vs actual state.** Run again with no changes → nothing happens = **idempotent**.
- `destroy` deletes only what **this folder's state** created. Not clicked things, not other folders.
- Delete one resource: remove its block + `apply` (normal) · `destroy -target=...` (emergency only).
- ⭐ **Blast radius** = how much breaks if something goes wrong. Separate states for dev/prod.
- `apply` morning, `destroy` evening → $0 overnight, same env next day.
- Terraform adds only what's **missing** (2nd bucket → "1 to add", not 2).
- Prod: almost never destroy. Daily work = `plan` + `apply`.
- ⭐ `.gitignore`: commit what you wrote, ignore generated + secret (`.terraform/`, `*.tfstate*`).
