package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/joho/godotenv"
)

type Task struct {
	TaskID    string `json:"task_id"`
	Process   string `json:"process"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	ExpiresAt int64  `json:"expires_at"`
}

func main() {
	_ = godotenv.Load(".env")
	table := os.Getenv("DDB_TABLE_NAME")
	webhook := os.Getenv("NOTIFY_WEBHOOK_URL")

	if table == "" || webhook == "" {
		fmt.Println("❌ Missing DDB_TABLE_NAME or NOTIFY_WEBHOOK_URL in .env")
		return
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess)

	for {
		fmt.Println("🔁 Agent48 scanning for unresolved tasks...")

		input := &dynamodb.ScanInput{
			TableName: aws.String(table),
		}

		result, err := svc.Scan(input)
		if err != nil {
			fmt.Println("❌ Failed to scan DynamoDB:", err)
			time.Sleep(3 * time.Second)
			continue
		}

		for _, item := range result.Items {
			var task Task
			err = dynamodbattribute.UnmarshalMap(item, &task)
			if err != nil {
				fmt.Println("❌ Failed to unmarshal item:", err)
				continue
			}

			if task.Status == "resolved" {
				continue
			}

			fmt.Printf("🚨 Resolving: %s (%s)\n", task.Process, task.Status)

			pid := extractPID(task.TaskID)

			if task.Status == "kill" {
				if strings.HasPrefix(task.Process, "cronjob:") {
					// Smart delete cron file
					path := strings.TrimPrefix(task.Process, "cronjob:")
					err := os.Remove(path)
					if err != nil {
						fmt.Printf("❌ Failed to delete cronjob file: %s (%v)\n", path, err)
					} else {
						fmt.Printf("🧹 Cronjob deleted: %s\n", path)
					}
				} else {
					err := exec.Command("kill", "-9", fmt.Sprintf("%d", pid)).Run()
					if err != nil {
						fmt.Printf("⚠️ Failed to kill PID %d: %v\n", pid, err)
					} else {
						fmt.Printf("☠️ Killed PID %d (%s)\n", pid, task.Process)
					}
				}
			} else {
				fmt.Printf("👁️ WATCH task resolved by logging only: %s\n", task.Process)
			}

			// ✅ Mark resolved
			task.Status = "resolved"
			task.Timestamp = time.Now().Format(time.RFC3339)

			updatedItem, _ := dynamodbattribute.MarshalMap(task)
			_, err = svc.PutItem(&dynamodb.PutItemInput{
				TableName: aws.String(table),
				Item:      updatedItem,
			})
			if err != nil {
				fmt.Println("❌ Failed to update task status:", err)
			} else {
				fmt.Printf("✅ Marked as resolved in DynamoDB: %s\n", task.TaskID)
			}

			// 🔔 Fire webhook
			NotifyWebhook(task, webhook)
		}

		time.Sleep(3 * time.Second)
	}
}

func extractPID(taskID string) int {
	parts := strings.Split(taskID, "-")
	if len(parts) < 2 {
		return 0
	}
	pid, _ := strconv.Atoi(parts[1])
	return pid
}


func NotifyWebhook(task Task, slackURL string) {
	payload := map[string]string{
		"text": fmt.Sprintf("🔔 *Agent48 RESOLVED TASK*\n• Process: `%s`\n• Status: `%s`\n• Time: %s",
			task.Process, task.Status, task.Timestamp),
	}
	jsonPayload, _ := json.Marshal(payload)

	// Slack/Discord Webhook
	resp1, err1 := http.Post(slackURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err1 == nil && resp1.StatusCode >= 200 && resp1.StatusCode < 300 {
		fmt.Println("📣 Slack/Discord webhook delivered.")
	} else {
		fmt.Println("⚠️ Slack webhook failed.")
	}

	// Optional PUSH_ENDPOINT
	pushURL := os.Getenv("PUSH_ENDPOINT")
	if pushURL != "" {
		resp2, err2 := http.Post(pushURL, "application/json", bytes.NewBuffer(jsonPayload))
		if err2 == nil && resp2.StatusCode >= 200 && resp2.StatusCode < 300 {
			fmt.Println("📡 PUSH endpoint delivered.")
		} else {
			fmt.Println("⚠️ PUSH endpoint failed or not set.")
		}
	}
}
