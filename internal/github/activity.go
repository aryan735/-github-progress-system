package github

import (
	"context"
	"time"

	gh "github.com/google/go-github/v60/github"
)

type Activity struct {
	Repo    string
	Type    string
	Message string
	SHA     string
	Date    time.Time
}

func FilterTodayActivities(
	activities []Activity,
) []Activity {

	now := time.Now()

	year, month, day := now.Date()

	var result []Activity

	for _, a := range activities {

		y, m, d := a.Date.Date()

		if y == year &&
			m == month &&
			d == day {

			result = append(result, a)
		}
	}

	return result
}
func (c *Client) GetUserEvents(
	username string,
	perPage int,
) ([]Activity, error) {

	ctx := context.Background()

	opts := &gh.ListOptions{
		PerPage: perPage,
	}

	events, _, err :=
		c.gh.Activity.ListEventsPerformedByUser(
			ctx,
			username,
			false,
			opts,
		)

	if err != nil {
		return nil, err
	}

	var result []Activity

	for _, e := range events {

		if e.CreatedAt == nil {
			continue
		}

		repoName := "unknown"

		if e.Repo != nil {
			repoName = e.Repo.GetName()
		}

		eventType := e.GetType()

		// Parse payload safely
		payload, err := e.ParsePayload()
		if err != nil {
			continue
		}

		// PushEvent contains commits
		if push, ok := payload.(*gh.PushEvent); ok {

			for _, commit := range push.Commits {

				result = append(result, Activity{
					Repo:    repoName,
					Type:    eventType,
					Message: commit.GetMessage(),
					SHA:     commit.GetSHA(),
					Date:    e.CreatedAt.Time,
				})
			}

			continue
		}

		// fallback for PR / issue / others
		result = append(result, Activity{
			Repo: repoName,
			Type: eventType,
			Date: e.CreatedAt.Time,
		})
	}

	return result, nil
}
