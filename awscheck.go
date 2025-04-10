package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
)

func main() {
	fmt.Println("⏳ Testing AWS session...")
	// Create a new AWS session
	sess, err := session.NewSession()
	if err != nil {
		fmt.Println("❌ AWS session error:", err)
		return
	}
	fmt.Println("✅ AWS session started successfully:", sess)
}
