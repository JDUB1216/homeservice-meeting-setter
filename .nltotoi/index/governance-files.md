# Governance File Registry

This file is the authoritative index of all governance files in the HomeService Meeting Setter repository. It is maintained by the NLT Governance Steward and validated by `.nltotoi/scripts/validate-governance.sh`.

## Registry

| File | Type | Required | Status | Purpose |
|---|---|---|---|---|
| `CLAUDE.md` | Core | ✅ | Present | Repo-specific agent context |
| `docs/active-threads.md` | Core | ✅ | Present | Tracks current work state |
| `docs/agent-log/README.md` | Core | ✅ | Present | Agent log directory init |
| `docs/agent-log/registrations/` | Directory | ✅ | Present | Agent registration records |
| `docs/agent-log/handoffs/` | Directory | ✅ | Present | Handoff records |
| `templates/agent-registration.json` | Template | ✅ | Present | Self-registration template |
| `templates/handoff-record.json` | Template | ✅ | Present | Handoff record template |
| `templates/escalation.md` | Template | ✅ | Present | Escalation template |
| `templates/intent-log.md` | Template | ✅ | Present | Intent log template |
| `SOPs/new-agent-onboarding.md` | SOP | ✅ | Present | New agent onboarding |
| `SOPs/repo-governance-setup.md` | SOP | ✅ | Present | Repo governance setup |
| `SOPs/incident-response.md` | SOP | ✅ | Present | Incident response |
| `ISSUE_TEMPLATE/agent-escalation.md` | Issue Template | ✅ | Present | GitHub escalation form |
| `ISSUE_TEMPLATE/governance-proposal.md` | Issue Template | ✅ | Present | GitHub governance proposal form |
| `.nltotoi/index/governance-files.md` | Index | ✅ | Present | This file |
| `.nltotoi/scripts/validate-governance.sh` | Script | ✅ | Present | Validation script |

## Conventions

- All `.md` governance files reference ORG-DEV-OTOI-1.0.3.
- All `.json` governance files are linted on commit.
- Governance files are committed to VCS — never gitignored.
- The `.nltotoi/` directory holds metadata and tooling only.

## Validation

Run the validation script to check compliance:

```bash
bash .nltotoi/scripts/validate-governance.sh
```

---

**Governance:** ORG-DEV-OTOI-1.0.3 | Solidarity Framework
**Authority:** Joshua W. Dorsey, Sr.
**Last updated:** 2026-09-13
