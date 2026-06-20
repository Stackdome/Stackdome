# GitHub OAuth — Frontend Support (Design)

**Date:** 2026-06-21
**Status:** Approved (design)
**Branch:** `worktree-feat+github-oauth-frontend` (off `main`)

## Problem

The backend already ships GitHub OAuth login (commit `9cf8124 feat: add auth infrastructure — API tokens, refresh tokens, GitHub OAuth`). The web SPA has no entry point for it: no "Sign in with GitHub" button, no callback page, no way to know whether OAuth is even enabled on the server. This adds that frontend support, plus one small backend endpoint so the frontend can detect whether OAuth is configured.

## Backend contract (already exists — do not change)

Routes are registered **only when** `Config.GitHubOAuth.Enabled()` (i.e. `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET` are both set):

| Route | Behavior |
|---|---|
| `GET /api/v1/auth/github?invite_token=<optional>` | Generates random `state`, stores it in `oauth_states` (invite_token encrypted), `307`-redirects the browser to GitHub's authorize URL. Scope `user:email`. |
| `GET /api/v1/auth/github/callback?code=&state=` | Validates+consumes `state` (10-min expiry), exchanges `code`, fetches the GitHub user + verified email, find-or-creates the user (joining the inviting org if `invite_token` present), returns **JSON** `RefreshTokenResponse { token, refresh_token }` with HTTP `200`. **No `user` object in the response.** |

Config env vars (operator-set, no defaults): `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `GITHUB_REDIRECT_URI`.

## Key decisions

1. **Flow — frontend route re-calls the backend callback via XHR.** Because the callback returns JSON (not a redirect), the browser cannot consume it from a top-level navigation. So `GITHUB_REDIRECT_URI` points at a **frontend** route, which reads `code`+`state` from its URL and calls the backend callback over axios. No change to the backend OAuth handler. The access token never appears in a URL — only the short-lived authorization `code` does.
2. **Button gating — a new `GET /api/v1/config` endpoint** exposes `{ github_oauth: boolean }`. The frontend fetches it on the auth pages and renders the button only when true. Chosen over a build-time env flag so the UI stays in sync with the server's actual config. Fail-closed: if the config fetch fails, hide the button.
3. **Scope — login + signup + invite-accept.** The button appears on all three; the invite-accept flow threads its `invite_token` into the initiate URL so an OAuth signup joins the inviting org.

## Architecture

### A. Backend slice (config endpoint only)

1. **OpenAPI spec** (`config/openapi/stackdome_api.yaml`): add public path `GET /api/v1/config` → `200` with schema `AppConfigResponse { github_oauth: boolean }`. The REST contract is the source of truth; regenerate the Go client (`make generate`) and the frontend types (`pnpm --prefix frontend generate:openapi-types`).
2. **Handler** `pkg/handlers/config_handler.go`: returns `{ github_oauth: cfg.GitHubOAuth.Enabled() }`. Wired through the environment like other handlers.
3. **Route registration** in `cmd/server/routes.go`: register **public** (no auth middleware), next to `/auth/login` and `/auth/refresh`.
4. **Deployment**: document that `GITHUB_REDIRECT_URI` must equal `<frontend-origin>/auth/github/callback`. Update `.env_template`.

### B. Frontend

5. **`frontend/src/api/config.ts`** — `getAppConfig(): Promise<{ github_oauth: boolean }>`. Typed off the generated `components["schemas"]["AppConfigResponse"]`, using the shared `api` client (pattern from `api/users.ts`).
6. **`frontend/src/api/auth-github.ts`**
   - `completeGitHubOAuth(code, state): Promise<{ token, refresh_token }>` → `api.get("/auth/github/callback", { params: { code, state } })`.
   - `githubOAuthUrl(inviteToken?): string` → `` `${import.meta.env.VITE_API_BASE_URL || "/api/v1"}/auth/github` `` + `?invite_token=` when present.
7. **`GitHubSignInButton`** (shared component, branded via `ui/button` + GitHub icon): `onClick → window.location.assign(githubOAuthUrl(inviteToken))`. Renders only when app config reports `github_oauth` true.
8. **New public route `/auth/github/callback`** → `GithubCallbackPage`:
   - Read `code` + `state` via `useSearchParams`.
   - On mount: `completeGitHubOAuth(code, state)` → store `authToken` + `refreshToken` → `getCurrentUser()` (now authorized) → `setAuthSession(token, user, refresh_token)` → `useCurrentUser().refresh()` → `navigate("/dashboard")`.
   - Missing `code`/`state`, or backend `400/401/500` → error state with `getErrorMessage` + "Back to sign in" link.
   - Registered outside `RequireAuth` in `frontend/src/App.tsx` (alongside `/sign-in`, `/sign-up`).
9. **Button placement**: `login-form.tsx`, `signup-form.tsx`, `invite-accept-form.tsx` (invite passes its `token` as `invite_token`). An "or" divider above each.

## Data flow

```
[Sign in with GitHub]
  → window.location → GET /api/v1/auth/github[?invite_token=…]   (browser nav)
  → GitHub consent
  → browser redirect → /auth/github/callback?code=…&state=…       (SPA page)
  → GithubCallbackPage: axios GET /api/v1/auth/github/callback?code&state
  → { token, refresh_token }  (no user)
  → store authToken + refreshToken
  → GET /api/v1/users/current → user
  → setAuthSession(token, user, refresh_token)
  → useCurrentUser().refresh()
  → navigate /dashboard
```

## Error handling

- App-config fetch fails → button hidden (fail-closed).
- Callback page missing `code`/`state` → error UI, no backend call.
- Backend callback error JSON → `getErrorMessage` → error UI with retry / "Back to sign in".
- Access token is never placed in a URL; only the short-lived auth `code` is.

## Testing

- **Frontend (vitest):** `GithubCallbackPage` — success path, backend-error path, missing-params path; `GitHubSignInButton` gating (config on/off); `githubOAuthUrl` builder with and without `invite_token`.
- **Backend:** `config_handler` unit test for enabled and disabled config.
- **Regen check:** confirm `make generate` / `generate:openapi-types` produce no unexpected diff beyond the new endpoint.

## Units (each single-purpose, independently testable)

- `api/config.ts` — fetch server feature flags.
- `api/auth-github.ts` — OAuth URL builder + callback completion.
- `GitHubSignInButton` — gated entry-point button.
- `GithubCallbackPage` — round-trip completion + session bootstrap.

## Out of scope (YAGNI)

- No backend change to the OAuth handler (no redirect-with-tokens variant).
- No other OAuth providers; `AppConfigResponse` starts with a single `github_oauth` flag (extensible later).
- No "link GitHub to existing account" settings UI — first-login find-or-create only, per existing backend behavior.
