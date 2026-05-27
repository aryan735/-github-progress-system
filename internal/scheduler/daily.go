package scheduler

import (
	"context"
	"fmt"

	"github.com/aryan735/-github-progress-system/internal/github"
)

// DailyJobTimeIST is when the GitHub Actions workflow runs (10:30 PM IST).
const DailyJobTimeIST = "22:30"

type DailyService struct {
	client *github.Client
	target github.ProgressLogTarget
}

func NewDailyService(client *github.Client, target github.ProgressLogTarget) *DailyService {
	return &DailyService{
		client: client,
		target: target,
	}
}

// Run scans all GitHub repos for today's commits and writes the log only
// to the configured progress repo (never to other repositories).
func (s *DailyService) Run(ctx context.Context) (*github.DailyProgressResult, error) {
	summary, err := s.client.CollectTodayCommits(ctx, s.target.FullName())
	if err != nil {
		return nil, fmt.Errorf("collect today commits: %w", err)
	}

	syncResult, err := s.client.SyncProgressLog(ctx, s.target, summary)
	if err != nil {
		return nil, fmt.Errorf("sync progress log to %s/%s: %w",
			s.target.Owner, s.target.Repo, err)
	}

	return &github.DailyProgressResult{
		Summary: summary,
		Sync:    syncResult,
	}, nil
}
