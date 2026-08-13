---
name: stackdome-deploy
description: Deploy an application to Stackdome with the current CLI and return its public URL.
license: Apache-2.0
compatibility: Requires the Stackdome CLI, a repository, a Stackdome destination URL, and an API token.
metadata:
  author: Stackdome
---

# Deploy with Stackdome

Install the maintained skill:

```bash
npx skills add stackdome/skills
```

Follow [CLI workflows](/reference/cli-workflows) for authentication, Stackfile setup, deployment, checks, and failure handling.

Use the [Stackfile reference](/reference/stackfile) for configuration fields. Confirm every command and flag with the installed CLI help:

```bash
stackdome --help
stackdome <command> --help
```

Use API tokens, not account passwords. Use structured output for automation. Do not leave `--follow` or `--watch` running unattended.
