package prospect

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_JSONArray(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prospects.json")
	data := `[{"id":"p1","business_name":"Test HVAC","vertical":"hvac","city":"Austin","gbp":{"rating":4.1,"review_count":10}}]`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	prospects, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(prospects) != 1 {
		t.Fatalf("expected 1 prospect, got %d", len(prospects))
	}
	if prospects[0].BusinessName != "Test HVAC" {
		t.Fatalf("expected Test HVAC, got %s", prospects[0].BusinessName)
	}
}

func TestLoad_Wrapper(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prospects.json")
	data := `{"prospects":[{"id":"p1","business_name":"Wrapped Co","vertical":"plumbing","city":"Denver","gbp":{"rating":4.0,"review_count":5}}]}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	prospects, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(prospects) != 1 {
		t.Fatalf("expected 1 prospect, got %d", len(prospects))
	}
}

func TestLoad_SingleObject(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prospects.json")
	data := `{"id":"p1","business_name":"Single Co","vertical":"roofing","city":"Seattle","gbp":{"rating":3.5,"review_count":8}}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	prospects, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(prospects) != 1 {
		t.Fatalf("expected 1 prospect, got %d", len(prospects))
	}
}

func TestLoad_CSV(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prospects.csv")
	data := "id,business_name,vertical,city,state,gbp_rating,gbp_review_count,gbp_photo_count\np1,Test HVAC,hvac,Austin,TX,4.1,10,5\n"
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	prospects, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(prospects) != 1 {
		t.Fatalf("expected 1 prospect, got %d", len(prospects))
	}
	if prospects[0].BusinessName != "Test HVAC" {
		t.Fatalf("expected Test HVAC, got %s", prospects[0].BusinessName)
	}
	if prospects[0].GBP.Rating != 4.1 {
		t.Fatalf("expected rating 4.1, got %f", prospects[0].GBP.Rating)
	}
}

func TestLoad_Empty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for empty file")
	}
}

func TestLoad_MissingID(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "noid.json")
	data := `[{"business_name":"No ID","vertical":"hvac","city":"Austin"}]`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing ID")
	}
	if !strings.Contains(err.Error(), "missing id or business_name") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoad_BadVertical(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "badvertical.json")
	data := `[{"id":"p1","business_name":"Bad","vertical":"unknown","city":"Austin"}]`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for bad vertical")
	}
	if !strings.Contains(err.Error(), "vertical must be") {
		t.Fatalf("unexpected error: %v", err)
	}
}