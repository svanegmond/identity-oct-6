# Project index

Taxonomy for this project's `$ARTIFACT_DIR/`. Authoritative definition of each directory: `rules/information-hygiene.mdc` § Authorized Taxonomy.

## Directories

| Path | Purpose |
|------|---------|
| `briefs/` | Design specifications and acceptance criteria |
| `decisions/` | Hard-to-reverse, cross-cutting decisions |
| `disciplines/` | Process and quality posture (checklists) |
| `demos/` | Proof Parade (Lead comprehension; reproduction appendix) |
| `_continuity/` | Session focus, handoff, blockers |

## Artifact kinds

Not all artifacts are the same kind. The distinction matters for where they live:

- **Recording why we chose X over Y** → `decisions/`. Hard-to-reverse, cross-cutting. Future Stewards must not rediscover these.
- **Recording what X is** → project artifact (flat file or directory). Meaning-bearing content: vocabulary, shapes, flows, runtime semantics.
- **Recording how to do X well** → `disciplines/`. Process and quality posture. Checklist-shaped.

When elicitation surfaces something that needs to be settled before the brief proceeds, pick the right kind and write it. Brief cites the artifact; it does not restate the content.

## Project artifacts

`concepts.md` is always present. Others appear when work touches the surface; promote `.md` → directory when a single file gets unwieldy.

Example artifacts (none required beyond `concepts.md`):

| File | Covers |
|------|--------|
| `concepts.md` | Core domain vocabulary and mental models |
| `interactions.md` | User or system interaction flows |
| `runtime.md` | Runtime semantics, lifecycle, concurrency |
| `integrations.md` | External system interfaces and contracts |
| `errors.md` | Error taxonomy, handling posture, recovery |
| `agents.md` | Agent roles, responsibilities, boundaries |
| `schemas.md` | Data shapes, wire formats, validation rules |

Add an artifact when work touches the surface. The list above is illustration, not enforcement.
