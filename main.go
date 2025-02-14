package main

import (
	"database/sql"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
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
	accessKey  string
	secretKey  string
)

func init() {
	region = common.ConfInfo["nhn.region"]
	bucketName = common.ConfInfo["nhn.storage.bucket.name"]
	dsn = common.ConfInfo["database.url"]
	endpoint = common.ConfInfo["nhn.storage.endpoint.url"]
	accessKey = common.ConfInfo["nhn.storage.accessKey"]
	secretKey = common.ConfInfo["nhn.storage.secretKey"]
}

func main() {
	go printKST()
	fmt.Println("Start redmine-upload Service..")
	// Ensure the keys are not empty
	if accessKey == "" || secretKey == "" {
		log.Fatalf("AccessKey or SecretKey is empty")
	}

	// Connect to MySQL database
	db, err := database.ConnectDB(dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Load the last checked time from file
	lastChecked, err := common.LoadLastCheckedTime()
	if err != nil {
		log.Fatalf("failed to load last checked time: %v", err)
	}
	if lastChecked.IsZero() {
		// If there is no last checked time, start from one week ago
		lastChecked = time.Now().Add(-7 * 24 * time.Hour)
	}

	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		S3ForcePathStyle: aws.Bool(true)}, // Use path-style addressing for compatibility with custom endpoints
	)
	if err != nil {
		log.Fatalf("failed to create AWS session: %v", err)
	}

	s3Client := s3.New(sess)

	targetTimes := []string{"09:00", "13:00"}
	lastRun := make(map[string]bool)
	for {
		now := time.Now()
		currentTimeStr := now.Format("15:04")

		for _, target := range targetTimes {
			if currentTimeStr == target {
				if !lastRun[target] {
					processOne(db, s3Client)
					lastRun[target] = true
				}
			} else {
				// 시간이 지나면 다음 실행을 위해 플래그를 초기화
				lastRun[target] = false
			}
		}

		lastChecked, err := common.LoadLastCheckedTime()
		if err != nil {
			log.Fatalf("failed to load last checked time: %v", err)
		}
		fmt.Println(lastChecked)

		processTwo(db, s3Client, lastChecked)

		// Sleep for the poll interval
		time.Sleep(pollInterval)
	}
}

func processOne(db *sql.DB, s3Client *s3.S3) {
	ImminentIssues, err := database.FetchImminentIssue(db)
	if err != nil {
		log.Printf("failed to fetch new issues: %v", err)
		return
	}

	if len(ImminentIssues) > 0 {
		err = action.ProcessImminentIssues(s3Client, bucketName, ImminentIssues)
		if err != nil {
			log.Printf("failed to process and upload issues: %v", err)
		}
	}
}

func processTwo(db *sql.DB, s3Client *s3.S3, lastChecked time.Time) {
	// Fetch new issues from MySQL
	issues, err := database.FetchNewIssues(db, lastChecked)
	if err != nil {
		log.Printf("failed to fetch new issues: %v", err)
		return
	}
	fmt.Println(len(issues))

	// Process and upload issues
	err = action.ProcessIssues(s3Client, bucketName, issues)
	if err != nil {
		log.Printf("failed to process and upload issues: %v", err)
	}

	// Update lastChecked time
	lastChecked = time.Now()
	err = common.SaveLastCheckedTime(lastChecked)
	if err != nil {
		log.Printf("failed to save last checked time: %v", err)
	}

}

func printKST() {
	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		panic(err)
	}

	// 프로그램 시작시..
	fmt.Println(time.Now().In(loc).Format("2006-01-02 15:04:05"))

	// 1시간마다..
	for range time.NewTicker(1 * time.Hour).C {
		fmt.Println(time.Now().In(loc).Format("2006-01-02 15:04:05"))
	}
}
