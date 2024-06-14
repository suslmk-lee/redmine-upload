package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	_ "strings"
	"time"
)

const (
	accessKey  = ""
	secretKey  = ""
	region     = "KR1" // Replace with your region
	service    = "s3"
	bucketName = "suslmk-storage"                             // Replace with your bucket name
	endpoint   = "kr1-api-object-storage.nhncloudservice.com" // Replace with your endpoint 	// Replace with your bucket name
	objectKey  = "your-object-key"
)

func sign(key []byte, message string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(message))
	return mac.Sum(nil)
}

func getSignatureKey(secret, date, region, service string) []byte {
	kDate := sign([]byte("AWS4"+secret), date)
	kRegion := sign(kDate, region)
	kService := sign(kRegion, service)
	kSigning := sign(kService, "aws4_request")
	return kSigning
}

func hashSHA256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func printRequestHeaders(req *http.Request) {
	fmt.Println("Request Headers:")
	for k, v := range req.Header {
		fmt.Printf("%s: %s\n", k, v)
	}
}

func listBuckets() {
	method := "GET"
	host := endpoint
	uri := "/"
	queryString := ""

	t := time.Now().UTC()
	amzDate := t.Format("20060102T150405Z")
	dateStamp := t.Format("20060102")

	payloadHash := hashSHA256("")
	canonicalHeaders := fmt.Sprintf("host:%s\nx-amz-content-sha256:%s\nx-amz-date:%s\n", host, payloadHash, amzDate)
	signedHeaders := "host;x-amz-content-sha256;x-amz-date"

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		method, uri, queryString, canonicalHeaders, signedHeaders, payloadHash)

	algorithm := "AWS4-HMAC-SHA256"
	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, region, service)
	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s",
		algorithm, amzDate, credentialScope, hashSHA256(canonicalRequest))

	signingKey := getSignatureKey(secretKey, dateStamp, region, service)
	signature := hex.EncodeToString(sign(signingKey, stringToSign))

	authorizationHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		accessKey, credentialScope, signedHeaders, signature)

	url := fmt.Sprintf("https://%s%s", host, uri)
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Add("x-amz-date", amzDate)
	req.Header.Add("x-amz-content-sha256", payloadHash)
	req.Header.Add("Authorization", authorizationHeader)

	printRequestHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	fmt.Printf("Response: %s\n", body)
}

func putObject() {
	method := "PUT"
	host := endpoint
	uri := fmt.Sprintf("/%s/%s", bucketName, objectKey)
	queryString := ""

	body := "This is the content of the object."
	t := time.Now().UTC()
	amzDate := t.Format("20060102T150405Z")
	dateStamp := t.Format("20060102")

	payloadHash := hashSHA256(body)
	canonicalHeaders := fmt.Sprintf("host:%s\nx-amz-content-sha256:%s\nx-amz-date:%s\n", host, payloadHash, amzDate)
	signedHeaders := "host;x-amz-content-sha256;x-amz-date"

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		method, uri, queryString, canonicalHeaders, signedHeaders, payloadHash)

	algorithm := "AWS4-HMAC-SHA256"
	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, region, service)
	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s",
		algorithm, amzDate, credentialScope, hashSHA256(canonicalRequest))

	signingKey := getSignatureKey(secretKey, dateStamp, region, service)
	signature := hex.EncodeToString(sign(signingKey, stringToSign))

	authorizationHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		accessKey, credentialScope, signedHeaders, signature)

	url := fmt.Sprintf("https://%s%s", host, uri)
	req, _ := http.NewRequest(method, url, strings.NewReader(body))
	req.Header.Add("x-amz-date", amzDate)
	req.Header.Add("x-amz-content-sha256", payloadHash)
	req.Header.Add("Authorization", authorizationHeader)
	req.Header.Add("Content-Length", fmt.Sprintf("%d", len(body)))

	printRequestHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	fmt.Printf("Response: %s\n", respBody)
}

func createBucket() {
	method := "PUT"
	host := endpoint
	uri := fmt.Sprintf("/%s", bucketName)
	queryString := ""

	t := time.Now().UTC()
	amzDate := t.Format("20060102T150405Z")
	dateStamp := t.Format("20060102")

	payloadHash := hashSHA256("")
	canonicalHeaders := fmt.Sprintf("host:%s\nx-amz-content-sha256:%s\nx-amz-date:%s\n", host, payloadHash, amzDate)
	signedHeaders := "host;x-amz-content-sha256;x-amz-date"

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		method, uri, queryString, canonicalHeaders, signedHeaders, payloadHash)

	algorithm := "AWS4-HMAC-SHA256"
	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, region, service)
	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s",
		algorithm, amzDate, credentialScope, hashSHA256(canonicalRequest))

	signingKey := getSignatureKey(secretKey, dateStamp, region, service)
	signature := hex.EncodeToString(sign(signingKey, stringToSign))

	authorizationHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		accessKey, credentialScope, signedHeaders, signature)

	url := fmt.Sprintf("https://%s%s", host, uri)
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Add("x-amz-date", amzDate)
	req.Header.Add("x-amz-content-sha256", payloadHash)
	req.Header.Add("Authorization", authorizationHeader)

	printRequestHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	fmt.Printf("Response: %s\n", respBody)
}

func main() {
	fmt.Println("Creating bucket:")
	createBucket()

	fmt.Println("\nListing buckets:")
	listBuckets()

	fmt.Println("\nPutting object:")
	putObject()
}
