// Package prospect loads prospect records and extracts the specific,
// factual observations about a business's Google Business Profile or website
// that drive email personalization.
package prospect

import (
	"fmt"
	"sort"
	"strings"

	"github.com/NeuroLift-Technologies/homeservice-meeting-setter/internal/model"
)

// verticalServices is the set of services we expect a well-maintained
// profile for each vertical to at least mention. Missing ones are a
// personalization hook ("you didn't list an emergency service").
var verticalServices = map[model.Vertical][]string{
	model.VerticalHVAC:     {"air conditioning", "ac repair", "heating", "furnace", "emergency service", "maintenance"},
	model.VerticalPlumbing: {"drain cleaning", "water heater", "leak repair", "emergency plumbing", "pipe repair"},
	model.VerticalRoofing:  {"roof repair", "roof replacement", "storm damage", "gutter", "roof inspection"},
}

// ExtractObservations derives 1..N concrete observations from a prospect's
// profile data. Every observation carries the literal facts it is based on.
// If nothing specific is found a baseline observation about review volume is
// produced so the pipeline never falls back to an unpersonalized email.
func ExtractObservations(p model.Prospect) []model.Observation {
	var obs []model.Observation
	add := func(kind, text string, facts ...string) {
		obs = append(obs, model.Observation{Kind: kind, Text: text, Facts: facts})
	}

	g := p.GBP

	// Photo count is a strong personalization signal and easy to fix.
	if g.PhotoCount > 0 && g.PhotoCount < 25 {
		add("photos",
			fmt.Sprintf("%s's Google Business Profile shows only %d photo%s", p.BusinessName, g.PhotoCount, plural(g.PhotoCount)),
			fmt.Sprintf("%d", g.PhotoCount))
	}

	// Review response rate shows whether they reply to reviews at all.
	if g.ReviewCount >= 5 && g.ReviewResponseRate >= 0 && g.ReviewResponseRate < 0.6 {
		add("review_response_rate",
			fmt.Sprintf("%s has responded to roughly %d%% of its %d Google reviews", p.BusinessName,
				int(g.ReviewResponseRate*100), g.ReviewCount),
			fmt.Sprintf("%d", int(g.ReviewResponseRate*100)), fmt.Sprintf("%d", g.ReviewCount))
	}

	// A middling rating is a specific, tractable opportunity.
	if g.Rating > 0 && g.Rating < 4.5 && g.ReviewCount >= 10 {
		add("rating",
			fmt.Sprintf("%s holds a %s-star Google rating across %d reviews", p.BusinessName,
				trimZero(g.Rating), g.ReviewCount),
			trimZero(g.Rating), fmt.Sprintf("%d", g.ReviewCount))
	}

	// Missing website while the profile gets traffic.
	if p.Website == "" && !g.HasWebsite && (g.ReviewCount > 0 || g.PhotoCount > 0) {
		add("no_website",
			fmt.Sprintf("%s's Google profile links to no website", p.BusinessName))
	}

	// Missing services for the vertical.
	if missing := MissingServices(p.Vertical, g.Services); len(missing) > 0 {
		top := missing
		if len(top) > 2 {
			top = top[:2]
		}
		add("missing_service",
			fmt.Sprintf("%s does not list %s on its Google Business Profile", p.BusinessName, joinQuoteList(top)),
			missing...)
	}

	// Profile with reviews but no videos: video listings are rare and
	// increase engagement.
	if g.ReviewCount >= 5 && g.VideoCount == 0 {
		add("no_videos",
			fmt.Sprintf("%s has %d Google reviews but no videos on its profile", p.BusinessName, g.ReviewCount),
			fmt.Sprintf("%d", g.ReviewCount))
	}

	// Baseline so we never send a truly generic email.
	if len(obs) == 0 {
		if g.ReviewCount > 0 {
			add("baseline",
				fmt.Sprintf("%s has %d Google review%s", p.BusinessName, g.ReviewCount, plural(g.ReviewCount)),
				fmt.Sprintf("%d", g.ReviewCount))
		} else {
			add("baseline",
				fmt.Sprintf("%s has a Google Business Profile in %s", p.BusinessName, p.City),
				fmt.Sprintf("%d", 0))
		}
	}

	// Prefer observations that carry hard numbers over the baseline.
	sort.SliceStable(obs, func(i, j int) bool {
		return obs[i].Kind != "baseline" && obs[j].Kind == "baseline"
	})
	return obs
}

// MissingServices returns the vertical's expected services that are not
// present (case-insensitive, subject match) in the provided list.
func MissingServices(v model.Vertical, provided []string) []string {
	expected := verticalServices[v]
	if len(expected) == 0 {
		return nil
	}
	has := func(supplied []string, need string) bool {
		for _, s := range supplied {
			if strings.Contains(strings.ToLower(s), need) {
				return true
			}
		}
		return false
	}
	var missing []string
	for _, need := range expected {
		if !has(provided, need) {
			missing = append(missing, need)
		}
	}
	return missing
}

// PickPersonalization selects the 1-2 observations best suited to open an
// email: rated by kind priority, then by presence of numeric facts.
func PickPersonalization(obs []model.Observation) []model.Observation {
	priority := map[string]int{
		"rating":              0,
		"review_response_rate": 1,
		"photos":              2,
		"missing_service":     3,
		"no_website":          4,
		"no_videos":           5,
		"baseline":            6,
	}
	out := append([]model.Observation(nil), obs...)
	sort.SliceStable(out, func(i, j int) bool {
		pi, pj := priority[out[i].Kind], priority[out[j].Kind]
		if pi != pj {
			return pi < pj
		}
		return len(out[i].Facts) > len(out[j].Facts)
	})
	if len(out) > 2 {
		out = out[:2]
	}
	return out
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// trimZero renders a rating without trailing zeros ("4.5", "4.1").
func trimZero(v float64) string {
	s := fmt.Sprintf("%.1f", v)
	return strings.TrimSuffix(s, ".0")
}

func joinQuoteList(items []string) string {
	quoted := make([]string, len(items))
	for i, s := range items {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return strings.Join(quoted, " or ")
}