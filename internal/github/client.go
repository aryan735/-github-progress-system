
package github

import (
	"context"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

type Client struct {
	gh *github.Client
}

func NewClient(token string) *Client  {
	ts:= oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
   tc:= oauth2.NewClient(context.Background(), ts)
   return &Client{gh: github.NewClient(tc)}
}