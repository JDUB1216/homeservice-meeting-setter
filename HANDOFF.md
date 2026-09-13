# Handoff — HomeService Meeting Setter

**Date:** 2026-09-13
**Status:** Alpha skeleton. Nothing has been written yet — this is a planning handoff, not a delivery handoff.

---

## What has been done

### 1. Codebase analysis (complete)

Every internal package has been read and its public API captured:

| Package | File | Public surface |
|---|---|---|
| `internal/agent` | `agent.go` (147 lines) | `Build(ctx, *engine.Engine, BuildOptions) (*agent.Agent, *Service, error)`, `Service{Engine, Sessions}`, `runPipeline(inv) iter.Seq2` |
| `internal/model` | `model.go` (156 lines) | `Vertical`, `Prospect`, `GBPProfile`, `Observation`, `GeneratedEmail`, `EmailStatus`, `ASFDKStatus`, `ApprovalStatus`, `ApprovalItem`, `ReplyClassification`, `Reply` |
| `internal/email` | `generate.go` (293 lines) | `Config`, `Generator`, `NewGenerator(cfg)`, `Generate(p, obs) (GeneratedEmail, bool)`, `ExtractObservations(p)`, `ValidateSpecificity(p, obs, body) bool` |
| `internal/prospect` | `loader.go` (181), `research.go` (172) | `Load(path) ([]Prospect, error)`, `LoadCSV`, `ExtractObservations(p)`, `PickPersonalization(obs)`, `MissingServices(v, provided)` |
| `internal/asfdk` | `asfdk.go` (139) | `Guard`, `NewGuard(userID, logPath) (*Guard, error)`, `GuardResult`, `SanitizeEmailInput`, `ValidateEmailOutput`, `LogBlock` |
| `internal/queue` | `queue.go` (236) | `ApprovalThreshold=50`, `Queue`, `Open(path) (*Queue, error)`, `Enqueue(e)`, `Approve(id)`, `Reject(id)`, `Edit(id, text)`, `MarkSent(id)` |
| `internal/reply` | `reply.go` (168) | Classification + booking-response draft |
| `internal/sender` | `sender.go` (143) | `MockSender`, `SMTPSender`, sent log |

### 2. Module and dependency verification (complete)

- `go.mod`: `module github.com/NeuroLift-Technologies/homeservice-meeting-setter`, `go 1.25.0`, requires `google.golang.org/adk v1.6.1` and `github.com/NeuroLift-Technologies/asfdk-go`.
- ADK-Go v1.6.1 module cache is present at `/home/joshd/go/pkg/mod/google.golang.org/adk@v1.6.1/`.
- `cmd/agent/` is **empty** — runner `main.go` has not been written.
- **go.sum has MISSING entries** (LSP confirmed): `google.golang.org/genai`, `go.opentelemetry.io/otel`, `github.com/google/uuid`, `rsc.io/omap`, `rsc.io/ordered`. Run `go mod tidy` to fix.

### 3. Engine contract extracted (complete)

`agent.go` is the sole consumer of `internal/engine`. The contract is fully locked:

```go
// Package-level loader (called as engine.LoadFile(path))
func LoadFile(path string) ([]model.Prospect, error)

// Engine methods (called on *engine.Engine)
func (e *Engine) Generate(p model.Prospect) (mail model.GeneratedEmail, ok bool)
func (e *Engine) GuardEmail(email model.GeneratedEmail, p model.Prospect) error
func (e *Engine) StatusSummary() string
```

Also: `Service.Engine *engine.Engine`, `Build(ctx, eng *engine.Engine, opts BuildOptions)`.

**Pipeline flow (from `runPipeline`):** `LoadFile` → for each prospect, `Generate` (skip if `!ok`) → `GuardEmail` (skip on error) → enqueue → yield events (`research`, `email`, `summary`, `status`).

---

## Real compile-time bugs in `agent.go` (LSP-verified, must fix before engine works)

The `gopls` LSP diagnostics on `agent.go` are authoritative (not bash replay). These bugs block `go build ./...` even AFTER `internal/engine` is created:

### BUG 1 — Wrong ADK import path for genai (line 17)
```
ERROR: could not import google.golang.org/adk/genai (no required module provides package)
```
ADK agent imports `google.golang.org/genai` (a **separate module**, NOT under `google.golang.org/adk`). Fix: change `"google.golang.org/adk/genai"` → `"google.golang.org/genai"` and add it to `go.mod`/`go.sum`.

### BUG 2 — Wrong import path for session/inmemory (line 19)
```
ERROR: could not import google.golang.org/adk/session/inmemory (no required module provides package)
```
No `session/inmemory` subpackage exists in ADK-Go v1.6.1. The in-memory session constructor lives **in-package** at `google.golang.org/adk/session` (e.g. `session.InMemoryService()`). Fix: remove the `session/inmemory` import, use the in-package constructor, and change `Sessions *inmemory.Service` to the correct type.

### BUG 3 — `agent.New` returns interface, not pointer (line 52)
```
ERROR: cannot use a (interface type agent.Agent) as *agent.Agent value in return
```
`agent.New` returns `agent.Agent` (interface). Fix: `return a, svc, nil` (drop the pointer) or change the return type from `*agent.Agent` to `agent.Agent`.

### BUG 4 — `eng` undefined in `runPipeline` (lines 87, 92, 102)
```
ERROR: undefined: eng
```
`runPipeline` is declared as `func runPipeline(inv agent.InvocationContext) iter.Seq2[*session.Event, error]` — it takes only `inv` and has no access to the `*engine.Engine` from `Build`. The engine must be captured as a closure or stored on `Service`. Fix: either make `runPipeline` a local closure inside `Build` that captures `eng`, or pass the engine another way.

### BUG 5 — `ev.Role` undefined (line 119)
```
ERROR: ev.Role undefined (type *session.Event has no field or method Role)
```
`session.Event` has no `Role` field in ADK-Go v1.6.1. Fix: use the correct field/method for the event role.

### BUG 6 — Missing go.sum entries
Run `go mod tidy` to pull in `google.golang.org/genai`, `go.opentelemetry.io/otel`, `github.com/google/uuid`, `rsc.io/omap`, `rsc.io/ordered`.

---

## What is still left (BLOCKING)

### P0 — `internal/engine/engine.go` (NOT WRITTEN)

The `internal/engine/` directory is **empty**. This is the one missing package that prevents the entire project from compiling. It must implement the contract above and compose the healthy siblings:

- `prospect.Load` / `prospect.ExtractObservations` — for `LoadFile`
- `email.NewGenerator` / `email.Generator.Generate` — for `Generate`
- `asfdk.NewGuard` / `asfdk.Guard` — for `GuardEmail`
- `queue.Open` / `queue.Queue.Enqueue` — for the enqueue step
- `model` types — for signatures and return values

Constructor needed: something that wires `email.Generator`, `*asfdk.Guard`, `*queue.Queue`, and counters into `*Engine`. The `Engine` struct fields are not yet defined (design decision for the next agent).

### P1 — `cmd/agent/main.go` (NOT WRITTEN)

The runner `main.go` referenced by `Makefile` (`go build -o bin/agent ./cmd/agent`, `make demo`) does not exist. Needs to call `agent.Build` and run the agent.

### P2 — Fix `agent.go` bugs (BLOCKING for build)

All 6 bugs listed above must be fixed before `go build ./...` passes. These are NOT optional — they are confirmed compile errors.

### P3 — `go mod tidy`

Run `go mod tidy` to restore missing `go.sum` entries after fixing the import paths (P2).

---

## Design decisions still open

1. **Engine struct fields** — what fields `Engine` holds (likely: `gen *email.Generator`, `guard *asfdk.Guard`, `q *queue.Queue`, processed/failed counters, `cfg`).
2. **Constructor signature** — how `New` (or equivalent) is called and by whom.
3. **LoadFile vs Load** — whether `engine.LoadFile` delegates directly to `prospect.Load` or wraps it.
4. **GuardEmail behavior** — whether it calls `asfdk.Guard.SanitizeEmailInput` + `ValidateEmailOutput`, sets `email.ASFDKStatus`, and returns an error when blocked/flagged.
5. **StatusSummary format** — what string it returns (processed count, failed count, queue state).
6. **runPipeline closure** — how the engine is captured (closure vs Service field vs package var).

---

## Environment notes

- **Bash tool output is unreliable** (replays polluted old transcripts). The `read` tool and the `gopls` LSP diagnostics are authoritative for file content and compile errors respectively.
- Repo root: `/home/joshd/Desktop/homeservice-meeting-setter`
- Module cache: `/home/joshd/go/pkg/mod/google.golang.org/adk@v1.6.1/`
- Go version: 1.25.0
- Runtime data lives in `data/` (gitignored)
- Files created in this session: `README.md`, `HANDOFF.md`
