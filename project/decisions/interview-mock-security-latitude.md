---
description: "Interview-mock latitude: credential-at-rest and JWT signing-key defaults are Implementor choice; argon2id and strong keys are not hard ACs."
date: 2026-10-06
---

# Decision: Interview-mock security latitude

## Problem

ENG-561 is an Identity service built as a LoginID **interview discussion object**, not a production security product. Hard ACs for argon2id password hashing or non-default JWT signing-key management would over-constrain a mock and mis-signal production readiness.

## Alternatives

- **Production bar:** Require argon2id (or similar) at rest and managed signing keys as acceptance criteria. Rejected for this task — wrong product framing (SK-3).
- **Interview latitude (chosen):** Credential-at-rest representation and JWT signing-key defaults stay Implementor latitude, including weak or plain ideas suitable for a local mock. Document explicitly so Plan/Implementation Review do not invent production ACs.

## Ruling

For ENG-561 (and this Identity interview object unless superseded):

1. Credential-at-rest hashing/storage is Implementor latitude. **argon2id may appear as an optional choice; it is not a required AC.**
2. JWT signing-key defaults (including weak/dev keys) are Implementor latitude. Non-default or rotated production key management is **not** required.
3. Reviewers treat missing argon2id / strong key management as **in latitude**, not defects, unless the Lead amends this decision.

## Consequences

- Brief NOT in Scope / Allowed latitude restates this ruling.
- `github.com/alexedwards/argon2id` stays optional on the dependency menu, not locked-required.
- Live VCs prove JWT issue/verify and auth gate behavior, not credential-hardening claims.

## References

- ENG-561 `project/briefs/identity-go-service.md` (SK-3)
- `project/decisions/rest-api-jwt-bearer.md`
