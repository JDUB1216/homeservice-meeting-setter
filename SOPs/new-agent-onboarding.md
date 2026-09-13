# New Agent Onboarding — SOP

## Purpose

This SOP guides a new coding agent (human or AI) through onboarding into the HomeService Meeting Setter repository.

## Pre-requisites

1. Access to `.github-private/NLT-DEV-OTOI.md`
2. Access to `AGENTS.md` at the repo root (or `Desktop/AGENTS.md` for global)
3. Go 1.25.0+ installed

## Steps

### 1. Read the canonical governance contract

> Have you read `NLT-DEV-OTOI.md` in `.github-private`? Focus on Sections 1, 4, 4.4, 5, and 8.

- Understand the 5-step session start protocol (Section 4.1)
- Understand the escalation triggers (Section 5)
- Understand the handoff requirements (Section 5)

### 2. Read AGENTS.md

> Have you read `AGENTS.md`? It defines the coordination protocol, guardrails, and internal file map.

- Coordination protocol
- Global file map
- Repo index

### 3. Read this repo's CLAUDE.md

> Have you read the `CLAUDE.md` in `/home/joshd/Desktop/homeservice-meeting-setter/`?

- Repo-specific context
- Build & run commands
- Agent coordination
- Commit format
- Guardrails

### 4. Read active threads

> Have you read `docs/active-threads.md`?

- Understand what work is in progress
- Do not duplicate or conflict with open threads

### 5. Self-register

Fill out `templates/agent-registration.json` with:
- `agent_name` — your identifier (e.g., `HERMES`, `CLAUDE`)
- `session_id` — unique session ID
- `working_repo` — `/home/joshd/Desktop/homeservice-meeting-setter`
- `scope` — what you're working on
- `notes` — anything relevant

Save to `docs/agent-log/registrations/YYYY-MM-DD-<agent>.json`.

### 6. Confirm task scope with human

Before writing any code, confirm:
- What the task is
- What's in scope vs. out of scope
- Any architectural constraints

### 7. Work

- Create a feature branch
- Make changes following commit format
- Run tests before pushing
- Open a PR (never push to main)

### 8. Handoff

At the end of the session:
1. Update `docs/active-threads.md`
2. Write a handoff record to `docs/agent-log/handoffs/`
3. Document open escalations in `docs/escalations/`
4. Summarize decisions made and pending

---

**Governance:** ORG-DEV-OTOI-1.0.3 | Solidarity Framework
**Escalation:** Joshua W. Dorsey, Sr. (`info@neuroliftsolutions.com`)
