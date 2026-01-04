package github

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v80/github"
	"golang.org/x/oauth2"

	"github.com/excircle/quik-version/internal/config"
)

// Client wraps the GitHub client with authentication
type Client struct {
	*github.Client
	token string
}

// NewClient creates a new GitHub client using auth flow:
// 1. Check for GITHUB_TOKEN environment variable
// 2. Check if Token is defined in quik.conf
// 3. Interactive prompt - ask user for Token
func NewClient(ctx context.Context) (*Client, error) {
	token, err := getToken()
	if err != nil {
		return nil, err
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return &Client{
		Client: client,
		token:  token,
	}, nil
}

// getToken retrieves the GitHub token using the auth flow
func getToken() (string, error) {
	// 1. Check GITHUB_TOKEN environment variable
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}

	// 2. Check token in config file
	if token := config.GetToken(); token != "" {
		return token, nil
	}

	// 3. Interactive prompt
	return promptForToken()
}

func promptForToken() (string, error) {
	fmt.Print("Enter GitHub token: ")
	reader := bufio.NewReader(os.Stdin)
	token, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read token: %w", err)
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return "", fmt.Errorf("token cannot be empty")
	}

	return token, nil
}

// GetToken returns the token used by this client
func (c *Client) GetToken() string {
	return c.token
}

// Tag represents a GitHub tag
type Tag struct {
	Name   string
	SHA    string
}

// ListTags fetches all tags from a GitHub repository
func (c *Client) ListTags(ctx context.Context, owner, repo string) ([]Tag, error) {
	var allTags []Tag
	opts := &github.ListOptions{PerPage: 100}

	for {
		tags, resp, err := c.Repositories.ListTags(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list tags: %w", err)
		}

		for _, tag := range tags {
			allTags = append(allTags, Tag{
				Name: tag.GetName(),
				SHA:  tag.GetCommit().GetSHA(),
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allTags, nil
}

// ParseRepoURL extracts owner and repo from a GitHub URL
func ParseRepoURL(url string) (owner, repo string, err error) {
	url = strings.TrimSuffix(url, ".git")
	url = strings.TrimPrefix(url, "https://github.com/")
	url = strings.TrimPrefix(url, "http://github.com/")
	url = strings.TrimPrefix(url, "git@github.com:")

	parts := strings.Split(url, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid GitHub URL format")
	}

	return parts[0], parts[1], nil
}

// PullRequest represents a created PR
type PullRequest struct {
	Number  int
	URL     string
	Title   string
}

// CreatePR creates a pull request from head branch to base branch
func (c *Client) CreatePR(ctx context.Context, owner, repo, title, body, head, base string) (*PullRequest, error) {
	pr := &github.NewPullRequest{
		Title: github.Ptr(title),
		Body:  github.Ptr(body),
		Head:  github.Ptr(head),
		Base:  github.Ptr(base),
	}

	created, _, err := c.PullRequests.Create(ctx, owner, repo, pr)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	return &PullRequest{
		Number: created.GetNumber(),
		URL:    created.GetHTMLURL(),
		Title:  created.GetTitle(),
	}, nil
}

// GetLatestCommitSHA gets the SHA of the latest commit on a branch
func (c *Client) GetLatestCommitSHA(ctx context.Context, owner, repo, branch string) (string, error) {
	ref, _, err := c.Git.GetRef(ctx, owner, repo, "refs/heads/"+branch)
	if err != nil {
		return "", fmt.Errorf("failed to get branch ref: %w", err)
	}
	return ref.GetObject().GetSHA(), nil
}

// CreateTag creates an annotated tag on a specific commit
func (c *Client) CreateTag(ctx context.Context, owner, repo, tagName, commitSHA, message string) error {
	// Create the tag object
	tag := github.CreateTag{
		Tag:     tagName,
		Message: message,
		Object:  commitSHA,
		Type:    "commit",
	}

	createdTag, _, err := c.Git.CreateTag(ctx, owner, repo, tag)
	if err != nil {
		return fmt.Errorf("failed to create tag object: %w", err)
	}

	// Create the reference pointing to the tag
	ref := github.CreateRef{
		Ref: "refs/tags/" + tagName,
		SHA: createdTag.GetSHA(),
	}

	_, _, err = c.Git.CreateRef(ctx, owner, repo, ref)
	if err != nil {
		return fmt.Errorf("failed to create tag reference: %w", err)
	}

	return nil
}
