# Repository Governance Setup — SOP

## Purpose

This SOP documents how to set up governance files in a new NeuroLift Technologies repository.

## Required files

| File | Required | Purpose |
|---|---|---|
| `CLAUDE.md` | ✅ | Repo-specific context for agents |
| `docs/active-threads.md` | ✅ | Tracks current work state |
| `docs/agent-log/README.md` | ✅ | Creates the agent-log directory |
| `docs/agent-log/registrations/` | ✅ | Directory for agent registration records |
| `docs/agent-log/handoffs/` | ✅ | Directory for handoff records |
| `templates/agent-registration.json` | ✅ | Self-registration template |
| `templates/handoff-record.json` | ✅ | Handoff record template |
| `templates/escalation.md` | ✅ | Escalation template |
| `templates/intent-log.md` | ✅ | Intent log template |
| `SOPs/new-agent-onboarding.md` | ✅ | New agent onboarding |
| `SOPs/repo-governance-setup.md` | ✅ | This file |
| `ISSUE_TEMPLATE/agent-escalation.md` | ✅ | GitHub escalation issue form |
| `ISSUE_TEMPLATE/governance-proposal.md` | ✅ | GitHub governance proposal form |
| `.nltotoi/index/governance-files.md` | ✅ | Governance file registry |
| `.nltotoi/scripts/validate-governance.sh` | ✅ | Validation script |

## Steps

1. **Create CLAUDE.md** at repo root. Reference ORG-DEV-OTOI-1.0.3 and this repo's context.
2. **Create `docs/active-threads.md`** with initial state.
3. **Create `docs/agent-log/README.md`** to initialize the directory.
4. **Create `docs/agent-log/registrations/` and `docs/agent-log/handoffs/`** directories (via `.gitkeep` or README).
5. **Create templates:** `agent-registration.json`, `handoff-record.json`, `escalation.md`, `intent-log.md`.
6. **Create SOPs:** `new-agent-onboarding.md`, `repo-governance-setup.md`.
7. **Create issue templates:** `agent-escalation.md`, `governance-proposal.md`.
8. **Create `.nltotoi/index/governance-files.md`** registry.
9. **Create `.nltotoi/scripts/validate-governance.sh`** validation script.
10. **Verify** by running the validation script.

## Notes

- All `.md` files should reference ORG-DEV-OTOI-1.0.3 where relevant.
- Governance files are NOT gitignored — they must be committed.
- The `.nltotoi/` directory is for governance metadata and tooling.

---

**Governance:** ORG-DEV-OTOI-1.0.3 | Solidarity Framework
**Escalation:** Joshua W. Dorsey, Sr. (`info@neuroliftsolutions.com`)
