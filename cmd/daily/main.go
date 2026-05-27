package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aryan735/-github-progress-system/internal/config"
	"github.com/aryan735/-github-progress-system/internal/github"
	"github.com/aryan735/-github-progress-system/internal/scheduler"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	client := github.NewGitHubClient(cfg.GithubToken)
	daily := scheduler.NewDailyService(client, cfg.ProgressLogTarget())

	result, err := daily.Run(context.Background())
	if err != nil {
		log.Fatalf("daily progress: %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		log.Fatalf("encode result: %v", err)
	}
}
