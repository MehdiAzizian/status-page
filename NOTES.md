# Notes

The important points only. ⭐ = asked in interviews.
**Before each session: read only the ⭐ lines (5 min).**

---

## 1. The project

- **Status page** = page showing if a company's services work (Website ✅ API ⚠️ Payments ❌). E.g. githubstatus.com.

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

# Phase 1: AWS + Terraform

## 12. AWS account & safety

- **Free plan** (new account): $100 credits, 6 months. Credits first, no surprise bills. Out of credits/days → account stops.
- ⭐ **Root** = owner (signup email), can do everything → emergencies only. **IAM** = service to create users/roles with limited permissions. Daily work → IAM user.
- ⭐ **Root account = master key** → MFA on it.
- ⭐ **AWS Organizations** = many accounts under one parent. **Centralized root access** = member accounts have no own root; parent controls it. My Free-plan account works like this.
- ⭐ **Least privilege** = give only the permissions needed.
- IAM user `terraform` = identity for programs (no console). **Access key** = its login for CLI. Secret shown once. Never in chat/code/git. Later replaced by **OIDC** (no stored keys).
- ⭐ **Budget alert** = email before money/credits burn silently. **Budgets do NOT stop spending** → you must act (destroy).
- ⭐ **Choose a region by:** law (EU data → EU), price, latency. We use **eu-north-1 (Stockholm)**.
- ⭐ **Resources are per region.** Console shows only the selected region → zombies hide in other regions.
- Every account starts with a **default VPC** and **default security group**. Normal. Ours are separate, made from code.

## 13. Cost lesson: the zombie disk

- ⭐ **Stop** = server paused, **disk still billed**. **Terminate** = server deleted forever.
- ⭐ Delete/stop a server → its disk (EBS) can stay → you keep paying. Same as Docker volumes.
- **Zombie resource** = forgotten leftover that still costs money (disk, snapshot, public IP).
- ⭐ **Delete on termination** (Storage tab): Yes → disk deleted with server. No → disk stays → zombie.
- Attached disk can't be deleted → delete the **instance**, root disk follows. Delete the parent, children follow.
- Find costs: Cost Explorer → Group by **Usage type**. No region prefix = us-east-1.
- `EBS:VolumeUsage.gp3` = disk storage, ~$0.08 per GB per month.
- Always verify after cleanup: Instances + Volumes empty.
- Blocking AWS at the bank ≠ cancel. Debt stays → account suspended.

## 14. EC2, EBS, S3

- ⭐ **EC2 instance** = virtual server (CPU, RAM, OS). **EBS volume** = its disk. Instance ≈ container (but a full computer) · EBS ≈ Docker volume.
- ⭐ **EBS (block storage)** = disk for 1 server, you pick size, pay full size even empty.
- ⭐ **S3 (object storage)** = no size, grows itself, pay only what you store. Backups, images, logs, Terraform state. Bucket names unique worldwide. Empty bucket = $0.

## 15. AWS CLI + Mac setup

- `aws configure` → keys in `~/.aws/credentials` (plain text, never in git) + region in `~/.aws/config`.
- ⭐ `aws sts get-caller-identity` = "who am I?" First command when AWS access fails.
- ⭐ **ARN** = unique ID of anything in AWS: `arn:aws:SERVICE::ACCOUNT:TYPE/NAME`.
- Terraform install: `brew tap hashicorp/tap && brew install hashicorp/tap/terraform`.
- `which X` = where is this program? `hash -r` = forget old paths. `file $(which X)` = its CPU type.

## 16. Terraform

- Why AWS: we need a real server on the internet. AWS rents servers (EC2) + network + firewall + disk.
- ⭐ **IaC (Infrastructure as Code)**: infra in files, not clicks → repeatable, reviewable, no forgotten clicks.
- ⭐ **Terraform = Compose for the cloud.** `main.tf` ≈ `compose.yaml` · `apply` ≈ `up` · `destroy` ≈ `down`.
- Terraform uses the `terraform` IAM user's key (from `aws configure`).
- **Provider** = plugin that talks to a cloud's API (`hashicorp/aws`).
- `resource "TYPE" "NICKNAME" { ... }` → e.g. `aws_s3_bucket` `test`.
- ⭐ Workflow: `init` (download provider) → `plan` (dry run) → `apply` (do it) → `destroy`.
- Plan symbols: `+` create · `~` change · `-` destroy. `(known after apply)` = AWS decides it (e.g. ARN).
- ⭐ **Always read `Plan: X to add, Y to change, Z to destroy` before apply.** Surprise destroy = danger.
- ⭐ `plan` compares 3 things: `main.tf` (desired) · `terraform.tfstate` (what TF made) · real AWS (actual).
- ⭐ **`terraform.tfstate`** = Terraform's memory. May contain secrets → never in git, never edit. Later: remote state in S3.
- ⭐ **Desired vs actual state.** No changes → nothing happens = **idempotent**.
- Terraform adds only what's **missing** (2nd bucket → "1 to add", not 2).
- `destroy` deletes only what **this folder's state** created. Not clicked things, not other folders.
- Delete one resource: remove its block + `apply` (normal) · `destroy -target=...` (emergency only).
- ⭐ **Blast radius** = how much breaks if something goes wrong. Separate states for dev/prod.
- `apply` morning, `destroy` evening → $0 overnight. Prod: almost never destroy. Daily work = `plan` + `apply`.
- ⭐ `aws_vpc.main.id` = reference → Terraform figures the order itself (dependency graph).
- Terraform reads **all `.tf` files** in the folder + **auto-loads `terraform.tfvars`**. Missing value → it asks.
- `variable` + `terraform.tfvars` = input values, kept out of git. `var.my_ip`.
- One value = `string`. Many = `list(string)` → `["a", "b"]`. `#` = comment.
- ⭐ `.gitignore`: commit what you wrote, ignore generated + secret (`.terraform/`, `*.tfstate*`, `*.tfvars`).
- `git check-ignore -v FILE` = which `.gitignore` line ignores this file.

## 17. Network

| Home | AWS |
|---|---|
| home network (192.168.x.x) | **VPC** (`10.0.0.0/16`) |
| part of the network | **Subnet** (`10.0.1.0/24`) |
| router (my phone hotspot) | **Internet Gateway (IGW)** |
| routing rules | **Route table** (`0.0.0.0/0` → IGW) + **association** to the subnet |
| laptop | **EC2** |

- ⭐ **Public subnet** = its route table sends `0.0.0.0/0` to an IGW. That's all.
- Route table: `10.0.0.0/16 → local` (inside VPC, automatic) · `0.0.0.0/0 → igw` (internet).
- ⭐ CIDR: `/16` ≈ 65k · `/24` = 256 · `/32` = 1 IP · `0.0.0.0/0` = everywhere. Bigger number = smaller range.
- ⭐ AWS reserves **5 IPs** per subnet → `/24` = 251 usable.
- VPC, subnet, IGW, route table = $0.
- ⭐ "How does a server reach the internet?" → public subnet + route `0.0.0.0/0` → IGW + public IP.

## 18. Security group (firewall)

- ⭐ **Security group** = firewall attached to a server. `ingress` = in, `egress` = out.
- Ours: 22 SSH (only my IP `/32`) · 80 HTTP · 443 HTTPS (everyone) · out all.
- ⭐ **Default deny**: port 5432 has no rule → blocked, without writing "block".
- ⭐ **Stateful**: visitor comes in on 80 → the page going back is allowed automatically.
- ⭐ **SG vs NACL**: SG = on server, stateful, allow only. NACL = on subnet, stateless, allow + deny.
- SG protects nothing until attached to the EC2.
- Hotspot IP changes → SSH fails → update tfvars + apply.
- ⭐ Real world: no port 22 → **SSM Session Manager** (connect via AWS, IP doesn't matter).

## 18b. EC2 server

- Cost (Stockholm, approx): `t4g.small` ~$0.02/h · public IP ~$0.005/h · 2h session ≈ $0.06 · 24/7 month ≈ $20 → destroy after session.
- `t4g` = ARM (like my Mac), cheaper. `small` = 2 GB RAM (Docker + Postgres + app).
- ⭐ **SSH key**: private key stays on Mac (`~/.ssh/status-page`), public key goes to server. No password.
- `ssh-keygen -t ed25519 -f ~/.ssh/status-page -N ""` · login: `ssh -i ~/.ssh/status-page ubuntu@IP`.
- ⭐ **AMI** = server image (OS). Like a Docker image, but for a whole server.
- `data` block = **look up** something existing (doesn't create). `output` = print a value after apply.
- EC2 needs: AMI · instance type · subnet · security group · key pair.

- ⭐ **SSH login**: public key on server (`~/.ssh/authorized_keys`) → server sends challenge → Mac signs with private key → server checks. Private key never leaves Mac.
- SSH ≠ TLS. Same idea (key pair). SSH = server login (port 22). TLS = HTTPS.
- Key types: `rsa` (old) · `ed25519` (modern standard). `-N ""` = no passphrase (ok for learning).
- `~/.ssh`: `id_*` private · `*.pub` public · `known_hosts` = servers I connected to.
- ⭐ **Terraform grammar**: `KIND "TYPE" "NAME" { arg = value }`. KIND: `resource` create · `data` look up · `variable` input · `output` print.
- Use another block: `TYPE.NAME.attribute` (e.g. `aws_subnet.public.id`).
- ⭐ **EC2 = 5 questions**: which OS (AMI) · how big (type) · where (subnet) · which firewall (SG) · how to log in (key).
- ⭐ Argument names → docs: registry.terraform.io → resource → Argument Reference. Know what you need, look up the name.

- ⭐ **Private key MAKES the signature (only on my Mac). Public key only CHECKS it (on the server).** Public key alone lets nobody in.
- `aws_key_pair` uploads `.pub` → AWS puts it in `/home/ubuntu/.ssh/authorized_keys` → `ssh -i PRIVATE_KEY` logs in.
- ⭐ Private key permissions must be `-rw-------` (600), or SSH refuses it.
- ⭐ **Argument** = setting I write (left of `=`). **Attribute** = value AWS gives back (e.g. `id`).
  `aws_subnet.public.id` = the `id` of the `aws_subnet` named `public` (= `subnet-00f1e...`).

- ⭐ First SSH → server fingerprint saved in `~/.ssh/known_hosts`. Different later → warning (maybe fake server).
- On the server: `aarch64` = ARM · private IP `10.0.1.x` from our subnet · `htop` = CPU/RAM live.

- Destroy = money (server + public IP billed hourly) + proof everything is in code. Prod is never destroyed; ours is a dev/learning env.
- `destroy` = all or nothing for the folder. Free things (VPC) rebuild in ~1 min anyway.

- `output` = print a value after apply. `terraform output` = show again. `-raw` = value only.
- ⭐ `ssh -i ~/.ssh/status-page ubuntu@$(terraform output -raw public_ip)` · `$( )` = run inside first, use result.
- ⭐ New server = **new random public IP**. Fixed IP = **Elastic IP** (needed for a domain name).
- Private IP also changes, but stays inside the subnet range (`10.0.1.x`).

- ⭐ Split by topic: `main.tf` (provider) · `network.tf` · `security.tf` · `server.tf` · `variables.tf` · `outputs.tf`. Terraform reads all `.tf` as one.
- `terraform fmt` = format code · `terraform validate` = syntax check (no AWS).
- ⭐ Refactor rule: code moved, not changed → plan must show the same result as before.

- ⭐ Linux text tools: `grep "x" file` (find lines, `-n` = line numbers) · `sed -n 'A,Bp' file` (print lines A–B) ·
  `head -n N` / `tail -n N` (first/last lines, `tail -f` = live log) · `cat` (whole file) · `>` save (overwrite) · `>>` append.
- ⭐ grep `^` = start of line · `-E 'a|b'` = a OR b. `grep -nE '^(resource|data)' main.tf` = only block starts.
- grep shows matching **lines only**. A block ends one line before the next block starts.
- Unknown file → read first: `less file` (space/arrows, `/word` search, `q` quit). Known structure → grep.
- `mv a b` = move or rename (here: replace main.tf with main.tmp).

- ⭐ **Remote state** in S3: state not only on my Mac → Mac dies / teammate runs TF → no zombies, no duplicates.
- State bucket is created **outside** this Terraform (else `destroy` deletes its own memory).
- ⭐ **Versioning** = S3 keeps old versions of the state → restore if broken.
- ⭐ **Locking** (`use_lockfile = true`) = only one apply at a time → state can't be corrupted.
- `backend "s3" { bucket, key, region, use_lockfile, encrypt }` inside `terraform { }` → `terraform init -migrate-state`.

## 19. My mistakes (Phase 1)

- Blocked AWS at my bank → bill stayed → past due. Real cause: stopped instance's disk.
- macOS upgrade → "Command Line Tools too outdated" → `xcode-select --install`. Read the error first.
- ⭐ `bad CPU type in executable` → old aws-cli was **x86_64** (Intel), Mac is **arm64** → reinstalled via brew. Same issue for Docker images and EC2 types.
- Old key from old account was still in `~/.aws/credentials` (`aws configure` showed `****ICMD`).
- aws-cli asked to install AI-agent skills/MCP → answered no.
- `ssh ubuntu@PUBLIC_IP` → placeholder, use the real IP from `terraform output`.
- `Attribute redefined` → set `my_ip` twice in tfvars → keep one line, comment the other with `#`.

---

# Interview answers (say them out loud)

**"How do you create a firewall on AWS?"**
> In AWS the firewall for a server is a **security group**. I create it with Terraform inside my VPC.
> I add inbound rules: port, protocol, and who is allowed. In my project: SSH port 22 only from my IP,
> and HTTP 80 and HTTPS 443 from everyone, because the website is public. Outbound I allow everything,
> so the server can download updates. Then I attach it to the EC2 instance.
> Everything not in a rule is blocked by default.

**"How does a server in AWS reach the internet?"**
> The server is in a **public subnet**. That subnet's route table sends all internet traffic (`0.0.0.0/0`)
> to an **Internet Gateway**, and the server has a public IP. The security group decides which ports are open.

**"Why Terraform instead of clicking in the console?"**
> Clicks can't be repeated or reviewed, and people forget to delete things. With Terraform the infrastructure
> is code in git: I can recreate it with `apply`, remove all of it with `destroy`, and `plan` shows every
> change before it happens. I destroy and rebuild my whole setup every session.

**"Image vs container?"**
> An image is a read-only template with the app and everything it needs. A container is a running copy
> of that image. I can run many containers from one image.

**"Where do you keep Terraform state, and why?"**
> In an S3 bucket, not on my laptop. If the state were only local and my laptop died, Terraform would forget
> what it created. With remote state the whole team uses the same state. I enabled **versioning** so I can
> restore an old state, and **locking** so two applies can't run at the same time and corrupt it.
> The state bucket is created outside that Terraform code, so `destroy` can't delete it.

---

# Phase 2: Ansible

## 20. What Ansible is

- ⭐ **Terraform creates** servers/networks. **Ansible configures** what's inside servers (users, packages, Docker, app).
- Why not SSH + type commands? Not repeatable, easy to forget, every server ends up different.
- **Inventory** = list of servers (IPs, groups like `[web]`).
- **Module** = one built-in action: `apt` install packages · `user` create user · `copy` copy file.
- **Task** = one use of a module. **Playbook** = YAML file, list of tasks, run in order.
- ⭐ **Agentless**: nothing installed on the server. Ansible connects over **SSH** and runs the tasks.
- ⭐ **Idempotent**: run twice → 2nd run changes nothing (checks first). A bash script would redo it.
- `ansible web -m ping` = check SSH + Python work. **Not** network ping.
- `ansible.cfg`: `inventory = inventory.ini` · `StrictHostKeyChecking=accept-new` = accept new server, warn if it changes.
- Before SSH/Ansible: `curl -4 ifconfig.me` → IP changed? update `terraform.tfvars` → apply.
- ⭐ Auto IP: `data "http" "my_ip" { url = "https://checkip.amazonaws.com" }` → `["${chomp(data.http.my_ip.response_body)}/32"]`. No more manual IP.
- `chomp()` = remove trailing newline · `"${x}/32"` = put a value inside text · new provider → `terraform init -upgrade`.
- `..` = parent folder · `&&` = run next only if previous succeeded · `mkdir` = make folder.
- Repo layout: app (root) · `infra/` Terraform creates server · `ansible/` Ansible configures it.
- ⭐ GitHub = **code** · S3 = **state** (list of created resources) · AWS = **real resources**.
- Inventory line = my ssh command in a file: `IP ansible_user=ubuntu ansible_ssh_private_key_file=~/.ssh/status-page`.
  Private key stays on my Mac; Ansible runs on my Mac and uses it like `ssh -i`.
- ⭐ How a task runs: SSH in → copy small Python program → run → delete → result back. Needs only SSH + Python on server.
- Mistake: named the file `invertory.ini` → Ansible "Unable to parse inventory" / "Could not match host pattern: web". "Can't find X" → check the file name first.
- `Operation timed out` on port 22 → wrong/old IP (server gone) or firewall. Compare inventory IP with `terraform output`.
- ⭐ Ansible reads `ansible.cfg` from the **current folder** → always run it from `ansible/`.
- ⭐ Terraform knows the IP → Terraform writes `inventory.ini` (`local_file`). Manual copy = goes stale after every new server.
- `local_file` = file on my **Mac**, not on AWS. Needs `terraform init -upgrade` (new provider `hashicorp/local`).
- `ping` → `SUCCESS` + `pong` = SSH works + Python found on server. Ansible itself is NOT installed there (agentless).
- Generated files (inventory.ini) → `.gitignore`.
