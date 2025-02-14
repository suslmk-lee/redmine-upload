package model

import "time"

// Issue represents an issue in the Redmine system
type Issue struct {
	ID             int       `json:"id"`
	LoginID        string    `json:"login"`
	JobID          int       `json:"job_id"`
	Status         string    `json:"status"`
	StatusID       int       `json:"status_id"`
	Assignee       string    `json:"assignee"`
	StartDate      time.Time `json:"start_date"`
	DueDate        time.Time `json:"due_date"`
	DoneRatio      int       `json:"done_ratio"`
	EstimatedHours float64   `json:"estimated_hours"`
	Priority       string    `json:"priority"`
	Author         string    `json:"author"`
	Email          string    `json:"email"`
	Subject        string    `json:"subject"`
	Description    string    `json:"description"`
	Commentor      string    `json:"commentor"`
	Property       string    `json:"property"`
	PropKey        string    `json:"prop_key"`
	OldValue       string    `json:"old_value"`
	Value          string    `json:"value"`
	Notes          string    `json:"notes"`
	CreatedOn      time.Time `json:"created_on"`
	UpdatedOn      time.Time `json:"updated_on"`
}
