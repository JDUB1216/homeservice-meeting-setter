// Package model defines the core data structures of the HomeService
// Meeting Setter agent: prospects, generated emails, approval queue items,
// send records and inbox replies.
package model

import "time"

// Vertical is the home-service trade a prospect operates in.
type Vertical string

const (
	VerticalHVAC     Vertical = "hvac"
	VerticalPlumbing Vertical = "plumbing"
	VerticalRoofing  Vertical = "roofing"
)

// Prospect is a local home-services business we want to book a discovery
// call with. GBP holds a snapshot of their Google Business Profile data.
type Prospect struct {
	ID           string     `json:"id"`
	BusinessName string     `json:"business_name"`
	Vertical     Vertical   `json:"vertical"`
	City         string     `json:"city"`
	State        string     `json:"state"`
	Website      string     `json:"website,omitempty"`
	OwnerName    string     `json:"owner_name,omitempty"`
	OwnerEmail   string     `json:"owner_email"`
	GBP          GBPProfile `json:"gbp"`
}

// GBPProfile is a snapshot of the data we collect from a business's Google
// Business Profile during research.
type GBPProfile struct {
	URL                 string   `json:"url,omitempty"`
	Rating              float64  `json:"rating"`
	ReviewCount         int      `json:"review_count"`
	ReviewResponseCount int      `json:"review_response_count,omitempty"`
	ReviewResponseRate  float64  `json:"review_response_rate"` // 0.0..1.0
	PhotoCount          int      `json:"photo_count"`
	VideoCount          int      `json:"video_count,omitempty"`
	HasWebsite          bool     `json:"has_website,omitempty"`
	Services            []string `json:"services,omitempty"`
	Categories          []string `json:"categories,omitempty"`
	BusinessHours       string   `json:"business_hours,omitempty"`
	RecentReviewMonth   string   `json:"recent_review_month,omitempty"`
}

// Observation is a single specific, factual finding extracted from a
// prospect's Google Business Profile or website. Facts holds the literal
// values that appear in the profile so the specificity checker can prove the
// email references real data.
type Observation struct {
	Kind  string   `json:"kind"`
	Text  string   `json:"text"`
	Facts []string `json:"facts"`
}

// EmailStatus tracks a generated email's lifecycle.
type EmailStatus string

const (
	EmailPendingApproval EmailStatus = "pending_approval"
	EmailApproved        EmailStatus = "approved"
	EmailRejected        EmailStatus = "rejected"
	EmailSent            EmailStatus = "sent"
)

// ASFDKStatus records the governance verdict for a generated email.
type ASFDKStatus string

const (
	ASFDKClean   ASFDKStatus = "clean"     // passed sanitization + output validation
	ASFDKFlagged ASFDKStatus = "flagged"   // SanitizeInput returned non-clean
	ASFDKInvalid ASFDKStatus = "invalid"   // ValidateOutput failed
)

// GeneratedEmail is the output of the email generation stage, after it has
// been run through the ASFDK governance guard.
type GeneratedEmail struct {
	ID           string      `json:"id"`
	ProspectID   string      `json:"prospect_id"`
	Subject      string      `json:"subject"`
	Body         string      `json:"body"`
	Observations []string    `json:"observations"`
	Specificity  float64     `json:"specificity"`
	ASFDKStatus  ASFDKStatus `json:"asfdk_status"`
	ASFDKReason  string      `json:"asfdk_reason,omitempty"`
	Status       EmailStatus `json:"status"`
	ApprovedBy   string      `json:"approved_by,omitempty"`
	ApprovedAt   *time.Time  `json:"approved_at,omitempty"`
	Edited       bool        `json:"edited,omitempty"`
	OriginalBody string      `json:"original_body,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	TimesSent    int         `json:"times_sent"`
}

// ApprovalStatus is the human-review state of an approval queue item.
type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "pending"
	ApprovalApproved ApprovalStatus = "approved"
	ApprovalRejected ApprovalStatus = "rejected"
)

// ApprovalItem wraps a generated email in the human approval queue.
type ApprovalItem struct {
	ID           string         `json:"id"`
	Position     int            `json:"position"`
	Email        GeneratedEmail `json:"email"`
	Status       ApprovalStatus `json:"status"`
	ReviewedBy   string         `json:"reviewed_by,omitempty"`
	ReviewedAt   *time.Time     `json:"reviewed_at,omitempty"`
	RejectReason string         `json:"reject_reason,omitempty"`
}

// SendRecord records one email delivery.
type SendRecord struct {
	ID         string    `json:"id"`
	EmailID    string    `json:"email_id"`
	ProspectID string    `json:"prospect_id"`
	To         string    `json:"to"`
	Subject    string    `json:"subject"`
	Provider   string    `json:"provider"`
	Status     string    `json:"status"` // "sent" | "error"
	MessageID  string    `json:"message_id,omitempty"`
	Error      string    `json:"error,omitempty"`
	SentAt     time.Time `json:"sent_at"`
}

// ReplyClassification is the outcome of lightweight reply triage.
type ReplyClassification string

const (
	ReplyPositive    ReplyClassification = "positive"
	ReplyNeutral     ReplyClassification = "neutral"
	ReplyNegative    ReplyClassification = "negative"
	ReplyOutOfOffice ReplyClassification = "out_of_office"
)

// Reply is an inbound reply (imported manually in v1). For positive replies
// the agent drafts a booking response.
type Reply struct {
	ID               string             `json:"id"`
	From             string             `json:"from"`
	To               string             `json:"to,omitempty"`
	Subject          string             `json:"subject,omitempty"`
	Body             string             `json:"body"`
	InReplyToEmailID string             `json:"in_reply_to_email_id,omitempty"`
	ReceivedAt       time.Time          `json:"received_at"`
	Classification   ReplyClassification `json:"classification"`
	Confidence       float64            `json:"confidence"`
	DraftResponse    string             `json:"draft_response,omitempty"`
	BookingLink      string             `json:"booking_link,omitempty"`
	Responded        bool               `json:"responded,omitempty"`
}