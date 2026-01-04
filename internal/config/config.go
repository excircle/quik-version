package config

import "github.com/spf13/viper"

// Config represents the full configuration structure
type Config struct {
	Version VersionConfig `mapstructure:"version"`
	Build   BuildConfig   `mapstructure:"build"`
	Storage StorageConfig `mapstructure:"storage"`
}

// VersionConfig holds version-related settings
type VersionConfig struct {
	GitURL string `mapstructure:"git_url"`
	Token  string `mapstructure:"token"`
}

// BuildConfig holds build-related settings
type BuildConfig struct {
	BuildManagement bool `mapstructure:"build_management"`
}

// StorageConfig holds storage-related settings
type StorageConfig struct {
	DBPath string `mapstructure:"db_path"`
}

// Load reads the configuration from Viper into a Config struct
func Load() (*Config, error) {
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

// GetGitURL returns the configured git URL
func GetGitURL() string {
	return viper.GetString("version.git_url")
}

// GetToken returns the configured token (may be empty)
func GetToken() string {
	return viper.GetString("version.token")
}

// GetBuildManagement returns whether build management is enabled
func GetBuildManagement() bool {
	return viper.GetBool("build.build_management")
}

// GetDBPath returns the configured database path
func GetDBPath() string {
	return viper.GetString("storage.db_path")
}
