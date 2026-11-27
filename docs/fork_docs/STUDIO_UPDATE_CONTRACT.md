---
owner: @Hamza-ali1223
backup: @SivanCola
status: active
reviewed: 2026-11-27
---

# Fork Studio Update Contract

## Scope

This contract governs how fork releases are discovered, described, verified, rolled back and shown in Reasonix Studio. It covers automatic updates for Windows amd64 NSIS installs and Linux amd64 installs owned by the `reasonix-studio` Debian package. Other install formats remain manual.

## Rules

| ID | Rule |
| --- | --- |
| U1 | The stable update line MUST use only strict `MAJOR.MINOR.PATCH` versions with matching `fork-vMAJOR.MINOR.PATCH` tags. Suffix and prerelease tags MUST NOT enter the update catalog or latest pointer. The catalog version MUST match its immutable release manifest. |
| U2 | Studio MUST read the fork-owned GitHub Releases catalog. Each catalog row MUST point to that release's immutable manifest and MUST NOT redirect updates to upstream. |
| U3 | Every in-app install start, including a user-clicked Install action, MUST be allowed only for NSIS-installed Windows amd64 and dpkg-owned `reasonix-studio` Linux amd64 installs. RPM, AppImage, Windows portable ZIP, other architectures and other install layouts MUST remain manual. Those installs MUST still be able to browse release notes and open the matching GitHub release page. |
| U4 | The release workflow MUST sign each auto-installable artifact with the fork's dedicated Minisign key. The private key and its password MUST be GitHub Actions secrets; the matching public key MUST be pinned by the updater and the privileged Linux verifier. SHA-256 checksums MUST remain an additional integrity check. |
| U5 | The catalog MUST retain every complete, signed, eligible, non-withdrawn release so a user can choose a prior version. Each row MUST keep its tag-pinned manifest; an update selected from the list MUST resolve to that version even after a newer release is published. Withdrawn releases MUST remain unavailable. Releases published before this contract takes effect are not rollback targets for the automatic updater. |
| U6 | Each eligible release MUST publish its generated Markdown release notes as UTF-8 plain text in an immutable `release-notes.txt` asset. The catalog's `notes` field MUST mark whether that asset exists. The host MUST derive the notes URL from the validated version and fixed fork repository; the Versions view MUST render those notes for the selected version, including a rollback target. Notes MUST be fetched from the fork release, never from the upstream mirror. |
| U7 | A release MUST become GitHub's latest release only after all required artifacts, signatures, manifest, notes and the complete catalog are present. Promotion MUST be serialized and require a version strictly newer than both the current latest release and the current eligible catalog. Published assets MUST NOT be replaced, including on a same-tag retry; an older tag MUST NOT move the latest pointer backwards. |
| U8 | Existing installations MUST complete one manual install of an updater-enabled fork build before automatic fork updates can work. The existing updater trusts the upstream key and catalog, so it cannot securely install the first release signed by the fork. The first updater catalog MAY bootstrap only from the known pre-updater release `fork-v4.0.2`. |
| U9 | Update modes MUST keep their current meaning: manual does no background fetch; notify reports discovery; download stages and verifies eligible releases; installing or restarting MUST require the user's action. |

## Enforcement

| Rule | Enforcement |
| --- | --- |
| U1–U2 | Catalog and manifest contract tests; immutable version-specific GitHub asset URLs; release workflow validation. |
| U3 | Platform and install-layout checks at the UI and install owner; unsupported installs MUST NOT stage or apply through any in-app path. |
| U4 | Embedded-key tests, signature verification tests, Linux helper verification and a publish job that fails on an absent or invalid signature. |
| U5 | Catalog generation preserves prior eligible rows and withdrawal records and resolves each requested version's manifest. |
| U6 | Release workflow generates and publishes notes; the host resolves the fixed fork tag URL and the Versions view renders latest and older release notes. |
| U7 | Serialized publish ordering and a latest-pointer guard reject incomplete releases, replaced assets and older-tag promotion. |
| U8 | Bootstrap instructions identify the required manual install before the first signed fork update. |
| U9 | Existing update-mode and restart-boundary tests remain in force. |
