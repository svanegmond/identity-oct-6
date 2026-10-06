# Disciplines

Dir holds process and quality posture: checklists Lead agreed to (engineering, deployment, security). Complements Helm plugin rules; defines product-specific "good," not pipeline mechanics.

## What belongs here

Concrete, reviewable expectations: errors, tests, observability, rollout safety, secrets handling. Checklist-shaped, not definitional.

## What does not belong here

- **Definitions, concepts, domain vocabulary** — project artifacts (`$ARTIFACT_DIR/concepts.md`, etc.).
- **Decision records** — `$ARTIFACT_DIR/decisions/`.
- **Design briefs** — `$ARTIFACT_DIR/briefs/`.

## Adding files

Create or update only with explicit Lead agreement. Names: `engineering.md`, `deployment.md`, `security.md`. Optional: `review-extensions.md` for reviewer extensions.
