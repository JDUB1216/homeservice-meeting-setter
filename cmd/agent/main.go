// Command agent is the runner for the HomeService Meeting Setter agent.
//
// Usage:
//
//	agent demo     run the deterministic happy path on data/prospects.json
//	agent status  print the current queue state
//	agent process <path>   research → generate → guard → enqueue a prospect file
package main

import (
	"fmt"
	"os"

	"github.com/NeuroLift-Technologies/homeservice-meeting-setter/internal/asfdk"
	"github.com/NeuroLift-Technologies/homeservice-meeting-setter/internal/engine"
	"github.com/NeuroLift-Technologies/homeservice-meeting-setter/internal/email"
)

const (
	defaultQueuePath   = "data/queue.json"
	defaultSecurityLog = "data/security.jsonl"
	defaultUserID      = asfdk.UserID
	defaultDataPath    = "data/prospects.json"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "agent: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	// Default subcommand: run the demo pipeline on the bundled prospects file.
	command := "demo"
	if len(args) > 0 {
		command = args[0]
	}

	eng, err := engine.New(engine.Config{
		QueuePath:       defaultQueuePath,
		SecurityLogPath: defaultSecurityLog,
		UserID:          defaultUserID,
		EmailConfig:     email.Config{},
	})
	if err != nil {
		return fmt.Errorf("engine: %w", err)
	}
	defer eng.Close()

	switch command {
	case "demo":
		return runDemo(eng)
	case "status":
		fmt.Println(eng.StatusSummary())
		return nil
	case "process":
		if len(args) < 2 {
			return fmt.Errorf("usage: agent process <path>")
		}
		return runProcess(eng, args[1])
	default:
		return fmt.Errorf("unknown command %q — try: demo, status, process <path>", command)
	}
}

// runDemo drives the deterministic happy path on the bundled prospects file.
func runDemo(eng *engine.Engine) error {
	path := defaultDataPath
	if _, err := os.Stat(path); err == nil {
		return runProcess(eng, path)
	}
	// No bundled data — print status and exit cleanly.
	fmt.Println(eng.StatusSummary())
	fmt.Println("no bundled prospects file at", path)
	return nil
}

// runProcess loads a prospect file and runs the full pipeline:
// research → generate → guard → enqueue. Each step is observable via the
// engine's counters and the queue state.
func runProcess(eng *engine.Engine, path string) error {
	prospects, err := engine.LoadFile(path)
	if err != nil {
		return fmt.Errorf("load %s: %w", path, err)
	}

	var processed, failed int
	for _, p := range prospects {
		email, ok := eng.Generate(p)
		if !ok {
			failed++
			continue
		}
		if err := eng.GuardEmail(email, p); err != nil {
			fmt.Fprintf(os.Stderr, "guard %s: %v\n", p.BusinessName, err)
			failed++
			continue
		}
		if err := eng.Enqueue(email); err != nil {
			fmt.Fprintf(os.Stderr, "enqueue %s: %v\n", p.BusinessName, err)
			failed++
			continue
		}
		processed++
		eng.IncrementProcessed()
		fmt.Printf("queued %s -> %s (specificity %.2f)\n", p.BusinessName, email.Subject, email.Specificity)
	}

	for i := 0; i < failed; i++ {
		eng.IncrementFailed()
	}

	fmt.Println()
	fmt.Println(eng.StatusSummary())
	fmt.Printf("processed=%d failed=%d queue.pending=%d\n", processed, failed, len(eng.Queue().Pending()))
	return nil
}