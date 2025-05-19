package bot

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/charleshuang3/cronjobs/internal/utils"
)

// TelegramConfig holds configuration for Telegram bot
type TelegramConfig struct {
	BotToken    string `yaml:"bot_token"`
	ChatID      int64  `yaml:"chat_id"`
	Webhook     string `yaml:"webhook,omitempty"`
	WebhookPort int    `yaml:"webhook_port,omitempty"`
	WebhookPath string `yaml:"webhook_path,omitempty"`
}

// Validate implements Config interface
func (c *TelegramConfig) Validate() error {
	if c.BotToken == "" {
		return fmt.Errorf("telegram bot token is required")
	}
	if c.ChatID == 0 {
		return fmt.Errorf("telegram chat ID is required")
	}
	if c.Webhook != "" {
		if _, err := url.Parse(c.Webhook); err != nil {
			return fmt.Errorf("invalid webhook URL: %v", err)
		}
		if !strings.HasPrefix(c.Webhook, "https://") {
			return fmt.Errorf("webhook URL must use HTTPS")
		}

		if c.WebhookPath == "" {
			return fmt.Errorf("webhook_path cannot be empty when webhook is set")
		}

		if c.WebhookPort == 0 {
			return fmt.Errorf("webhook_port cannot be empty when webhook is set")
		}
	}

	return nil
}

// TelegramBot implements the Bot interface for Telegram
type TelegramBot struct {
	config     *TelegramConfig
	bot        *gotgbot.Bot
	dispatcher *ext.Dispatcher
	scheduler  JobScheduler
	updater    *ext.Updater
}

// NewTelegramBot creates a new TelegramBot
func NewTelegramBot(config *TelegramConfig, scheduler JobScheduler) (*TelegramBot, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	bot, err := gotgbot.NewBot(config.BotToken, &gotgbot.BotOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			fmt.Printf("error handling update: %v\n", err)
			return ext.DispatcherActionNoop
		},
	})

	updater := ext.NewUpdater(dispatcher, &ext.UpdaterOpts{
		ErrorLog: nil,
	})

	return &TelegramBot{
		config:     config,
		bot:        bot,
		dispatcher: dispatcher,
		scheduler:  scheduler,
		updater:    updater,
	}, nil
}

// Send implements Bot interface
func (t *TelegramBot) Send(message string) error {
	chatID := t.config.ChatID

	// Escape the message before sending
	escapedMessage := t.messageEscape(message)

	_, err := t.bot.SendMessage(chatID, escapedMessage, &gotgbot.SendMessageOpts{
		ParseMode: "MarkdownV2",
	})
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}

	return nil
}

// messageEscape escapes MarkdownV2 special characters in a message
func (t *TelegramBot) messageEscape(input string) string {
	escapers := []struct {
		old string
		new string
	}{
		{".", "\\."},
		{"-", "\\-"},
		{"_", "\\_"},
	}
	for _, escaper := range escapers {
		input = strings.ReplaceAll(input, escaper.old, escaper.new)
	}
	return input
}

// HandleListCommand implements Bot interface
func (t *TelegramBot) HandleListCommand(ctx context.Context) (string, error) {
	jobs := t.scheduler.ListJobs()
	if len(jobs) == 0 {
		return "No jobs scheduled", nil
	}

	var sb strings.Builder
	sb.WriteString("*Scheduled jobs:*\n\n")
	for _, job := range jobs {
		sb.WriteString(fmt.Sprintf("*%s* \\(Schedule: `%s`\\)\n",
			job.Name,
			job.Cron,
		))
	}
	return sb.String(), nil
}

// HandleRunCommand implements Bot interface
func (t *TelegramBot) HandleRunCommand(ctx context.Context, jobName string) (string, error) {
	if err := t.scheduler.RunJobByName(jobName); err != nil {
		return "", fmt.Errorf("failed to run job %s: %w", jobName, err)
	}

	return fmt.Sprintf("Job *%s* started", jobName), nil
}

// HandleStatusCommand implements Bot interface
func (t *TelegramBot) HandleStatusCommand(ctx context.Context) (string, error) {
	runningJobs := t.scheduler.RunningJobs()
	if len(runningJobs) == 0 {
		return "No jobs are currently running", nil
	}

	var sb strings.Builder
	sb.WriteString("*Running jobs:*\n\n")
	for jobName, duration := range runningJobs {
		formattedDuration := utils.PrettyPrintDuration(duration)
		sb.WriteString(fmt.Sprintf("*%s*: Running for %s\n", jobName, formattedDuration))
	}
	return sb.String(), nil
}

// StartWebhook implements Bot interface
func (t *TelegramBot) StartWebhook(ctx context.Context) error {
	if t.config.Webhook == "" {
		return nil
	}

	// Add command handlers
	t.dispatcher.AddHandler(handlers.NewCommand("list", t.handleListCommandWebhook))
	t.dispatcher.AddHandler(handlers.NewCommand("run", t.handleRunCommandWebhook))
	t.dispatcher.AddHandler(handlers.NewCommand("status", t.handleStatusCommandWebhook))

	commands := []gotgbot.BotCommand{
		{Command: "list", Description: "List scheduled jobs"},
		{Command: "status", Description: "Show status of running jobs"},
	}
	if _, err := t.bot.SetMyCommands(commands, &gotgbot.SetMyCommandsOpts{Scope: gotgbot.BotCommandScopeDefault{}}); err != nil {
		return fmt.Errorf("failed to set bot commands: %w", err)
	}

	// Start webhook server
	if err := t.updater.StartWebhook(t.bot, t.config.WebhookPath, ext.WebhookOpts{ListenAddr: fmt.Sprintf(":%d", t.config.WebhookPort)}); err != nil {
		return fmt.Errorf("failed to start webhook: %w", err)
	}

	// Set webhook in Telegram
	if _, err := t.bot.SetWebhook(t.config.Webhook, &gotgbot.SetWebhookOpts{}); err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}

	return nil
}

// StopWebhook implements Bot interface
func (t *TelegramBot) StopWebhook(ctx context.Context) error {
	if t.updater != nil {
		return t.updater.Stop()
	}
	return nil
}

func (t *TelegramBot) handleListCommandWebhook(bot *gotgbot.Bot, ctx *ext.Context) error {
	chatID := t.config.ChatID

	if ctx.EffectiveChat.Id != chatID {
		return nil
	}

	resp, err := t.HandleListCommand(context.TODO())
	if err != nil {
		_, _ = ctx.EffectiveMessage.Reply(bot, fmt.Sprintf("Error: `%v`", t.messageEscape(err.Error())), &gotgbot.SendMessageOpts{
			ParseMode: "MarkdownV2",
		})
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(bot, t.messageEscape(resp), &gotgbot.SendMessageOpts{
		ParseMode: "MarkdownV2",
	})
	return err
}

func (t *TelegramBot) handleRunCommandWebhook(bot *gotgbot.Bot, ctx *ext.Context) error {
	chatID := t.config.ChatID

	if ctx.EffectiveChat.Id != chatID {
		return nil
	}

	args := ctx.Args()
	if len(args) != 2 {
		_, _ = ctx.EffectiveMessage.Reply(bot, "Error: run command requires exactly one argument", nil)
		return fmt.Errorf("run command requires exactly one argument")
	}

	resp, err := t.HandleRunCommand(context.TODO(), args[1])
	if err != nil {
		_, _ = ctx.EffectiveMessage.Reply(bot, fmt.Sprintf("Error: `%v`", t.messageEscape(err.Error())), &gotgbot.SendMessageOpts{
			ParseMode: "MarkdownV2",
		})
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(bot, t.messageEscape(resp), &gotgbot.SendMessageOpts{
		ParseMode: "MarkdownV2",
	})
	return err
}

func (t *TelegramBot) handleStatusCommandWebhook(bot *gotgbot.Bot, ctx *ext.Context) error {
	chatID := t.config.ChatID

	if ctx.EffectiveChat.Id != chatID {
		return nil
	}

	resp, err := t.HandleStatusCommand(context.TODO())
	if err != nil {
		_, _ = ctx.EffectiveMessage.Reply(bot, fmt.Sprintf("Error: `%v`", t.messageEscape(err.Error())), &gotgbot.SendMessageOpts{
			ParseMode: "MarkdownV2",
		})
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(bot, t.messageEscape(resp), &gotgbot.SendMessageOpts{
		ParseMode: "MarkdownV2",
	})
	return err
}

// SendJobExecuteResult implements Bot interface
func (t *TelegramBot) SendJobExecuteResult(result JobExecuteResult) error {
	resStr := success
	if result.Error != nil {
		resStr = failed
	}

	durationStr := utils.PrettyPrintDuration(result.Duration)

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf(notificationFormat, resStr, result.JobName, durationStr))

	remaining := maxMessageLen - msg.Len()

	// some program (eg. ls) output the error message to stdout, so we send output even captureOutput is false.
	if result.CaptureOutput || result.Error != nil {
		output := fmt.Sprintf(outputFormat, result.Output)
		// we don't have enough space for the output, so we'll send it via file.
		if len(output) > remaining {
			file := gotgbot.InputFileByReader(
				fmt.Sprintf("%s.log", result.JobName),
				strings.NewReader(result.Output),
			)
			if _, err := t.bot.SendDocument(t.config.ChatID, file, nil); err != nil {
				return fmt.Errorf("send file error: %v", err)
			}
		} else {
			msg.WriteString(output)
			remaining = maxMessageLen - msg.Len()
		}
	}

	if result.Error != nil {
		str := result.Error.Error()
		errorStr := fmt.Sprintf(errorFormat, str)
		if len(errorStr) > remaining {
			file := gotgbot.InputFileByReader(
				fmt.Sprintf("%s.error.log", result.JobName),
				strings.NewReader(str),
			)
			if _, err := t.bot.SendDocument(t.config.ChatID, file, nil); err != nil {
				return fmt.Errorf("send file error: %v", err)
			}
		} else {
			msg.WriteString(errorStr)
		}
	}

	// Send notification regardless of job success/failure
	if nErr := t.Send(msg.String()); nErr != nil {
		if result.Error != nil {
			return fmt.Errorf("job error: %v; notification error: %v", result.Error, nErr)
		}
		return fmt.Errorf("notification error: %v", nErr)
	}

	return nil
}

const (
	notificationFormat = "%s %s\n⏱️: %s"
	outputFormat       = "\nOutput:\n```\n%s\n```"
	errorFormat        = "\nError:\n```\n%s\n```"
	success            = "✅"
	failed             = "❌"
	maxMessageLen      = 4000 // Telegram API limit is 4096.
)
