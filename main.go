package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/statusapp-io/statusapp-agent/checks"
)

const version = "1.0.0"

type config struct {
	apiURL       string
	agentKey     string
	instanceID   string
	pollInterval time.Duration
	concurrency  int
}

type pollResponse struct {
	Monitors []checks.Monitor `json:"monitors"`
}

type checkResult struct {
	MonitorID    string `json:"monitorId"`
	Status       string `json:"status"`
	ResponseTime *int   `json:"responseTime,omitempty"`
	StatusCode   *int   `json:"statusCode,omitempty"`
	Error        string `json:"error,omitempty"`
}

func main() {
	cfg := loadConfig()

	log.Printf("StatusApp Agent v%s starting", version)
	log.Printf("  API:        %s", cfg.apiURL)
	log.Printf("  Instance:   %s", cfg.instanceID)
	log.Printf("  Poll:       %s", cfg.pollInterval)
	log.Printf("  Concurrency: %d", cfg.concurrency)

	// Send initial heartbeat
	if err := sendHeartbeat(cfg); err != nil {
		log.Fatalf("Initial heartbeat failed — check your agent key and API URL: %v", err)
	}
	log.Println("Connected successfully")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	heartbeatTicker := time.NewTicker(30 * time.Second)
	pollTicker := time.NewTicker(cfg.pollInterval)

	defer heartbeatTicker.Stop()
	defer pollTicker.Stop()

	// Initial poll
	poll(cfg)

	for {
		select {
		case <-heartbeatTicker.C:
			if err := sendHeartbeat(cfg); err != nil {
				log.Printf("Heartbeat failed: %v", err)
			}
		case <-pollTicker.C:
			poll(cfg)
		case <-stop:
			log.Println("Shutting down...")
			return
		}
	}
}

func loadConfig() config {
	apiURL := os.Getenv("STATUSAPP_API_URL")
	if apiURL == "" {
		apiURL = os.Getenv("API_URL")
	}
	if apiURL == "" {
		log.Fatal("STATUSAPP_API_URL environment variable is required")
	}

	agentKey := os.Getenv("STATUSAPP_AGENT_KEY")
	if agentKey == "" {
		agentKey = os.Getenv("AGENT_KEY")
	}
	if agentKey == "" {
		log.Fatal("STATUSAPP_AGENT_KEY environment variable is required")
	}

	instanceID := os.Getenv("STATUSAPP_INSTANCE_ID")
	if instanceID == "" {
		// Generate a stable instance ID from hostname
		hostname, _ := os.Hostname()
		if hostname == "" {
			hostname = "unknown"
		}
		instanceID = fmt.Sprintf("%s-%d", hostname, os.Getpid())
	}

	pollInterval := 30 * time.Second
	if v := os.Getenv("STATUSAPP_POLL_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d >= 10*time.Second {
			pollInterval = d
		}
	}

	concurrency := 5
	if v := os.Getenv("STATUSAPP_CONCURRENCY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 50 {
			concurrency = n
		}
	}

	return config{
		apiURL:       apiURL,
		agentKey:     agentKey,
		instanceID:   instanceID,
		pollInterval: pollInterval,
		concurrency:  concurrency,
	}
}

func sendHeartbeat(cfg config) error {
	hostname, _ := os.Hostname()
	body, _ := json.Marshal(map[string]string{
		"instanceId":   cfg.instanceID,
		"agentVersion": version,
		"hostname":     hostname,
	})

	req, err := http.NewRequest("POST", cfg.apiURL+"/api/agent/heartbeat", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("X-Agent-Key", cfg.agentKey)
	req.Header.Set("X-Instance-Id", cfg.instanceID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}

	return nil
}

func poll(cfg config) {
	req, err := http.NewRequest("GET", cfg.apiURL+"/api/agent/poll", nil)
	if err != nil {
		log.Printf("Poll error: %v", err)
		return
	}
	req.Header.Set("X-Agent-Key", cfg.agentKey)
	req.Header.Set("X-Instance-Id", cfg.instanceID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Poll error: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		log.Printf("Poll error: HTTP %d: %s", resp.StatusCode, string(b))
		return
	}

	var pr pollResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		log.Printf("Poll decode error: %v", err)
		return
	}

	if len(pr.Monitors) == 0 {
		return
	}

	log.Printf("Received %d monitor(s) to check", len(pr.Monitors))

	// Execute checks with bounded concurrency
	sem := make(chan struct{}, cfg.concurrency)
	var mu sync.Mutex
	var results []checkResult

	var wg sync.WaitGroup
	for _, m := range pr.Monitors {
		wg.Add(1)
		go func(mon checks.Monitor) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			r := checks.Run(mon)

			mu.Lock()
			results = append(results, checkResult{
				MonitorID:    mon.ID,
				Status:       r.Status,
				ResponseTime: r.ResponseTime,
				StatusCode:   r.StatusCode,
				Error:        r.Error,
			})
			mu.Unlock()

			status := "UP"
			if r.Status != "UP" {
				status = r.Status
			}
			rt := 0
			if r.ResponseTime != nil {
				rt = *r.ResponseTime
			}
			if r.Error != "" {
				log.Printf("  [%s] %s %s — %s (%dms) error=%s", mon.Type, status, mon.Name, mon.URL, rt, r.Error)
			} else {
				log.Printf("  [%s] %s %s — %s (%dms)", mon.Type, status, mon.Name, mon.URL, rt)
			}
		}(m)
	}
	wg.Wait()

	if len(results) > 0 {
		submitResults(cfg, results)
	}
}

func submitResults(cfg config, results []checkResult) {
	body, _ := json.Marshal(map[string]interface{}{
		"instanceId": cfg.instanceID,
		"results":    results,
	})

	req, err := http.NewRequest("POST", cfg.apiURL+"/api/agent/results", bytes.NewReader(body))
	if err != nil {
		log.Printf("Submit error: %v", err)
		return
	}
	req.Header.Set("X-Agent-Key", cfg.agentKey)
	req.Header.Set("X-Instance-Id", cfg.instanceID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Submit error: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		log.Printf("Submit error: HTTP %d: %s", resp.StatusCode, string(b))
		return
	}

	var result map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	log.Printf("Submitted %d result(s), accepted: %v", len(results), result["accepted"])
}
