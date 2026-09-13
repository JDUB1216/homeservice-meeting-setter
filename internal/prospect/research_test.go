package prospect

import (
	"testing"

	"github.com/NeuroLift-Technologies/homeservice-meeting-setter/internal/model"
)

func TestExtractObservations_Photos(t *testing.T) {
	p := model.Prospect{
		ID:           "p1",
		BusinessName: "Test HVAC",
		Vertical:     model.VerticalHVAC,
		City:         "Austin",
		GBP: model.GBPProfile{
			PhotoCount: 10,
		},
	}
	obs := ExtractObservations(p)
	if len(obs) == 0 {
		t.Fatal("expected observations")
	}
	if obs[0].Kind != "photos" {
		t.Fatalf("expected photos kind, got %s", obs[0].Kind)
	}
	if len(obs[0].Facts) == 0 || obs[0].Facts[0] != "10" {
		t.Fatalf("expected fact '10', got %v", obs[0].Facts)
	}
}

func TestExtractObservations_Rating(t *testing.T) {
	p := model.Prospect{
		ID:           "p2",
		BusinessName: "Cool Plumbing",
		Vertical:     model.VerticalPlumbing,
		City:         "Denver",
		GBP: model.GBPProfile{
			Rating:              4.2,
			ReviewCount:         15,
			ReviewResponseRate:  1.0, // avoid review_response_rate observation
		},
	}
	obs := ExtractObservations(p)
	if len(obs) == 0 {
		t.Fatal("expected observations")
	}
	// Find the rating observation (may not be first if other observations exist)
	found := false
	for _, o := range obs {
		if o.Kind == "rating" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected rating kind in observations, got kinds: %v", kinds(obs))
	}
}

func kinds(obs []model.Observation) []string {
	out := make([]string, len(obs))
	for i, o := range obs {
		out[i] = o.Kind
	}
	return out
}

func TestExtractObservations_MissingService(t *testing.T) {
	p := model.Prospect{
		ID:           "p3",
		BusinessName: "Best Roofing",
		Vertical:     model.VerticalRoofing,
		City:         "Seattle",
		GBP: model.GBPProfile{
			Services: []string{"roof repair"},
		},
	}
	obs := ExtractObservations(p)
	if len(obs) == 0 {
		t.Fatal("expected observations")
	}
	if obs[0].Kind != "missing_service" {
		t.Fatalf("expected missing_service kind, got %s", obs[0].Kind)
	}
}

func TestExtractObservations_Baseline(t *testing.T) {
	// Use a prospect with no specific signals: no photos, no rating issues,
	// full review response rate, website present, all services listed,
	// videos present. This forces the baseline observation.
	p := model.Prospect{
		ID:           "p4",
		BusinessName: "Empty Co",
		Vertical:     model.VerticalHVAC,
		City:         "Portland",
		Website:      "https://example.com",
		GBP: model.GBPProfile{
			PhotoCount:          30,  // >= 25, no photos observation
			Rating:              4.8,  // >= 4.5, no rating observation
			ReviewCount:         5,
			ReviewResponseRate:  1.0,  // >= 0.6, no review_response_rate
			HasWebsite:          true, // has website, no no_website
			Services:            []string{"air conditioning", "ac repair", "heating", "furnace", "emergency service", "maintenance"},
			VideoCount:          1, // has videos, no no_videos
		},
	}
	obs := ExtractObservations(p)
	if len(obs) == 0 {
		t.Fatal("expected observations")
	}
	if obs[0].Kind != "baseline" {
		t.Fatalf("expected baseline kind, got %s (kinds: %v)", obs[0].Kind, kinds(obs))
	}
}

func TestMissingServices(t *testing.T) {
	missing := MissingServices(model.VerticalHVAC, []string{"heating"})
	if len(missing) == 0 {
		t.Fatal("expected missing services")
	}
	for _, m := range missing {
		if m == "heating" {
			t.Fatal("heating should not be missing")
		}
	}
}

func TestPickPersonalization_Priority(t *testing.T) {
	obs := []model.Observation{
		{Kind: "baseline", Text: "baseline", Facts: []string{"5"}},
		{Kind: "rating", Text: "rating", Facts: []string{"4.2", "15"}},
	}
	picked := PickPersonalization(obs)
	if len(picked) == 0 {
		t.Fatal("expected picked observations")
	}
	if picked[0].Kind != "rating" {
		t.Fatalf("expected rating first, got %s", picked[0].Kind)
	}
}

func TestPickPersonalization_MaxTwo(t *testing.T) {
	obs := []model.Observation{
		{Kind: "rating", Text: "r1", Facts: []string{"4.2"}},
		{Kind: "photos", Text: "p1", Facts: []string{"10"}},
		{Kind: "no_website", Text: "w1", Facts: []string{}},
	}
	picked := PickPersonalization(obs)
	if len(picked) > 2 {
		t.Fatalf("expected at most 2, got %d", len(picked))
	}
}