# How Casbin Works in Stackdome

Casbin is a policy engine. It answers one question: **"does this request match any policy?"**

Think of it as a database query: you have rows (policies), and for each incoming request, Casbin checks if any row matches. If any row matches, access is granted. If no row matches, access is denied.

## The Four Sections of a Casbin Model

### 1. Request Definition

```ini
[request_definition]
r = sub, org, resource, action
```

This defines the **shape of the question** you ask Casbin. Every time your code calls `enforce()`, you pass exactly these four values:

```go
enforcer.Enforce("user-123", "org-abc", "stacks/stack-777", "read")
//                 sub        org        resource              action
```

These are just parameter names — labels that the matcher references as `r.sub`, `r.org`, `r.resource`, `r.action`.

### 2. Policy Definition

```ini
[policy_definition]
p = sub, org, resource, action
```

This defines the **shape of each row in your policy table**. Policies are stored in the database and look like this:

```
p, viewer, *, stacks/*, read
p, editor, *, stacks/*, write
p, user-123, org-abc, stacks/stack-777, *
```

Each row is an **allow rule**: "this subject can perform this action on this resource in this org."

There is no deny. If no policy matches, access is denied by default.

The request definition and policy definition have the same field names here, but they don't have to be identical. The matcher decides how to compare request fields against policy fields.

### 3. Role Definition

```ini
[role_definition]
g = _, _, _    # user, role, org
```

This defines a **grouping table** that maps users to roles, scoped by org:

```
g, user-123, viewer, org-abc      # user-123 IS a viewer IN org-abc
g, user-456, editor, org-abc      # user-456 IS an editor IN org-abc
g, user-789, admin, org-abc       # user-789 IS an admin IN org-abc
```

The function `g(r.sub, p.sub, r.org)` in the matcher asks: "does the requesting user (`r.sub`) have the role named in the policy (`p.sub`) within the requested org (`r.org`)?"

Without roles, you'd need a separate policy row for every user. Roles let you write policies for `viewer` once and assign many users to it.

### 4. Matcher

```ini
[matchers]
m = (g(r.sub, p.sub, r.org) || r.sub == p.sub) \
    && (r.org == p.org || p.org == "*") \
    && keyMatch2(r.resource, p.resource) \
    && (r.action == p.action || p.action == "*")
```

This is the **comparison logic**. For each incoming request, Casbin iterates through every policy row and evaluates the matcher. If any row makes the matcher return `true`, access is granted.

A policy row matches when **ALL four conditions are true**:

```
Condition 1: IDENTITY
  g(r.sub, p.sub, r.org)    -> user has this role in this org (via grouping table)
  || r.sub == p.sub          -> OR policy is a direct grant to this specific user

Condition 2: ORG SCOPE
  r.org == p.org             -> policy is for this org
  || p.org == "*"            -> OR policy applies to all orgs

Condition 3: RESOURCE
  keyMatch2(r.resource, p.resource)  -> resource pattern matches
                                        "stacks/stack-777" matches "stacks/*"
                                        "stacks/stack-777" matches "stacks/stack-777"

Condition 4: ACTION
  r.action == p.action       -> action matches exactly
  || p.action == "*"         -> OR policy allows all actions
```

## Policy Types

### Role Policies — what each role can do

These define permissions for roles, not individual users. The `org = *` means "these permissions apply in whichever org the user holds this role":

```
p, viewer, *, stacks/*, read
p, viewer, *, stacks/*, list
p, viewer, *, secrets/*, read

p, editor, *, stacks/*, read
p, editor, *, stacks/*, write
p, editor, *, stacks/*, create
p, editor, *, stacks/*, logs

p, admin, *, *, *
```

### Grouping Policies — who has which role

These assign users to roles within specific orgs:

```
g, user-123, viewer, org-abc
g, user-456, editor, org-abc
g, user-789, admin, org-abc
```

### Resource-Level Grants — direct access to specific resources

These grant a specific user access to a specific resource, bypassing the role system:

```
p, user-123, org-abc, stacks/stack-777, *
p, user-456, org-abc, addons/postgres/pg-001, read
```

These are used for:
- **Ownership** — auto-created when a user creates a resource
- **Sharing** — admin grants a viewer write access to one specific stack
- **Scoped access** — CI token needs access to only certain resources

## Walkthrough Examples

### Setup

For all examples below, assume this state:

```
# Roles
g, user-123, viewer, org-abc
g, user-456, editor, org-abc
g, user-789, admin, org-abc

# Role permissions
p, viewer, *, stacks/*, read
p, viewer, *, stacks/*, list
p, viewer, *, secrets/*, read
p, viewer, *, secrets/*, list

p, editor, *, stacks/*, read
p, editor, *, stacks/*, list
p, editor, *, stacks/*, write
p, editor, *, stacks/*, create
p, editor, *, stacks/*, logs
p, editor, *, secrets/*, *
p, editor, *, addons/postgres/*, *

p, admin, *, *, *

# Direct grants (ownership)
p, user-456, org-abc, stacks/stack-new-1, *
```

---

### Example 1: Viewer reads a stack

```
enforce("user-123", "org-abc", "stacks/stack-999", "read")
```

Casbin checks each policy row:

```
Policy: p, viewer, *, stacks/*, read

  1. IDENTITY: g("user-123", "viewer", "org-abc")
     -> look up grouping table
     -> found: g, user-123, viewer, org-abc
     -> YES

  2. ORG SCOPE: "org-abc" == "*"
     -> YES (wildcard matches any org)

  3. RESOURCE: keyMatch2("stacks/stack-999", "stacks/*")
     -> YES (wildcard pattern match)

  4. ACTION: "read" == "read"
     -> YES

  All four conditions true -> MATCH -> ALLOWED
```

---

### Example 2: Viewer tries to delete a stack

```
enforce("user-123", "org-abc", "stacks/stack-999", "delete")
```

Casbin checks each policy row:

```
Policy: p, viewer, *, stacks/*, read
  -> ACTION: "delete" == "read" -> NO
  -> SKIP

Policy: p, viewer, *, stacks/*, list
  -> ACTION: "delete" == "list" -> NO
  -> SKIP

Policy: p, viewer, *, secrets/*, read
  -> RESOURCE: keyMatch2("stacks/stack-999", "secrets/*") -> NO
  -> SKIP

... (all remaining viewer policies fail on action or resource mismatch)

Policy: p, user-456, org-abc, stacks/stack-new-1, *
  -> IDENTITY: "user-123" == "user-456" -> NO
  -> SKIP

No policy matched -> DENIED
```

---

### Example 3: Viewer reads a stack in a different org

```
enforce("user-123", "org-xyz", "stacks/stack-001", "read")
```

```
Policy: p, viewer, *, stacks/*, read

  1. IDENTITY: g("user-123", "viewer", "org-xyz")
     -> look up grouping table
     -> no entry for user-123 as viewer in org-xyz
     -> NO

  Also check: "user-123" == "viewer" -> NO

  First condition fails -> SKIP

... (no policy matches because user-123 has no role in org-xyz)

No policy matched -> DENIED
```

The user is a viewer in `org-abc` but has no role in `org-xyz`. Org isolation is enforced by the grouping table.

---

### Example 4: Editor creates a stack

```
enforce("user-456", "org-abc", "stacks/", "create")
```

```
Policy: p, editor, *, stacks/*, create

  1. IDENTITY: g("user-456", "editor", "org-abc")
     -> found: g, user-456, editor, org-abc
     -> YES

  2. ORG SCOPE: "org-abc" == "*" -> YES

  3. RESOURCE: keyMatch2("stacks/", "stacks/*") -> YES

  4. ACTION: "create" == "create" -> YES

  MATCH -> ALLOWED
```

After creation, the system auto-creates an ownership grant:

```
p, user-456, org-abc, stacks/stack-new-1, *
```

---

### Example 5: Editor views logs of their own stack

```
enforce("user-456", "org-abc", "stacks/stack-new-1", "logs")
```

Two policies match — either one is sufficient:

```
Path 1 (via role):
  Policy: p, editor, *, stacks/*, logs
  -> g("user-456", "editor", "org-abc") -> YES
  -> keyMatch2("stacks/stack-new-1", "stacks/*") -> YES
  -> "logs" == "logs" -> YES
  -> MATCH

Path 2 (via ownership grant):
  Policy: p, user-456, org-abc, stacks/stack-new-1, *
  -> "user-456" == "user-456" -> YES (direct match)
  -> "org-abc" == "org-abc" -> YES
  -> keyMatch2("stacks/stack-new-1", "stacks/stack-new-1") -> YES
  -> "logs" == "*" -> YES
  -> MATCH
```

---

### Example 6: Viewer with a resource-level override

An admin decides to share `stack-777` with user-123 (who is only a viewer). A new policy is added:

```
p, user-123, org-abc, stacks/stack-777, *
```

Now user-123 can delete that specific stack:

```
enforce("user-123", "org-abc", "stacks/stack-777", "delete")

Policy: p, viewer, *, stacks/*, read -> action mismatch -> SKIP
Policy: p, viewer, *, stacks/*, list -> action mismatch -> SKIP

Policy: p, user-123, org-abc, stacks/stack-777, *
  -> "user-123" == "user-123" -> YES (direct grant)
  -> "org-abc" == "org-abc" -> YES
  -> keyMatch2("stacks/stack-777", "stacks/stack-777") -> YES
  -> "delete" == "*" -> YES
  -> MATCH -> ALLOWED
```

But they still can't delete a different stack:

```
enforce("user-123", "org-abc", "stacks/stack-888", "delete")

Policy: p, user-123, org-abc, stacks/stack-777, *
  -> RESOURCE: keyMatch2("stacks/stack-888", "stacks/stack-777") -> NO
  -> SKIP

No match -> DENIED
```

---

### Example 7: Admin does anything

```
enforce("user-789", "org-abc", "addons/postgres/pg-123", "delete")

Policy: p, admin, *, *, *
  -> g("user-789", "admin", "org-abc") -> YES
  -> "org-abc" == "*" -> YES
  -> keyMatch2("addons/postgres/pg-123", "*") -> YES (wildcard matches all)
  -> "delete" == "*" -> YES
  -> MATCH -> ALLOWED
```

The single `p, admin, *, *, *` policy grants admin access to all resources and all actions in any org where they hold the admin role.

---

### Example 8: API token with scoped access

API token scopes are NOT checked by Casbin. They are checked in the `PermissionService` before Casbin is called.

User-456 (editor) creates a token with scopes `["stacks:read", "stacks:logs"]`.

**Request via token: read a stack**

```
Step 1 (PermissionService - scope check):
  Token scopes: ["stacks:read", "stacks:logs"]
  Requested: resource="stacks", action="read"
  -> "stacks:read" in scopes -> PASS

Step 2 (Casbin - as user-456):
  enforce("user-456", "org-abc", "stacks/stack-new-1", "read")
  -> editor role grants stacks read -> MATCH -> ALLOWED
```

**Request via token: delete a stack**

```
Step 1 (PermissionService - scope check):
  Token scopes: ["stacks:read", "stacks:logs"]
  Requested: resource="stacks", action="delete"
  -> "stacks:delete" not in scopes, no wildcard -> DENIED
  -> Never reaches Casbin
```

**Request via token: exec into a stack**

```
Step 1 (PermissionService - scope check):
  Token scopes: ["stacks:read", "stacks:logs"]
  Requested: resource="stacks", action="exec"
  -> "stacks:exec" not in scopes -> DENIED
```

The token acts as a ceiling on the user's permissions. It can restrict but never expand access.

## Key Concepts Summary

| Concept | What it does | Example |
|---------|-------------|---------|
| Request (`r`) | The question being asked | "Can user-123 read stacks/stack-777 in org-abc?" |
| Policy (`p`) | An allow rule | "viewers can read stacks/*" |
| Grouping (`g`) | Maps users to roles per org | "user-123 is a viewer in org-abc" |
| Matcher (`m`) | Compares request against each policy | All four conditions must be true |
| `keyMatch2` | Pattern matching for resources | `stacks/*` matches `stacks/stack-777` |
| `g()` function | Role lookup in grouping table | `g("user-123", "viewer", "org-abc")` |
| Direct grant | Policy targeting a specific user | `p, user-123, org-abc, stacks/stack-777, *` |

## How Policies Get Created

| Event | Policy created | Type |
|-------|---------------|------|
| User joins org as viewer | `g, user-123, viewer, org-abc` | Grouping |
| User role changes to editor | Remove old grouping, add `g, user-123, editor, org-abc` | Grouping |
| User creates a stack | `p, user-123, org-abc, stacks/stack-new, *` | Direct grant |
| Admin shares a resource | `p, user-456, org-abc, stacks/stack-777, read` | Direct grant |
| Resource is deleted | Remove all policies matching `stacks/stack-id` | Cleanup |
| User removed from org | Remove grouping `g, user-123, *, org-abc` | Cleanup |

## Performance

Casbin loads all policies into memory and caches enforcement results using `SyncedCachedEnforcer`. It does not query the database on each `Enforce()` call. The linear scan through in-memory policies handles tens of thousands of policies in sub-millisecond time.

For larger scale (100K+ policies), use filtered policy loading to keep the in-memory set scoped to the relevant org:

```go
enforcer.LoadFilteredPolicy(&gormadapter.Filter{
    V1: []string{orgID, "*"},
})
```
