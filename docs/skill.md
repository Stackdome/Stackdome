---
name: stackdome-deploy
description: Deploy a repository to Stackdome with the CLI, then verify the released application and return its public URL.
license: Apache-2.0
compatibility: Requires the Stackdome CLI, a repository, a Stackdome destination URL, and an API token.
metadata:
  author: Stackdome
---

# Deploy with Stackdome

Install the maintained deploy skill before starting:

```bash
npx skills add stackdome/skills
```

Use the [AI-agent workflow](/guides/ai-agents) for the complete operating contract and the [Stackfile reference](/reference/stackfile) for configuration grammar.

## Deployment workflow

1. Inspect the repository for its build, startup command, ports, environment variables, dependencies, and storage needs.
2. Run `stackdome init`; edit the generated `stackfile.yaml` from repository evidence.
3. Run `stackdome validate` after every edit and correct all reported errors.
4. Authenticate only with an API token: run `stackdome login --url https://<your-stackdome-host> --token <api-token>`, then confirm the active host and scope with `stackdome whoami -o json`.
5. Deploy with `stackdome deploy --wait -o json`. Retain the non-empty `release.id` as the deployed release ID. Treat any non-zero exit or `release.state` other than `Released` as a failed deployment.
6. Run `stackdome status -o json`. Report that deployed release as serving only when `converged_release.id` equals the deployed release ID, `converged_release.state` is `Released`, and `converged_release.health` is `ok`. Claim it is still the newest release only when `latest_release.id` also equals the deployed release ID and `latest_release.state` is `Released`. For a public service, run `stackdome open -o json` and return a URL from `urls`.

Use `stackdome --help` and `stackdome <command> --help` to verify commands and flags. Use `-o json` for automation. Never request the user's account password.

## Failure routing

| Failure stage | First command |
| --- | --- |
| Stackfile validation | `stackdome validate` |
| Image build | `stackdome build list`, then `stackdome build info <id>` and `stackdome build logs <id>` |
| Release progression | `stackdome release events <id> --follow` |
| Runtime readiness | `stackdome status -o json`; use `stackdome status --conditions` for condition history |
| Application runtime | `stackdome logs <resource> --tail 100` |

Use the retained deployed release ID for status equality checks and release events, and use the build ID from `stackdome build list` for build inspection.
