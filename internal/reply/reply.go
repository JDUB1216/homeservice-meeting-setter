// Package reply performs lightweight triage on inbound replies: classification
// (positive / neutral / negative / out-of-office) and, for positive replies,
// a concrete booking-response draft. Replies are imported manually in v1.
package reply

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/NeuroLift-Technologies/homeservice-meeting-setter/internal/model"
)

// Lexicons drive a scoring classifier. Keys are matched case-insensitively
// as substrings of the normalized reply text.
var lexicons = map[model.ReplyClassification][]string{
	model.ReplyPositive: {
		"interested", "sounds good", "let's do it", "lets do", "yes", "sure",
		"book", "schedule", "available", "thursday", "friday", "monday",
		"tuesday", "wednesday", "what times", "free", "15 minutes",
		"15-minute", "go ahead", "let's talk", "let's chat", "count me in",
		"would love", "sign me up", "set it up", "call me", "give me a call",
	},
	model.ReplyNegative: {
		"no thank", "no thanks", "not interested", "unsubscribe",
		"remove me", "stop contacting", "don't contact", "do not contact",
		"take me off", "scam", "spam", "not for us", "not for me",
		"waste of time", "leave me alone",
	},
	model.ReplyOutOfOffice: {
		"out of office", "out of the office", "on vacation", "on holiday",
		"away from", "will return", "autoreply", "auto reply", "auto-reply",
		"on leave", "ooo", "currently out",
	},
}

// Classify returns a classification and a confidence in [0,1].
func Classify(r model.Reply) (model.ReplyClassification, float64) {
	text := strings.ToLower(r.Subject + " " + r.Body)
	text = normalize(text)

	scores := map[model.ReplyClassification]float64{}
	best, bestScore := model.ReplyNeutral, 0.0
	for cls, keys := range lexicons {
		respOccur := 0
		for _, k := range keys {
			nk := strings.Count(text, k)
			if !strings.HasPrefix(k, "ooo") && nk > 0 {
				respOccur++
			}
			scores[cls] += float64(nk)
			if cls == model.ReplyOutOfOffice && nk > 0 {
				scores[cls] += 1.5
			}
		}
		if scores[cls] > bestScore {
			best, bestScore = cls, scores[cls]
		}
		_ = respOccur
	}

	// An out-of-office hit dominates because it is structurally unambiguous.
	if scores[model.ReplyOutOfOffice] > 0 {
		best, bestScore = model.ReplyOutOfOffice, min(0.98, 0.6+scores[model.ReplyOutOfOffice]*0.1)
	}
	// Negativity wins over positive language to avoid overbooking someone
	// who said "thanks but no thanks" (contains both "thanks" and "no").
	if scores[model.ReplyNegative] > 0 && scores[model.ReplyNegative] >= scores[model.ReplyPositive] {
		best, bestScore = model.ReplyNegative, min(0.98, 0.5+scores[model.ReplyNegative]*0.12)
	}
	if bestScore == 0 {
		return model.ReplyNeutral, 0.25
	}
	return best, min(0.99, 0.4+bestScore*0.12)
}

// DraftResponse builds a booking reply for a positive reply. prospects is a
// lookup of prospect by ID for personalization; when absent a generic but
// still concrete draft is produced.
func DraftResponse(r model.Reply, ownerName, businessName, bookingLink string) string {
	if ownerName == "" {
		ownerName = "there"
	} else {
		ownerName = strings.Split(ownerName, " ")[0]
	}
	if businessName == "" {
		businessName = "your business"
	}
	if bookingLink == "" {
		bookingLink = "https://cal.homeservice.example/maya"
	}
	return fmt.Sprintf(`Great to hear, %s!

Here's the rundown for %s's Google-presence review: it takes about 15 minutes, over the phone, and I'll share three concrete, easy fixes specific to your profile.

Two options this week:
- Thursday at 10:00am or 3:30pm
- Friday at 11:00am

Pick whichever suits — book straight into a slot here: %s

If those times don't work, reply with two that do and I'll confirm.

Best,
Maya (HomeService Meeting Setter)`, ownerName, businessName, bookingLink)
}

// ImportLoad reads replies from a JSON file: an array, a single object, or a
// {"replies":[...]} wrapper.
func ImportLoad(path string) ([]model.Reply, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var list []model.Reply
	if err := json.Unmarshal(data, &list); err == nil && len(list) > 0 {
		return list, nil
	}
	var one model.Reply
	if err := json.Unmarshal(data, &one); err == nil && one.From != "" {
		return []model.Reply{one}, nil
	}
	var wrapped struct {
		Replies []model.Reply `json:"replies"`
	}
	if err := json.Unmarshal(data, &wrapped); err == nil {
		return wrapped.Replies, nil
	}
	return nil, fmt.Errorf("could not parse %s as replies JSON", path)
}

// Save persists processed replies.
func Save(replies []model.Reply, path string) error {
	data, err := json.MarshalIndent(replies, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func normalize(s string) string {
	s = strings.ToLower(s)
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// EnsureReverseTime is a helper so callers always have a non-zero timestamp.
func EnsureReverseTime(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now().UTC()
	}
	return t
}