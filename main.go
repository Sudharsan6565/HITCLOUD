package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/shirou/gopsutil/process"
)

func main() {
	fmt.Println("🔁 Agent47 Recon + Verdict + Cloud Logging")
	_ = godotenv.Load(".env")

	for {
		procs, _ := process.Processes()
		fmt.Printf("\n%-8s %-8s %-6s %-6s %-10s %s\n", "PID", "PPID", "CPU%", "MEM%", "TAG", "CMD")

		for _, p := range procs {
			cpu, _ := p.CPUPercent()
			if cpu < 5.0 {
				continue
			}
			mem, _ := p.MemoryPercent()
			pid := p.Pid
			ppid, _ := p.Ppid()
			cmd, _ := p.Exe()
			if cmd == "" {
				cmd, _ = p.Name()
			}

			tag := "-"
			if IsWhitelisted(cmd) {
				tag = "SAFE"
				LogToDynamoDB(pid, cmd, "safe")
				SendToCloudWatch(fmt.Sprintf("SAFE: %s [%d]", cmd, pid), "Agent47")
			} else if IsBlacklisted(cmd) {
				tag = "KILL"
				go killProcess(pid, cmd)
				LogToDynamoDB(pid, cmd, "killed")
				SendToCloudWatch(fmt.Sprintf("KILL: %s [%d]", cmd, pid), "Agent47")
			} else {
				tag = "SCAN"
				go func(cmd string, pid int32) {
					fmt.Printf("📡 Sending to OpenAI: %s [%d]\n", cmd, pid)
					verdict := GetOpenAIVerdict(cmd)
					switch verdict {
					case "kill":
						go killProcess(pid, cmd)
						LogToDynamoDB(pid, cmd, "killed")
						SendToCloudWatch(fmt.Sprintf("💀 AI Verdict: KILL %s [%d]", cmd, pid), "Agent47")
						AppendToList("blacklist.json", cmd)
					case "safe":
						LogToDynamoDB(pid, cmd, "safe")
						SendToCloudWatch(fmt.Sprintf("✅ AI Verdict: SAFE %s [%d]", cmd, pid), "Agent47")
						AppendToList("whitelist.json", cmd)
					default:
						LogToDynamoDB(pid, cmd, "watch")
						SendToCloudWatch(fmt.Sprintf("👁️ AI Verdict: WATCH %s [%d]", cmd, pid), "Agent47")
						AppendToList("watchlist.json", cmd)
					}
				}(cmd, pid)
				tag = "WATCH"
			}

			fmt.Printf("%-8d %-8d %-6.1f %-6.1f %-10s %s\n", pid, ppid, cpu, mem, tag, cmd)
		}
		time.Sleep(4 * time.Second)
	}
}

func killProcess(pid int32, name string) {
	fmt.Printf("☠️  Killing blacklisted process: %s [%d]\n", name, pid)
	cmd := exec.Command("kill", "-9", strconv.Itoa(int(pid)))
	err := cmd.Run()
	if err != nil {
		fmt.Printf("❌ Failed to kill: %v\n", err)
	}
}

func IsWhitelisted(cmd string) bool {
	data, _ := os.ReadFile("whitelist.json")
	return strings.Contains(string(data), cmd)
}

func IsBlacklisted(cmd string) bool {
	data, _ := os.ReadFile("blacklist.json")
	return strings.Contains(string(data), cmd)
}

func AppendToList(filename, cmd string) {
	data, _ := os.ReadFile(filename)
	if strings.Contains(string(data), cmd) {
		return // already present
	}
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err == nil {
		f.WriteString(fmt.Sprintf("\n# %s\n", cmd))
		f.Close()
	}
}
