package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NeuroLift-Technologies/homeservice-meeting-setter/internal/email"
	"github.com/NeuroLift-Technologies/homeservice-meeting-setter/internal/model"
	"github.com/NeuroLift-Technologies/homeservice-meeting-setter/internal/prospect"
)

func tempEngine(t *testing.T) (*Engine, func()) {
	t.Helper()
	dir := t.TempDir()
	eng, err := New(Config{
		QueuePath:       filepath.Join(dir, "queue.json"),
		SecurityLogPath: filepath.Join(dir, "security.jsonl"),
		UserID:          "test-user",
		EmailConfig:     email.Config{},
	})
	if err != nil {
		t.Fatalf("New engine failed: %v", err)
	}
	return eng, func() { eng.Close() }
}

func TestLoadFile_JSONArray(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prospects.json")
	data := `[{"id":"p1","business_name":"Test HVAC","vertical":"hvac","city":"Austin","gbp":{"rating":4.1,"review_count":10}}]`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	prospects, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}
	if len(prospects) != 1 {
		t.Fatalf("expected 1 prospect, got %d", len(prospects))
	}
	if prospects[0].BusinessName != "Test HVAC" {
		t.Fatalf("expected Test HVAC, got %s", prospects[0].BusinessName)
	}
}

func TestLoadFile_Wrapper(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prospects.json")
	data := `{"prospects":[{"id":"p1","business_name":"Wrapped Co","vertical":"plumbing","city":"Denver","gbp":{"rating":4.0,"review_count":5}}]}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	prospects, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}
	if len(prospects) != 1 {
		t.Fatalf("expected 1 prospect, got %d", len(prospects))
	}
}

func TestLoadFile_Invalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	data := `{"invalid":true}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadFile(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGenerate_Pipeline(t *testing.T) {
	eng, cleanup := tempEngine(t)
	defer cleanup()

	p := model.Prospect{
		ID:           "p1",
		BusinessName: "PipeReady Plumbing",
		Vertical:     model.VerticalPlumbing,
		City:         "Austin",
		GBP: model.GBPProfile{
			Rating:      4.1,
			ReviewCount: 23,
		},
	}
	email, ok := eng.Generate(p)
	if !ok {
		t.Fatal("expected generation to succeed")
	}
	if email.Subject == "" {
		t.Fatal("expected non-empty subject")
	}
	if !strings.Contains(email.Body, p.BusinessName) {
		t.Fatalf("body must reference business name")
	}
}

func TestGuardEmail_Clean(t *testing.T) {
	eng, cleanup := tempEngine(t)
	defer cleanup()

	email := model.GeneratedEmail{
		Subject: "Hello",
		Body:    "This is a clean email with no issues.",
	}
	if err := eng.GuardEmail(email, model.Prospect{}); err != nil {
		t.Fatalf("expected clean guard, got: %v", err)
	}
}

func TestEnqueue(t *testing.T) {
	eng, cleanup := tempEngine(t)
	defer cleanup()

	email := model.GeneratedEmail{
		Subject: "Test",
		Body:    "Body",
		Status:  model.EmailPendingApproval,
	}
	if err := eng.Enqueue(email); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	if len(eng.Queue().Pending()) != 1 {
		t.Fatal("expected 1 pending item")
	}
}

func TestStatusSummary(t *testing.T) {
	eng, cleanup := tempEngine(t)
	defer cleanup()

	summary := eng.StatusSummary()
	if summary == "" {
		t.Fatal("expected non-empty summary")
	}
	if !strings.Contains(summary, "processed=") {
		t.Fatal("summary should contain processed=")
	}
}

func TestIncrementCounters(t *testing.T) {
	eng, cleanup := tempEngine(t)
	defer cleanup()

	eng.IncrementProcessed()
	eng.IncrementFailed()
	summary := eng.StatusSummary()
	if !strings.Contains(summary, "processed=1") {
		t.Fatalf("expected processed=1, got: %s", summary)
	}
	if !strings.Contains(summary, "failed=1") {
		t.Fatalf("expected failed=1, got: %s", summary)
	}
}

// Ensure the package compiles with the prospect import used.
var _ = prospect.Load