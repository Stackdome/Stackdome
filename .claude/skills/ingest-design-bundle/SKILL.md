---
name: ingest-design-bundle
description: Fetch a Claude Design share URL (api.anthropic.com/v1/design/h/...), unpack the handoff bundle, and emit a Design-reference artifact for brainstorming to fold into a spec. Use when the user pastes a Claude Design link or asks to fetch/implement a Claude design.
---

> Spine note: superpowers is the workflow spine. `ingest-design-bundle` is inbound design intake — it feeds `superpowers:brainstorming` and gates nothing.

# Ingest Design Bundle

## Overview

A Claude Design share URL points at a **gzip tar handoff bundle**, not a webpage. This skill fetches it through the only fetch path available in this repo (the context-mode sandbox), unpacks it, reads it in the order the bundle's own README dictates, reconciles its tokens against the live brand source, and emits a single Design-reference artifact for downstream brainstorming.

**Core principle:** The bundle is the source of truth. Read it, never render it.

## When to Use

Trigger on either:
- A URL matching `api.anthropic.com/v1/design/h/` (with or without `https://`).
- A request to fetch or implement a Claude design (e.g. "implement this Claude design", "pull the design Claude made").

Do NOT use for Figma links, screenshots, or generic design references — only Claude Design share bundles.

## No Spine Precondition

This is **inbound intake**. It does NOT require a spec, plan, or any superpowers spine step to run first. It produces an input that `superpowers:brainstorming` later consumes (see Composition).

## Fetch — Sandbox Only

`curl` and `wget` are intercepted/blocked in this repo. `WebFetch` is denied. The **only** fetch path is the context-mode sandbox.

The endpoint returns binary: content-type `application/gzip`, roughly hundreds of KB, a gzip-compressed tar archive. Do not try to read it as text or as HTML.

Recipe:

**Single sandbox context.** The fetch+write AND the `tar` extraction MUST run in the SAME context-mode sandbox invocation. The sandbox filesystem is not the host filesystem — if you write the buffer in `ctx_execute` and then run `tar` in host Bash, the temp file does not exist there and extraction fails with a cryptic "not found". Keep every step below inside one `ctx_execute` call (or one sandbox shell invocation).

1. Fetch the bytes and write them to a temp file with `ctx_execute` (javascript):

```javascript
const url = "https://api.anthropic.com/v1/design/h/<id>"; // the pasted share URL
const res = await fetch(url);
const buf = Buffer.from(await res.arrayBuffer());
const fs = require("fs");
const out = "/tmp/design-bundle.tar.gz";
fs.writeFileSync(out, buf);
console.log(`status=${res.status} bytes=${buf.length} ct=${res.headers.get("content-type")} -> ${out}`);
```

   Expect `status=200`, a content-type of `application/gzip`, and a non-trivial byte count. If you get HTML or a tiny body, the URL is wrong or expired — stop and ask the user for a fresh link.

2. **Integrity guard — verify BEFORE extracting.** In the same sandbox, assert the byte length is non-trivial and that the archive is a valid gzip tar by *listing* it first. `tar -tzf` must succeed and print entries:

```bash
test "$(stat -c%s /tmp/design-bundle.tar.gz 2>/dev/null || stat -f%z /tmp/design-bundle.tar.gz)" -gt 1024 || { echo "CORRUPT: body too small"; exit 1; }
tar -tzf /tmp/design-bundle.tar.gz | head -n 20
```

   If `tar -tzf` errors or lists zero entries, the download is corrupt (commonly a silently-mangled binary body from a sandbox fetch polyfill). STOP — report a corrupt download and do NOT proceed to extract or read any files.

3. Only after the listing succeeds, unpack — in the same sandbox:

```bash
mkdir -p /tmp/design-bundle && tar -xzf /tmp/design-bundle.tar.gz -C /tmp/design-bundle
```

4. List the tree so you know what the bundle ships (use the sandbox if the listing is long).

## Read Order — Honor the Bundle's README

The bundle ships its own `README.md` with instructions. Follow them. The standard order is:

1. **`README.md`** — first, always. It states how the bundle wants to be consumed and the constraints below.
2. **`chats/*.md`** — the design conversation: design intent and, critically, *where the user landed* (the last accepted direction overrides earlier exploration).
3. **The primary `project/` file the chat last iterated on** — then follow its imports, plus `tokens.css` and `acceptance.jsx`.

Distill intent from the chats, not from your own assumptions about the visuals.

## Hard Rule — No Screenshots, No Rendering

Do **NOT** screenshot the bundle, open it in a browser, or render any component to "see" it. The bundle README forbids it. Every dimension, color, spacing, and layout decision is already explicit in the source files (`tokens.css`, the `project/` JSX/CSS, `acceptance.jsx`). Read the source; reading is authoritative, rendering is not.

## Brand Reconciliation

`tokens.css` in the bundle is typically a copy of `frontend/src/index.css` at design time. Before treating any token as new:

- Diff bundle `tokens.css` against the live `frontend/src/index.css`.
- Reuse existing brand tokens by name. Map bundle values onto the live tokens.
- Do **NOT** fork the token set, and do **NOT** introduce raw hex/rgb values. Anything that looks new is usually an existing token under a different name or a drift to flag, not a new color to hardcode.

## Output Artifact

Write exactly one file:

```
dev-docs/design-refs/YYYY-MM-DD-<topic>-design-reference.md
```

(`<topic>` = short kebab slug from the design subject; `YYYY-MM-DD` = today.) It MUST contain:

1. **Component inventory** — every component/screen in the bundle, with its source path.
2. **Token mapping** — a table mapping bundle `tokens.css` ↔ `frontend/src/index.css` (token name, bundle value, live token, match/drift/new).
3. **Intent summary** — distilled from `chats/*.md`, ending with where the user landed.
4. **Acceptance criteria** — extracted from `acceptance.jsx`.
5. **Source-file map** — bundle path → what it defines, so brainstorming can cite it.

Do not paste raw bundle contents into chat — write the artifact and report its path.

## Composition

This skill has no spine precondition and gates nothing. Its output is an input: `superpowers:brainstorming` consumes the Design-reference artifact and folds a **"Design reference"** section into the resulting spec. State the artifact path explicitly when handing off so brainstorming can pick it up.

## Common Mistakes

- Treating the URL as a webpage / passing it to WebFetch → blocked. It is a gzip tar; use the sandbox fetch recipe.
- Rendering or screenshotting a component → forbidden by the bundle README; read the source instead.
- Inventing new colors because `tokens.css` "looks different" → diff against `frontend/src/index.css`; reuse existing tokens.
- Skipping the chats → you lose where the user actually landed and implement an abandoned direction.
- Dumping bundle files into chat → write the single artifact and report its path only.
