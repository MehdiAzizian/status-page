# Notes — details (exact commands & names)

## Phase 3: CI/CD
- ⭐ **CI** = every push → code is **tested + built** automatically.
- ⭐ **CD** = if CI is green → code is **deployed** automatically.
- Before: `git push` → `ansible-playbook` by hand → server built the image itself.

### 1. CI tests my code 🧪
On every push, GitHub lends me a fresh Ubuntu machine and runs `.github/workflows/ci.yml` → `go test`.
- `on` = when · `jobs` = what · `runs-on` = which machine · `uses` = ready-made step (like an Ansible module) · `run` = my own shell command.
- 🟡 running → 🟢 passed / 🔴 failed (+ email). Broken code stops **before** the server.
- Free for public repos (private = limited minutes).
- Alternatives: **GitLab CI** (common in Italy), **Jenkins**, Azure DevOps, CircleCI. Same idea: trigger → steps → pass/fail.

### 2. CI builds the image 📦
If tests pass (`needs: test`), GitHub builds the Docker image and pushes it to **GHCR** (GitHub's image storage, like Docker Hub).
- ⭐ Build **once**, test it, deploy the **same** image. Server stops wasting CPU on builds.
- Tag = commit ID (`github.sha`) → every version stays in GHCR → ⭐ **rollback = run an old tag**.
- `ubuntu-24.04-arm` runner because my server is ARM (t4g). `packages: write` = allowed to push. `GITHUB_TOKEN` = password GitHub makes per run.
- My GHCR package is **public** → server pulls without login.
- ⭐ CI catches only what I test: broken `compose.yaml` indent stayed 🟢. Check: `docker compose config -q`.

### 3. The problem: how does GitHub reach my server? 🚧
- GitHub's IPs change every run; SSH (22) is allowed only from my Mac. Opening 22 to the world ❌.
- ⭐ Firewall blocks traffic coming **IN**. Going **OUT** is allowed.
- Server downloading from GHCR = OUT → already fine. Only the "deploy NOW" message is the problem.

### 4. The answer: SSM 📬
⭐ **SSM** (AWS Systems Manager) = AWS runs commands on my server. No SSH, no port 22.
- **SSM agent** = small program already on Ubuntu AWS images (`snap services amazon-ssm-agent`). It asks AWS (OUT) "any command for me?" → runs it.
- GitHub leaves the order at AWS → agent picks it up → server pulls the image from GHCR.
- SSM only **delivers the order**; the server fetches the image itself.
- Bonus: solves my own changing-IP problem too.
- Agent started before it had a role → stayed offline → `snap restart amazon-ssm-agent` → `aws ssm describe-instance-information` → `Online` ✅.

### 5. Permissions (IAM) 🪪
AWS services (SSM, S3, IAM… at `*.amazonaws.com`) trust nobody without an identity — not even programs on my own EC2 (anyone can rent one). "Calling AWS" = sending a request to them, like every `aws ...` command.
- ⭐ IAM **user** = for a person, stored key (`~/.aws`). IAM **role** = for a machine/service, no stored key, temporary keys from AWS.
- 3 identities:
  - my Mac → user `terraform`
  - server → role with `AmazonSSMManagedInstanceCore`
  - GitHub → user `status-page-github`
- Server role = 3 pieces: **role** (who: EC2) + **policy attachment** (what: SSM) + **instance profile** (attaches it to the server). `1 changed` in place → same server, same IP.

### 6. OIDC was blocked ⛔
⭐ **OIDC** = GitHub shows AWS a signed ID ("I am repo MehdiAzizian/status-page, branch main") → AWS gives a 1-hour key. No stored key. Like "Login with Google".
- `403 … explicit deny in a service control policy` → ⭐ **SCP** = rule from the parent organization, above my account. Even admin can't override. Free-plan account lives inside AWS's org.
- Plan B: IAM user with a key in **GitHub Secrets** (encrypted, shown as `***` in logs).
- Allowed only: `ssm:SendCommand` on **my server** + `GetCommandInvocation` (read result) + `ec2:DescribeInstances` (find server, read-only) = ⭐ **least privilege**, small blast radius.
- Less safe than OIDC (key doesn't expire) → rotate it, `destroy` deletes it.
- Key went AWS → GitHub without showing on screen: `read -r KEY SECRET <<< "$(aws iam create-access-key …)"` → `gh secret set` → `unset`.

### 7. Deploy 🚀
- `compose.yaml` app: `image: ghcr.io/mehdiazizian/status-page:${TAG:-latest}` + keep `build: .` for my Mac. `:-latest` = fallback if TAG empty.
- Deploy job (`needs: build`): login with Secrets → find server by tag `Name=status-page` → SSM runs: `git pull` → `docker compose pull app` → `up -d --no-build` → `wait` (server fails = 🔴).
- ⭐ Apply the permission **first**, then push (the pipeline uses it right away).
- ⭐ Proof: server `docker ps` → IMAGE `ghcr.io/…:<SHA>` = same SHA as the GitHub run. Self-built image was called `status-page-app`.
- Only `app` restarted; `db` + `caddy` untouched → Compose restarts only what changed.
- Terraform builds the **server**, never images.

### 8. Rollback ⏪
- `rollback.yml` with `on: workflow_dispatch` = **Run workflow** button in Actions, input = **full** commit SHA.
- Same SSM command as deploy, minus `git pull`.
- Commit ID: `git log` (full) · `--oneline` (short) · GitHub run list · GHCR tags.
- Proof: tag went `ce0c70c…` → `5e9283c…`, new app container, db + caddy untouched. ~30 seconds.

### My mistakes (Phase 3)
- `ci.ylm` instead of `ci.yml` → `mv`. Tip: `mkdir -p .github/workflows`.
- `image:`/`build:` with 2 spaces instead of 4 → `mapping values are not allowed`.
- Deleted resource still used in `outputs.tf` → `Reference to undeclared resource`. My first edit wasn't saved → check with `cat`.
- Ran ansible from `infra/` → "No inventory was parsed".


