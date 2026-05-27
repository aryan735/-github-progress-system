package github

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	gh "github.com/google/go-github/v60/github"
)

type Commit struct {
	SHA     string    `json:"sha"`
	Message string    `json:"message"`
	Author  string    `json:"author"`
	Date    time.Time `json:"date"`
}

type RepoCommitSummary struct {
	FullName    string   `json:"full_name"`
	CommitCount int      `json:"commit_count"`
	Commits     []Commit `json:"commits"`
}

type TodaySummary struct {
	Date             string              `json:"date"`
	Username         string              `json:"username"`
	TotalRepos       int                 `json:"total_repos"`
	ReposWithCommits int                 `json:"repos_with_commits"`
	TotalCommits     int                 `json:"total_commits"`
	Repos            []RepoCommitSummary `json:"repos"`
	Summary          string              `json:"summary"`
}

func startOfToday() time.Time {
	now := time.Now()
	y, m, d := now.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, now.Location())
}

func isToday(t time.Time) bool {
	now := time.Now()
	y, m, d := now.Date()
	cy, cm, cd := t.In(now.Location()).Date()
	return y == cy && m == cm && d == cd
}

func (c *Client) listAuthenticatedUserRepos(ctx context.Context) ([]*gh.Repository, error) {
	var all []*gh.Repository

	opts := &gh.RepositoryListByAuthenticatedUserOptions{
		Sort: "updated",
		ListOptions: gh.ListOptions{
			PerPage: 100,
		},
	}

	for {
		repos, resp, err := c.gh.Repositories.ListByAuthenticatedUser(ctx, opts)
		if err != nil {
			return nil, err
		}

		all = append(all, repos...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return all, nil
}

func (c *Client) listRepoCommitsSince(
	ctx context.Context,
	owner, repo string,
	since time.Time,
) ([]*gh.RepositoryCommit, error) {
	var all []*gh.RepositoryCommit

	opts := &gh.CommitsListOptions{
		Since: since,
		ListOptions: gh.ListOptions{
			PerPage: 100,
		},
	}

	for {
		commits, resp, err := c.gh.Repositories.ListCommits(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}

		all = append(all, commits...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return all, nil
}

func commitDate(commit *gh.RepositoryCommit) (time.Time, bool) {
	if commit == nil || commit.Commit == nil {
		return time.Time{}, false
	}

	if commit.Commit.Author != nil && commit.Commit.Author.Date != nil {
		return commit.Commit.Author.Date.Time, true
	}

	if commit.Commit.Committer != nil && commit.Commit.Committer.Date != nil {
		return commit.Commit.Committer.Date.Time, true
	}

	return time.Time{}, false
}

func toCommit(commit *gh.RepositoryCommit) (Commit, bool) {
	date, ok := commitDate(commit)
	if !ok || !isToday(date) {
		return Commit{}, false
	}

	msg := ""
	if commit.Commit != nil {
		msg = commit.Commit.GetMessage()
	}

	author := ""
	if commit.Commit != nil && commit.Commit.Author != nil {
		author = commit.Commit.Author.GetName()
	}

	return Commit{
		SHA:     commit.GetSHA(),
		Message: strings.TrimSpace(msg),
		Author:  author,
		Date:    date,
	}, true
}

func generateSummary(username string, totalRepos int, repos []RepoCommitSummary) TodaySummary {
	totalCommits := 0
	for _, r := range repos {
		totalCommits += r.CommitCount
	}

	sort.Slice(repos, func(i, j int) bool {
		return repos[i].CommitCount > repos[j].CommitCount
	})

	now := time.Now()
	summary := TodaySummary{
		Date:             now.Format("2006-01-02"),
		Username:         username,
		TotalRepos:       totalRepos,
		ReposWithCommits: len(repos),
		TotalCommits:     totalCommits,
		Repos:            repos,
	}

	if totalCommits == 0 {
		summary.Summary = fmt.Sprintf("%s made no commits today across %d repos.", username, totalRepos)
		return summary
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s made %d commit(s) today across %d repo(s):\n",
		username, totalCommits, len(repos))

	for _, r := range repos {
		fmt.Fprintf(&b, "- %s: %d commit(s)\n", r.FullName, r.CommitCount)
		for _, commit := range r.Commits {
			firstLine := commit.Message
			if idx := strings.IndexByte(firstLine, '\n'); idx >= 0 {
				firstLine = firstLine[:idx]
			}
			fmt.Fprintf(&b, "  • %s (%s)\n", firstLine, shortSHA(commit.SHA))
		}
	}

	summary.Summary = strings.TrimSpace(b.String())
	return summary
}

func shortSHA(sha string) string {
	if len(sha) >= 7 {
		return sha[:7]
	}
	return sha
}

// CollectTodayCommits lists repos, fetches commits since today, and builds a summary.
func (c *Client) CollectTodayCommits(ctx context.Context) (*TodaySummary, error) {
	user, _, err := c.gh.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("get authenticated user: %w", err)
	}

	repos, err := c.listAuthenticatedUserRepos(ctx)
	if err != nil {
		return nil, fmt.Errorf("list repos: %w", err)
	}

	since := startOfToday()
	var repoSummaries []RepoCommitSummary

	for _, repo := range repos {
		owner := repo.GetOwner().GetLogin()
		name := repo.GetName()

		commits, err := c.listRepoCommitsSince(ctx, owner, name, since)
		if err != nil {
			continue
		}

		var todayCommits []Commit
		for _, commit := range commits {
			c, ok := toCommit(commit)
			if ok {
				todayCommits = append(todayCommits, c)
			}
		}

		if len(todayCommits) > 0 {
			repoSummaries = append(repoSummaries, RepoCommitSummary{
				FullName:    repo.GetFullName(),
				CommitCount: len(todayCommits),
				Commits:     todayCommits,
			})
		}
	}

	result := generateSummary(user.GetLogin(), len(repos), repoSummaries)
	return &result, nil
}
