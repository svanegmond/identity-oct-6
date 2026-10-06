# Proof Parade: [Task Name]

**Brief**: [link to brief]
**Task**: [task ref]
**Date**: [YYYY-MM-DD]
**Scaffold drafted by**: Steward
**Evidence captured by**: Implementor

<!-- Proof Parade: Lead-comprehension artifact. Evidence Index maps every AC and VC to What-landed and/or an Exhibit (not appendix-only). ACs: named tests. VCs: the surface driven live — screenshot, scripted interaction recording, request/response against a running server, terminal session, stored state read back. Reproduction appendix points at /tmp/helm-ir-battery/<ref>/<sha>.json + tests + how to reconstruct each VC. Steward scaffolds contracts→slots; Implementor fills. Lead Made audits understanding. -->

## Evidence Index

| Criterion | Planned evidence | Captured evidence | What it proves |
|---|---|---|---|
| AC-1 | [Steward: What-landed pointer and/or Exhibit medium] | [Implementor: What-landed § / Exhibit name — not appendix-only] | [contract claim] |
| TP-1 | [Steward: named test / battery check] | [Implementor: What-landed pointer or appendix command ref] | [claim] |
| VC-1 | [Steward: Exhibit medium — walkthrough / payload / diagram / media] | [Implementor: Exhibit name] | [landed-contract comprehension] |

Every AC and every VC gets an Index row whose Captured column points at **What landed** and/or an **Exhibit**. The Reproduction appendix is not a substitute for an Index row.

**AC captured evidence** is a named test. **VC captured evidence** is the thing itself, driven: a screenshot of the page, a recording of the scripted interaction, a request and its response against a running server, a terminal session, a row read back out of the store. A test name in a VC cell is not proof.

## What landed

[Steward scaffolds one narrative page from the brief's contracts. Implementor fills with landed shapes.]

- **[Contract / AC]:** [before → after, or "here is the thing working"]
- **[Interaction designed]:** [Mermaid or short trace when the brief designed a network/flow]

Write for a reader who already holds the brief's intent and wants consequences, not a re-litigation of evidence.

## Exhibits

### Exhibit: [name — page screenshot / browser-interaction recording / request-response against a running server / terminal session / stored-state readback / diagram / live URL]

**What the Lead should see/feel:** [one paragraph]
**Maps to:** [AC-# / VC-#]
**Captured:** [paths / embeds / links]

![or embed / link media here](screenshots/or-video.mp4)

```text
[full payload or other output *when the output is the exhibit*]
```

Inline raw output survives only where the output *is* the exhibit (a payload worth reading, a refusal message worth quoting). Otherwise point at the Reproduction appendix.

## Reproduction appendix

Per-criterion pointers for cold replay. AC truth lives in named tests and the battery. VC truth lives in the live-probe exhibit; the appendix says how to reconstruct it. Paste transcripts when the output *is* the exhibit (a refusal sentence, a tool response).

- **Battery report:** `/tmp/helm-ir-battery/<ref>/<sha>.json` — [blocking_green / key checks]
- **Named tests:** [module::test that permanently guards each contract]
- **VC drivers:** [one exact task-worktree-root command per single-surface VC; match `implementation_handoff.vc_drivers`]
- **Commands (non-battery):** [exact commands for anything not battery-covered]

```
[optional: short command + outcome when not captured by battery]
```

## Known Gaps

- [What this doesn't prove, why]
