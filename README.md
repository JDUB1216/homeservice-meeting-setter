# HomeService Meeting Setter

A Google ADK-Go agent that researches home-service prospects, generates highly specific first-touch cold emails, guards them through ASFDK, and queues them for human approval.

## Status

**Alpha — skeleton in place, engine not yet implemented.** See [Handoff](HANDOFF.md) for what's done and what's missing.

## What it does

1. **Research** — Loads a prospect record (JSON or CSV), extracts concrete observations from their Google Business Profile.
2. **Generate** — Builds a personalized subject/body from those observations (deterministic templates, no LLM). Rejects anything below a specificity bar.
3. **Guard** — Runs every email through ASFDK (prompt defense + output validation). Blocks or flags are logged to a security event log.
4. **Queue** — Holds emails in a file-backed approval queue. The first 50 approved emails are individually gated; after that the queue flips to monitoring mode (auto-approve, still logged).

## Project layout

```
.
├── go.mod
├── Makefile               # build, run, demo, test, vet, fmt, tidy, clean
├── .gitignore
├── cmd/agent/             # (empty) runner main.go not yet written
├── data/                  # runtime data (queue.json, sent.json, security.jsonl, ...)
└── internal/
    ├── agent/             # ADK agent wiring; runPipeline drives the happy path
    ├── asfdk/             # ASFDK Guard wrapper (sanitization, validation, security log)
    ├── email/             # email.Generator — deterministic template-based generation
    ├── engine/            # (empty) — the missing package; see Handoff
    ├── model/             # Prospect, GeneratedEmail, ApprovalItem, Observation, etc.
    ├── prospect/          # loader.go (JSON/CSV) + research.go (ExtractObservations)
    ├── queue/             # file-backed approval queue, ApprovalThreshold=50
    ├── reply/             # lightweight triage of inbound replies (v1: manual import)
    └── sender/            # MockSender + SMTPSender delivery; sent log
```

## Quick start

```bash
make build    # -> bin/agent
make demo     # ./bin/agent demo
make test     # go test ./...
```

The agent reads a `process <path>` command (path = JSON or CSV prospect file), runs research → generate → guard → enqueue, and yields session events. A bare `status` command reports the queue state.

## Key types

- `model.Prospect` — business + GBP snapshot
- `model.GeneratedEmail` — the email object (subject, body, observations, specificity, ASFDK status)
- `model.Observation` — a single concrete fact extracted from the profile
- `email.Generator` — composes emails from observations; `Generate(p, obs) (GeneratedEmail, bool)`
- `asfdk.Guard` — single seam into ASFDK; `NewGuard(userID, logPath)`
- `queue.Queue` — file-backed queue; `Open(path)` → `Enqueue(email)`, `Approve/Reject/Edit(id)`

## Dependencies

- `google.golang.org/adk v1.6.1` — Google ADK-Go
- `github.com/NeuroLift-Technologies/asfdk-go` — ASFDK governance/safety layer

## Governance

Every agent action is governed by the NeuroLift TOI/OTOI contract via ASFDK. The security event log lives at `data/security.jsonl` by default.
