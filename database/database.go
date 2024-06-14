package database

import (
	"database/sql"
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

// FetchNewIssues fetches new issues from the MySQL database
func FetchNewIssues(db *sql.DB, lastChecked time.Time) ([]model.Issue, error) {
	formattedTime := lastChecked.Format("2006-01-02 15:04:05")
	query := "SELECT id, subject, description, created_on, updated_on FROM bitnami_redmine.issues WHERE updated_on > ?"
	rows, err := db.Query(query, formattedTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []model.Issue
	for rows.Next() {
		var issue model.Issue
		if err := rows.Scan(&issue.ID, &issue.Subject, &issue.Description, &issue.CreatedOn, &issue.UpdatedOn); err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}
	return issues, nil
}
