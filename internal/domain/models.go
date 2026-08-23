package domain

import "time"

type EventStatus string

const (
	EventDraft    EventStatus = "draft"
	EventOpen     EventStatus = "open"
	EventClosed   EventStatus = "closed"
	EventArchived EventStatus = "archived"
)

type RegistrationStatus string

const (
	RegistrationPending   RegistrationStatus = "pending"
	RegistrationApproved  RegistrationStatus = "approved"
	RegistrationRejected  RegistrationStatus = "rejected"
	RegistrationCancelled RegistrationStatus = "cancelled"
)

type ReviewDecision string

const (
	DecisionApprove ReviewDecision = "approve"
	DecisionReject  ReviewDecision = "reject"
)

type BatchStep string

const (
	BatchCreated   BatchStep = "created"
	BatchPrepared  BatchStep = "prepared"
	BatchRanked    BatchStep = "ranked"
	BatchConfirmed BatchStep = "confirmed"
	BatchArchived  BatchStep = "archived"
	BatchCancelled BatchStep = "cancelled"
)

type Event struct {
	ID                  string      `json:"id"`
	Name                string      `json:"name"`
	RegistrationOpensAt time.Time   `json:"registration_opens_at"`
	Deadline            time.Time   `json:"deadline"`
	Capacity            int         `json:"capacity"`
	Status              EventStatus `json:"status"`
	CreatedAt           time.Time   `json:"created_at"`
	UpdatedAt           time.Time   `json:"updated_at"`
}

type Registration struct {
	ID          string             `json:"id"`
	EventID     string             `json:"event_id"`
	Name        string             `json:"name"`
	Phone       string             `json:"phone"`
	Note        string             `json:"note"`
	Score       int                `json:"score"`
	SubmittedAt time.Time          `json:"submitted_at"`
	Status      RegistrationStatus `json:"status"`
	ReviewedBy  string             `json:"reviewed_by"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type ReviewRecord struct {
	ID             string         `json:"id"`
	EventID        string         `json:"event_id"`
	RegistrationID string         `json:"registration_id"`
	Decision       ReviewDecision `json:"decision"`
	Reason         string         `json:"reason"`
	Reviewer       string         `json:"reviewer"`
	ReviewedAt     time.Time      `json:"reviewed_at"`
}

type AuditEvent struct {
	ID             string    `json:"id"`
	EventID        string    `json:"event_id"`
	RegistrationID string    `json:"registration_id"`
	Action         string    `json:"action"`
	Actor          string    `json:"actor"`
	Detail         string    `json:"detail"`
	OccurredAt     time.Time `json:"occurred_at"`
}

type ExportJob struct {
	ID          string    `json:"id"`
	EventID     string    `json:"event_id"`
	RequestedBy string    `json:"requested_by"`
	Format      string    `json:"format"`
	RowCount    int       `json:"row_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type Batch struct {
	ID              string    `json:"id"`
	EventID         string    `json:"event_id"`
	Step            BatchStep `json:"step"`
	Cursor          int       `json:"cursor"`
	Total           int       `json:"total"`
	CancelRequested bool      `json:"cancel_requested"`
	UpdatedAt       time.Time `json:"updated_at"`
}
