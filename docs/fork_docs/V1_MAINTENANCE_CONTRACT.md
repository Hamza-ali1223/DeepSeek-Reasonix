---
owner: @Hamza-ali1223
backup: @SivanCola
status: active
reviewed: 2026-11-27
---

# 1.x maintenance boundary

## Scope

This contract defines which Reasonix version line this fork maintains and what a maintainer MUST restore before enabling 1.x monitoring in this repository.

## Rules

| ID | Rule |
| --- | --- |
| V1 | This fork MUST maintain the `studio` branch and its Reasonix 2.x line. |
| V2 | This fork MUST NOT schedule or dispatch 1.x monitoring while it does not maintain a forked `main-v2` branch. |
| V3 | Before 1.x monitoring is enabled, the fork MUST contain the `main-v2` branch and both target workflows, `app-memory.yml` and `diagnostic-overhead.yml`, on its default branch. |
| V4 | A maintainer restoring 1.x monitoring MUST review each workflow's permissions, required secrets, and targets before enabling its schedule. |
| V5 | The README MUST identify 1.x as upstream-maintained and link to this contract while this fork's support remains limited to 2.x. |

## Enforcement

1. Fetch the upstream 1.x branch and publish it to this fork:

   ```bash
   git fetch https://github.com/esengine/DeepSeek-Reasonix.git main-v2
   git push origin FETCH_HEAD:refs/heads/main-v2
   ```

2. In a pull request to `studio`, restore or port `.github/workflows/app-memory.yml` and `.github/workflows/diagnostic-overhead.yml` from `main-v2`. Both workflow files MUST also exist on the default branch so GitHub can dispatch them.
3. Review the restored workflows' permissions, secrets, and branch targets. Restore the scheduler only after those requirements are met.
4. After the pull request merges, verify each dispatch against this fork and the `main-v2` ref:

   ```bash
   gh workflow run app-memory.yml --repo Hamza-ali1223/DeepSeek-Reasonix --ref main-v2
   gh workflow run diagnostic-overhead.yml --repo Hamza-ali1223/DeepSeek-Reasonix --ref main-v2
   ```

5. Confirm both runs start successfully before restoring their scheduled triggers.
