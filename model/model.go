package model

import "time"

// Issue represents an issue in the Redmine system
type Issue struct {
	ID          int       `json:"id"`
	Subject     string    `json:"subject"`
	Description string    `json:"description"`
	CreatedOn   time.Time `json:"created_on"`
	UpdatedOn   time.Time `json:"updated_on"`
}
