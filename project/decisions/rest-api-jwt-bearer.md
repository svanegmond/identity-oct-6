---
description: "REST API authenticates callers with JWT bearer tokens issued after local credential check."
date: 2026-10-06
---

# Decision: REST API JWT bearer authentication

## Problem

The Identity Go service REST surface must authenticate callers before profile search/retrieve and other protected routes. The source prompt required a Lead lock on scheme before implementation. Leaving auth unspecified would block Phase B and invite incompatible Implementor guesses.

## Alternatives

- **O1 JWT bearer (chosen):** After successful local credential check, issue a JWT; callers send `Authorization: Bearer <token>`; middleware verifies before protected handlers. Aligns with LoginID verify-token spirit (local issue+verify) without adopting LoginID SDK or passkeys.
- **O2 API key:** Static key on requests. Simpler, but weak interview discussion of token lifecycle and verify-before-route.
- **O3 HTTP basic:** Username/password on every request. Couples every call to credential material; poor fit for separate auth vs identity/PII conversation.

## Ruling

**Auth D1 = O1.** Protected REST routes require a valid JWT in the `Authorization: Bearer` header. Tokens are issued only after a successful local user-credential check. Optional JWT scope claims are allowed for discussion; they are not required for the minimum auth gate. Binding on ENG-561 and downstream Identity REST work in this workspace.

## Consequences

- Phase A REST/OpenAPI contract must document bearer auth, login/token issue, and verify-before-handler behavior.
- Implementor uses `github.com/golang-jwt/jwt/v5` (locked dependency menu).
- Signing-key defaults for this interview mock follow `project/decisions/interview-mock-security-latitude.md` — production key management is not required.
- Passkeys/FIDO2 and LoginID cloud token APIs remain out of scope.

## References

- ENG-561 `project/briefs/identity-go-service.md`
- `project/tmp/agent-prompt-identity-go.md`
- `project/decisions/interview-mock-security-latitude.md`
- https://docs.loginid.io
