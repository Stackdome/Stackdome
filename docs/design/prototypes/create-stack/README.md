# Create-a-stack — interactive prototype

The agreed design for the create-new-stack flow, as a clickable prototype.
**This is the reference the Figma board and the code are both built from.**

| | |
|---|---|
| **Live** | https://claude.ai/code/artifact/6d649912-8686-4915-b5e9-501eeb821e40 |
| **Audit that produced it** | https://claude.ai/code/artifact/9b332f32-91a3-4b06-aa7f-fe1ec2f61287 |
| **Figma plan** | `docs/design/create-stack-figma-plan.md` |
| **Rules it obeys** | `DESIGN-PRODUCT.md` |
| Built | August 2026, branch `graphite-pass-2` |

## Running it locally

```bash
node docs/design/prototypes/create-stack/build.mjs   # writes prototype.html
open docs/design/prototypes/create-stack/prototype.html
```

`prototype.source.html` carries `__GEIST__` / `__MONO__` placeholders; the build
inlines the two product typefaces from `pkg/web/dist/assets/` so the prototype is
set in exactly the faces the product ships. **`prototype.html` is generated — do
not edit it, and do not commit it** (178 KB of base64).

## What it settles

| Decision | |
|---|---|
| **A page, not a dialog** | Lives at `/stacks/new` inside the app shell |
| **Five peer starting points** | Repository · ready-made app · compose file · building blocks · blank canvas. All visible at once, one always active |
| **Cards on top, work below** | Options are actions, not navigation. A chooser rail put 476px — 37% of the width — of nav-shaped list beside the 240px sidebar |
| **The flow ends on the canvas** | `Create stack` makes a **draft**. Deploy happens on the canvas, because blocking errors have to be seen first |
| **Blank is the odd one out** | Dashed edge, and it explains what an empty canvas gives you rather than jumping there |
| **Config belongs to a resource** | Branch, port, Dockerfile path and build context live in the node inspector, not a form step |

## Things it is deliberately careful about

- **No orange anywhere.** §7 — orange belongs to the website and to thresholds.
- **No uppercase labels.** §11 — including the block registry's own `SERVICES` / `DATABASES`.
- **One filled button per screen**, and `Deploy` is the only pill.
- **Empty states carry no box.**
- **Push-to-deploy is never promised** — `pkg/services/github_webhook_service.go`
  drops push events (*"push (and everything else) is accepted and dropped for
  now"*). The flow claims per-PR preview environments instead, which are real.

## Functionality parity with today's wizard

Checked against `frontend/src/pages/stacks/components/wizard/` and its data:

- 12 blocks in 3 categories, and the **same block can be added repeatedly**
  (`addBlockToStack` de-duplicates with `uniqueName()` → `postgres`, `postgres-2`)
- Managed add-ons can be linked, with their own empty state
- Templates show every field the record carries — badge, category, version,
  long description, website and docs
- Git offers both **Connected provider** and **Public URL**
- Compose supports file upload as well as paste, and previews the parse

## Scope note — the canvas

The prototype demonstrates the whole flow **through to a deployed stack**, but
the canvas half — nodes, wires, the node inspector, deploy — is **a separate
project** and is not being taken to Figma with the starting points. The
illustration node in `glyph / architecture` is a visual device for empty states
only; it is not the basis for a canvas UI component.

## Layout options still open

The chrome carries a **Layout** switcher with four variants of the starting-point
chooser — `Cards on top` (chosen), `Left rail`, `Gallery`, `Tabs`. Kept so the
comparison is reproducible; only `Cards on top` needs to reach Figma.
