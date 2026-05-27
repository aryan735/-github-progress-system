package github

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	gh "github.com/google/go-github/v60/github"
)

const progressLogFileHeader = "# Developer Progress Log\n\nDaily GitHub activity tracked automatically.\n\n"

type ProgressLogTarget struct {
	Owner  string
	Repo   string
	Branch string
	Path   string
}

type ProgressLogSyncResult struct {
	Path          string `json:"path"`
	CommitMessage string `json:"commit_message"`
	CommitSHA     string `json:"commit_sha,omitempty"`
	Created       bool   `json:"created"`
}

type DailyProgressResult struct {
	Summary *TodaySummary        `json:"summary"`
	Sync    *ProgressLogSyncResult `json:"sync"`
}

func firstLine(message string) string {
	message = strings.TrimSpace(message)
	if idx := strings.IndexByte(message, '\n'); idx >= 0 {
		return strings.TrimSpace(message[:idx])
	}
	return message
}

func flatCommits(repos []RepoCommitSummary) []Commit {
	var all []Commit
	for _, r := range repos {
		all = append(all, r.Commits...)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].Date.Before(all[j].Date)
	})
	return all
}

func repoNames(repos []RepoCommitSummary) []string {
	names := make([]string, len(repos))
	for i, r := range repos {
		names[i] = r.FullName
	}
	sort.Strings(names)
	return names
}

func FormatProgressLogEntry(summary *TodaySummary) string {
	var b strings.Builder

	fmt.Fprintf(&b, "## %s\n\n", summary.Date)

	if summary.TotalCommits == 0 {
		fmt.Fprintf(&b, "Status: No GitHub commits today\n\n")
		fmt.Fprintf(&b, "Repositories worked on:\n- none\n\n")
		fmt.Fprintf(&b, "Total commits: 0\n\n")
		fmt.Fprintf(&b, "Notes:\n")
		fmt.Fprintf(&b, "- No commits detected across GitHub today.\n")
		fmt.Fprintf(&b, "- Daily progress log updated automatically.\n")
		return b.String()
	}

	fmt.Fprintf(&b, "Status: Productive day\n\n")
	fmt.Fprintf(&b, "Repositories worked on:\n")
	for _, name := range repoNames(summary.Repos) {
		fmt.Fprintf(&b, "- %s\n", name)
	}

	fmt.Fprintf(&b, "\nTotal commits: %d\n\n", summary.TotalCommits)
	fmt.Fprintf(&b, "Commits:\n")
	for _, commit := range flatCommits(summary.Repos) {
		fmt.Fprintf(&b, "- %s\n", firstLine(commit.Message))
	}

	return b.String()
}

func ProgressLogCommitMessage(summary *TodaySummary) string {
	if summary.TotalCommits == 0 {
		return fmt.Sprintf("docs: add no-activity log for %s", summary.Date)
	}
	return fmt.Sprintf("docs: update progress log for %s", summary.Date)
}

func upsertProgressLogSection(existing, date, entry string) string {
	entry = strings.TrimRight(entry, "\n") + "\n"

	if existing == "" {
		return progressLogFileHeader + entry
	}

	marker := "## " + date
	if idx := strings.Index(existing, marker); idx >= 0 {
		tail := existing[idx+len(marker):]
		if endRel := strings.Index(tail, "\n## "); endRel >= 0 {
			end := idx + len(marker) + endRel + 1
			return existing[:idx] + entry + existing[end:]
		}
		return existing[:idx] + entry
	}

	if strings.HasPrefix(existing, "# Developer Progress Log") {
		headerEnd := strings.Index(existing, "\n\n")
		if headerEnd >= 0 {
			insertAt := headerEnd + 2
			return existing[:insertAt] + entry + "\n" + existing[insertAt:]
		}
	}

	return progressLogFileHeader + entry + "\n" + existing
}

func (c *Client) getProgressLogFile(
	ctx context.Context,
	target ProgressLogTarget,
) (content string, sha string, err error) {
	opts := &gh.RepositoryContentGetOptions{Ref: target.Branch}
	file, _, resp, err := c.gh.Repositories.GetContents(
		ctx,
		target.Owner,
		target.Repo,
		target.Path,
		opts,
	)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return "", "", nil
		}
		return "", "", err
	}

	if file == nil {
		return "", "", nil
	}

	decoded, err := file.GetContent()
	if err != nil {
		return "", "", err
	}

	return decoded, file.GetSHA(), nil
}

func (c *Client) SyncProgressLog(
	ctx context.Context,
	target ProgressLogTarget,
	summary *TodaySummary,
) (*ProgressLogSyncResult, error) {
	entry := FormatProgressLogEntry(summary)
	message := ProgressLogCommitMessage(summary)

	existing, sha, err := c.getProgressLogFile(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("read progress log: %w", err)
	}

	newContent := upsertProgressLogSection(existing, summary.Date, entry)
	opts := &gh.RepositoryContentFileOptions{
		Message: gh.String(message),
		Content: []byte(newContent),
		Branch:  gh.String(target.Branch),
	}

	result := &ProgressLogSyncResult{
		Path:          target.Path,
		CommitMessage: message,
	}

	if sha == "" {
		resp, _, err := c.gh.Repositories.CreateFile(
			ctx,
			target.Owner,
			target.Repo,
			target.Path,
			opts,
		)
		if err != nil {
			return nil, fmt.Errorf("create progress log: %w", err)
		}
		result.Created = true
		if resp != nil {
			result.CommitSHA = resp.GetSHA()
		}
		return result, nil
	}

	opts.SHA = gh.String(sha)
	resp, _, err := c.gh.Repositories.UpdateFile(
		ctx,
		target.Owner,
		target.Repo,
		target.Path,
		opts,
	)
	if err != nil {
		return nil, fmt.Errorf("update progress log: %w", err)
	}

	if resp != nil {
		result.CommitSHA = resp.GetSHA()
	}

	return result, nil
}
