# Notes

The important points only. ⭐ = asked in interviews.

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

## 9. Debugging

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

## 10. Commands

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
