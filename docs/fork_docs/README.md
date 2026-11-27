---
owner: @Hamza-ali1223
backup: @SivanCola
status: active
reviewed: 2026-11-27
---

# Fork documentation

## Purpose

Documents that are true of this fork alone. Upstream's `docs/` describes the shared product; a fact that holds only here
belongs in this folder, so the shared tree stays portable back upstream.

`docs/DOCS_STANDARD.md` governs every file below `docs/`, this folder included. Read it before adding a file here.

## Files

| File | Answers |
| --- | --- |
| `FORK_STATE.md` | What this fork ships, what is broken on it, what work is open |
| `STUDIO_UPDATE_CONTRACT.md` | Which fork releases Studio may discover, verify, install, roll back to and show notes for |
| `MAINTAINER.md` | How its maintainer decides, what is refused, what a review looks for |
| `V1_MAINTENANCE_CONTRACT.md` | Which version line this fork maintains, and what must be restored before enabling 1.x monitoring |
| `agent+planner-mode-spec.md` | What explicit planner routing, consultation, and role settings must do |
| `PLANNER_RESEARCH_BUDGET_CONTRACT.md` | What bounds a planning turn's research, what the planner is told, and what its last round may still do |
| `PLANNER_CONSULTATION_LIMITS_CONTRACT.md` | What bounds one consultation and one turn's consultation count |
| `PLANNER_APPROVAL_MODE_SWITCH_CONTRACT.md` | What the work mode becomes once a plan is accepted, and when the card is always shown |
| `PLANNER_CACHE_HIT_CONTRACT.md` | What Studio's cache-hit reading counts, and what a low reading does not establish |

## Rules a new file here is most likely to break

| Rule | Detail | Source |
| --- | --- | --- |
| Header | First line `---`; keys `owner`, `backup`, `status`, `reviewed`, in that order | `DOCS_STANDARD.md` §3 |
| Owner pair | MUST equal the first and second handle of the matching `.github/CODEOWNERS` rule | H1, `tools/repolint` `doc-owner` |
| Status | `active` or `deprecated` only; a deprecated file links its replacement on the first line | H2 |
| Reviewed | `YYYY-MM-DD`, the last date the owner confirmed the file matches the code | H3 |
| Type | A document opens with `Purpose`; a contract has `Scope`, `Rules`, `Enforcement`; a runbook has `Owners`, `Preconditions`, `Steps`, `Verification`, `Recovery` | `DOCS_STANDARD.md` §5 |
| Shape | Numbered rules, ordered steps or tables — not prose paragraphs | W1 |
| Width | A paragraph or list item MUST stay within 320 display columns; a CJK character counts as two | W2 |
| Verbs | MUST, SHOULD and MAY for requirements; anything else reads as information | W3 |
| Tense | Current behaviour only. History and rationale belong in the commit message | W4 |
| Commands | Fenced code block, copyable as-is | W5 |
| Language | English. Only `README.zh-CN.md`, `docs/GUIDE.zh-CN.md` and `docs/CLI.zh-CN.md` keep a Chinese copy | L1, L2 |
| Removal | Delete a document in the change that removes what it describes; no `archive/` directory | C1, C2 |

## Checking a change

`make` is absent on this machine, so run the recipe directly:

```bash
go run ./tools/repolint -changed
```

`-changed` reports only what the worktree altered or added. A bare `go run ./tools/repolint` reports the whole tree,
including debt the change never touched.
