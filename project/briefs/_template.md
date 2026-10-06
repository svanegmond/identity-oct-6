---
description: "Required. One sentence: what this brief contracts for."
date: YYYY-MM-DD
ref: TASK-ID
---

# Brief: [Task Name]

First durable write is a land-ready contract (`rules/roles/steward-elicitation.mdc` § Land-ready draft): template sections filled, Binary Forks or premise-forcing `N/A`, Parade scaffolded at `{artifact_root}/demos/{ref}-{brief_stem}.md`, brief file under `{artifact_root}`. A Problem/Approach stub is only for an open elicitation hop. Pre-flight lives in `roles/steward` § Pre-Flight Checklist — do not duplicate it here.

## Problem

What's broken or missing.

## Approach

System behavior after work—not file edits. Implementor picks those. Outcomes to verify, not steps. Two shapes:

**Default shape** (single coherent outcome): flat bullets, each beginning "After this work, …".

**Sequenced shape** (work spans distantly-related surfaces — modules, packages, methodology surfaces that read independently — that still share Implementor context: same model fit, same architectural decisions, same review window): h3 phase sub-sections, each phase producing one self-contained outcome that begins "After this phase, …". Each phase lands in its own commit; all phases ship within a single `implementor_transition` cycle. Inline snippet:

````
### Phase A — <phase name>

After this phase, <outcome>.

### Phase B — <phase name>

After this phase, <outcome>.
````

Recover from `git log -- project/briefs/` for sequenced-shape examples, or reach for `## Task Graph` (external decomposition) only when sequenced phases won't fit — see `rules/roles/steward-pipeline.mdc` § Brief size sanity.

## Seams

The contract boundaries this work is built and tested at. A **seam** *(Michael Feathers)* is a place where you can alter behavior without editing in that place — the location at which a module's interface lives, and the public boundary you test at. Where to put the seam is its own design decision, distinct from what goes behind it.

Steward declares these in Planning; the Implementor tests at them and escalates rather than introducing one; Implementation Review checks the work landed there. Guidance for *choosing* seams — deep narrow seams, `domain → persistence port → wire projection` — stays in `rules/engineering-preferences.mdc`. This section records the choice, not the argument for the preference.

- **Existing seam used**: [`module` / interface — what crosses it, why it already fits]
- **New seam introduced**: [N/A | `module` / interface — what crosses it, what it separates, and what second adapter justifies it. One adapter means a hypothetical seam; two adapters means a real one. Don't introduce a seam unless something actually varies across it.]
- **Not a seam here**: [Optional. Boundaries deliberately not drawn, so the Implementor doesn't invent them.]

*Omit only when the work introduces no new code path — a prose- or docs-only brief.*

## Task Graph (when sequenced phases won't fit)

Escape valve for when sequenced phases (see § Approach) would share too little Implementor context to land in one cycle — different parts of the system, different model fits, different review windows, or one phase blocking on external evidence (spike, decision, dependency). Coordinates child briefs as a DAG; each child task ships under its own Implementation Review cycle. Omit when § Approach (flat bullets or sequenced phases) covers the work.

Each task names:
- **Outcome**: [What becomes true after this task]
- **Boundary**: [Owned contract or artifact boundary]
- **Depends on**: [Real prerequisite tasks/decisions/artifacts]
- **Proof**: [What evidence closes this task]
- **Non-goals**: [What this task must not absorb]

## Artifacts

Project artifacts, decisions, and disciplines this brief reads or updates. Steward settles all artifact questions before plan-review.

- **Consulted**: [N/A | `path` — what it contributes]
- **Create/update**: [N/A | `path` — what changes and why]

## Context Payload

Fully serializable, stateless payload for cold-booted agent.

- **Target Files**: [Optional. Name files only when Steward has architectural guidance on where to work — e.g. "touch only the MCP layer, not the domain model."]
- **Required Context**: [Reference artifacts, related tasks, snippets, existing code/modules providing insight or examples]
- **Discussion Decisions**: [Decisions implementor/reviewer must not rediscover]
- **Dependency / Technology Decisions**: [Approved choices, rejected alternatives, why]
- **External References**: [Docs, links, external APIs, SDKs, deps, constraints]

## Demonstration Plan

Where **VCs** land. The brief's proof contract is **AC + VC** — that *is* the test plan. ACs are proven by the named automated tests below. VCs are proven here, live: stand the thing up, drive it the way a user or caller would, and capture what came back.

Pick the medium the surface actually has:

| Surface | Live demonstration |
|---|---|
| Web UI | Screenshot the rendered page; record the scripted browser interaction |
| API / RPC | Call-and-response against a running server — request and response body |
| CLI | Terminal session: invocation and output |
| Data / migration | Before and after state read out of the real store |
| Job / pipeline / daemon | Trigger it; capture logs, emitted artifacts, health |
| Deployed environment | Reach the deployed URL or cluster and show it answering |

- **Proof Parade path:** `$ARTIFACT_DIR/demos/[task-ref]-[brief-file-stem].md`
- [ ] [For each VC: how the situation gets constructed, what gets driven, and the exhibit slot it fills. See `roles/implementor § Proof Parade`.]

## Test Plan

Named automated tests that **back the ACs**. Type, what tested, expected behavior. This list is not a third proof channel and does not prove a VC.

**Stable IDs:** Reference as TP-1, TP-2, … in MCP tool arguments (e.g. `implementor_escalate` optional `test`). When § Approach uses sequenced phases, group entries under matching `### Phase A` / `### Phase B` sub-headings; TP-N numbering continues across phase sub-headings and does not restart per phase.

- [ ] **TP-1** [Type]: [Function/component] [expected behavior]

## Acceptance Criteria

Static / unit-tested contract. A function that accepts only a named parameter, a validation rule, a precedence order, an error type, a schema field, an absent code path — those are ACs. They do not become VCs because they matter. Proof is a named test (TP-*). A passing test is the observation for an AC; it is not the observation for a VC.

Binary, falsifiable checkboxes. All met = in-process contract fidelity.

**Attestation hygiene:** Keep every AC box **`[ ]`** through Planning and Plan Review. Only Implementor marks met (`[x]`) in the handoff cycle after work is done — MCP rejects submit when ACs are pre-ticked.

**Stable IDs:** Reference as AC-1, AC-2, … in MCP tools (`plan_review_pass` / `plan_review_block` / `implementation_review_pass` / `implementation_review_block`, `implementor_escalate` use `criterion` as AC-*). When § Approach uses sequenced phases, group ACs under matching `### Phase A` / `### Phase B` sub-headings — numbering continues across phase sub-headings; AC-N / TP-N / VC-N do not restart per phase, and IDs are not phase-prefixed, so stable IDs stay stable across MCP tool consumers.

- [ ] **AC-N** Proof Parade demonstrates each new boundary/seam with inspectable evidence.

## Validation Criteria

E2E, live-demonstrated contract. If the claim's truth lives outside the process — a rendered page, a running server, a real filesystem or database, a deployed environment, a message bus, a third-party integration, the machine's own git or clock — it is a VC. Construct the situation and try the thing. Captured evidence is the request and its response, the screenshot, the recording, the terminal session, the stored row. A test name in a VC cell is not proof, because the test replaced the very thing in question with a stub.

Agent-verifiable by default; single-surface VCs are driven live at Implementation Review; integration-shaped VCs are Steward-orchestrated before Validating. The Lead's walk exhibits already-driven criteria — not first-run verification. Use VC-* IDs. Every briefed task routes through Validating after Implementation Review.

*Omit only when the work has no surface outside the process. Do not omit a VC and then "prove" the live claim with a unit test.*

- [ ] **VC-N** Shape: single-surface (default) | integration-shaped | human-judgment. Medium: page screenshot | scripted browser-interaction recording | request/response transcript against a running server | terminal session | before/after state read from the real store | logs or emitted artifacts from a triggered job | equivalent inspectable capture. [What the Lead should see/feel] → [where it lands in Proof Parade]. When shape is integration-shaped, name the deploy or live data-plane crossing. When shape is human-judgment, name the human decision.

## Tracking

- **Task ref**: [e.g., TEAM-42 — task in system of record. Created during /plan.]
- **Task body (tracker)**: pointer only — ≤5 lines naming the brief path (`brief_path`); do not paste the contract into the issue body.

## NOT in Scope

Excluded or deferred work with one-line rationale. Omit if nothing deferred. Also record allowed implementation latitude — what Implementor may choose internally.

- [Deferred item]: [One-line rationale]. [Optional follow-up task ref]
- **Allowed latitude**: [What Implementor may choose internally without escalating]

## Error & Failure Map

For each new codepath that can fail, error & failure mapping from Plan Review.

*Omit when task doesn't introduce new failure-prone codepaths.*

## References

Pointers to related decisions, docs, code.