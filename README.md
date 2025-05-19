# Go Cron Job Service

A flexible cron job scheduler with Telegram bot integration for notifications and control.

## Features

- Schedule and run cron jobs with customizable schedules
- Telegram bot integration for notifications and control
- Support for command execution and monitoring
- Optional webhook support for enhanced bot interaction
- Real-time job execution notifications
- Command-based job management

## Bot Commands

The following commands are available through the Telegram bot:

- `/list` - List all scheduled jobs with their names and schedules
- `/run <name>` - Run a specific job immediately by its name
- `/status` - Show currently running jobs and their duration

## Configuration

Create a YAML configuration file with the following structure:

```yaml
bot: telegram
bot.telegram:
  bot_token: your_bot_token_here
  chat_id: your_chat_id_here
  # Optional webhook URL (must be HTTPS), used to tell telegram
  webhook: "https://example.com/tgbot"
  # Required if webhook is set, used to setup listening
  webhook_path: "tgbot"
  # Required if webhook is set, used to setup listening
  webhook_port: 8080

jobs:
  # Run backup at 5 AM daily
  - name: daily-backup
    cron: "0 0 5 * * *"
    program: rsync
    args:
      - "-av"
      - "/path/from"
      - "/path/to"

  # Run system update check every 6 hours
  - name: system-update
    cron: "0 0 */6 * * *"
    program: apt-get
    args:
      - "update"
```

### Configuration Fields

#### Bot Configuration
- `bot`: The type of bot to use (currently only 'telegram' is supported)
- `bot.telegram`: Telegram-specific configuration
  - `bot_token`: Your Telegram bot token from BotFather
  - `chat_id`: Your Telegram chat id
  - `webhook`: (Optional) HTTPS URL where Telegram will send updates

#### Job Configuration
- `jobs`: List of jobs to schedule
  - `name`: Name of the job, must be unique, format `[0-9a-zA-Z_-]+`
  - `cron`: Cron expression in format "second minute hour day month weekday"
  - `program`: The program to execute
  - `args`: List of arguments to pass to the program

## Running the Service

1. Create a config file based on the example above
2. Run the service:
```bash
./cronjobs /path/to/config.yaml
```

## Building from Source

```bash
go build -o cronjobs cmd/cronjobs/main.go
```

## Docker Support

Build the Docker image:
```bash
docker build -t go-cron-job .
```

Run with Docker:
```bash
docker run -v /path/to/config.yaml:/etc/cronjobs/config.yaml go-cron-job
```

## Webhook Setup

To enable webhook mode:

1. Ensure you have a domain with SSL certificate (required by Telegram)
2. Add the webhook URL to your config file:
   ```yaml
   bot.telegram:
     webhook: "https://your.domain.com/path"
   ```
3. Make sure your domain is accessible
4. The service will automatically set up the webhook with Telegram

The webhook server will listen on port 8443 (required by Telegram).

## Security Notes

- Only the configured Telegram chat id can interact with the bot
- Webhook mode requires HTTPS (SSL/TLS certificate)
- Keep your bot token secret and secure
- Be careful with job permissions, as they run with the service's privileges
