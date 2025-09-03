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
	// 파일 읽기/파싱에 실패해도 서비스를 중단하지 않고, 경고 로그 후 기본값(10분 전)으로 계속 진행합니다.
	lastChecked, err := common.LoadLastCheckedTime()
	if err != nil {
		log.Printf("WARN: failed to load last checked time: %v. Using default time.", err)
	}
	if lastChecked.IsZero() {
		log.Println("INFO: Last checked time is zero or invalid, setting to 10 minutes ago.")
		lastChecked = time.Now().Add(-10 * time.Minute)
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

	// 특정 시간에 한번만 실행되는 작업의 스케줄링을 더 안정적으로 변경합니다.
	targetTimes := []string{"09:00", "13:00"}
	lastRunDay := make(map[string]int) // map[targetTime] -> day of year
	for {
		now := time.Now()

		// processOne: 특정 시간에 실행되는 작업
		currentDay := now.YearDay()
		for _, target := range targetTimes {
			// 현재 시간이 목표 시간이고, 오늘 아직 실행되지 않았다면 실행합니다.
			if now.Format("15:04") == target && lastRunDay[target] != currentDay {
				log.Printf("INFO: Running scheduled task 'processOne' for %s", target)
				processOne(db, s3Client)
				lastRunDay[target] = currentDay
			}
		}

		// processTwo: 주기적으로 새로운 이슈를 확인하는 작업의 로직을 메인 루프로 통합합니다.
		// 1. DB에서 새로운 이슈를 가져옵니다.
		issues, err := database.FetchNewIssues(db, lastChecked)
		if err != nil {
			log.Printf("ERROR: failed to fetch new issues: %v", err)
			time.Sleep(pollInterval) // 오류 발생 시 잠시 대기 후 재시도
			continue
		}

		// 2. 새로운 이슈가 있으면 처리하고 업로드합니다.
		if len(issues) > 0 {
			log.Printf("INFO: Found %d new issue(s) to process.", len(issues))
			err = action.ProcessIssues(s3Client, bucketName, issues)
			if err != nil {
				log.Printf("ERROR: failed to process and upload issues: %v", err)
				time.Sleep(pollInterval) // 오류 발생 시 잠시 대기 후 재시도 (시간 업데이트 없이)
				continue
			}
		}

		// 3. 모든 처리가 성공하면, 마지막 확인 시간을 현재 시간으로 갱신하고 파일에 저장합니다.
		newLastChecked := time.Now()
		if err := common.SaveLastCheckedTime(newLastChecked); err != nil {
			log.Printf("ERROR: failed to save last checked time: %v. Will retry processing in next cycle.", err)
		} else {
			lastChecked = newLastChecked // 파일 저장이 성공했을 때만 메모리의 시간도 갱신
		}

		// Sleep for the poll interval
		time.Sleep(pollInterval)
	}
}

func processOne(db *sql.DB, s3Client *s3.S3) {
	ImminentIssues, err := database.FetchImminentIssue(db)
	if err != nil {
		log.Printf("ERROR: failed to fetch imminent issues: %v", err)
		return
	}

	if len(ImminentIssues) > 0 {
		err = action.ProcessImminentIssues(s3Client, bucketName, ImminentIssues)
		if err != nil {
			log.Printf("ERROR: failed to process and upload imminent issues: %v", err)
		}
	}
}

func printKST() {
	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		// panic은 전체 프로그램을 중단시키므로, 로그만 남기고 해당 고루틴을 종료하는 것이 더 안전합니다.
		log.Printf("ERROR: could not load location 'Asia/Seoul': %v", err)
		return
	}

	// 프로그램 시작시..
	fmt.Println(time.Now().In(loc).Format("2006-01-02 15:04:05"))

	// 1시간마다..
	for range time.NewTicker(1 * time.Hour).C {
		fmt.Println(time.Now().In(loc).Format("2006-01-02 15:04:05"))
	}
}
