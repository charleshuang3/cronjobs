package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/charleshuang3/cronjobs/internal/bot"
	"github.com/charleshuang3/cronjobs/internal/config"
	"github.com/charleshuang3/cronjobs/internal/scheduler"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s /path/to/config.yaml\n", os.Args[0])
	}

	// Load configuration
	cfg, err := config.Load(os.Args[1])
	if err != nil {
		log.Fatalf("Failed to load configuration: %v\n", err)
	}

	// Create scheduler
	sched := scheduler.New(nil) // temporarily nil, will be set after bot creation

	// Create bot
	var botInstance bot.Bot
	switch cfg.BotType {
	case "telegram":
		botInstance, err = bot.NewTelegramBot(cfg.TelegramConfig, sched)
		if err != nil {
			log.Fatalf("Failed to create Telegram bot: %v\n", err)
		}
	default:
		log.Fatalf("Unsupported bot type: %s\n", cfg.BotType)
	}

	// Set bot in scheduler
	sched.SetBot(botInstance)

	// Add jobs to scheduler
	if err := sched.AddJobs(cfg.Jobs); err != nil {
		log.Fatalf("Failed to add jobs: %v\n", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start webhook if configured
	if err := botInstance.StartWebhook(ctx); err != nil {
		log.Fatalf("Failed to start webhook: %v\n", err)
	}

	// Start the scheduler
	sched.Start()

	log.Println("Cron job service started successfully")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	log.Println("Shutting down...")
	sched.Stop()
	if err := botInstance.StopWebhook(ctx); err != nil {
		log.Printf("Error stopping webhook: %v\n", err)
	}
}
