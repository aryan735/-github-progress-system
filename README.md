# GitHub Progress System

Automatically tracks your daily GitHub commits across all repositories and writes a summary to a separate progress-log repository.

**Application repo:** [aryan735/-github-progress-system](https://github.com/aryan735/-github-progress-system)  
**Progress log repo:** [aryan735/engineering-progress-log](https://github.com/aryan735/engineering-progress-log)

---

## What it does

Every day (10:30 PM IST via GitHub Actions):

1. Lists all your GitHub repositories
2. Collects commits made **today** (IST timezone)
3. Builds a markdown summary (productive day or no-activity)
4. Commits the summary to **`engineering-progress-log.md`** in the [engineering-progress-log](https://github.com/aryan735/engineering-progress-log) repo only

This repo holds the **application code**. Your daily logs live in a **separate repo** so history stays clean and public-friendly.

---

## Example log output

**Productive day:**

```markdown
## 2026-05-27

Status: Productive day

Repositories worked on:
- aryan735/my-api
- aryan735/my-frontend

Total commits: 4

Commits:
- add github sdk client
- parse GitHub push payload safely
- filter today commits
- update progress summary
```

**No activity:**

```markdown
## 2026-05-28

Status: No GitHub commits today

Repositories worked on:
- none

Total commits: 0

Notes:
- No commits detected across GitHub today.
- Daily progress log updated automatically.
```

---

## Project structure

```
cmd/
  daily/          # CLI used by GitHub Actions (one-shot job)
  server/         # Optional HTTP API for local testing
internal/
  config/         # Config loader (file + environment variables)
  github/         # GitHub API client, commit pipeline, log sync
  handler/        # HTTP route handlers
  scheduler/      # Daily job orchestration
.github/workflows/
  daily-progress.yml   # Scheduled workflow (10:30 PM IST)
```

---

## Setup for your own account

### 1. Fork or clone this repo

```bash
git clone https://github.com/aryan735/-github-progress-system.git
cd -github-progress-system
```

### 2. Create a progress-log repository

Create an empty repo on GitHub (e.g. `engineering-progress-log`) where daily summaries will be written.

### 3. Create a GitHub Personal Access Token

- Go to **GitHub → Settings → Developer settings → Personal access tokens**
- Create a token with **`repo`** scope (needed to read all repos and write the log file)

### 4. Local configuration

Copy the example config and add your token:

```bash
cp internal/config/config.example.yml internal/config/config.yml
```

Edit `internal/config/config.yml`:

```yaml
github_token: "ghp_your_token_here"

progress_log:
  owner: your-github-username
  repo: engineering-progress-log
  branch: main
  path: engineering-progress-log.md
```

Or use environment variables (recommended for CI):

| Variable | Description | Example |
|----------|-------------|---------|
| `GH_PROGRESS_TOKEN` | GitHub PAT with `repo` scope | `ghp_...` |
| `PROGRESS_LOG_REPO_URL` | Target repo for daily logs | `https://github.com/you/engineering-progress-log` |
| `PROGRESS_LOG_PATH` | Markdown file path in that repo | `engineering-progress-log.md` |
| `PROGRESS_LOG_BRANCH` | Branch to commit to | `main` |
| `TZ` | Timezone for "today" | `Asia/Kolkata` |

### 5. GitHub Actions secret

In **this application repo**, add:

- **Settings → Secrets → Actions → New secret**
- Name: `GH_PROGRESS_TOKEN`
- Value: your PAT

The workflow already sets `PROGRESS_LOG_REPO_URL` to the engineering progress log repo.

### 6. Run locally

**One-shot daily job (same as the workflow):**

```bash
export GH_PROGRESS_TOKEN=ghp_your_token
export PROGRESS_LOG_REPO_URL=https://github.com/you/engineering-progress-log
go run ./cmd/daily
```

**Optional API server:**

```bash
go run ./cmd/server
```

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/activity/today` | Preview today's commits (JSON) |
| POST | `/progress/daily` | Run full pipeline and sync log |

---

## How the scheduler works

```
List authenticated user
        ↓
List all repos (skip progress-log repo)
        ↓
For each repo → commits since start of today (IST)
        ↓
Build summary markdown
        ↓
Create/update engineering-progress-log.md in target repo
        ↓
Commit: docs: update progress log for YYYY-MM-DD
     or docs: add no-activity log for YYYY-MM-DD
```

The workflow runs automatically at **10:30 PM IST** (`30 17 * * *` UTC).  
You can also trigger it manually: **Actions → Daily Progress Log → Run workflow**.

---

## Requirements

- Go 1.25+
- GitHub PAT with `repo` scope
- A separate empty repository for progress logs

---

## License

Use freely for personal progress tracking. Adapt the progress-log repo URL and file path to your own setup.
