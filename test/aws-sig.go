package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	awsAccessKeyID     = ""
	awsSecretAccessKey = ""
	region             = "KR1"
	service            = "s3"
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

func main() {
	// Define the request parameters
	method := "GET"
	host := "your-bucket-name.s3.amazonaws.com"
	uri := "/your-object-key"
	queryString := ""

	// Get the current time and format it
	t := time.Now().UTC()
	amzDate := t.Format("20060102T150405Z")
	dateStamp := t.Format("20060102")

	// Create canonical request
	canonicalHeaders := fmt.Sprintf("host:%s\nx-amz-date:%s\n", host, amzDate)
	signedHeaders := "host;x-amz-date"
	payloadHash := hashSHA256("")

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		method, uri, queryString, canonicalHeaders, signedHeaders, payloadHash)

	// Create the string to sign
	algorithm := "AWS4-HMAC-SHA256"
	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, region, service)
	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s",
		algorithm, amzDate, credentialScope, hashSHA256(canonicalRequest))

	// Calculate the signature
	signingKey := getSignatureKey(awsSecretAccessKey, dateStamp, region, service)
	signature := hex.EncodeToString(sign(signingKey, stringToSign))

	// Add signing information to the request
	authorizationHeader := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm, awsAccessKeyID, credentialScope, signedHeaders, signature)

	// Create the request
	url := fmt.Sprintf("https://%s%s", host, uri)
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Add("x-amz-date", amzDate)
	req.Header.Add("Authorization", authorizationHeader)

	// Print the request (for demonstration purposes)
	fmt.Println("Request URL:", url)
	fmt.Println("Request Headers:")
	for k, v := range req.Header {
		fmt.Printf("%s: %s\n", k, strings.Join(v, ","))
	}

	// Send the request (optional, for demonstration purposes)
	// client := &http.Client{}
	// resp, err := client.Do(req)
	// if err != nil {
	//     log.Fatalf("Failed to make request: %v", err)
	// }
	// defer resp.Body.Close()
	// body, _ := io.ReadAll(resp.Body)
	// fmt.Printf("Response: %s\n", body)
}
