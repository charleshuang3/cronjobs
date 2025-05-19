package bot

import (
	"context"
	"time"
)

// Command represents a bot command with its arguments
type Command struct {
	Name string
	Args []string
}

// JobExecuteResult holds the result of a job execution
type JobExecuteResult struct {
	JobName       string
	Duration      time.Duration
	Output        string
	CaptureOutput bool
	Error         error
}

// Bot defines the interface for bot implementations
type Bot interface {
	// Send sends a message
	Send(message string) error

	// StartWebhook starts the webhook server if webhook is configured
	StartWebhook(ctx context.Context) error

	// StopWebhook stops the webhook server
	StopWebhook(ctx context.Context) error

	// HandleListCommand processes the 'list' command
	HandleListCommand(ctx context.Context) (string, error)

	// HandleRunCommand processes the 'run' command with specified arguments
	HandleRunCommand(ctx context.Context, jobName string) (string, error)

	// HandleStatusCommand processes the 'status' command
	HandleStatusCommand(ctx context.Context) (string, error)

	// SendJobExecuteResult sends the execution result of a job
	SendJobExecuteResult(result JobExecuteResult) error
}

// Config is the interface that bot configs must implement
type Config interface {
	// Validate validates the configuration
	Validate() error
}

// JobScheduler interface defines methods that the scheduler must implement for bot commands
type JobScheduler interface {
	ListJobs() []JobInfo
	RunJobByName(name string) error
	RunningJobs() map[string]time.Duration // Added method to get running jobs with their durations
}

// JobInfo holds information about a scheduled job
type JobInfo struct {
	Name          string
	Program       string
	Args          []string
	Cron          string
	CaptureOutput bool
}
