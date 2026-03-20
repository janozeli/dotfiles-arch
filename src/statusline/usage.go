package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	usageCacheFile = filepath.Join(os.TempDir(), "claude-usage-cache.json")
	usageLockFile  = filepath.Join(os.TempDir(), "claude-usage-lock.json")
)

const (
	cacheTTL    = 60 // seconds
	minInterval = 30 // seconds
)

type UsageCache struct {
	FiveHour   UsageTier  `json:"five_hour"`
	SevenDay   UsageTier  `json:"seven_day"`
	ExtraUsage ExtraUsage `json:"extra_usage"`
	CachedAt   int64      `json:"cached_at"`
}

type UsageTier struct {
	Utilization float64 `json:"utilization"`
	ResetsAt    string  `json:"resets_at"`
}

type ExtraUsage struct {
	IsEnabled    bool    `json:"is_enabled"`
	UsedCredits  float64 `json:"used_credits"`
	MonthlyLimit float64 `json:"monthly_limit"`
	Utilization  float64 `json:"utilization"`
}

type UsageLock struct {
	LastFetchAt  int64 `json:"last_fetch_at"`
	BackoffUntil int64 `json:"backoff_until"`
}

type UsageResponse struct {
	FiveHour   UsageTier  `json:"five_hour"`
	SevenDay   UsageTier  `json:"seven_day"`
	ExtraUsage ExtraUsage `json:"extra_usage"`
}

func loadUsageCache() (UsageCache, bool) {
	var cache UsageCache
	needFetch := true

	data, err := os.ReadFile(usageCacheFile)
	if err == nil {
		if json.Unmarshal(data, &cache) == nil {
			if time.Now().Unix()-cache.CachedAt < cacheTTL {
				needFetch = false
			}
		}
	}

	if needFetch {
		data, err := os.ReadFile(usageLockFile)
		if err == nil {
			var lock UsageLock
			if json.Unmarshal(data, &lock) == nil {
				now := time.Now().Unix()
				if lock.BackoffUntil > now {
					needFetch = false
				} else if now-lock.LastFetchAt < minInterval {
					needFetch = false
				}
			}
		}
	}

	return cache, needFetch
}

func fetchUsage() (*UsageResponse, int) {
	homeDir, _ := os.UserHomeDir()
	credsFile := filepath.Join(homeDir, ".claude", ".credentials.json")
	data, err := os.ReadFile(credsFile)
	if err != nil {
		return nil, 0
	}

	var creds struct {
		ClaudeAiOauth struct {
			AccessToken string `json:"accessToken"`
		} `json:"claudeAiOauth"`
	}
	if json.Unmarshal(data, &creds) != nil || creds.ClaudeAiOauth.AccessToken == "" {
		return nil, 0
	}

	out, err := runCmd(5*time.Second, "", "curl",
		"-s", "-w", "\n%{http_code}",
		"-H", "Authorization: Bearer "+creds.ClaudeAiOauth.AccessToken,
		"-H", "Content-Type: application/json",
		"-H", "anthropic-beta: oauth-2025-04-20",
		"https://api.anthropic.com/api/oauth/usage",
	)
	if err != nil {
		return nil, 0
	}

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) < 2 {
		return nil, 0
	}
	code := 0
	fmt.Sscanf(lines[len(lines)-1], "%d", &code)
	if code == 429 {
		return nil, 429
	}
	if code != 200 {
		return nil, code
	}

	body := strings.Join(lines[:len(lines)-1], "\n")
	var resp UsageResponse
	if json.Unmarshal([]byte(body), &resp) != nil {
		return nil, 0
	}
	return &resp, 200
}

func saveUsage() {
	now := time.Now().Unix()
	lock := UsageLock{LastFetchAt: now}
	if lockData, err := json.Marshal(lock); err == nil {
		os.WriteFile(usageLockFile, lockData, 0644)
	}

	resp, code := fetchUsage()
	if code == 429 {
		lock.BackoffUntil = now + 300
		if lockData, err := json.Marshal(lock); err == nil {
			os.WriteFile(usageLockFile, lockData, 0644)
		}
		return
	}
	if resp == nil {
		return
	}

	cache := UsageCache{
		FiveHour:   resp.FiveHour,
		SevenDay:   resp.SevenDay,
		ExtraUsage: resp.ExtraUsage,
		CachedAt:   now,
	}
	if cacheData, err := json.Marshal(cache); err == nil {
		os.WriteFile(usageCacheFile, cacheData, 0644)
	}
}

func usageSegment(cache UsageCache) string {
	var parts []string
	usageURL := "https://claude.ai/settings/usage"

	if cache.FiveHour.Utilization > 0 {
		util := math.Round(cache.FiveHour.Utilization)
		color := thresholdColor(util)
		reset := formatResetCompact(cache.FiveHour.ResetsAt)
		parts = append(parts, osc8Link(usageURL, CPurple+"\uf1da Session:"+Rst+" "+color+fmt.Sprintf("%.0f%%%s", util, reset))+Rst)
	}

	if cache.SevenDay.Utilization > 0 {
		util := math.Round(cache.SevenDay.Utilization)
		color := thresholdColor(util)
		reset := formatResetCompact(cache.SevenDay.ResetsAt)
		parts = append(parts, osc8Link(usageURL, CPurple+"\uf073 Weekly:"+Rst+" "+color+fmt.Sprintf("%.0f%%%s", util, reset))+Rst)
	}

	if cache.ExtraUsage.IsEnabled {
		used := cache.ExtraUsage.UsedCredits
		limit := cache.ExtraUsage.MonthlyLimit
		util := cache.ExtraUsage.Utilization
		color := thresholdColor(util)
		parts = append(parts, color+osc8Link(usageURL, fmt.Sprintf("Extra: $%.2f/$%.2f", used, limit))+Rst)
	}

	return strings.Join(parts, Sep)
}
