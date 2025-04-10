package main

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

func SendToCloudWatch(message string, groupName string) {
	logStream := "agent47-stream"
	region := os.Getenv("AWS_REGION")
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := cloudwatchlogs.New(sess)

	// Ensure log group exists
	_, _ = svc.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(groupName),
	})

	// Ensure stream exists
	_, _ = svc.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(groupName),
		LogStreamName: aws.String(logStream),
	})

	input := &cloudwatchlogs.PutLogEventsInput{
		LogGroupName:  aws.String(groupName),
		LogStreamName: aws.String(logStream),
		LogEvents: []*cloudwatchlogs.InputLogEvent{
			{
				Message:   aws.String(message),
				Timestamp: aws.Int64(time.Now().UnixNano() / int64(time.Millisecond)),
			},
		},
	}

	_, err := svc.PutLogEvents(input)
	if err != nil {
		fmt.Printf("❌ CloudWatch error: %v\n", err)
	} else {
		fmt.Println("🛰️  Logged to CloudWatch.")
	}
}
