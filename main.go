package main

import (
	"log"
	"time"

	"redmine-upload/action"
	"redmine-upload/common"
	"redmine-upload/database"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	pollInterval = 10 * time.Second
)

var (
	bucketName string
	region     string
	dsn        string
	endpoint   string
)

func init() {
	region = common.ConfInfo["nhn.storage.region"]
	bucketName = common.ConfInfo["nhn.storage.endpoint.url"]
	dsn = common.ConfInfo["database.url"]
	endpoint = common.ConfInfo["nhn.storage.endpoint.url"]
}

func main() {
	// Connect to MySQL database
	db, err := database.ConnectDB(dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		S3ForcePathStyle: aws.Bool(true)},
	)
	if err != nil {
		log.Fatalf("failed to create AWS session: %v", err)
	}

	s3Client := s3.New(sess)
	lastChecked := time.Now().Add(-7 * 24 * time.Hour)

	for {
		// Fetch new issues from MySQL
		issues, err := database.FetchNewIssues(db, lastChecked)
		if err != nil {
			log.Printf("failed to fetch new issues: %v", err)
			continue
		}

		// Process and upload issues
		err = action.ProcessIssues(s3Client, bucketName, issues)
		if err != nil {
			log.Printf("failed to process and upload issues: %v", err)
		}

		// Update lastChecked time
		lastChecked = time.Now()

		// Sleep for the poll interval
		time.Sleep(pollInterval)
	}
}
