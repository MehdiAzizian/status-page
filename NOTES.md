# Notes

The important points only. ⭐ = asked in interviews.
**Before each session: read only the ⭐ lines (5 min).**

---

## 0. The story so far (Phase 0)

1. **The app**: small Go web server (port 8081) + PostgreSQL. Shows if Website / Login / API work.
2. **Config**: DB address comes from env vars (`DATABASE_URL`), not code → same code everywhere.
3. **Tests**: `go test` checks `validStatus` → later CI runs it on every push.
4. **Docker**: Dockerfile → image (27 MB) → container. Runs the same on Mac and AWS.
5. **Compose**: one file starts app + db together. App finds DB at `db:5432`.
6. **Volume**: DB data lives in `pgdata`, survives container delete.
7. **Restart policy**: containers come back after reboot.

⭐ **Flow in one line:** code → Dockerfile → image → Compose (app + db + volume) → `docker compose up -d`.

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

## The story (Phase 1)

1. **Account + safety**: new AWS account, MFA, budget alert, IAM user `terraform` (not root).
2. **Zombie lesson**: a stopped server's disk kept billing me → always terminate + check volumes.
3. **Terraform**: infra as code → `init` · `plan` · `apply` · `destroy`. State = its memory.
4. **Network**: VPC → public subnet → Internet Gateway → route table. Like my home network.
5. **Firewall**: security group → 22 only from my IP, 80/443 for everyone.
6. **Server**: EC2 (Ubuntu ARM) + my SSH key → I log in with `ssh -i`.
7. **Destroy + apply**: everything rebuilt from code, new IP each time.
8. **Clean code + remote state**: split `.tf` files, state in S3 (versioning + locking).

⭐ **Flow in one line:** `terraform apply` → VPC + subnet + IGW + SG + EC2 → `ssh` in → `terraform destroy` at night.

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

# Phase 2: Ansible — the story

### 1. The goal
Terraform gave me an empty Ubuntu server. Now I must fill it (Docker, app, security) — without typing commands by hand.
- ⭐ **Terraform creates** servers/networks. **Ansible configures** what's inside (users, packages, Docker, app).
- Why not SSH + type commands? Not repeatable, easy to forget, every server ends up different.
- Repo layout: app (root) · `infra/` Terraform creates the server · `ansible/` Ansible configures it.
- ⭐ GitHub = **code** · S3 = **state** (list of created resources) · AWS = **real resources**.

### 2. How Ansible works
- **Inventory** = list of servers (IPs, groups like `[web]`).
- **Module** = one ready-made action (a Python script): `apt` install · `user` users · `copy` files.
- **Task** = one use of a module. **Playbook** = YAML file, list of tasks, run in order.
- ⭐ **Agentless**: nothing installed on the server. Ansible connects over **SSH**.
- ⭐ How a task runs: SSH in → copy small Python script → run → delete → result back. Needs only SSH + Python on the server. (That's why `history` on the server is empty.)
- ⭐ **Idempotent**: run twice → 2nd run changes nothing (checks first). A bash script would redo it.
- Linux command (`sudo`, `ls`, `apt`) = program on the server. Ansible module = script inside Ansible on my Mac. `sudo` has no module → `become: true`.
- ⭐ Don't memorize modules. Google "ansible <thing> module" → **Examples** at the bottom of the docs page. Know the CONCEPT.

### 3. First contact
- Inventory line = my ssh command in a file: `IP ansible_user=ubuntu ansible_ssh_private_key_file=~/.ssh/status-page`. Private key stays on my Mac; Ansible uses it like `ssh -i`.
- `ansible.cfg`: `inventory = inventory.ini` · `StrictHostKeyChecking=accept-new` = accept new server, warn if it changes.
- ⭐ Ansible reads `ansible.cfg` from the **current folder** → always run it from `ansible/`.
- `ansible web -m ping` = check SSH + Python. **Not** network ping. `SUCCESS` + `pong` = works.
- `ansible web -a "cmd"` = run one command on all servers (`-b` = with sudo). Always says `CHANGED` (can't check state). `rc=0` = success.
- `..` = parent folder · `&&` = run next only if previous succeeded · `mkdir -p a/b` = make folders.
- Mistake: named the file `invertory.ini` → "Unable to parse inventory" / "Could not match host pattern: web". "Can't find X" → check the file name first.

### 4. The IP problem (twice!)
**a) Server IP changes** → new server after destroy/apply = new IP → inventory stale → timeout.
- ⭐ Terraform knows the IP → Terraform writes `inventory.ini` (`local_file`). Manual copy = goes stale.
- `local_file` = file on my **Mac**, not on AWS. New provider `hashicorp/local` → `terraform init -upgrade`.
- Generated files (inventory.ini) → `.gitignore`. In `.gitignore` a trailing `/` = folder only.

**b) My Mac IP changes** (new Wi-Fi) → SG allows SSH only from the old IP → timeout, but server is running.
- ⭐ Auto IP: `data "http" "my_ip" { url = "https://checkip.amazonaws.com" }` → `["${chomp(data.http.my_ip.response_body)}/32"]`.
- `chomp()` = remove trailing newline · `"${x}/32"` = put a value inside text.
- Fix = `terraform apply` → changes only the SG rule. Server + app keep running. No Ansible rerun.
- `Operation timed out` on 22 → old IP in inventory, or my IP changed. Compare with `terraform output` / `curl -4 ifconfig.me`.
- Real fix later: **SSM** (no port 22 at all).

### 5. First playbook
- ⭐ YAML = only 3 rules: `key: value` (pair) · `- item` (list) · indent 2 spaces (inside). Space after `:`, no tabs.
- Playbook: `hosts` = which group · `become: true` = sudo · `tasks` = jobs. Run: `ansible-playbook playbook.yml`.
- Task shape: `- name:` + module + module's arguments (indented inside).
- `Gathering Facts` = Ansible reads server info first (OS, IP…). Always `ok`.
- RECAP: `ok` = already right · `changed` = Ansible did something · `unreachable` = SSH failed.
- `apt update_cache` always says `changed` → `cache_valid_time: 3600` (skip if updated < 1h ago).
- Mistake: `paybook.yml` typo → `mv`.

### 6. Install Docker
- `apt` module: `name:` list of packages (`docker.io`, `docker-compose-v2`) + `state: present` = install if missing. 1st run `changed`, 2nd `ok`.
- `apt` because the server is Ubuntu (like `brew` on Mac). Red Hat/Amazon Linux → `dnf`. Generic → `package`.
- ⭐ Use real modules, not `shell:` → modules check state first (idempotent), `shell` always says `changed`.
- `docker` group = its users run `docker` without `sudo`. Module `user`, `append: true` keeps old groups.

### 7. Deploy the app
- Same files as on my Mac — nothing rewritten. Code is on GitHub (public) → server clones it.
- ⭐ Deploy = Ansible: `git` clone from GitHub → `community.docker.docker_compose_v2` up (idempotent `compose up -d`).
- `community.docker.` = module from an extra collection (comes with brew's ansible).
- `buildx isn't installed` warning → harmless, Compose uses old builder. Fix: apt package `docker-buildx`.
- `docker ps` PORTS: `0.0.0.0:8081->8081` = published · `5432/tcp` without `->` = DB not published, only reachable inside the Docker network ⭐.
- `localhost` = the machine I run the command on (server via SSH/Ansible, my Mac in my browser).
- ⭐ IP = which machine. Port = which program on it (22 SSH · 80 HTTP · 443 HTTPS · 8081 my app · 5432 Postgres).
- Update app = `git push` → `ansible-playbook` (git pulls new commit, compose recreates). Phase 3 automates this.

### 8. Open it to the internet: Caddy
SG allows only 22/80/443, app listens on 8081 → put a proxy on 80.
- ⭐ **Reverse proxy** = single public entry on 80/443 → forwards to `app:8081`. App + DB stay hidden. Does HTTPS later. Routes many apps on one port.
- `Caddyfile`: `:80 { reverse_proxy app:8081 }`. Caddy = one more service in `compose.yaml` → starts with `compose up`.
- Why Caddy: 3 lines of config + free automatic HTTPS. Alternatives: Nginx (most common), Traefik, AWS ALB.
- Opening 80 to everyone = normal for a website. Only Caddy listens there.
- Mistake: `Cadyfile` typo → caught before push.

### 9. Survive a reboot
- ⭐ Chain: server boots → systemd starts Docker → Docker restarts containers (`restart: unless-stopped`).
- Test: `ansible web -b -a "reboot"` → `uptime` shows `up 1 min` → site still works.

### 10. Lock it down
- ⭐ SSH hardening = `PasswordAuthentication no` + `PermitRootLogin no` → only my key works, nothing to guess.
- ⭐ 2 firewalls: AWS SG (outside) + `ufw` (inside the server) = **defense in depth**. One mistake ≠ open server.
- ufw: allow 22 **before** enabling, or I lock myself out.
- Handler = runs only if a task with `notify` changed (restart ssh only when config changed).
- `loop` + `{{ item }}` = same task for each value.

### ⭐ The whole flow in one line
Terraform creates server + writes inventory → Ansible: Docker → clone repo → compose up (app + db + Caddy) → harden SSH + ufw → site live on :80, survives reboot.

# Phase 3: CI/CD — in one page

**The goal:** I only do `git push`. Testing, building and deploying happen alone.
**CI** = every push is tested and built automatically. **CD** = if CI is green, it's deployed automatically.

**1. CI tests my code** 🧪
On every push, GitHub lends me a fresh Ubuntu machine and runs `go test` from `.github/workflows/ci.yml`. Red = broken code stops there, before the server. Free for public repos. Alternatives: GitLab CI (common in Italy), Jenkins.

**2. CI builds the image** 📦
If tests pass, GitHub builds the Docker image on an **ARM** machine (my server is ARM) and uploads it to **GHCR**, GitHub's image storage. The tag is the commit ID, so every old version stays there. Build once, deploy the same image. ⚠️ CI only checks what I test: a broken `compose.yaml` stayed green.

**3. The problem: how does GitHub reach my server?** 🚧
GitHub's IPs change every run, and SSH is allowed only from my Mac. Key fact: firewalls block traffic coming **in**, but going **out** is allowed. So the server downloading from GHCR is fine — only the "deploy now" message can't get in.

**4. The answer: SSM** 📬
An **agent** already on my server keeps asking AWS "any command for me?" (going out). GitHub leaves the order at AWS, the agent runs it, and the server pulls the image itself. No port 22. It stayed offline until I restarted it after giving it a role.

**5. Permissions (IAM)** 🪪
AWS trusts nobody without an identity — not even programs on my own server. A **user** has a stored key (for people), a **role** gets temporary keys (for machines). My Mac = user `terraform`, the server = a role allowed to use SSM, GitHub = user `status-page-github`.

**6. OIDC was blocked** ⛔
OIDC = GitHub logs into AWS with a signed ID instead of a key, like "Login with Google". An organization rule (**SCP**) above my account blocked it — even admin can't override that. Plan B: a key in **GitHub Secrets** that can do only one thing on one server (least privilege). It never appeared on my screen.

**7. Deploy** 🚀
`compose.yaml` now names the GHCR image with `${TAG}`. After a green build, GitHub tells SSM: `git pull` → download the new image → restart the app. I gave the permission first (`apply`), then pushed. Proof: `docker ps` on the server shows the same commit ID as the GitHub run; db and caddy weren't touched.

**8. Rollback** ⏪
A **Run workflow** button: paste an old full commit ID, and the old image comes back in ~30 seconds. Same as deploy, without `git pull`.

**My mistakes:** `ci.ylm` typo · 2 spaces instead of 4 in `compose.yaml` · `outputs.tf` still pointing to a deleted role (edit not saved) · ran Ansible from `infra/`.

### ⭐ The whole flow
**push → test → build → GHCR → GitHub logs into AWS → SSM → server pulls image → app updated**
**Broken? → Rollback button → old tag → fixed**

### 🎤 Say it in an interview
> "On every push, GitHub Actions tests the code, builds a Docker image tagged with the commit SHA and pushes it to GHCR. Then it deploys through AWS SSM, so no SSH port is open. Rollback is one click: redeploy the previous tag."

> "I planned OIDC, but an organization SCP blocked it, so I used a least-privilege key in GitHub Secrets. In a company I'd ask the platform team to allow OIDC."

📎 Exact commands and names: [NOTES-details.md](NOTES-details.md)
