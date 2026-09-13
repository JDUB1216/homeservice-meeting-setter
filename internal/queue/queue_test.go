package queue

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/NeuroLift-Technologies/homeservice-meeting-setter/internal/model"
)

func tempQueuePath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "queue.json")
}

func TestOpen_NewFile(t *testing.T) {
	path := tempQueuePath(t)
	q, err := Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	if q == nil {
		t.Fatal("expected non-nil queue")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("queue file not created: %v", err)
	}
}

func TestEnqueue_Pending(t *testing.T) {
	path := tempQueuePath(t)
	q, err := Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	email := model.GeneratedEmail{
		ID:           "em_1",
		ProspectID:   "p1",
		Subject:      "Test",
		Body:         "Body",
		Specificity:  0.8,
		Status:       model.EmailPendingApproval,
	}
	if err := q.Enqueue(email); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	pending := q.Pending()
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending, got %d", len(pending))
	}
	if pending[0].Email.ID != "em_1" {
		t.Fatalf("expected em_1, got %s", pending[0].Email.ID)
	}
	if pending[0].Position != 1 {
		t.Fatalf("expected position 1, got %d", pending[0].Position)
	}
}

func TestApprove(t *testing.T) {
	path := tempQueuePath(t)
	q, err := Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	email := model.GeneratedEmail{ID: "em_1", Status: model.EmailPendingApproval}
	if err := q.Enqueue(email); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	count, err := q.Approve("em_1", "human")
	if err != nil {
		t.Fatalf("Approve failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
	if q.ApprovedCount() != 1 {
		t.Fatalf("expected approved count 1, got %d", q.ApprovedCount())
	}
	pending := q.Pending()
	if len(pending) != 0 {
		t.Fatalf("expected 0 pending after approve, got %d", len(pending))
	}
}

func TestReject(t *testing.T) {
	path := tempQueuePath(t)
	q, err := Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	email := model.GeneratedEmail{ID: "em_1", Status: model.EmailPendingApproval}
	if err := q.Enqueue(email); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	if err := q.Reject("em_1", "not specific"); err != nil {
		t.Fatalf("Reject failed: %v", err)
	}
	// Rejected items should not appear in pending.
	pending := q.Pending()
	if len(pending) != 0 {
		t.Fatalf("expected 0 pending after reject, got %d", len(pending))
	}
}

func TestEdit(t *testing.T) {
	path := tempQueuePath(t)
	q, err := Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	email := model.GeneratedEmail{ID: "em_1", Subject: "Old", Body: "Old body", Status: model.EmailPendingApproval}
	if err := q.Enqueue(email); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	if err := q.Edit("em_1", "human", "New Subject", "New body"); err != nil {
		t.Fatalf("Edit failed: %v", err)
	}
	items := q.Items()
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Email.Subject != "New Subject" {
		t.Fatalf("expected new subject, got %s", items[0].Email.Subject)
	}
	if !items[0].Email.Edited {
		t.Fatal("expected edited flag")
	}
}

func TestMarkSent(t *testing.T) {
	path := tempQueuePath(t)
	q, err := Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	email := model.GeneratedEmail{ID: "em_1", Status: model.EmailPendingApproval}
	if err := q.Enqueue(email); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	if _, err := q.Approve("em_1", "human"); err != nil {
		t.Fatalf("Approve failed: %v", err)
	}
	if err := q.MarkSent("em_1"); err != nil {
		t.Fatalf("MarkSent failed: %v", err)
	}
	items := q.Items()
	if items[0].Email.Status != model.EmailSent {
		t.Fatalf("expected sent status, got %s", items[0].Email.Status)
	}
}

func TestIsMonitorMode(t *testing.T) {
	path := tempQueuePath(t)
	q, err := Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	if q.IsMonitorMode() {
		t.Fatal("should not be in monitor mode initially")
	}
	// Approve 50 emails to cross threshold.
	for i := 0; i < 50; i++ {
		email := model.GeneratedEmail{ID: fmt.Sprintf("em_%d", i), Status: model.EmailPendingApproval}
		if err := q.Enqueue(email); err != nil {
			t.Fatalf("Enqueue %d failed: %v", i, err)
		}
		if _, err := q.Approve(email.ID, "human"); err != nil {
			t.Fatalf("Approve %d failed: %v", i, err)
		}
	}
	if !q.IsMonitorMode() {
		t.Fatal("should be in monitor mode after 50 approvals")
	}
}