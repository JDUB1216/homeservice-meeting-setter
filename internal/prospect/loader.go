package prospect

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/NeuroLift-Technologies/homeservice-meeting-setter/internal/model"
)

// Axis values for validating prospect load input.
const (
	ErrNotProspect = "missing id or business_name"
	ErrBadVertical = "vertical must be hvac, plumbing or roofing"
)

// Load reads prospect records from a file path. Supported formats are JSON
// (a single object, an array, or a {"prospects":[...]} wrapper) and CSV with
// the header documented in LoadCSV.
func Load(path string) ([]model.Prospect, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	raw, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	text := strings.TrimSpace(string(raw))
	if text == "" {
		return nil, fmt.Errorf("prospect file %s is empty", path)
	}
	if strings.HasSuffix(strings.ToLower(path), ".csv") {
		// Re-open for the csv reader on the specific file handle unchanged.
		r := csv.NewReader(strings.NewReader(text))
		return decodeCSV(r)
	}

	var dec struct {
		Prospects []model.Prospect `json:"prospects"`
	}
	if err := json.Unmarshal(raw, &dec); err == nil && dec.Prospects != nil {
		return validateAll(dec.Prospects)
	}

	var single model.Prospect
	if err := json.Unmarshal(raw, &single); err == nil && single.ID != "" {
		return validateAll([]model.Prospect{single})
	}

	var many []model.Prospect
	if err := json.Unmarshal(raw, &many); err == nil {
		return validateAll(many)
	}

	return nil, fmt.Errorf("could not parse %s as prospect JSON or CSV", path)
}

// LoadCSV parses prospects from a CSV reader.
//
// Header:
//
//	id,business_name,vertical,city,state,website,owner_name,owner_email,
//	gbp_rating,gbp_review_count,gbp_review_response_count,gbp_photo_count,
//	gbp_video_count,gbp_has_website,gbp_services,gbp_business_hours
//
// gbp_services is a pipe-separated list (CSV-safe without quoting).
func LoadCSV(r *csv.Reader) ([]model.Prospect, error) {
	return decodeCSV(r)
}

func decodeCSV(r *csv.Reader) ([]model.Prospect, error) {
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("csv must have a header and at least one row")
	}
	header := records[0]
	col := func(name string) int {
		for i, h := range header {
			if strings.EqualFold(strings.TrimSpace(h), name) {
				return i
			}
		}
		return -1
	}
	idx := map[string]int{
		"id":                          col("id"),
		"business_name":               col("business_name"),
		"vertical":                    col("vertical"),
		"city":                        col("city"),
		"state":                       col("state"),
		"website":                     col("website"),
		"owner_name":                  col("owner_name"),
		"owner_email":                 col("owner_email"),
		"gbp_rating":                  col("gbp_rating"),
		"gbp_review_count":            col("gbp_review_count"),
		"gbp_review_response_count":   col("gbp_review_response_count"),
		"gbp_photo_count":             col("gbp_photo_count"),
		"gbp_video_count":             col("gbp_video_count"),
		"gbp_has_website":             col("gbp_has_website"),
		"gbp_services":                col("gbp_services"),
		"gbp_business_hours":          col("gbp_business_hours"),
	}
	get := func(row []string, key string) string {
		i := idx[key]
		if i < 0 || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}
	getInt := func(row []string, key string) int {
		v, _ := strconv.Atoi(get(row, key))
		return v
	}
	getFloat := func(row []string, key string) float64 {
		v, _ := strconv.ParseFloat(get(row, key), 64)
		return v
	}
	getBool := func(row []string, key string) bool {
		s := strings.ToLower(get(row, key))
		return s == "true" || s == "1" || s == "yes"
	}

	var out []model.Prospect
	for _, row := range records[1:] {
		respRate := 0.0
		rc := getInt(row, "gbp_review_count")
		if rr := getInt(row, "gbp_review_response_count"); rr > 0 && rc > 0 {
			respRate = float64(rr) / float64(rc)
		}
		var services []string
		if s := get(row, "gbp_services"); s != "" {
			services = strings.Split(s, "|")
		}
		out = append(out, model.Prospect{
			ID:           get(row, "id"),
			BusinessName: get(row, "business_name"),
			Vertical:     model.Vertical(get(row, "vertical")),
			City:         get(row, "city"),
			State:        get(row, "state"),
			Website:      get(row, "website"),
			OwnerName:    get(row, "owner_name"),
			OwnerEmail:   get(row, "owner_email"),
			GBP: model.GBPProfile{
				Rating:              getFloat(row, "gbp_rating"),
				ReviewCount:         rc,
				ReviewResponseCount: getInt(row, "gbp_review_response_count"),
				ReviewResponseRate:  respRate,
				PhotoCount:          getInt(row, "gbp_photo_count"),
				VideoCount:          getInt(row, "gbp_video_count"),
				HasWebsite:          getBool(row, "gbp_has_website"),
				Services:            services,
				BusinessHours:       get(row, "gbp_business_hours"),
			},
		})
	}
	return validateAll(out)
}

func validateAll(in []model.Prospect) ([]model.Prospect, error) {
	for _, p := range in {
		if strings.TrimSpace(p.ID) == "" || strings.TrimSpace(p.BusinessName) == "" {
			return nil, fmt.Errorf(ErrNotProspect)
		}
		switch p.Vertical {
		case model.VerticalHVAC, model.VerticalPlumbing, model.VerticalRoofing:
		default:
			return nil, fmt.Errorf("%s: %q", ErrBadVertical, p.Vertical)
		}
	}
	return in, nil
}