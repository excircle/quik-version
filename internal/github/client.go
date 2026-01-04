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
