package database

import (
	"database/sql"
	"fmt"
	"redmine-upload/common"
	"redmine-upload/model"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// ConnectDB connects to the MySQL database
func ConnectDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func FetchImminentIssue(db *sql.DB) ([]model.Issue, error) {
	var issues []model.Issue

	query := `
         SELECT i.id as 'job_id', i.status_id, is2.name, u.firstname, u.lastname, u.login, i.start_date, i.due_date, 
                i.done_ratio, i.estimated_hours,
			(SELECT e.name FROM bitnami_redmine.enumerations e WHERE e.type = 'IssuePriority' AND i.priority_id = e.id) AS priority,
			(SELECT b.firstname FROM bitnami_redmine.users b WHERE i.author_id = b.id) AS author,
			(select ea.address from bitnami_redmine.email_addresses ea where u.id = ea.user_id) as email,
			i.subject, i.description, i.updated_on 
		  FROM bitnami_redmine.issues i
		  JOIN bitnami_redmine.issue_statuses is2 ON i.status_id = is2.id
		  JOIN bitnami_redmine.users u ON i.assigned_to_id = u.id
		 WHERE i.due_date = ?
		   AND i.done_ratio <> 100
		ORDER BY i.updated_on  desc`

	toDayDate := common.TodayDate("2006-01-02")
	rows, err := db.Query(query, toDayDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var issue model.Issue
		var assigneeFirstName, assigneeLastName, commentorFirstName sql.NullString
		var estimatedHours sql.NullFloat64
		var dueDate sql.NullTime
		if err := rows.Scan(
			&issue.JobID, &issue.StatusID, &issue.Status, &assigneeFirstName, &assigneeLastName, &issue.LoginID, &issue.StartDate, &dueDate,
			&issue.DoneRatio, &estimatedHours,
			&issue.Priority,
			&issue.Author,
			&issue.Email,
			&issue.Subject, &issue.Description, &issue.UpdatedOn,
		); err != nil {
			return nil, err
		}
		issue.Assignee = fmt.Sprintf("%s %s", assigneeFirstName.String, assigneeLastName.String)
		issue.Commentor = commentorFirstName.String
		if estimatedHours.Valid {
			issue.EstimatedHours = estimatedHours.Float64
		} else {
			issue.EstimatedHours = 0
		}
		if dueDate.Valid {
			issue.DueDate = dueDate.Time
		} else {
			issue.DueDate = time.Time{}
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// FetchNewIssues fetches new issues from the MySQL database
func FetchNewIssues(db *sql.DB, lastChecked time.Time) ([]model.Issue, error) {
	formattedTime := lastChecked.Format("2006-01-02 15:04:05")
	query := `
        SELECT j.id, i.id as 'job_id', is2.name, u.firstname, u.lastname, i.start_date, i.due_date, i.done_ratio, i.estimated_hours,
        (SELECT e.name FROM bitnami_redmine.enumerations e WHERE e.type = 'IssuePriority' AND i.priority_id = e.id) AS priority,
        (SELECT b.firstname FROM bitnami_redmine.users b WHERE i.author_id = b.id) AS author,
        i.subject, i.description,
        (SELECT b.firstname FROM bitnami_redmine.users b WHERE j.user_id = b.id) AS commentor,
        j.notes, j.created_on
        FROM bitnami_redmine.issues i
        JOIN bitnami_redmine.issue_statuses is2 ON i.status_id = is2.id
        JOIN bitnami_redmine.users u ON i.assigned_to_id = u.id
        JOIN bitnami_redmine.journals j ON i.id = j.journalized_id
        WHERE j.created_on > ?
        ORDER BY j.created_on DESC`

	rows, err := db.Query(query, formattedTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []model.Issue
	for rows.Next() {
		var issue model.Issue
		var assigneeFirstName, assigneeLastName, commentorFirstName sql.NullString
		var estimatedHours sql.NullFloat64
		var dueDate sql.NullTime
		if err := rows.Scan(
			&issue.ID, &issue.JobID, &issue.Status, &assigneeFirstName, &assigneeLastName, &issue.StartDate, &dueDate,
			&issue.DoneRatio, &estimatedHours, &issue.Priority, &issue.Author, &issue.Subject,
			&issue.Description, &commentorFirstName, &issue.Notes, &issue.CreatedOn,
		); err != nil {
			return nil, err
		}
		issue.Assignee = fmt.Sprintf("%s %s", assigneeFirstName.String, assigneeLastName.String)
		issue.Commentor = commentorFirstName.String
		if estimatedHours.Valid {
			issue.EstimatedHours = estimatedHours.Float64
		} else {
			issue.EstimatedHours = 0
		}
		if dueDate.Valid {
			issue.DueDate = dueDate.Time
		} else {
			issue.DueDate = time.Time{}
		}
		issues = append(issues, issue)
	}
	return issues, nil
}
