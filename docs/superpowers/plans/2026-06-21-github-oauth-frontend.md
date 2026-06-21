# GitHub OAuth Frontend Support Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add "Sign in with GitHub" to the web SPA (login, signup, invite-accept) plus a public config endpoint so the button only shows when the server has OAuth configured.

**Architecture:** The backend OAuth handler is unchanged. Its callback returns JSON (`RefreshTokenResponse { token, refreshToken }`), so `GITHUB_REDIRECT_URI` points at a frontend route (`/auth/github/callback`) that reads `code`/`state` from its URL and calls the backend callback via XHR, then bootstraps the session (store tokens → fetch current user → `setAuthSession`). A new public `GET /api/v1/config` exposes `{ github_oauth: bool }` to gate the button.

**Tech Stack:** Go (gorilla/mux, net/http), React 19 + react-router-dom v7 + Vite + Tailwind v4, vitest + @testing-library/react, lucide-react, OpenAPI (openapi-typescript).

**Source spec:** `docs/superpowers/specs/2026-06-21-github-oauth-frontend-design.md` · **PRD:** issue #101 (slices S1–S3).

---

## File Structure

**Slice S1 — config endpoint + gating signal**
- Modify: `config/openapi/stackdome_api.yaml` (new path `/api/v1/config` + schema `AppConfigResponse`)
- Create: `pkg/handlers/config_handler.go`, `pkg/handlers/config_handler_test.go`
- Modify: `cmd/server/routes.go` (register public `/config`)
- Modify: `.env_template` (GitHub OAuth vars)
- Modify: `frontend/src/api/types/openapi.d.ts` (regenerated)
- Create: `frontend/src/api/config.ts`, `frontend/src/hooks/use-app-config.ts` (+ tests)

**Slice S2 — GitHub sign-in on login (core round-trip)**
- Create: `frontend/src/api/auth-github.ts` (+ test)
- Create: `frontend/src/components/auth/github-sign-in-button.tsx` (+ test)
- Create: `frontend/src/pages/auth/github-callback.tsx` (+ test)
- Modify: `frontend/src/App.tsx` (public route), `frontend/src/pages/login/components/login-form.tsx`

**Slice S3 — signup + invite**
- Modify: `frontend/src/pages/signup/components/signup-form.tsx`, `frontend/src/pages/signup/components/invite-accept-form.tsx`

Each slice is independently shippable. S2 consumes S1's `useAppConfig`; S3 reuses S2's `GitHubSignInButton`.

---

# Slice S1 — Config endpoint + gating signal

## Task 1: Add `/api/v1/config` to the OpenAPI contract

**Files:**
- Modify: `config/openapi/stackdome_api.yaml`

- [ ] **Step 1: Add the path.** In the `paths:` block, right after the `/api/v1/user-signup:` entry (before the next path), add:

```yaml
  /api/v1/config:
    get:
      summary: Get public application configuration
      description: Returns feature flags the web client needs before authentication, such as whether GitHub OAuth is enabled.
      responses:
        '200':
          description: Application configuration
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/AppConfigResponse'
        '500':
          description: Internal server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
```

- [ ] **Step 2: Add the schema.** In `components:` → `schemas:`, add a new entry at the same 4-space indentation as sibling schemas (e.g. right after the `schemas:` line, before the first existing schema):

```yaml
    AppConfigResponse:
      type: object
      properties:
        github_oauth:
          type: boolean
          description: True when GitHub OAuth login is configured on the server.
```

- [ ] **Step 3: Regenerate frontend types** (required — the FE consumes `components["schemas"]["AppConfigResponse"]`).

Run: `pnpm --prefix frontend generate:openapi-types`
Expected: `frontend/src/api/types/openapi.d.ts` now contains `AppConfigResponse: { github_oauth?: boolean; }`.

- [ ] **Step 4: Regenerate the Go client for contract parity** (keeps `pkg/api/openapi` in sync; needs Docker). The backend handler in Task 2 deliberately uses a local struct, so this is NOT a compile blocker — if Docker is unavailable, note it and proceed.

Run: `make generate`
Expected: regenerated `pkg/api/openapi`; review the diff is limited to the new schema/path.

- [ ] **Step 5: Verify the spec is valid YAML and types generated.**

Run: `grep -n "AppConfigResponse" frontend/src/api/types/openapi.d.ts`
Expected: at least one match.

- [ ] **Step 6: Commit.**

```bash
git add config/openapi/stackdome_api.yaml frontend/src/api/types/openapi.d.ts pkg/api/openapi
git commit -m "feat(api): add public GET /api/v1/config (github_oauth flag) to OpenAPI"
```

## Task 2: Backend config handler (TDD)

**Files:**
- Create: `pkg/handlers/config_handler.go`
- Test: `pkg/handlers/config_handler_test.go`
- Modify: `cmd/server/routes.go`

Design note: the handler takes a plain `bool`, not the config object — a deep, isolated interface that is trivially testable and never changes. The route wiring passes `Config.GitHubOAuth.Enabled()`. Response uses a local struct matching the OpenAPI schema (`{"github_oauth": bool}`) via the existing `writeJSONResponse` helper (`pkg/handlers/handler_helper.go:206`).

- [ ] **Step 1: Write the failing test.** Create `pkg/handlers/config_handler_test.go`:

```go
package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConfigHandler_Get(t *testing.T) {
	cases := []struct {
		name    string
		enabled bool
	}{
		{"enabled", true},
		{"disabled", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewConfigHandler(ConfigHandlerSpec{GitHubOAuthEnabled: tc.enabled})
			req := httptest.NewRequest(http.MethodGet, "/api/v1/config", nil)
			rec := httptest.NewRecorder()

			h.Get(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", rec.Code)
			}
			var body struct {
				GithubOauth bool `json:"github_oauth"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if body.GithubOauth != tc.enabled {
				t.Fatalf("expected github_oauth=%v, got %v", tc.enabled, body.GithubOauth)
			}
		})
	}
}
```

- [ ] **Step 2: Run the test to verify it fails.**

Run: `go test ./pkg/handlers/ -run TestConfigHandler_Get`
Expected: FAIL (compile error — `NewConfigHandler` / `ConfigHandlerSpec` undefined).

- [ ] **Step 3: Implement the handler.** Create `pkg/handlers/config_handler.go`:

```go
package handlers

import "net/http"

type ConfigHandlerSpec struct {
	GitHubOAuthEnabled bool
}

type configHandler struct {
	githubOAuthEnabled bool
}

func NewConfigHandler(spec ConfigHandlerSpec) *configHandler {
	return &configHandler{githubOAuthEnabled: spec.GitHubOAuthEnabled}
}

// appConfigResponse mirrors the AppConfigResponse schema in the OpenAPI spec.
type appConfigResponse struct {
	GithubOauth bool `json:"github_oauth"`
}

func (h *configHandler) Get(w http.ResponseWriter, r *http.Request) {
	writeJSONResponse(w, http.StatusOK, appConfigResponse{GithubOauth: h.githubOAuthEnabled})
}
```

- [ ] **Step 4: Run the test to verify it passes.**

Run: `go test ./pkg/handlers/ -run TestConfigHandler_Get`
Expected: PASS (both subtests).

- [ ] **Step 5: Register the public route.** In `cmd/server/routes.go`, near the `authenticationRouter` setup (where `/auth/login` and `/auth/refresh` are registered), register `/config` directly on `apiV1Router` so it is public (no Bearer middleware). Add:

```go
	configHandler := handlers.NewConfigHandler(handlers.ConfigHandlerSpec{
		GitHubOAuthEnabled: s.environment.Environment().Config.GitHubOAuth.Enabled(),
	})
	apiV1Router.HandleFunc("/config", configHandler.Get).Methods(http.MethodGet)
```

(`apiV1Router := mainRouter.PathPrefix("/api/v1").Subrouter()` already exists; auth is applied per-subrouter, e.g. `authenticatedUserRouter`, so `apiV1Router`-level routes are unauthenticated — same as `/auth/login`.)

- [ ] **Step 6: Verify the server still builds.**

Run: `go build ./...`
Expected: no errors.

- [ ] **Step 7: Commit.**

```bash
git add pkg/handlers/config_handler.go pkg/handlers/config_handler_test.go cmd/server/routes.go
git commit -m "feat(api): config handler exposing github_oauth enabled flag"
```

## Task 3: Document the GitHub OAuth env vars

**Files:**
- Modify: `.env_template`

- [ ] **Step 1: Append the OAuth vars.** Add to `.env_template` (after the existing `GHACCESS_TOKEN=` line):

```
# GitHub OAuth login (optional). When CLIENT_ID + CLIENT_SECRET are set, the
# /api/v1/auth/github routes register and the web UI shows "Continue with GitHub".
# GITHUB_REDIRECT_URI must equal <frontend-origin>/auth/github/callback and match
# the GitHub OAuth app's registered Authorization callback URL.
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
GITHUB_REDIRECT_URI=
```

- [ ] **Step 2: Commit.**

```bash
git add .env_template
git commit -m "docs(env): document GitHub OAuth vars + frontend redirect URI"
```

## Task 4: Frontend `getAppConfig` + `useAppConfig` (TDD)

**Files:**
- Create: `frontend/src/api/config.ts`
- Create: `frontend/src/hooks/use-app-config.ts`
- Test: `frontend/src/hooks/tests/use-app-config.test.tsx`

- [ ] **Step 1: Add the API wrapper.** Create `frontend/src/api/config.ts` (mirrors `api/users.ts`):

```ts
import api from "./client";
import type { components } from "../api/types/openapi";

export type AppConfigResponse = components["schemas"]["AppConfigResponse"];

export async function getAppConfig(): Promise<AppConfigResponse> {
  const response = await api.get("/config");
  return response.data;
}
```

- [ ] **Step 2: Write the failing hook test.** Create `frontend/src/hooks/tests/use-app-config.test.tsx`:

```tsx
// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { useAppConfig } from "@/hooks/use-app-config";
import * as configApi from "@/api/config";

vi.mock("@/api/config");

describe("useAppConfig", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns githubOAuth true when the server reports it enabled", async () => {
    vi.mocked(configApi.getAppConfig).mockResolvedValue({ github_oauth: true });
    const { result } = renderHook(() => useAppConfig());
    await waitFor(() => expect(result.current.githubOAuth).toBe(true));
  });

  it("fails closed (githubOAuth false) when the config request errors", async () => {
    vi.mocked(configApi.getAppConfig).mockRejectedValue(new Error("boom"));
    const { result } = renderHook(() => useAppConfig());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.githubOAuth).toBe(false);
  });
});
```

- [ ] **Step 3: Run it to verify it fails.**

Run: `pnpm --prefix frontend test:run src/hooks/tests/use-app-config.test.tsx`
Expected: FAIL (cannot resolve `@/hooks/use-app-config`).

- [ ] **Step 4: Implement the hook.** Create `frontend/src/hooks/use-app-config.ts`:

```ts
import { useEffect, useState } from "react";
import { getAppConfig } from "@/api/config";

export function useAppConfig() {
  const [githubOAuth, setGithubOAuth] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;
    getAppConfig()
      .then((cfg) => {
        if (active) setGithubOAuth(Boolean(cfg.github_oauth));
      })
      .catch(() => {
        if (active) setGithubOAuth(false); // fail-closed: hide the button on error
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  return { githubOAuth, loading };
}
```

- [ ] **Step 5: Run it to verify it passes.**

Run: `pnpm --prefix frontend test:run src/hooks/tests/use-app-config.test.tsx`
Expected: PASS (both cases).

- [ ] **Step 6: Commit.**

```bash
git add frontend/src/api/config.ts frontend/src/hooks/use-app-config.ts frontend/src/hooks/tests/use-app-config.test.tsx
git commit -m "feat(frontend): getAppConfig + useAppConfig (fail-closed gating)"
```

---

# Slice S2 — GitHub sign-in on login (core round-trip)

## Task 5: `auth-github.ts` — URL builder + callback completion (TDD)

**Files:**
- Create: `frontend/src/api/auth-github.ts`
- Test: `frontend/src/api/tests/auth-github.test.ts`

Design note: `completeGitHubOAuth` uses bare `axios` (not the shared `api` instance) so this pre-session bootstrap call bypasses the auth interceptors — same pattern as `api/auth-refresh.ts`. The callback returns `RefreshTokenResponse`, whose fields are `token` and **`refreshToken`** (camelCase — confirmed in the generated types and `auth-refresh.ts`).

- [ ] **Step 1: Write the failing test.** Create `frontend/src/api/tests/auth-github.test.ts`:

```ts
// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";
import axios from "axios";
import { githubOAuthUrl, completeGitHubOAuth } from "@/api/auth-github";

vi.mock("axios");

describe("githubOAuthUrl", () => {
  it("builds the base initiate URL", () => {
    expect(githubOAuthUrl()).toBe("/api/v1/auth/github");
  });
  it("appends invite_token when provided", () => {
    expect(githubOAuthUrl("abc123")).toBe("/api/v1/auth/github?invite_token=abc123");
  });
});

describe("completeGitHubOAuth", () => {
  beforeEach(() => vi.clearAllMocks());
  it("GETs the backend callback with code+state and returns the token data", async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { token: "t", refreshToken: "r" } });
    const result = await completeGitHubOAuth("code1", "state1");
    expect(axios.get).toHaveBeenCalledWith("/api/v1/auth/github/callback", {
      params: { code: "code1", state: "state1" },
    });
    expect(result).toEqual({ token: "t", refreshToken: "r" });
  });
});
```

- [ ] **Step 2: Run it to verify it fails.**

Run: `pnpm --prefix frontend test:run src/api/tests/auth-github.test.ts`
Expected: FAIL (cannot resolve `@/api/auth-github`).

- [ ] **Step 3: Implement.** Create `frontend/src/api/auth-github.ts`:

```ts
import axios from "axios";
import type { components } from "../api/types/openapi";

export type RefreshTokenResponse = components["schemas"]["RefreshTokenResponse"];

const API_BASE = import.meta.env.VITE_API_BASE_URL || "/api/v1";

export function githubOAuthUrl(inviteToken?: string): string {
  const base = `${API_BASE}/auth/github`;
  return inviteToken ? `${base}?invite_token=${encodeURIComponent(inviteToken)}` : base;
}

// Bare axios (not the shared `api` instance) so this pre-session call bypasses
// the auth interceptors — same rationale as api/auth-refresh.ts.
export async function completeGitHubOAuth(
  code: string,
  state: string,
): Promise<RefreshTokenResponse> {
  const res = await axios.get(`${API_BASE}/auth/github/callback`, { params: { code, state } });
  return res.data;
}
```

- [ ] **Step 4: Run it to verify it passes.**

Run: `pnpm --prefix frontend test:run src/api/tests/auth-github.test.ts`
Expected: PASS.

- [ ] **Step 5: Commit.**

```bash
git add frontend/src/api/auth-github.ts frontend/src/api/tests/auth-github.test.ts
git commit -m "feat(frontend): auth-github URL builder + callback completion"
```

## Task 6: `GitHubSignInButton` (TDD)

**Files:**
- Create: `frontend/src/components/auth/github-sign-in-button.tsx`
- Test: `frontend/src/components/auth/tests/github-sign-in-button.test.tsx`

Design note: the component self-gates via `useAppConfig` (returns `null` when disabled) and includes its own "or" divider, so each form needs only a one-line insertion. Uses the existing `ui/button` `inverse` variant (used by the login/signup submit buttons) and the lucide `Github` icon.

- [ ] **Step 1: Write the failing test.** Create `frontend/src/components/auth/tests/github-sign-in-button.test.tsx`:

```tsx
// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { GitHubSignInButton } from "@/components/auth/github-sign-in-button";
import * as useAppConfigModule from "@/hooks/use-app-config";

vi.mock("@/hooks/use-app-config");

describe("GitHubSignInButton", () => {
  beforeEach(() => vi.clearAllMocks());

  it("renders nothing when GitHub OAuth is disabled", () => {
    vi.mocked(useAppConfigModule.useAppConfig).mockReturnValue({ githubOAuth: false, loading: false });
    const { container } = render(<GitHubSignInButton />);
    expect(container).toBeEmptyDOMElement();
  });

  it("renders the button when enabled and navigates with invite_token on click", () => {
    vi.mocked(useAppConfigModule.useAppConfig).mockReturnValue({ githubOAuth: true, loading: false });
    const assignSpy = vi.spyOn(window.location, "assign").mockImplementation(() => {});
    render(<GitHubSignInButton inviteToken="inv1" />);
    const btn = screen.getByRole("button", { name: /github/i });
    btn.click();
    expect(assignSpy).toHaveBeenCalledWith("/api/v1/auth/github?invite_token=inv1");
  });
});
```

- [ ] **Step 2: Run it to verify it fails.**

Run: `pnpm --prefix frontend test:run src/components/auth/tests/github-sign-in-button.test.tsx`
Expected: FAIL (cannot resolve the component).

- [ ] **Step 3: Implement.** Create `frontend/src/components/auth/github-sign-in-button.tsx`:

```tsx
import { Github } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useAppConfig } from "@/hooks/use-app-config";
import { githubOAuthUrl } from "@/api/auth-github";

interface GitHubSignInButtonProps {
  inviteToken?: string;
}

export function GitHubSignInButton({ inviteToken }: GitHubSignInButtonProps) {
  const { githubOAuth } = useAppConfig();
  if (!githubOAuth) return null;

  return (
    <div>
      <div className="my-4 flex items-center gap-3">
        <div className="flex-1 border-t border-border" />
        <span className="text-xs text-muted-foreground">or</span>
        <div className="flex-1 border-t border-border" />
      </div>
      <Button
        type="button"
        variant="inverse"
        className="w-full"
        onClick={() => window.location.assign(githubOAuthUrl(inviteToken))}
      >
        <Github className="h-4 w-4" />
        Continue with GitHub
      </Button>
    </div>
  );
}
```

- [ ] **Step 4: Run it to verify it passes.**

Run: `pnpm --prefix frontend test:run src/components/auth/tests/github-sign-in-button.test.tsx`
Expected: PASS (both cases).

- [ ] **Step 5: Commit.**

```bash
git add frontend/src/components/auth/github-sign-in-button.tsx frontend/src/components/auth/tests/github-sign-in-button.test.tsx
git commit -m "feat(frontend): self-gating GitHubSignInButton with divider"
```

## Task 7: `GithubCallbackPage` + route (TDD)

**Files:**
- Create: `frontend/src/pages/auth/github-callback.tsx`
- Test: `frontend/src/pages/auth/tests/github-callback.test.tsx`
- Modify: `frontend/src/App.tsx`

Design note: completion order is store-tokens → fetch user → `setAuthSession` → context `refresh()` → navigate. Tokens are stored to `localStorage` first so `getCurrentUser()` (which goes through the `api` instance) carries the Bearer header. A `ranRef` guard prevents the effect's double-run under React StrictMode from firing two callback exchanges (the backend consumes `state` once).

- [ ] **Step 1: Write the failing test.** Create `frontend/src/pages/auth/tests/github-callback.test.tsx`:

```tsx
// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Routes, Route } from "react-router-dom";
import type { User } from "@/api/users";
import GithubCallbackPage from "@/pages/auth/github-callback";
import * as authGithub from "@/api/auth-github";
import * as usersApi from "@/api/users";
import * as common from "@/helpers/common";

const navigateMock = vi.fn();
vi.mock("react-router-dom", async (orig) => {
  const actual = await orig<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => navigateMock };
});
vi.mock("@/api/auth-github");
vi.mock("@/api/users");
vi.mock("@/hooks/use-current-user", () => ({
  useCurrentUser: () => ({ refresh: vi.fn().mockResolvedValue(undefined) }),
}));

function renderAt(url: string) {
  return render(
    <MemoryRouter initialEntries={[url]}>
      <Routes>
        <Route path="/auth/github/callback" element={<GithubCallbackPage />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("GithubCallbackPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  it("completes the round trip and navigates to the dashboard", async () => {
    vi.mocked(authGithub.completeGitHubOAuth).mockResolvedValue({ token: "tok", refreshToken: "ref" });
    vi.mocked(usersApi.getCurrentUser).mockResolvedValue({ id: "u1", email: "a@b.c" } as User);
    const setSession = vi.spyOn(common, "setAuthSession").mockImplementation(() => {});

    renderAt("/auth/github/callback?code=c1&state=s1");

    await waitFor(() => expect(navigateMock).toHaveBeenCalledWith("/dashboard", { replace: true }));
    expect(setSession).toHaveBeenCalledWith("tok", { id: "u1", email: "a@b.c" }, "ref");
  });

  it("shows an error and does not call the backend when code/state are missing", async () => {
    renderAt("/auth/github/callback");
    await screen.findByText(/missing authorization code or state/i);
    expect(authGithub.completeGitHubOAuth).not.toHaveBeenCalled();
  });

  it("shows the backend error message when the exchange fails", async () => {
    vi.mocked(authGithub.completeGitHubOAuth).mockRejectedValue(new Error("invalid state parameter"));
    renderAt("/auth/github/callback?code=c1&state=s1");
    await screen.findByText(/invalid state parameter/i);
  });
});
```

- [ ] **Step 2: Run it to verify it fails.**

Run: `pnpm --prefix frontend test:run src/pages/auth/tests/github-callback.test.tsx`
Expected: FAIL (cannot resolve the page).

- [ ] **Step 3: Implement the page.** Create `frontend/src/pages/auth/github-callback.tsx`:

```tsx
import { useEffect, useRef, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { completeGitHubOAuth } from "@/api/auth-github";
import { getCurrentUser } from "@/api/users";
import { setAuthSession } from "@/helpers/common";
import { useCurrentUser } from "@/hooks/use-current-user";
import { getErrorMessage } from "@/api/client";

export default function GithubCallbackPage() {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const { refresh } = useCurrentUser();
  const [error, setError] = useState<string | null>(null);
  const ranRef = useRef(false);

  useEffect(() => {
    if (ranRef.current) return; // StrictMode double-invoke guard: consume state once
    ranRef.current = true;

    const code = params.get("code");
    const state = params.get("state");
    if (!code || !state) {
      setError("Missing authorization code or state. Please try signing in again.");
      return;
    }

    (async () => {
      try {
        const { token, refreshToken } = await completeGitHubOAuth(code, state);
        if (!token) throw new Error("OAuth response missing access token");
        localStorage.setItem("authToken", token);
        if (refreshToken) localStorage.setItem("refreshToken", refreshToken);
        const user = await getCurrentUser();
        setAuthSession(token, user, refreshToken);
        await refresh();
        navigate("/dashboard", { replace: true });
      } catch (err) {
        setError(getErrorMessage(err));
      }
    })();
  }, [params, navigate, refresh]);

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 p-6 text-center">
      {error ? (
        <>
          <p className="rounded-sm border border-danger-border bg-danger-bg px-3 py-2 text-sm text-danger">
            {error}
          </p>
          <Link to="/sign-in" className="text-sm underline underline-offset-4">
            Back to sign in
          </Link>
        </>
      ) : (
        <p className="text-sm text-muted-foreground">Completing GitHub sign-in…</p>
      )}
    </div>
  );
}
```

- [ ] **Step 4: Run it to verify it passes.**

Run: `pnpm --prefix frontend test:run src/pages/auth/tests/github-callback.test.tsx`
Expected: PASS (all three cases).

- [ ] **Step 5: Register the public route.** In `frontend/src/App.tsx`, add the import and a route in the public section (alongside `/sign-in`, `/sign-up`, outside `RequireAuth`):

Import (with the other page imports near the top):
```tsx
import GithubCallbackPage from "@/pages/auth/github-callback"
```

Route (right after `<Route path="/sign-up" element={<Signup />} />`):
```tsx
      <Route path="/auth/github/callback" element={<GithubCallbackPage />} />
```

- [ ] **Step 6: Verify the app type-checks.**

Run: `pnpm --prefix frontend exec tsc -b`
Expected: no errors.

- [ ] **Step 7: Commit.**

```bash
git add frontend/src/pages/auth/github-callback.tsx frontend/src/pages/auth/tests/github-callback.test.tsx frontend/src/App.tsx
git commit -m "feat(frontend): GitHub OAuth callback page + public route"
```

## Task 8: Wire the button into the login form

**Files:**
- Modify: `frontend/src/pages/login/components/login-form.tsx`

- [ ] **Step 1: Add the import** (with the other imports at the top of `login-form.tsx`):

```tsx
import { GitHubSignInButton } from "@/components/auth/github-sign-in-button";
```

- [ ] **Step 2: Render the button** inside the form, immediately after the submit `<Button …>…</Button>` element:

```tsx
      <GitHubSignInButton />
```

- [ ] **Step 3: Verify type-check + existing login tests still pass.**

Run: `pnpm --prefix frontend exec tsc -b && pnpm --prefix frontend test:run src/pages/login`
Expected: no type errors; existing login tests PASS.

- [ ] **Step 4: Commit.**

```bash
git add frontend/src/pages/login/components/login-form.tsx
git commit -m "feat(frontend): Continue with GitHub on the login form"
```

---

# Slice S3 — Signup + invite

## Task 9: Wire the button into signup + invite-accept forms

**Files:**
- Modify: `frontend/src/pages/signup/components/signup-form.tsx`
- Modify: `frontend/src/pages/signup/components/invite-accept-form.tsx`

Design note: the signup form has no invite, so no token. The invite-accept form already receives the invite `token` as a prop — pass it so the OAuth signup joins the inviting org (the backend `/auth/github?invite_token=` path encrypts it into the OAuth state).

- [ ] **Step 1: Signup form — import** (top of `signup-form.tsx`):

```tsx
import { GitHubSignInButton } from "@/components/auth/github-sign-in-button";
```

- [ ] **Step 2: Signup form — render** immediately after the submit `<Button type="submit" …>…</Button>`:

```tsx
        <GitHubSignInButton />
```

- [ ] **Step 3: Invite-accept form — import** (top of `invite-accept-form.tsx`):

```tsx
import { GitHubSignInButton } from "@/components/auth/github-sign-in-button";
```

- [ ] **Step 4: Invite-accept form — render** immediately after the submit `<Button …>…</Button>`, passing the invite token (the component receives `token` via its `InviteAcceptFormProps`):

```tsx
        <GitHubSignInButton inviteToken={token} />
```

- [ ] **Step 5: Verify type-check + existing signup/invite tests still pass.**

Run: `pnpm --prefix frontend exec tsc -b && pnpm --prefix frontend test:run src/pages/signup`
Expected: no type errors; existing signup/invite tests PASS.

- [ ] **Step 6: Commit.**

```bash
git add frontend/src/pages/signup/components/signup-form.tsx frontend/src/pages/signup/components/invite-accept-form.tsx
git commit -m "feat(frontend): Continue with GitHub on signup + invite-accept"
```

---

# Final Verification

- [ ] **Backend tests + build.**

Run: `go test ./pkg/handlers/ -run TestConfigHandler_Get && go build ./...`
Expected: PASS, no build errors.

- [ ] **Frontend lint + types + full test suite.**

Run: `pnpm --prefix frontend lint && pnpm --prefix frontend exec tsc -b && pnpm --prefix frontend test:run`
Expected: clean lint, no type errors, all tests green.

- [ ] **Manual end-to-end (requires a GitHub OAuth app).**
  1. Set `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `GITHUB_REDIRECT_URI=http://localhost:5173/auth/github/callback` in `.env`; set the same callback URL in the GitHub OAuth app.
  2. `mage run` (API) + `pnpm --prefix frontend dev` (Vite at :5173).
  3. Visit `/sign-in` → confirm "Continue with GitHub" appears (it is hidden if the vars are unset — verify both states).
  4. Click it → GitHub consent → redirected to `/auth/github/callback` → lands on `/dashboard`, logged in.
  5. Repeat from an invite link (`/sign-up?invite_token=…`) → confirm the new user joins the inviting org.
  6. Force an error (tamper the `state` query param) → confirm the error message + "Back to sign in" link render.

- [ ] **Use the Playwright MCP** against `http://localhost:5173` to drive steps 3–6 and screenshot the button + callback states.

---

## Self-Review (completed during authoring)

- **Spec coverage:** S1 (config endpoint + gating + env docs) → Tasks 1–4; S2 (login button, url builder, callback, session) → Tasks 5–8; S3 (signup + invite + invite_token) → Task 9. All six condensed PRD stories covered.
- **Field-name correctness:** OAuth callback uses `RefreshTokenResponse { token, refreshToken }` (camelCase) — verified against generated types and `auth-refresh.ts`; the callback page and `completeGitHubOAuth` destructure `refreshToken`, not `refresh_token`.
- **Type consistency:** `getAppConfig` → `{ github_oauth }`; `useAppConfig` → `{ githubOAuth, loading }`; `GitHubSignInButton` consumes `useAppConfig` and `githubOAuthUrl`; callback consumes `completeGitHubOAuth` + `getCurrentUser` + `setAuthSession`. Names align across tasks.
- **No placeholders:** every code/test step contains complete content and an exact run command.
