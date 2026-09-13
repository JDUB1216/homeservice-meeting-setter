# Incident Response — SOP

## Purpose

This SOP defines how to respond to incidents in the HomeService Meeting Setter repository. Incidents include security events, governance violations, production issues, and agent malfunctions.

## Incident tiers

| Tier | Severity | Examples | Response time |
|---|---|---|---|
| **P0** | Critical | Active security breach, data leak, production outage | Immediate |
| **P1** | High | ASFDK guard bypass, credential exposure, governance violation | < 1 hour |
| **P2** | Medium | Agent producing non-compliant commits, unclear scope | < 4 hours |
| **P3** | Low | Documentation gaps, minor process drift | < 24 hours |

## Roles

| Role | Responsibility |
|---|---|
| **Incident commander** | Coordinates response, communicates status |
| **Technical lead** | Investigates root cause, implements fix |
| **Governance steward** | Ensures OTOI compliance during response |

## Response steps

### 1. Detect

- Identify the incident through monitoring, alerts, or agent reports.
- Classify severity tier.

### 2. Contain

- Stop the immediate harm (pause agent, revoke access, disable integration).
- Preserve evidence — do not delete logs or state files.

### 3. Escalate

- **P0/P1:** Escalate immediately to Joshua W. Dorsey, Sr. (`info@neuroliftsolutions.com`).
- **P2/P3:** File an escalation using `templates/escalation.md` or `ISSUE_TEMPLATE/agent-escalation.md`.

### 4. Investigate

- Review logs (`data/security.jsonl`, agent session history, git history).
- Document findings in an intent log (`templates/intent-log.md`).

### 5. Resolve

- Apply fix on a feature branch.
- Get human review and approval.
- Merge via PR.

### 6. Post-incident

- Write a handoff record.
- Update `docs/active-threads.md`.
- File a governance proposal if the incident revealed a gap (`ISSUE_TEMPLATE/governance-proposal.md`).

## ASFDK-specific incidents

### Security event log (`data/security.jsonl`)

- Review all `SecurityValidationFailure`, `SecurityInjectionAttempt`, `SecurityLengthExceeded` events.
- If the guard blocked content, investigate why the content was generated.
- If the guard missed content, escalate immediately — the guard is the safety layer.

### Guard health

- Run `Guard.Health()` periodically.
- If health check fails, escalate — the governance layer is compromised.

---

**Governance:** ORG-DEV-OTOI-1.0.3 | Solidarity Framework
**Escalation:** Joshua W. Dorsey, Sr. (`info@neuroliftsolutions.com`)
