# HomeService Meeting Setter

A Google ADK-Go agent that researches home-service prospects, generates highly specific first-touch cold emails, guards them through ASFDK, and queues them for human approval.

## Status

**Alpha — skeleton complete, engine implemented, tests passing.** See [Handoff](HANDOFF.md) for full history.

## What it does

1. **Research** — Loads a prospect record (JSON or CSV), extracts concrete observations from their Google Business Profile.
2. **Generate** — Builds a personalized subject/body from those observations (deterministic templates, no LLM). Rejects anything below a specificity bar.
3. **Guard** — Runs every email through ASFDK (prompt defense + output validation). Blocks or flags are logged to a security event log.
4. **Queue** — Holds emails in a file-backed approval queue. The first 50 approved emails are individually gated; after that the queue flips to monitoring mode (auto-approve, still logged).
5. **Send** — Delivers approved emails via MockSender (demo) or SMTPSender (production). Every attempt is recorded in the sent log.
6. **Reply** — Lightweight triage of inbound replies (positive / neutral / negative / out-of-office) with auto-drafted booking responses for positive replies.

## Project layout

```
.
├── go.mod
├── Makefile               # build, run, demo, test, vet, fmt, tidy, clean
├── .gitignore
├── cmd/agent/             # runner main.go — demo / status / process <path>
├── data/                  # runtime data (queue.json, sent.json, security.jsonl, ...)
└── internal/
    ├── agent/             # ADK agent wiring; runPipeline drives the happy path
    ├── asfdk/             # ASFDK Guard wrapper (sanitization, validation, security log)
    ├── email/             # email.Generator — deterministic template-based generation
    ├── engine/            # composition root — wires research → generate → guard → queue
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
- `sender.Sender` — transport interface; `MockSender` (demo) or `SMTPSender` (production)
- `reply.Classify` — keyword-based reply classifier; `reply.DraftResponse` — booking reply draft

## Testing

```
make test    # 28 tests across 4 packages (email, engine, prospect, queue)
```

## Dependencies

- `google.golang.org/adk v1.6.1` — Google ADK-Go
- `github.com/NeuroLift-Technologies/asfdk-go` — ASFDK governance/safety layer

## Governance

Every agent action is governed by the NeuroLift TOI/OTOI contract via ASFDK. The security event log lives at `data/security.jsonl` by default.
