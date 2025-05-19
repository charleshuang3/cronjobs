package bottesting

import (
	"context"
	"fmt"

	"github.com/charleshuang3/cronjobs/internal/bot"
)

// StubBot is a stub implementation of the bot.Bot interface for testing
type StubBot struct {
	SentMessage      string
	JobExecuteResult bot.JobExecuteResult
	ShouldError      bool
	ListJobsOutput   string
	RunJobOutput     string
	StatusOutput     string
}

func (sb *StubBot) Send(message string) error {
	if sb.ShouldError {
		return fmt.Errorf("stub bot error")
	}
	sb.SentMessage = message
	return nil
}

func (sb *StubBot) SendJobExecuteResult(result bot.JobExecuteResult) error {
	if sb.ShouldError {
		return fmt.Errorf("stub bot error")
	}
	sb.JobExecuteResult = result
	return nil
}

func (sb *StubBot) StartWebhook(ctx context.Context) error {
	if sb.ShouldError {
		return fmt.Errorf("stub bot error starting webhook")
	}
	return nil
}

func (sb *StubBot) StopWebhook(ctx context.Context) error {
	if sb.ShouldError {
		return fmt.Errorf("stub bot error stopping webhook")
	}
	return nil
}

func (sb *StubBot) HandleListCommand(ctx context.Context) (string, error) {
	if sb.ShouldError {
		return "", fmt.Errorf("stub bot error listing")
	}
	return sb.ListJobsOutput, nil
}

func (sb *StubBot) HandleRunCommand(ctx context.Context, jobName string) (string, error) {
	if sb.ShouldError {
		return "", fmt.Errorf("stub bot error running")
	}
	return sb.RunJobOutput, nil

}

func (sb *StubBot) HandleStatusCommand(ctx context.Context) (string, error) {
	if sb.ShouldError {
		return "", fmt.Errorf("stub bot error getting status")
	}
	return sb.StatusOutput, nil
}
