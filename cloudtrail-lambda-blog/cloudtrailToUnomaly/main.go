package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type unomalyPostBody struct {
	Source    string `json:"source"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

type unomalyCfg struct {
	Endpoint  string
	UseSSL    bool
	BatchSize int
}

func (cfg *unomalyCfg) getEndpointFromEnv() {
	cfg.Endpoint = os.Getenv("UNOMALY_API_ENDPOINT")
}

func (cfg *unomalyCfg) getBatchSizeFromEnv() error {
	size, err := strconv.ParseInt(os.Getenv("UNOMALY_BATCH_SIZE"), 0, 0)
	if err != nil {
		return err
	}
	cfg.BatchSize = int(size)
	return nil
}

var uCfg = new(unomalyCfg)

func postToUnomaly(url string, payload []byte) error {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("Wrong status code returned from Unomaly API")
	}

	return nil
}

func handler(ctx context.Context, s3Event events.S3Event) error {
	uCfg := new(unomalyCfg)
	uCfg.getEndpointFromEnv()
	fmt.Println("Unomaly Endpoint: ", uCfg.Endpoint)
	uCfg.getBatchSizeFromEnv()
	fmt.Println("Unomaly Batch Size: ", uCfg.BatchSize)

	svc := s3.New(session.New())

	var events []*unomalyPostBody

	for _, record := range s3Event.Records {
		res, err := svc.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(record.S3.Bucket.Name),
			Key:    aws.String(record.S3.Object.Key),
		})
		if err != nil {
			return err
		}
		defer res.Body.Close()

		buf := new(bytes.Buffer)
		buf.ReadFrom(res.Body)
		var data map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
			return err
		}

		// un-nesting the records we want from the file in S3...
		rcrds := data["Records"].([]interface{})
		for _, r := range rcrds {
			j, err := json.Marshal(r)
			if err != nil {
				return err
			}

			var x map[string]interface{}
			if err = json.Unmarshal(j, &x); err != nil {
				return err
			}
			eventTime := x["eventTime"]
			eventSource := x["eventSource"]

			e := &unomalyPostBody{
				Source:    eventSource.(string),
				Message:   string(j),
				Timestamp: eventTime.(string),
			}
			events = append(events, e)
		}
	}

	fmt.Println("Number of messages to send: ", len(events))

	for len(events) > 0 {
		fmt.Println("Events left to process: ", len(events))
		chunkSize := uCfg.BatchSize
		if len(events) <= chunkSize {
			chunkSize = len(events)
		}

		take := events[:chunkSize]
		events = events[chunkSize:]
		reqBody, err := json.Marshal(take)
		if err != nil {
			return err
		}

		if err := postToUnomaly(uCfg.Endpoint, reqBody); err != nil {
			return err
		}
		fmt.Printf("Sending %d events to Unomaly endpoint\n", len(take))
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
