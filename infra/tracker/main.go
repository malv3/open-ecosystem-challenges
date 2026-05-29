package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	allowedType = "open-ecosystem-challenges"
	dtAPIPath   = "/api/v2/bizevents/ingest"
)

var validActions = map[string]bool{
	"codespace.created":      true,
	"codespace.initialized":  true,
	"smoketest.completed":    true,
	"verification.completed": true,
}

type bizEvent struct {
	Type         string   `json:"type"`
	Action       string   `json:"action"`
	Adventure    string   `json:"adventure"`
	Level        string   `json:"level"`
	GithubUser   string   `json:"github.user"`
	GithubRepo   string   `json:"github.repo"`
	CodespaceID  string   `json:"codespace.id"`
	Status       string   `json:"status,omitempty"`
	FailedChecks []string `json:"failed_checks,omitempty"`
}

func (e *bizEvent) validate() string {
	if e.Type != allowedType {
		return "type must be " + allowedType
	}
	if !validActions[e.Action] {
		return "unknown action: " + e.Action
	}
	if e.Adventure == "" {
		return "adventure is required"
	}
	if e.Level == "" {
		return "level is required"
	}
	if e.GithubUser == "" {
		return "github.user is required"
	}
	if e.GithubRepo == "" {
		return "github.repo is required"
	}
	if e.CodespaceID == "" {
		return "codespace.id is required"
	}
	return ""
}

func handler(dtURL, dtToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		var event bizEvent
		if err := json.Unmarshal(body, &event); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		if msg := event.validate(); msg != "" {
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, dtURL+dtAPIPath, bytes.NewReader(body))
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Authorization", "Api-Token "+dtToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("dynatrace ingest error: %v", err)
			http.Error(w, "failed to forward event", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			respBody, _ := io.ReadAll(resp.Body)
			log.Printf("dynatrace returned %d: %s", resp.StatusCode, respBody)
			http.Error(w, "dynatrace ingest failed", http.StatusBadGateway)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func main() {
	dtURL := os.Getenv("DT_TENANT_URL")
	dtToken := os.Getenv("DT_API_TOKEN")
	if dtURL == "" || dtToken == "" {
		log.Fatal("DT_TENANT_URL and DT_API_TOKEN must be set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", handler(dtURL, dtToken))
	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
