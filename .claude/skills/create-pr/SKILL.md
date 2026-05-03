---
name: create-pr
description: >-
  TRIGGER when the user asks to create, open, make, submit, or update a PR or
  pull request for the current branch. Pushes the branch (with upstream if
  needed) and opens the PR with a generated title and body, or updates an
  existing PR if one already exists for the branch. Also use when the user
  says "ship this", "open it up for review", or implies the work is ready
  for code review. Runs a hard pre-flight gate (frontend lint+test+tsc,
  backend tidy+vet+test) and refuses to push if anything fails. Do NOT use
  for committing changes (that's a different workflow) or for already-merged
  PRs.
---

Create a pull request for the current branch. Pushes (with upstream if needed),
then either creates a new PR or updates the existing one.

User input: $ARGUMENTS

Parse the input as follows:
- If empty: detect the base branch (see detection logic below).
- Otherwise: use the entire input as the **base branch** name.

---

## Phase 0 — Pre-flight checks (HARD GATE)

Before pushing or opening the PR, detect what changed and run the matching
verification commands. **If any check fails, stop — do not push, do not open
the PR.** Report the failure to the user and let them fix it. This gate exists
so PRs land in a state reviewers can trust without re-running the suite locally.

### Step 0.1: Detect changed scopes

```bash
BASE="<detected base>"
git diff --name-only "origin/${BASE}..HEAD"
```

Classify each changed path:
- **Frontend scope:** any path under `frontend/`.
- **Backend scope:** any other tracked path that affects Go code — `*.go`, `go.mod`, `go.sum`, `cmd/`, `internal/`, `pkg/`, anything else outside `frontend/`, `docs/`, `.claude/`, `*.md`, image fixtures, and other doc/asset paths.
- **Docs/assets only:** changes confined to `docs/`, `*.md`, `.claude/`, screenshots, and similar non-build paths. Skip both check suites — there's nothing to verify.

A PR can be both frontend AND backend; run the matching suite for each scope present.

### Step 0.2: Frontend checks (only if frontend scope)

```bash
cd frontend
pnpm lint
pnpm test:run
pnpm exec tsc -b
cd -
```

Pass criteria:
- `pnpm lint`: zero new errors compared to the base branch. Pre-existing baseline lint errors are acceptable as long as the diff doesn't add to them — diff against `origin/<base>` for the changed files if needed to confirm.
- `pnpm test:run`: all tests pass.
- `pnpm exec tsc -b`: zero new errors in the touched files. The repo has known pre-existing TS errors; a regression is one whose path matches a file changed in this PR.

If any of these fail, **stop** and report:
```
Pre-flight failed: <which check>. Fix locally before re-running this skill.
<paste the failing output snippet>
```

### Step 0.3: Backend checks (only if backend scope)

```bash
go mod tidy
git diff --exit-code go.mod go.sum
go vet ./...
go test ./...
```

Pass criteria:
- `go mod tidy` followed by `git diff --exit-code go.mod go.sum`: no diff. If `go.mod` / `go.sum` aren't tidy, **stop** and tell the user to commit the tidied files.
- `go vet ./...`: zero issues.
- `go test ./...`: all tests pass.

If a project-specific Makefile target supersedes any of these (e.g. `make test`, `make lint`), prefer that — but only if it actually exists. Don't invent targets.

If any check fails, **stop** and report the same way as the frontend gate.

### Step 0.4: Both scopes

If both frontend and backend scopes are present, run both suites. Both must pass before proceeding to Phase 1.

---

## Phase 1 — Analysis

### Step 1: Branch and base detection

```bash
git rev-parse --abbrev-ref HEAD
git ls-remote --heads origin main master develop 2>/dev/null
```

If the user did not provide a base branch, pick the first one that exists on
`origin` from this list, in order:
1. `main`
2. `master`
3. `develop`

If none exist on the remote, fall back to the repo's GitHub default branch:
```bash
gh repo view --json defaultBranchRef --jq .defaultBranchRef.name
```

### Step 2: Gather commits and diff

Run in a single Bash call so all the context lands together:
```bash
BRANCH=$(git rev-parse --abbrev-ref HEAD)
BASE="<detected base>"
echo "=== COMMITS ==="
git log --oneline --no-merges "origin/${BASE}..HEAD"
echo "=== DIFF STAT ==="
git diff --stat "origin/${BASE}..HEAD"
echo "=== STATUS ==="
git status --short
```

If there are **no commits** ahead of the base branch, say "No commits ahead of
`<base>` — nothing to create." and **stop**.

If you need deeper understanding of specific changes, read key modified files
with `git diff origin/<base>..HEAD -- <path>`.

### Step 3: Check for an existing PR

```bash
BRANCH=$(git rev-parse --abbrev-ref HEAD)
gh pr list --head "$BRANCH" --json url,title,state,number
```

If a PR exists, you'll be **updating** it (`gh pr edit`) instead of creating a
new one — the rest of the flow is the same.

### Step 4: Generate PR content

Analyze the commits and diff to determine:

**PR Title** — format rules:
- Type prefix in title case: `Feature:`, `Fix:`, `Chore:`, `Refactor:`, `Docs:`, `Test:`, `Perf:`, `CI:`
- Rest in normal sentence case
- Under 72 chars total
- Branch-name hints map to prefixes: `feature/` → Feature, `fix/` → Fix, `chore/` → Chore, `refactor/` → Refactor, `docs/` → Docs.

**Writing style** — these rules matter because PRs that read like documentation
get skimmed and PRs that read like a teammate explaining their work get
reviewed:
- Write like you're explaining to a teammate, not documenting for a spec.
- "What this does" = the elevator pitch (1-2 sentences, high-level why).
- "Changes" = concrete what changed (no overlap with the summary above).
- No file paths, function names, or class names unless they ARE the change.
- No per-line prefixes (`fix:` / `feat:`) on bullets — the PR title already has the category.
- Keep change bullets short, one line each, past tense, max 5. Combine related items if needed.
- Test-step language: action-first, short. "Configure the database addon" beats "Configure a Postgres database addon with the connectionString credential field bound to DATABASE_URL".
- Only add a Screenshots section if actual screenshots are being included — never an empty Screenshots heading.

**Issue linking rules:**
- Use `Closes #123` if the PR fully resolves the issue, `Relates to: #123` if partial.
- Multiple parents: `Relates to: #123, #456`.
- Sub-issues: list the issue numbers ONLY — don't repeat the title (GitHub auto-renders titles from issue references).
  ```
  Sub-issues:
  - #124
  - #125
  ```

**Conditional sections — include ONLY when applicable:**
- **References**: include when external spec/design links (a design doc in `docs/`, a Figma URL, a GitHub issue) are available from the conversation context. Skip if no external references exist.
- **Architecture**: include when the PR introduces new entities, permission models, complex flows, or changes relationships between entities. Use mermaid `erDiagram` for entity models and `sequenceDiagram` for flows. Skip for bug fixes, config changes, or UI-only work.
- **API Reference**: include when the PR adds or modifies HTTP endpoints. Use a markdown table with Method, Route, Auth, Request, Response columns. Skip for internal-only changes.
- **Screenshots**: include only when a dev server is running and pages can be captured (e.g. via Playwright MCP or a screenshot tool the user has set up). Take screenshots of key UI changes. Skip entirely if no dev server is available or the PR has no UI changes.

**Main PR body** — use this template, including the emoji prefixes in headings.
Sections after "What this does" are conditional — omit any that don't apply:

````markdown
## 📝 What this does
<1-2 sentence elevator pitch — what changed and why it matters>

## 📎 References
<omit entire section if no spec/design links available>
- **Spec**: [title](path or url)
- **Design**: [title](url)

<Closes #issue OR Relates to: #issue — omit if no related issue>

<Sub-issues: — omit if none>
<- #num>

## 🏗️ Architecture
<omit entire section if no new entities, models, or flows>
### Entity Model
<mermaid erDiagram or bullet list>
### Flows
<mermaid sequenceDiagram or description>

## 🔌 API Reference
<omit entire section if no endpoint changes>
| Method | Route | Auth | Request | Response |
|--------|-------|------|---------|----------|
| **POST** | `/api/...` | `required` | `{ body }` | `{ response }` |

## 🔀 Changes
- <what changed, past tense, no prefixes, max 5 bullets>

## 📸 Screenshots
<omit entire section if no UI changes or no dev server available>
<screenshots inline>

## 🧪 How to test
- [ ] <short action-first step>
````

---

## Phase 2 — Create or update the PR

### Step 1: Push the branch

```bash
BRANCH=$(git rev-parse --abbrev-ref HEAD)
git push -u origin "$BRANCH"
```

`-u` is a no-op once upstream is set, so it's safe to always pass.

### Step 2: Create or update

**If an existing PR was found:** update it.
```bash
gh pr edit <number> --title "<TITLE>" --body "$(cat <<'PREOF'
<MAIN_BODY>
PREOF
)"
```

**If no existing PR:** create it.
```bash
gh pr create --base <base> --title "<TITLE>" --body "$(cat <<'PREOF'
<MAIN_BODY>
PREOF
)"
```

### Step 3: Output the result

Print the PR URL in this format:
```
PR created: <url>
```
or
```
PR updated: <url>
```

---

## Important rules

0. **Pre-flight gate is non-negotiable** (Phase 0). If frontend or backend code changed, the matching check suite must pass before pushing. A broken build doesn't deserve a reviewer's time. The only exception is docs/asset-only PRs.
1. **Always use the heredoc form** (`cat <<'PREOF' ... PREOF`) for PR bodies so markdown renders correctly and special characters (backticks, dollar signs, exclamation marks) don't get interpreted by the shell.
2. If there are no commits ahead of the base branch, say so and **stop**.
3. If the branch name gives hints about the change type (`feature/`, `fix/`, `chore/`), use that to inform the prefix choice.
4. Don't ask the user to review the PR content before creating — just create or update it. Re-running the skill updates it.
5. Don't create duplicate PRs — always check for an existing one first and use `gh pr edit` to update.
6. **Headings include emoji prefixes** exactly as shown (📝, 📎, 🏗️, 🔌, 🔀, 📸, 🧪). The emojis make sections scannable; don't drop them.
7. **No interactive steps** — don't ask questions, request screenshots, or wait for user input. Run all steps autonomously.
8. **No author signature lines** in the PR body unless the user explicitly asks (e.g., the project's `CLAUDE.md` has a Co-Authored-By convention for commits, but PR bodies stay clean).

## Why these rules

- **Heredocs over `--body "..."`**: PR bodies frequently contain backticks (code spans), dollar signs (env var examples), and `!` (history expansion). Quoted heredocs sidestep all of it.
- **Update existing PRs**: opening a duplicate PR fragments review history and wastes the reviewer's attention. The single source of truth stays the original PR.
- **Conditional sections**: a PR with empty `## 🏗️ Architecture` or `## 🔌 API Reference` headings reads as ceremonial noise. Better to omit the section than to leave it bare.
- **Action-first test steps**: reviewers skim test plans. "Configure the data source" gives the verb up front; "Configure a data source with the X mode pointing at..." buries it.
