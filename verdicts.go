package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

var mu sync.Mutex

type Verdicts struct {
	Whitelist map[string]bool `json:"whitelist"`
	Blacklist map[string]bool `json:"blacklist"`
	Watchlist map[string]bool `json:"watchlist"`
}

var verdicts Verdicts

func LoadVerdicts() {
	verdicts = Verdicts{
		Whitelist: make(map[string]bool),
		Blacklist: make(map[string]bool),
		Watchlist: make(map[string]bool),
	}
	loadJSON("whitelist.json", &verdicts.Whitelist)
	loadJSON("blacklist.json", &verdicts.Blacklist)
	loadJSON("watchlist.json", &verdicts.Watchlist)
}

func loadJSON(file string, target *map[string]bool) {
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("⚠️  Could not read %s: %v\n", file, err)
		return
	}
	var raw struct {
		Items []string `json:"whitelist"` // dynamic key, reused
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		fmt.Printf("⚠️  Could not parse %s: %v\n", file, err)
		return
	}
	for _, item := range raw.Items {
		(*target)[item] = true
	}
}

func SaveVerdict(tag, cmd string) {
	mu.Lock()
	defer mu.Unlock()

	switch tag {
	case "SAFE":
		verdicts.Whitelist[cmd] = true
		writeJSON("whitelist.json", verdicts.Whitelist, "whitelist")
	case "KILL":
		verdicts.Blacklist[cmd] = true
		writeJSON("blacklist.json", verdicts.Blacklist, "blacklist")
	case "WATCH":
		verdicts.Watchlist[cmd] = true
		writeJSON("watchlist.json", verdicts.Watchlist, "watchlist")
	}
}

func writeJSON(file string, store map[string]bool, label string) {
	items := []string{}
	for k := range store {
		items = append(items, k)
	}
	payload := map[string][]string{label: items}
	data, _ := json.MarshalIndent(payload, "", "  ")
	os.WriteFile(file, data, 0644)
}
