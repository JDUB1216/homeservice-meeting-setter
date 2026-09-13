# AGENTS.md — HomeService Meeting Setter

This repository is governed by **NeuroLift Technologies ORG-DEV-OTOI-1.0.3**. See `CLAUDE.md` for project-specific context and `NLT-DEV-OTOI.md` in `.github-private` for the canonical contract.

## Mandatory Reading Order

1. Read this AGENTS.md file
2. Read `CLAUDE.md` for project-specific conventions
3. Read `docs/active-threads.md` at session start — do not duplicate or conflict with in-progress threads
4. Verify PR Review Hermes bot is active before any merge operations

## Agent Coordination

- Self-register using `templates/agent-registration.json` before significant work.
- Write handoff records to `docs/agent-log/handoffs/` using `templates/handoff-record.json`.
- Escalate to Joshua W. Dorsey, Sr. (`info@neuroliftsolutions.com`) when scope is unclear, architectural decisions are needed, or blockers arise.

## Commit Format

```
[AGENT_NAME] type(scope): description
```

Allowed types: `feat`, `fix`, `docs`, `refactor`, `chore`, `test`, `ci`

Example: `[HERMES] fix(vuln): upgrade golang.org/x/* to resolve govulncheck findings (ORG-DEV-OTOI-1.0.3)`

## PR Review Hermes Bot

All PRs must pass all required status checks before merge:
- OSSAR-Scan (security vulnerabilities)
- Check Contact Email Compliance
- Scan PR for Credential Exposure (SOP-NLT-003)
- Scan for Governance Incidents (SOP-NLT-003)
- Validate (governance compliance)
- Validate Agent Commit Format (SOP-NLT-001)
- Check Agent Handoff Record (SOP-NLT-001)
- Any repo-specific required checks

## Guardrails

- No LLM provider lock-in
- No architecture decisions without human sign-off
- No production deployments without explicit approval
- No credential storage in code or VCS
- PR-only workflow — feature branches + Pull Requests, never push to `main`
- No OTOI self-amendment

## Build & Run

```bash
make build    # -> bin/agent
make demo     # ./bin/agent demo
make test     # go test ./...
make vet      # go vet ./...
make fmt      # gofmt -l -w .
make tidy     # go mod tidy
```

## Project Context

- **Module:** `github.com/NeuroLift-Technologies/homeservice-meeting-setter`
- **Go version:** 1.26.0
- **Pipeline:** `process <path>` → research → generate → ASFDK guard → enqueue → yield events
- **Key packages:** `internal/agent`, `internal/engine`, `internal/model`, `internal/prospect`, `internal/email`, `internal/asfdk`, `internal/queue`, `internal/reply`, `internal/sender`
- **Dependencies:** `google.golang.org/adk v1.6.1`, `github.com/NeuroLift-Technologies/asfdk-go`
- **Runtime data:** `data/` (gitignored — queue.json, sent.json, security.jsonl, replies.json)

## Governance Files

See `.nltotoi/index/governance-files.md` for the full registry.
