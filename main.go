package main

import (
	"log"
	"schedule-bot/bot"
	"schedule-bot/config"
	"schedule-bot/database"
	"schedule-bot/parser"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize database
	db, err := database.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	// Start the Telegram bot
	bot.StartBot(cfg.TelegramToken, db)

	// Parse groups and schedules
	groups, err := parser.ParseGroups()
	if err != nil {
		log.Fatal(err)
	}

	schedules, err := parser.ParseSchedules(groups)
	if err != nil {
		log.Fatal(err)
	}

	// Save schedules to the database
	if err := database.SaveSchedules(db, schedules); err != nil {
		log.Fatal(err)
	}
}
