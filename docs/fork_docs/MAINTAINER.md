---
owner: @Hamza-ali1223
backup: @SivanCola
status: active
reviewed: 2026-11-17
---

# Maintainer profile

## Purpose

The compass this fork is maintained by: the rules that decide what merges, what is refused, and what a review looks for.
Read it before proposing a change. A change that contradicts a rule here needs that rule amended first, in the same
conversation, rather than contradicted quietly in the diff.

## 1. Decision rules

1. A tool result MUST be unambiguous. Where an earlier command in one call can fail while the process still exits `0`,
   the reported result may mean the opposite of what happened, so the host refuses that shape and asks for the work split
   into separate calls. One extra tool call is cheap; an ambiguous result is not.
2. A judgement MUST read structure — a shell AST, a type, a tool schema — never wording. A phrase table misfires on real
   input, and a misfiring judgement inside a gate is worse than no gate. Security allow-lists are the exception and stay.
3. A failure MUST carry an identity: a sentinel error inside Go, a dotted code across HTTP, a typed reason on a card.
   A message is where the knowledge of *why* goes to die.
4. Whoever first knows why an operation failed owns its class. A transport status is not a domain identity, and a display
   string is not a classification.
5. Ask only for what is dangerous. Recursive shell deletes and the dedicated file tools raise a card; protected targets
   are refused unasked; `mv`, `cp`, non-recursive `rm` and scripts stay uncarded by decision, not by oversight.
6. A metric is its definition, not its name. Before a number becomes a finding, state what it counts, from which anchor,
   over which population, and what it excludes.
7. A constructed state is not a reachable one. Name who in production reaches a state, through which transitions, before
   a test that assigns a private field is reported as a defect.
8. Fix at the owning layer, with the smallest change that removes the duplicate path rather than adding a flag beside it.
9. Never widen a gate baseline to land a change. Fix the code; a ratchet only goes down, and lowering one is asked for
   explicitly and justified.
10. Change the contract first, then the code. The documentation edit lands before the behaviour it describes.

## 2. Refused

| Proposal | Why it is refused |
| --- | --- |
| Dropping the shell-shape gates so no command is refused for its shape | It buys one fewer tool call with an ambiguous result |
| Deriving a failure class from a message, a status code, or an external server's prose | Domain, internal and dependency failures need three different next steps |
| Declaring a capability missing because a search found nothing | Say what is missing; a guessed substitute is worse than an absent one |
| Restating a value another owner already holds — a vendor rate, a config default | A table written twice drifts |
| Adding a `bool` to bound one more case | Independent flags multiply into states no type records as legal; group by lifetime |
| A check that ran after the fact standing in for observation | Observation proves what an action did; verification proves what the result satisfies |

## 3. Commits and pull requests

| Rule | Detail |
| --- | --- |
| Merge style | Rebase and merge: every commit lands as written, so a tidy branch is the deliverable |
| One concern | One concern per commit, conventional prefix, a body that says why |
| Review feedback | Amend rather than stack; one force-push per round; never a commit named for review repair |
| Type line | One or more of Feature, Bug fix, Clean-up, Security; Security is never folded into another kind |
| Body | Cause, Blast radius, Neighbouring behaviours tested, Why this layer |
| Release note | `release-notes/studio/unreleased/<pull number>-<slug>.md`, bilingual |
| Diff | Only the files the purpose needs; no stray changes from another branch |

## 4. Review priorities

| Order | Ask |
| --- | --- |
| 1 | Does the boundary hold — permissions, sandbox, approval, wire contracts? |
| 2 | Do the tests fail without the change, and does the neighbourhood still pass? |
| 3 | Does the evidence claim what was verified, and nothing more? |
| 4 | Is the diff the smallest one that fixes the owning layer? |
| 5 | Did the doc or contract change land before the code it describes? |
