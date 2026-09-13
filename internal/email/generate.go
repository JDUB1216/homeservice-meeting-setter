// Package email generates highly personalized first-touch cold emails for a
// home-services prospect. Generation is strictly template-based (deterministic,
// no LLM call) and is held to an explicit specificity bar: an email that does
// not reference actual facts from the prospect's profile is rejected by the
// generator itself, not just by downstream governance.
package email

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/NeuroLift-Technologies/homeservice-meeting-setter/internal/model"
	"github.com/NeuroLift-Technologies/homeservice-meeting-setter/internal/prospect"
)

// Config controls the generation parameters and the sender persona used for
// the sign-off and CTA.
type Config struct {
	// AgentName is the first-touch signer (e.g. a named SDR persona).
	AgentName string
	// CompanyName appears in the signature block.
	CompanyName string
	// BookingLink is the scheduling URL used in the soft CTA.
	BookingLink string
	// MinSpecificity is the minimum specificity score (0..1) an email must
	// reach before generation is accepted. "specific" means the body carries
	// hard facts from the prospect's profile and names the business.
	MinSpecificity float64
}

func (c Config) withDefaults() Config {
	if c.AgentName == "" {
		c.AgentName = "Maya (HomeService Meeting Setter)"
	}
	if c.CompanyName == "" {
		c.CompanyName = "HomeService Meeting Setter"
	}
	if c.BookingLink == "" {
		c.BookingLink = "https://cal.homeservice.example/maya"
	}
	if c.MinSpecificity <= 0 {
		c.MinSpecificity = 0.6
	}
	return c
}

// Generator composes personalized emails from prospect observations.
type Generator struct {
	cfg Config
}

var (
	reNumber = regexp.MustCompile(`\d`)
	reWord   = regexp.MustCompile(`[a-zA-Z0-9]+`)
)

// NewGenerator returns a Generator with sane defaults for unset fields.
func NewGenerator(cfg Config) *Generator {
	return &Generator{cfg: cfg.withDefaults()}
}

// ExtractObservations is a passthrough to prospect.ExtractObservations so the
// engine can call research and generation through one seam.
func (g *Generator) ExtractObservations(p model.Prospect) []model.Observation {
	return prospect.ExtractObservations(p)
}

// Generate builds a subject/body pair from the prospect's observations.
//
// The returned ok is false when the generator could not produce an email that
// meets the specificity bar for the given observations — the pipeline must
// treat that as a "do not send" outcome and must not fall back to a generic
// message.
func (g *Generator) Generate(p model.Prospect, obs []model.Observation) (mail model.GeneratedEmail, ok bool) {
	picked := prospect.PickPersonalization(obs)
	if len(picked) == 0 {
		return mail, false
	}

	subject := g.subject(p, picked)
	head, tail := g.paragraph(p, picked)
	body := strings.Join([]string{head, tail, g.signoff()}, "\n\n")

	// Enforced specificity: hard facts + business name must survive into the
	// body, otherwise reject.
	score := specificityScore(p, picked, body)
	if score < g.cfg.MinSpecificity || !referencesBusiness(p, body) {
		return mail, false
	}

	obsText := make([]string, 0, len(picked))
	for _, o := range picked {
		obsText = append(obsText, o.Text)
	}

	return model.GeneratedEmail{
		ID:           newID(p.ID),
		ProspectID:   p.ID,
		Subject:      subject,
		Body:         body,
		Observations: obsText,
		Specificity:  score,
		Status:       model.EmailPendingApproval,
		CreatedAt:    time.Now().UTC(),
	}, true
}

// ValidateSpecificity exposes the specificity gate for tests and the ASFDK
// guard: true only when the body names the business and carries a hard fact.
func (g *Generator) ValidateSpecificity(p model.Prospect, obs []model.Observation, body string) bool {
	picked := prospect.PickPersonalization(obs)
	if len(picked) == 0 {
		return false
	}
	return specificityScore(p, picked, body) >= g.cfg.MinSpecificity && referencesBusiness(p, body)
}

func (g *Generator) subject(p model.Prospect, picked []model.Observation) string {
	switch picked[0].Kind {
	case "photos":
		return fmt.Sprintf("Quick note about %s's Google presence", p.BusinessName)
	case "rating":
		return fmt.Sprintf("A thought on %s's Google rating", p.BusinessName)
	case "review_response_rate":
		return fmt.Sprintf("%s on Google — a quick observation", p.BusinessName)
	case "missing_service":
		return fmt.Sprintf("A small gap on %s's Google listing", p.BusinessName)
	case "no_website":
		return fmt.Sprintf("Your Google profile has no website link", p.BusinessName)
	case "no_videos":
		return fmt.Sprintf("Noticed %s has no videos on Google", p.BusinessName)
	default:
		return fmt.Sprintf("Quick note about %s", p.BusinessName)
	}
}

// paragraph returns (observation, implication) rendered for the picked
// observations. Facts are injected verbatim from obs.Facts.
func (g *Generator) paragraph(p model.Prospect, picked []model.Observation) (string, string) {
	facts := []string{}
	for _, o := range picked {
		facts = append(facts, o.Facts...)
	}
	get := func(i int) string {
		if i < len(facts) {
			return facts[i]
		}
		return ""
	}

	head := factSentence(p, picked[0], get, 0)
	imp := implication(p.BusinessName, picked[0], get, 0)

	if len(picked) == 2 {
		head += " " + factSentence(p, picked[1], get, len(picked[0].Facts))
		imp += " " + implication(p.BusinessName, picked[1], get, len(picked[0].Facts))
	}
	return head, imp
}

func factSentence(p model.Prospect, o model.Observation, get func(int) string, base int) string {
	switch o.Kind {
	case "photos":
		return fmt.Sprintf("Hi %s, I came across %s's Google Business Profile and noticed it lists only %s photo%s.",
			greet(p.OwnerName), p.BusinessName, get(base), pluralFrom(get(base)))
	case "rating":
		return fmt.Sprintf("Hi %s, I noticed %s currently shows a %s-star Google rating across %s reviews.",
			greet(p.OwnerName), p.BusinessName, get(base), get(base+1))
	case "review_response_rate":
		return fmt.Sprintf("Hi %s, scanning %s's Google Business Profile, it looks like you've responded to roughly %s%% of your %s reviews.",
			greet(p.OwnerName), p.BusinessName, get(base), get(base+1))
	case "missing_service":
		return fmt.Sprintf("Hi %s, quick observation: %s doesn't currently list %s on its Google Business Profile.",
			greet(p.OwnerName), p.BusinessName, strings.Join(o.Facts[:min(2, len(o.Facts))], " or "))
	case "no_website":
		return fmt.Sprintf("Hi %s, %s has a solid Google presence in %s, but the profile links to no website.",
			greet(p.OwnerName), p.BusinessName, p.City)
	case "no_videos":
		return fmt.Sprintf("Hi %s, %s has %s Google reviews, but no video content on its profile.",
			greet(p.OwnerName), p.BusinessName, get(base))
	default:
		if n := get(base); n != "" {
			return fmt.Sprintf("Hi %s, %s has %s Google reviews — a healthy amount of local signal.",
				greet(p.OwnerName), p.BusinessName, n)
		}
		return fmt.Sprintf("Hi %s, %s has a Google Business Profile in %s.",
			greet(p.OwnerName), p.BusinessName, p.City)
	}
}

func implication(business string, o model.Observation, get func(int) string, base int) string {
	switch o.Kind {
	case "photos":
		return fmt.Sprintf("Profiles with 25+ photos tend to receive measurably more profile clicks and calls than ones with a handful — and adding photos is free and takes minutes. A quick pass through that %s photo set would close the gap.",
			busybody(business))
	case "rating":
		return "At that review volume, every single review visibly moves the average — one or two more 5-star reviews would meaningfully change what new customers see next to the listing."
	case "review_response_rate":
		return "Replying to reviews is one of the cheapest local-rank and trust improvements a service business can make, and your current reply rate leaves a lot of that on the table."
	case "missing_service":
		return fmt.Sprintf("That term is one of the most-searched phrases in your category, and competitors who list it get found for it. Adding it takes under a minute on the profile.")
	case "no_website":
		return "Without a website link, Google has far less to work with when deciding which business answers nearby searches — and it often surfaces a competitor instead."
	case "no_videos":
		return "Video listings engage far better than text alone, and at your review count the profile is already getting visibility worth pairing with a short walkthrough clip."
	default:
		return "It's worth spending 15 minutes making sure that review momentum converts into booked work."
	}
}

func (g *Generator) signoff() string {
	closing := "If it's helpful, I can walk you through a quick review of your Google presence — a few concrete, easy fixes specific to your profile, in about 15 minutes. No cost, no obligation."
	return fmt.Sprintf("%s\n\nBest,\n%s\n%s\n%s",
		closing, g.cfg.AgentName, g.cfg.CompanyName, g.cfg.BookingLink)
}

// specificityScore 0..1: how much of the email is provably grounded in the
// prospect's own data. It rewards numeric facts that appear verbatim in the
// body, plus city/vertical mentions. Pure boilerplate scores 0.
func specificityScore(p model.Prospect, picked []model.Observation, body string) float64 {
	lower := strings.ToLower(body)
	hit := 0.0
	weight := 0.0

	for _, o := range picked {
		for _, f := range o.Facts {
			weight += 1
			if reNumber.MatchString(f) && strings.Contains(lower, strings.ToLower(f)) {
				hit += 1
			} else if f == "0" && strings.Contains(lower, "google") {
				hit += 0.5
			}
		}
	}

	// Business name and city are positional anchors.
	weight += 1
	if strings.Contains(lower, strings.ToLower(p.BusinessName)) {
		hit += 1
	}
	weight += 0.5
	if strings.Contains(lower, strings.ToLower(p.City)) {
		hit += 0.5
	}
	// Vertical word anchors (e.g. "ac repair") found in the body.
	weight += 1
	if strings.Contains(lower, strings.ToLower(string(p.Vertical))) {
		hit += 0.5
	}

	if weight == 0 {
		return 0
	}
	return hit / weight
}

// referencesBusiness enforces that the email is not a broadcast: it must name
// the actual business.
func referencesBusiness(p model.Prospect, body string) bool {
	return strings.Contains(strings.ToLower(body), strings.ToLower(p.BusinessName))
}

func greet(owner string) string {
	if owner != "" {
		return fmt.Sprintf("%s (%s)", owner, "owner")
	}
	return "there"
}

func pluralFrom(s string) string {
	if s == "" || s == "1" {
		return ""
	}
	return "s"
}

func busybody(name string) string {
	return name
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// newID allocates a stable, readable ID for generated emails.
func newID(prospectID string) string {
	return fmt.Sprintf("em_%s_%d", prospectID, time.Now().UTC().UnixNano())
}