package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/excircle/quik-version/internal/db"
)

const configFileName = "quik.conf"

type quikConfig struct {
	Version struct {
		GitURL string `yaml:"git_url"`
		Token  string `yaml:"token,omitempty"`
	} `yaml:"version"`
	Build struct {
		BuildManagement bool `yaml:"build_management"`
	} `yaml:"build"`
	Storage struct {
		DBPath string `yaml:"db_path"`
	} `yaml:"storage"`
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Quik Version configuration",
	Long: `Initialize Quik Version by creating quik.conf and qv.db.

This command will:
- Check for existing quik.conf and prompt to overwrite or skip
- Prompt for git_url
- Prompt for GitHub token
- Create qv.db with the required schema`,
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		// Check for existing quik.conf
		if _, err := os.Stat(configFileName); err == nil {
			fmt.Printf("%s already exists. Overwrite? (y/n): ", configFileName)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Skipping config creation.")
				return nil
			}
		}

		// Prompt for git_url
		fmt.Print("Enter git URL (e.g., https://github.com/user/repo): ")
		gitURL, _ := reader.ReadString('\n')
		gitURL = strings.TrimSpace(gitURL)
		if gitURL == "" {
			return fmt.Errorf("git URL cannot be empty")
		}

		// Prompt for GitHub token
		fmt.Print("Enter GitHub token (leave empty to use GITHUB_TOKEN env var): ")
		token, _ := reader.ReadString('\n')
		token = strings.TrimSpace(token)

		// Prompt for save token preference
		saveToken := false
		if token != "" {
			fmt.Print("Save token to config file? (y/n): ")
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			saveToken = response == "y" || response == "yes"
		}

		// Prompt for db path
		fmt.Print("Enter database path (leave empty for current directory): ")
		dbPath, _ := reader.ReadString('\n')
		dbPath = strings.TrimSpace(dbPath)

		// Create config
		config := quikConfig{}
		config.Version.GitURL = gitURL
		if saveToken {
			config.Version.Token = token
		}
		config.Build.BuildManagement = false
		config.Storage.DBPath = dbPath

		// Write config file
		configData, err := yaml.Marshal(&config)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		if err := os.WriteFile(configFileName, configData, 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}
		fmt.Printf("Created %s\n", configFileName)

		// Check for existing database
		if db.Exists() {
			fmt.Printf("qv.db already exists. Overwrite? (y/n): ")
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Skipping database creation.")
				return nil
			}
			// Remove existing database
			if err := os.Remove(db.GetDBPath()); err != nil {
				return fmt.Errorf("failed to remove existing database: %w", err)
			}
		}

		// Create and initialize database
		database, err := db.Open()
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer database.Close()

		if err := database.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}

		if err := database.SetConfigState(gitURL); err != nil {
			return fmt.Errorf("failed to set config state: %w", err)
		}

		fmt.Printf("Created %s\n", db.GetDBPath())
		fmt.Println("Initialization complete!")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
