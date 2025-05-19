package config

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"github.com/charleshuang3/cronjobs/internal/bot"
	"gopkg.in/yaml.v3"
)

var jobNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// JobConfig represents a single cron job configuration
type JobConfig struct {
	Name          string   `yaml:"name"`
	Cron          string   `yaml:"cron"`
	Program       string   `yaml:"program"`
	Args          []string `yaml:"args"`
	CaptureOutput bool     `yaml:"capture_output"`
}

// Config represents the main configuration structure
type Config struct {
	BotType string `yaml:"bot"`

	// Bot configs
	TelegramConfig *bot.TelegramConfig `yaml:"bot.telegram"`

	// Jobs configuration
	Jobs []JobConfig `yaml:"jobs"`
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.BotType == "" {
		return fmt.Errorf("bot type is required")
	}

	if c.BotType != "telegram" {
		return fmt.Errorf("unsupported bot type: %s", c.BotType)
	}

	if c.BotType == "telegram" {
		if c.TelegramConfig == nil {
			return fmt.Errorf("telegram configuration is required")
		}
		if err := c.TelegramConfig.Validate(); err != nil {
			return fmt.Errorf("invalid telegram configuration: %w", err)
		}
	}

	if len(c.Jobs) == 0 {
		return fmt.Errorf("at least one job must be configured")
	}

	// Track job names to ensure uniqueness
	jobNames := make(map[string]bool)
	for i, job := range c.Jobs {
		if err := validateJob(job, jobNames); err != nil {
			return fmt.Errorf("invalid job at index %d: %w", i, err)
		}
		jobNames[job.Name] = true
	}

	return nil
}

// validateJob checks if a single job configuration is valid
func validateJob(job JobConfig, existingNames map[string]bool) error {
	if job.Name == "" {
		return fmt.Errorf("job name is required")
	}

	if !jobNameRegex.MatchString(job.Name) {
		return fmt.Errorf("job name '%s' must only contain letters, numbers, underscores, and hyphens", job.Name)
	}

	if existingNames[job.Name] {
		return fmt.Errorf("duplicate job name '%s'", job.Name)
	}

	if job.Cron == "" {
		return fmt.Errorf("cron expression is required")
	}

	if job.Program == "" {
		return fmt.Errorf("program is required")
	}

	// Check if program exists in PATH or as a full path
	if _, err := exec.LookPath(job.Program); err != nil {
		// Check if it's a full path and executable
		fileInfo, err := os.Stat(job.Program)
		if err != nil {
			return fmt.Errorf("program '%s' not found in PATH or as a valid file", job.Program)
		}
		if fileInfo.IsDir() {
			return fmt.Errorf("program '%s' is a directory, not an executable", job.Program)
		}
		if mode := fileInfo.Mode(); !mode.IsRegular() || mode&0111 == 0 {
			return fmt.Errorf("program '%s' is not executable", job.Program)
		}
	}

	return nil
}
