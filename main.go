package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

const (
	bucketName   = "your-s3-bucket-name"
	region       = "your-region"
	dsn          = "root:ahsus123@tcp(localhost:3307)/"
	pollInterval = 10 * time.Second
)

type Issue struct {
	ID          int       `json:"id"`
	Subject     string    `json:"subject"`
	Description string    `json:"description"`
	CreatedOn   time.Time `json:"created_on"`
	UpdatedOn   time.Time `json:"updated_on"`
}

func fetchNewIssues(db *sql.DB, lastChecked time.Time) ([]Issue, error) {
	// Use the proper format for MySQL DATETIME
	formattedTime := lastChecked.Format("2006-01-02 15:04:05")
	query := "SELECT id, subject, description, created_on, updated_on FROM bitnami_redmine.issues WHERE updated_on > ?"
	rows, err := db.Query(query, formattedTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []Issue
	for rows.Next() {
		var issue Issue
		if err := rows.Scan(&issue.ID, &issue.Subject, &issue.Description, &issue.CreatedOn, &issue.UpdatedOn); err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}
	return issues, nil
}

func createCloudEvent(issue Issue) (cloudevents.Event, error) {
	event := cloudevents.NewEvent()
	event.SetID(uuid.New().String())
	event.SetSource("redmine/issues")
	event.SetType("com.example.issue")
	event.SetTime(time.Now())

	if err := event.SetData(cloudevents.ApplicationJSON, issue); err != nil {
		return event, err
	}
	return event, nil
}

func uploadToS3(s3Client *s3.S3, data []byte, key string) error {
	input := &s3.PutObjectInput{
		Body:   bytes.NewReader(data),
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}

	_, err := s3Client.PutObject(input)
	return err
}

func main() {
	// Connect to MySQL database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		log.Fatalf("failed to create AWS session: %v", err)
	}

	s3Client := s3.New(sess)
	lastChecked := time.Now()

	for {
		// Fetch new issues from MySQL
		issues, err := fetchNewIssues(db, lastChecked)
		if err != nil {
			log.Printf("failed to fetch new issues: %v", err)
			continue
		}

		for _, issue := range issues {
			// Create CloudEvent for each issue
			event, err := createCloudEvent(issue)
			if err != nil {
				log.Printf("failed to create CloudEvent: %v", err)
				continue
			}

			// Convert CloudEvent to JSON
			data, err := json.Marshal(event)
			if err != nil {
				log.Printf("failed to marshal CloudEvent: %v", err)
				continue
			}

			// Generate a unique key for the S3 object
			key := fmt.Sprintf("issues/%d.json", issue.ID)

			// Upload event data to S3
			err = uploadToS3(s3Client, data, key)
			if err != nil {
				log.Printf("failed to upload data to S3: %v", err)
				continue
			}
		}

		// Update lastChecked time
		lastChecked = time.Now()

		// Sleep for the poll interval
		time.Sleep(pollInterval)
	}
}
