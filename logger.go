package main

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
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
	_ = godotenv.Load(".env")
	table := os.Getenv("DDB_TABLE_NAME")

	entry := LogEntry{
		TaskID:    fmt.Sprintf("proc-%d", pid),
		Process:   proc,
		Status:    status,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess)

	av, err := dynamodbattribute.MarshalMap(entry)
	if err != nil {
		fmt.Println("❌ Failed to marshal log entry:", err)
		return
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(table),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		fmt.Printf("❌ Failed to write to DynamoDB: %v\n", err)
	} else {
		fmt.Printf("✅ Logged to DynamoDB: %+v\n", entry)
	}
}
