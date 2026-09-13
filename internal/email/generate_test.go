package email

import (
	"strings"
	"testing"

	"github.com/NeuroLift-Technologies/homeservice-meeting-setter/internal/model"
	"github.com/NeuroLift-Technologies/homeservice-meeting-setter/internal/prospect"
)

func TestGenerate_SpecificityPass(t *testing.T) {
	g := NewGenerator(Config{MinSpecificity: 0.6})
	p := model.Prospect{
		ID:           "p1",
		BusinessName: "PipeReady Plumbing",
		Vertical:     model.VerticalPlumbing,
		City:         "Austin",
		OwnerName:    "John",
		GBP: model.GBPProfile{
			Rating:      4.1,
			ReviewCount: 23,
		},
	}
	obs := prospect.ExtractObservations(p)
	email, ok := g.Generate(p, obs)
	if !ok {
		t.Fatal("expected generation to succeed")
	}
	if email.Subject == "" {
		t.Fatal("expected non-empty subject")
	}
	if email.Body == "" {
		t.Fatal("expected non-empty body")
	}
	if !strings.Contains(email.Body, p.BusinessName) {
		t.Fatalf("body must reference business name: %s", p.BusinessName)
	}
	if email.Specificity < 0.6 {
		t.Fatalf("specificity %.2f below threshold", email.Specificity)
	}
}

func TestGenerate_RejectsLowSpecificity(t *testing.T) {
	g := NewGenerator(Config{MinSpecificity: 0.9})
	p := model.Prospect{
		ID:           "p2",
		BusinessName: "Empty Co",
		Vertical:     model.VerticalHVAC,
		City:         "Portland",
		GBP:          model.GBPProfile{},
	}
	obs := prospect.ExtractObservations(p)
	_, ok := g.Generate(p, obs)
	if ok {
		t.Fatal("expected generation to fail for low specificity")
	}
}

func TestValidateSpecificity(t *testing.T) {
	g := NewGenerator(Config{MinSpecificity: 0.6})
	p := model.Prospect{
		ID:           "p3",
		BusinessName: "Cool HVAC",
		Vertical:     model.VerticalHVAC,
		City:         "Denver",
		GBP: model.GBPProfile{
			Rating:      3.9,
			ReviewCount: 42,
		},
	}
	obs := prospect.ExtractObservations(p)
	if !g.ValidateSpecificity(p, obs, "Hi there, Cool HVAC has 42 Google reviews in Denver.") {
		t.Fatal("expected valid specificity")
	}
}

func TestSubject_Kinds(t *testing.T) {
	g := NewGenerator(Config{})
	tests := []struct {
		kind string
		p    model.Prospect
	}{
		{"photos", model.Prospect{BusinessName: "Test", GBP: model.GBPProfile{PhotoCount: 5}}},
		{"rating", model.Prospect{BusinessName: "Test", GBP: model.GBPProfile{Rating: 4.0, ReviewCount: 10}}},
		{"review_response_rate", model.Prospect{BusinessName: "Test", GBP: model.GBPProfile{ReviewResponseRate: 0.3, ReviewCount: 10}}},
		{"missing_service", model.Prospect{BusinessName: "Test", GBP: model.GBPProfile{Services: []string{"other"}}}},
		{"no_website", model.Prospect{BusinessName: "Test", Website: ""}},
		{"no_videos", model.Prospect{BusinessName: "Test", GBP: model.GBPProfile{ReviewCount: 10, VideoCount: 0}}},
		{"default", model.Prospect{BusinessName: "Test"}},
	}
	for _, tt := range tests {
		obs := prospect.ExtractObservations(tt.p)
		if len(obs) == 0 {
			t.Fatalf("no obs for kind %s", tt.kind)
		}
		subject := g.subject(tt.p, prospect.PickPersonalization(obs))
		if subject == "" {
			t.Fatalf("empty subject for kind %s", tt.kind)
		}
	}
}