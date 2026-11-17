---
owner: @Hamza-ali1223
backup: @SivanCola
status: active
reviewed: 2026-11-17
---

# Fork state

## Purpose

What this fork is like today: what it ships, what does not work on it, what work is open, and the commands that answer
whether `studio` is healthy. Change this file in the pull that changes one of those answers.

## 1. What ships

Every push to `studio` tags `fork-v<MAJOR>.<MINOR>.<PATCH>` and publishes a full release marked latest
(`.github/workflows/fork-auto-release.yml`). A push touching only `**.md`, `docs/**` or issue templates releases nothing.

| Artifact | Contents |
| --- | --- |
| `reasonix-windows-amd64.exe`, `reasonix-windows-arm64.exe` | CLI for Windows, cross-built with `CGO_ENABLED=0` |
| `ReasonixStudio-windows-amd64-installer.exe`, `ReasonixStudio-windows-amd64.zip` | Studio for Windows, installer and portable |
| `SHA256SUMS` | Digests over every `.exe` and `.zip` in the release |

1. macOS and Linux are built from source. No artifact is published for either.
2. The Windows binaries are unsigned (`.github/workflows/fork-release.yml`).
3. The bump is a patch by default; `release:minor` and `release:major` on the merged pull size it, and
   `release:ships-with-next-release` holds a small change for the next release instead of cutting its own.

## 2. Known broken

Both features that reach a release host still name upstream, so neither can find this fork's releases.

| Symptom | Cause | Source |
| --- | --- | --- |
| The in-app update check never offers a newer build | `StudioMirror = "https://dl.reasonix.io"` | `internal/platform/update/manifest.go:35` |
| `reasonix upgrade` lists upstream's releases, not `fork-v*` | `ghOwner = "esengine"` | `internal/frontend/cli/upgrade.go:33`, `internal/platform/releaseasset/cli.go:28` |
| The CLI install hint links upstream's release page | text only | `internal/base/i18n/messages_en.go:749`, `messages_zh.go:749`, `messages_zh_tw.go:672` |
| `/feedback` names upstream's repository as the public venue | text only | `internal/session/control/feedback_command.go:30` |

No test catches either failure: each one runs against a fake server or a stub catalog, so a check that reaches nothing
still passes.

## 3. Open work

| Kind | Items |
| --- | --- |
| Defect in shipped code | #22 — a write card's folder answer also exempts deletes inside it |
| Upstream ports | #7 to #18, one issue per `upstream/*` branch on `origin` |
| Porting hazard | Some upstream commits are already on `studio` under another hash; check the file and the symbol before porting |

## 4. How to check

```bash
gh run list --branch studio --limit 5
go test ./...
go run ./tools/repolint -changed
gh release list --limit 3
```

1. A red job on Linux or macOS is a real defect, not a platform quirk: the source is one codebase.
2. A green Windows job may mean the test skipped, because the OS jail tests skip there.
3. Merging into `studio` does not re-run an open pull request's checks; close and reopen it to pick up the new base.

## 5. Update trigger

Refresh this file when an artifact changes, a known-broken row is fixed or a new one appears, an open issue closes, or a
command in section 4 stops being the one that answers.
