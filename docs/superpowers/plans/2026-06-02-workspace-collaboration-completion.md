# Workspace Collaboration — Completion Plan & Handoff

> **Scope update (2026-06-02, executed):** Product decision to **shelve the Users + Teams UI** for now (nav hidden, `/settings/*` routes redirect home; page code kept in-repo, unrouted). Net delivered this session:
> - **Slice A** — team-role helpers (`roleInTeam`/`canWrite`/`canWriteAnyTeam`) on the current-user context. ✅
> - **Slice G (new)** — hide Users + Teams (nav + route guard). ✅
> - **Slice B** — re-point resource writes (stacks/secrets/object-stores/addons) to team-scoped endpoints. New resources target the **org default team** (resolved via `useResourceTeams`/`listTeams`, not `user.teams`, so OrgAdmins who aren't members still resolve). ✅ (create→201, delete→204 verified live)
> - **Slice C** — Viewer read-only enforcement: **hide** (not disable) mutating controls; create gated on `canWriteAnyTeam`, per-row/detail on `canWrite(resource.team_id)`. ✅ (verified live as Viewer + admin)
> - **Slices D & E dropped** — demote / team-delete verification belonged to the now-shelved Users/Teams UI. (Note: `mage migrate` this session applied `add_timestamps_to_volumes`, which is the **#84** fix; and `create_stack_connections`, which the local DB was missing — that was the root of an observed secret-delete 500, a stale-DB issue, not a code defect.)

**Date:** 2026-06-02
**Scope:** Close out PRD [#62](https://github.com/Stackdome/stackdome/issues/62) **+ Viewer access enforcement** (hide/disable mutating controls for team Viewers).
**Parent PRD:** #62 — Workspace Collaboration (Users, Teams & Invite Acceptance, frontend).
**Branch / PR:** `worktree-feat+workspace-collaboration` → PR [#87](https://github.com/Stackdome/stackdome/pull/87).
**Out of scope:** stack connections/topology UI, the 404-write-layer beyond what Viewer enforcement needs is included here; pure connections feature is a separate effort.

---

## 0. Orientation — read first

- **Repo:** `Stackdome/stackdome`. Work in the worktree `…/.claude/worktrees/feat+workspace-collaboration` (do NOT cd to the main checkout).
- **Dev servers:** frontend Vite on `http://localhost:5173`; backend api-server on `:8000`.
  - Frontend hot-reloads. Backend does NOT — after any backend change run `mage build && mage run`. **The running binary can be stale**; if a backend behaviour looks wrong, confirm the binary was rebuilt after the last `git` change.
- **Login (admin / OrgAdmin):** `akshay@stackdome.dev` / `Password123!`.
  - Test members already in DB: `teammate2@example.com`, `teammate3@example.com` (both `Password123!`). `teammate3` is OrgMember/Developer on `default`; `teammate2` is currently OrgAdmin (demote is broken — see #88).
- **Verify everything live with Playwright MCP** against `:5173`, as both an admin and a member session. UI feedback loop is mandatory, not just at the end.
- **HARD CONSTRAINT: no backend changes in this PR.** Backend defects are filed as issues for separate AFK agents (#84, #88). Frontend-only here.
- **Design system:** use `frontend/src/index.css` tokens + `@/components/branded` and `@/components/ui` primitives. No raw hex, no off-scale type.
- **Pre-flight before pushing:** `pnpm --prefix frontend lint`, `pnpm --prefix frontend test:run`, `pnpm --prefix frontend exec tsc -b`. The PR uses the `create-pr` skill which gates on these.

---

## 1. Current state (already done on this branch)

- **#62 slices 1–7 built** (Users page, invite dialog + state machine, pending/user row actions, Teams list, Team detail, invite acceptance on `/sign-up`).
- **Auth/nav OrgMember lockout fixed** (`fix(auth)` commit): `frontend/src/api/client.ts` now logs out only on **401** (was 401 **or** 403 → members were ejected by an RBAC 403 from the sidebar's `/clusters` fetch). `app-sidebar.tsx` hides Clusters/Domains for non-admins; `App.tsx` route-guards `/clusters`,`/domains` with `RequireAdmin`; `login-form.tsx` calls `refresh()` post-login so role-gated nav is correct without a manual reload.
- **Team rename implemented** (`eaef221`) — was a dead button in both the list row menu and the detail header; added `RenameTeamDialog`, wired both, detail-page rename navigates to the new name-slug. Satisfies #62 story 29.
- **Merged `origin/main`** (`ac58fe2`) and **regenerated OpenAPI types** (`2886144`); removed the stacks env-from-secret/addon form feature (main's connections redesign deleted those API fields) to keep the branch compiling. **This is stacks-feature fallout, not #62.**
- **Green:** `tsc -b` baseline only (`src/api/postgres-backups.ts:11` is a known pre-existing drift, leave it), `pnpm test:run` = 255 passing.

---

## 2. Known defects — filed, BACKEND, out of scope (block two #62 stories)

| Issue | #62 story | Symptom | Root cause |
|---|---|---|---|
| [#84](https://github.com/Stackdome/stackdome/issues/84) | Story 31 — delete team | `DELETE /teams/{name}` → **500** | `volumes` table missing `created_at`; team-delete dependency check lists volumes ordered by it |
| [#88](https://github.com/Stackdome/stackdome/issues/88) | Story 22 — demote OrgAdmin | demote → **204 but role unchanged** | `dbUserStore.Update` (`pkg/stores/pgstore/users.go:67`) GORM `.Updates(struct)` drops zero-value `Role=NoRole("")` |

The **frontend for both is correct**. Slices D & E below are just E2E verification once these land.

---

## 3. Technical facts the next context needs (verified this session)

- **Frontend permission model today = one boolean.** `CurrentUserProvider` (`frontend/src/contexts/current-user-context.tsx`) exposes only `isOrgAdmin = user.role === "OrgAdmin"`. The user object carries `teams: [{ team_id, team_name, role: "Developer"|"Viewer", default_team }]` but **nothing reads team roles**. So Developer and Viewer look identical to the UI.
- **Backend authz is per-team.** Developer-on-team-X can write team-X resources; Viewer-on-team-X is read-only; OrgAdmin can write any team. A user can be Developer in one team and Viewer in another simultaneously.
- **Resources carry `team_id`.** Confirmed in the regenerated spec (e.g. `Stack` has `team_id`). This is what makes per-resource Viewer gating possible frontend-side. (Confirm whether responses also carry `team_name`; team-scoped endpoint URLs use `team_name`, so if only `team_id` is present, map via the teams list.)
- **WRITES currently 404.** The frontend's `createStack`/`updateStack`/`createSecret`/`updateSecret`/`deleteSecret`/object-store/addon writes POST/PUT/DELETE to **org-scoped** paths (`/organizations/{org}/stacks…`) which are **GET-only** → 404 for everyone (incl. admin). The real CRUD lives at **team-scoped** endpoints (`/organizations/{org}/teams/{team_name}/stacks…`), which the frontend never calls. Verified: org-scoped `POST/PUT/DELETE` → 404.
- **Org-scoped LIST self-filters by membership** (`ListForCurrentUser` pattern): a member's `GET /organizations/{org}/stacks` returns only their teams' resources (200). Clusters/Domains are OrgAdmin-only (403 for members) and are now nav-hidden + route-guarded.

---

## 4. The plan — vertical slices

Ordered by dependency. Each slice is a thin end-to-end path, independently verifiable. **A → B → C** are the build work; **D, E** are external-blocked verification; **F** is close-out.

### Slice A — Team-role helpers on the current-user context  *(foundation, frontend-only, no deps)*
**What:** Extend `CurrentUserValue` with team-role resolution derived from `user.teams`. Add `roleInTeam(teamId | teamName): "Developer" | "Viewer" | undefined` and `canWrite(teamRef): boolean` (= `isOrgAdmin || roleInTeam(...) === "Developer"`).
**Done when:** any component can ask "can this user write in team X" without reading raw `user.teams`.
**Acceptance:**
- [ ] `roleInTeam` returns the correct role for a team the user belongs to, `undefined` otherwise.
- [ ] `canWrite` returns true for OrgAdmin regardless of team; true for Developer-in-team; false for Viewer-in-team and non-members.
- [ ] Unit tests cover OrgAdmin, Developer, Viewer, multi-team (Developer in A / Viewer in B), and non-member.
**Notes:** pure addition to `current-user-context.tsx` + a test. No UI change yet.

### Slice B — Re-point resource writes to team-scoped endpoints  *(fixes the 404 write layer; prerequisite for Developer writes to actually work)*
**What:** Change the resource write API wrappers to call the **team-scoped** CRUD endpoints using each resource's team. Covers stacks, secrets, object-stores, postgres addons (create/update/delete). Resolve the target team from the resource's `team_id` (→ `team_name` via the teams list if needed); for create, the team is chosen by the user/context.
**Done when:** a Developer (and an OrgAdmin) can create/edit/delete a stack and a secret in their team through the UI — operations that currently 404.
**Acceptance:**
- [ ] Create / edit / delete a secret in a team you're a Developer of → 2xx, reflected in the list (was 404).
- [ ] Same for a stack (create via the create page, edit via detail, delete).
- [ ] No org-scoped write calls remain (`grep` for `api.post/put/delete('/organizations/${orgId}/(stacks|secrets|object-stores|addons)'` returns nothing).
- [ ] Reads still use the org-scoped list (which self-filters) — unchanged.
**Notes:** heaviest slice; it's stacks-feature plumbing but is the prerequisite for Viewer gating to be meaningful (you can only gate controls that actually function). Confirm the team source: does the list response carry `team_name`, or only `team_id`? If only `team_id`, add a `team_id → team_name` lookup from `useTeamOptions()`/listTeams. Watch the "which team does a NEW resource go to" question — there may need to be a team selector on the create flow if not already implied by context.

### Slice C — Viewer read-only enforcement on resource pages  *(the headline of this scope)*
**What:** Using `canWrite(resource.team_id)` from Slice A, hide or disable every mutating control for users who can't write the relevant team: create buttons, edit/delete row actions, deploy actions, form submits. Pages: stacks (list / create / detail), secrets, object-stores, addons. A Viewer sees data, never a mutate affordance; a Developer/OrgAdmin sees them.
**Done when:** logged in as a Viewer, the resource pages are read-only; as a Developer/OrgAdmin they're writable.
**Acceptance:**
- [ ] As a Viewer (make one: invite to a team with role Viewer, accept), the Secrets/Stacks/Object-Stores/Addons pages show no Add/Edit/Delete/Deploy controls (hidden or disabled-with-tooltip — pick one and be consistent).
- [ ] As a Developer in the same team, those controls are present and work (depends on Slice B).
- [ ] "Create" entry points (e.g. Add Secret, Create Stack) are gated by whether the user can write in *any* team (else there's nowhere to create) — decide and document the rule for users with mixed roles.
- [ ] Per-row gating uses the row's `team_id`, not a global flag (correct for multi-team users).
- [ ] Verified live with Playwright for both a Viewer and a Developer session.
**Notes:** Slice C is independently valuable even before Slice B (hiding controls a Viewer shouldn't see), but the Developer-side acceptance needs B to actually function. If B slips, ship C's hide-for-Viewers half and note Developer writes are still 404-blocked by B.

### Slice D — Verify demote end-to-end  *(blocked by #88; no frontend build)*
**What:** Once #88 lands, confirm story 22 works through the UI.
**Done when:** demoting an OrgAdmin via the user-row menu actually flips their role and the row updates.
**Acceptance:**
- [ ] Promote a member → OrgAdmin, then demote into a team with a role; the row reflects OrgMember and the team membership appears. (Note: the demote UI is a Radix `Select` nested inside a `DropdownMenu` in `user-row-menu.tsx` — if it misbehaves once the backend works, consider moving the demote sub-form into a proper `Dialog`.)

### Slice E — Verify team delete end-to-end  *(blocked by #84; no frontend build)*
**What:** Once #84 lands, confirm story 31.
**Done when:** deleting a non-default team via type-to-confirm removes it.
**Acceptance:**
- [ ] Create a throwaway team, delete it with the type-the-name confirm → row gone, success toast (no 500).
- [ ] The delete dialog surfaces a real error inline if the API fails (it currently stays open silently on 500 — worth hardening).

### Slice F — #62 E2E pass, test-data cleanup, PR finalize  *(close-out)*
**What:** Run the manual test plan, clean up the artifacts this session left, finalize PR #87.
**Done when:** the test plan passes (minus externally-blocked rows), the DB is clean, PR is review-ready.
**Acceptance:**
- [ ] Walk the manual test plan (in vault: `Stackdome/superpowers/test-plans/2026-05-27-…-test-plan.md`) end-to-end as admin + member.
- [ ] Delete leftover test data once #84 lands: users `Tee Two`/teammate2, `Three Member`/teammate3; teams `quality-eng`, `deltest`. (Team deletes are blocked by #84 until it's fixed.)
- [ ] `lint` + `test:run` + `tsc -b` green (baseline `postgres-backups.ts:11` only).
- [ ] PR #87 description updated to cover the auth fix, rename, merge/regen, and Viewer enforcement.

---

## 5. Verification recipe (Playwright)

- **Make a clean OrgMember/Viewer for testing** (the demote path is broken, so don't rely on it): as admin, `POST /organizations/{org}/invites` with `{ email, team_name, role: "Viewer", expires_in_days: 1 }`, grab the returned `token`, then `POST /api/v1/user-signup` with `{ name, email, password, invite_token }`. The new user is OrgMember with the chosen team role. (Do this via `browser_evaluate` fetch calls — same-origin from `:5173`.)
- **Switch sessions** by `localStorage.clear()` then logging in through the UI at `/sign-in` (don't hand-inject tokens — exercise the real flow, which also tests the post-login `refresh()`).
- **Org id** for this dev DB: `c2e2a498-48bb-4a68-9b6c-37e5df643179` (re-fetch from `/users/current` to be safe).
- Toasts auto-dismiss fast — install a `MutationObserver` on `document.body` before the action if you need to capture toast text.

---

## 6. References

- **PRD:** #62 (full user stories, slices 1–7). Design bundle: `https://api.anthropic.com/v1/design/h/xMpi42Gva5l6WtOMB6vSww`.
- **Backend bugs:** #84 (team delete), #88 (demote).
- **Manual test plan:** vault `Stackdome/superpowers/test-plans/2026-05-27-workspace-collaboration-test-plan.md` (also a results log dated 2026-05-30 alongside it).
- **Original plan:** vault `Stackdome/superpowers/plans/2026-05-18-workspace-collaboration.md`.
- **Key files:** `frontend/src/contexts/current-user-context.tsx`, `frontend/src/api/{client,stacks,secrets,object-stores,addons}.ts`, `frontend/src/pages/{stacks,secrets,object-stores,addons,users,teams}/`, `frontend/src/components/app-sidebar.tsx`, `frontend/src/App.tsx`.
