package action

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"redmine-upload/model"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

// CreateCloudEvent creates a CloudEvent for the given issue
func CreateCloudEvent(issue model.Issue) (cloudevents.Event, error) {
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

// UploadToS3 uploads the given data to S3 with the specified key
func UploadToS3(s3Client *s3.S3, bucketName string, data []byte, key string) error {
	input := &s3.PutObjectInput{
		Body:   bytes.NewReader(data),
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}

	_, err := s3Client.PutObject(input)
	return err
}

// ProcessIssues processes a list of issues by creating CloudEvents and uploading them to S3
func ProcessIssues(s3Client *s3.S3, bucketName string, issues []model.Issue) error {
	for _, issue := range issues {
		// Create CloudEvent for each issue
		event, err := CreateCloudEvent(issue)
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
		err = UploadToS3(s3Client, bucketName, data, key)
		if err != nil {
			log.Printf("failed to upload data to S3: %v", err)
			continue
		} else {
			fmt.Printf("Upload Issue :: [%s] %s", issue.Author, key)
		}
	}
	return nil
}
