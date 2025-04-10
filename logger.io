package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/joho/godotenv"
)

type LogEntry struct {
	TaskID    string `json:"task_id"`
	Process   string `json:"process"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

func LogToDynamoDB(pid int32, proc string, status string) {
	// Try to load .env
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("⚠️  .env not loaded, using system environment variables")
	}

	// Env vars
	table := os.Getenv("DDB_TABLE_NAME")
	logGroup := "Agent47-Logs"
	logStream := "Agent47Stream"

	if table == "" {
		fmt.Println("❌ DDB_TABLE_NAME not found in environment")
		return
	}

	entry := LogEntry{
		TaskID:    fmt.Sprintf("proc-%d", pid),
		Process:   proc,
		Status:    status,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// DynamoDB logic
	svc := dynamodb.New(sess)
	av, err := dynamodbattribute.MarshalMap(entry)
	if err != nil {
		fmt.Println("❌ Failed to marshal log entry:", err)
		return
	}

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(table),
	})
	if err != nil {
		fmt.Printf("❌ Failed to write to DynamoDB: %v\n", err)
	} else {
		fmt.Printf("✅ Logged to DynamoDB: %+v\n", entry)
	}

	// CloudWatch logic
	cw := cloudwatchlogs.New(sess)

	// Create stream if needed
	_, err = cw.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(logGroup),
		LogStreamName: aws.String(logStream),
	})
	if err != nil && !strings.Contains(err.Error(), "ResourceAlreadyExistsException") {
		fmt.Println("⚠️  CloudWatch stream creation error:", err)
		return
	}

	// Log entry
	_, err = cw.PutLogEvents(&cloudwatchlogs.PutLogEventsInput{
		LogEvents: []*cloudwatchlogs.InputLogEvent{
			{
				Message:   aws.String(fmt.Sprintf("Agent47: %s [%d] → %s", proc, pid, status)),
				Timestamp: aws.Int64(time.Now().Unix() * 1000),
			},
		},
		LogGroupName:  aws.String(logGroup),
		LogStreamName: aws.String(logStream),
	})

	if err != nil {
		fmt.Println("❌ CloudWatch log error:", err)
	} else {
		fmt.Println("🟢 CloudWatch log success")
	}
}
