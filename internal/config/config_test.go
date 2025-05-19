package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		content   string
		wantErr   bool
		checkFunc func(*testing.T, *Config)
	}{
		{
			name: "valid config with webhook",
			content: `bot: telegram
bot.telegram:
  bot_token: "test_token"
  chat_id: 123456
  webhook: "https://example.com/tgbot"
  webhook_path: "tgbot"
  webhook_port: 8080

jobs:
  - name: "daily-backup"
    cron: "0 0 5 * * *"
    program: "ls"
    args:
      - "/path/"

  - name: "system-update"
    cron: "0 0 */6 * * *"
    program: "echo"
    args:
      - "update"`,
			wantErr: false,
			checkFunc: func(t *testing.T, cfg *Config) {
				if cfg.BotType != "telegram" {
					t.Errorf("expected bot type telegram, got %s", cfg.BotType)
				}
				if cfg.TelegramConfig.BotToken != "test_token" {
					t.Errorf("expected bot token test_token, got %s", cfg.TelegramConfig.BotToken)
				}
				if cfg.TelegramConfig.ChatID != 123456 {
					t.Errorf("expected chat ID 123456, got %d", cfg.TelegramConfig.ChatID)
				}
				if cfg.TelegramConfig.Webhook != "https://example.com/tgbot" {
					t.Errorf("expected webhook https://example.com/tgbot, got %s", cfg.TelegramConfig.Webhook)
				}
				if len(cfg.Jobs) != 2 {
					t.Errorf("expected 2 jobs, got %d", len(cfg.Jobs))
				}
				// Check first job
				if cfg.Jobs[0].Name != "daily-backup" {
					t.Errorf("expected job name daily-backup, got %s", cfg.Jobs[0].Name)
				}
				if cfg.Jobs[0].Cron != "0 0 5 * * *" {
					t.Errorf("expected cron 0 0 5 * * *, got %s", cfg.Jobs[0].Cron)
				}
				if cfg.Jobs[0].Program != "ls" {
					t.Errorf("expected program rsync, got %s", cfg.Jobs[0].Program)
				}
				if len(cfg.Jobs[0].Args) != 1 {
					t.Errorf("expected 3 args for first job, got %d", len(cfg.Jobs[0].Args))
				}
			},
		},
		{
			name: "valid config without webhook",
			content: `bot: telegram
bot.telegram:
  bot_token: "test_token"
  chat_id: 123456

jobs:
  - name: "backup"
    cron: "0 0 5 * * *"
    program: "ls"
    args:
      - "/path/to"`,
			wantErr: false,
			checkFunc: func(t *testing.T, cfg *Config) {
				if cfg.TelegramConfig.Webhook != "" {
					t.Error("expected empty webhook")
				}
				if len(cfg.Jobs) != 1 {
					t.Errorf("expected 1 job, got %d", len(cfg.Jobs))
				}
			},
		},
		{
			name: "invalid webhook URL",
			content: `bot: telegram
bot.telegram:
  bot_token: "test_token"
  chat_id: 123456
  webhook: "not-a-url"

jobs:
  - name: "test"
    cron: "0 0 5 * * *"
    program: "rsync"
    args:
      - "-av"`,
			wantErr: true,
		},
		{
			name: "missing job name",
			content: `bot: telegram
bot.telegram:
  bot_token: "test_token"
  chat_id: 123456

jobs:
  - cron: "0 0 5 * * *"
    program: "rsync"
    args:
      - "-av"`,
			wantErr: true,
		},
		{
			name: "invalid job name",
			content: `bot: telegram
bot.telegram:
  bot_token: "test_token"
  chat_id: 123456

jobs:
  - name: "invalid name!"
    cron: "0 0 5 * * *"
    program: "rsync"
    args:
      - "-av"`,
			wantErr: true,
		},
		{
			name: "duplicate job names",
			content: `bot: telegram
bot.telegram:
  bot_token: "test_token"
  chat_id: 123456

jobs:
  - name: "backup"
    cron: "0 0 5 * * *"
    program: "rsync"
    args:
      - "-av"
  - name: "backup"
    cron: "0 0 6 * * *"
    program: "rsync"
    args:
      - "-av"`,
			wantErr: true,
		},
		{
			name: "non-HTTPS webhook URL",
			content: `bot: telegram
bot.telegram:
  bot_token: "test_token"
  chat_id: 123456
  webhook: "http://example.com/webhook"

jobs:
  - name: "test"
    cron: "0 0 5 * * *"
    program: "rsync"
    args:
      - "-av"`,
			wantErr: true,
		},
		{
			name: "no webhook port",
			content: `bot: telegram
bot.telegram:
  bot_token: "test_token"
  chat_id: 123456
  webhook: "https://example.com/webhook"
  webhook_path: "webhook"

jobs:
  - name: "test"
    cron: "0 0 5 * * *"
    program: "rsync"
    args:
      - "-av"`,
			wantErr: true,
		},
		{
			name: "no webhook path",
			content: `bot: telegram
bot.telegram:
  bot_token: "test_token"
  chat_id: 123456
  webhook: "https://example.com/webhook"
  webhook_port: 8080

jobs:
  - name: "test"
    cron: "0 0 5 * * *"
    program: "rsync"
    args:
      - "-av"`,
			wantErr: true,
		},
		{
			name: "missing bot token",
			content: `bot: telegram
bot.telegram:
  chat_id: 123456

jobs:
  - name: "test"
    cron: "0 0 5 * * *"
    program: "rsync"
    args:
      - "-av"`,
			wantErr: true,
		},
		{
			name: "missing user ID",
			content: `bot: telegram
bot.telegram:
  bot_token: "test_token"

jobs:
  - name: "test"
    cron: "0 0 5 * * *"
    program: "rsync"
    args:
      - "-av"`,
			wantErr: true,
		},
		{
			name: "invalid bot type",
			content: `bot: unknown
bot.telegram:
  bot_token: "test_token"
  chat_id: 123456

jobs:
  - name: "test"
    cron: "0 0 5 * * *"
    program: "rsync"
    args:
      - "-av"`,
			wantErr: true,
		},
		{
			name: "missing cron expression",
			content: `bot: telegram
bot.telegram:
  bot_token: "test_token"
  chat_id: 123456

jobs:
  - name: "test"
    program: "rsync"
    args:
      - "-av"`,
			wantErr: true,
		},
		{
			name: "missing program",
			content: `bot: telegram
bot.telegram:
  bot_token: "test_token"
  chat_id: 123456

jobs:
  - name: "test"
    cron: "0 0 5 * * *"
    args:
      - "-av"`,
			wantErr: true,
		},
		{
			name: "no jobs configured",
			content: `bot: telegram
bot.telegram:
  bot_token: "test_token"
  chat_id: 123456

jobs: []`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			configPath := filepath.Join(tmpDir, "config.yaml")
			err := os.WriteFile(configPath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("failed to write test config: %v", err)
			}

			// Load and validate config
			cfg, err := Load(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && tt.checkFunc != nil {
				tt.checkFunc(t, cfg)
			}
		})
	}
}
