package model

import "time"

// Issue represents an issue in the Redmine system
type Issue struct {
	ID             int       `json:"id"`
	Status         string    `json:"status"`
	Assignee       string    `json:"assignee"`
	StartDate      time.Time `json:"start_date"`
	DueDate        time.Time `json:"due_date"`
	DoneRatio      int       `json:"done_ratio"`
	EstimatedHours float64   `json:"estimated_hours"`
	Priority       string    `json:"priority"`
	Author         string    `json:"author"`
	Subject        string    `json:"subject"`
	Description    string    `json:"description"`
	Commentor      string    `json:"commentor"`
	Notes          string    `json:"notes"`
	CreatedOn      time.Time `json:"created_on"`
}
