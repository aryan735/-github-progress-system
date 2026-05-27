package main

import (
	"log"

	"github.com/aryan735/-github-progress-system/internal/config"
	"github.com/aryan735/-github-progress-system/internal/github"
	"github.com/aryan735/-github-progress-system/internal/handler"
	"github.com/aryan735/-github-progress-system/internal/scheduler"
	"github.com/labstack/echo/v4"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	client := github.NewGitHubClient(cfg.GithubToken)
	daily := scheduler.NewDailyService(client, cfg.ProgressLogTarget())
	h := handler.New(client, daily)

	e := echo.New()
	h.Register(e)

	e.Logger.Fatal(e.Start(":8080"))
}
