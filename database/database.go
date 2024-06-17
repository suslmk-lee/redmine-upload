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

	// -- 일감 조회 / note 포함
	//select i.id, is2.name, u.firstname, u.lastname, i.start_date, i.due_date, i.done_ratio, i.estimated_hours,
	//(select e.name from bitnami_redmine.enumerations e where e.type = "IssuePriority" and i.priority_id = e.id) as priority,
	//(select b.firstname from bitnami_redmine.users b where i.author_id = b.id) as author,
	//i.subject, i.description,
	//(select b.firstname from bitnami_redmine.users b where j.user_id = b.id) as commentor,
	//j.notes, j.created_on
	//  from bitnami_redmine.issues i,
	//  	bitnami_redmine.issue_statuses is2,
	//  	bitnami_redmine.users u,
	//  	bitnami_redmine.journals j
	// where i.status_id = is2.id
	//   and i.assigned_to_id = u.id
	//   and i.id = j.journalized_id
	//order by j.created_on desc;

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
