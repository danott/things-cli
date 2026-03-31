package things

import "time"

type Status string

const (
	StatusOpen      Status = "open"
	StatusCompleted Status = "completed"
	StatusCanceled  Status = "canceled"
)

type Todo struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Status           Status          `json:"status"`
	Notes            string          `json:"notes,omitempty"`
	TagNames         string          `json:"tagNames,omitempty"`
	DueDate          *time.Time      `json:"dueDate,omitempty"`
	ActivationDate   *time.Time      `json:"activationDate,omitempty"`
	CreationDate     *time.Time      `json:"creationDate,omitempty"`
	ModificationDate *time.Time      `json:"modificationDate,omitempty"`
	CompletionDate   *time.Time      `json:"completionDate,omitempty"`
	CancellationDate *time.Time      `json:"cancellationDate,omitempty"`
	ProjectID        string          `json:"projectID,omitempty"`
	ProjectName      string          `json:"projectName,omitempty"`
	AreaID           string          `json:"areaID,omitempty"`
	AreaName         string          `json:"areaName,omitempty"`
	StartBucket      int             `json:"startBucket,omitempty"` // 0=morning, 1=evening (Today view only)
	ChecklistItems   []ChecklistItem `json:"checklistItems,omitempty"`
}

// LogbookOptions controls filtering and pagination for logbook queries.
type LogbookOptions struct {
	Limit  int
	Offset int
	Since  *time.Time
	Until  *time.Time
}

type Project struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   Status `json:"status"`
	Notes    string `json:"notes,omitempty"`
	AreaName string `json:"areaName,omitempty"`
}

type Area struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ChecklistItem struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status Status `json:"status"`
}

type Tag struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ParentTag string `json:"parentTag,omitempty"`
}

type List struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
