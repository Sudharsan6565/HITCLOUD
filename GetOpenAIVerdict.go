package main

import (
	"context"
	"fmt"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

func GetOpenAIVerdict(cmd string) string {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	req := openai.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: "You're an AI security agent. Respond only with 'kill', 'safe', or 'watch' when given a process command.",
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Process command: %s", cmd),
			},
		},
	}

	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		fmt.Printf("⚠️ OpenAI error: %v\n", err)
		return "watch"
	}

	reply := resp.Choices[0].Message.Content
	switch reply {
	case "kill", "safe", "watch":
		return reply
	default:
		return "watch"
	}
}
