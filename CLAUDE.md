# CLAUDE.md — HomeService Meeting Setter

This repository is governed by **NeuroLift Technologies ORG-DEV-OTOI-1.0.3**. All work here must comply with the canonical contract at `NLT-DEV-OTOI.md` in `.github-private`.

## Repo-specific context

- **Module:** `github.com/NeuroLift-Technologies/homeservice-meeting-setter`
- **Go version:** 1.25.0
- **Pipeline:** `process <path>` → research → generate → ASFDK guard → enqueue → yield events
- **Key packages:** `internal/agent`, `internal/engine`, `internal/model`, `internal/prospect`, `internal/email`, `internal/asfdk`, `internal/queue`, `internal/reply`, `internal/sender`
- **Dependencies:** `google.golang.org/adk v1.6.1`, `github.com/NeuroLift-Technologies/asfdk-go`
- **Runtime data:** `data/` (gitignored — queue.json, sent.json, security.jsonl, replies.json)

## Build & run

```bash
make build    # -> bin/agent
make demo     # ./bin/agent demo
make test     # go test ./...
make tidy     # go mod tidy
```

## Agent coordination

- Read `docs/active-threads.md` at session start — do not duplicate or conflict with in-progress threads.
- Self-register using `templates/agent-registration.json` before significant work.
- Write handoff records to `docs/agent-log/handoffs/` using `templates/handoff-record.json`.
- Escalate to Joshua W. Dorsey, Sr. (`info@neuroliftsolutions.com`) when scope is unclear, architectural decisions are needed, or blockers arise.

## Commit format

```
[AGENT_NAME] type(scope): description
```

Allowed types: `feat`, `fix`, `docs`, `refactor`, `chore`, `test`, `ci`

Example: `[HERMES] feat(engine): wire ASFDK guard into pipeline (ORG-DEV-OTOI-1.0.3)`

## Guardrails

- No LLM provider lock-in
- No architecture decisions without human sign-off
- No production deployments without explicit approval
- No credential storage in code or VCS
- PR-only workflow — feature branches + Pull Requests, never push to `main`
- No OTOI self-amendment

## Governance files

See `.nltotoi/index/governance-files.md` for the full registry.
