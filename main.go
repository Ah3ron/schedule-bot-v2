package main

import (
	"log"
	"schedule-bot/bot"
	"schedule-bot/config"
	"schedule-bot/database"
	"schedule-bot/parser"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Ошибка загрузки конфигурации:", err)
	}

	db, err := database.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	allSchedules, err := parser.ParseAllSchedules()
	if err != nil {
		log.Fatal("Ошибка парсинга расписаний:", err)
	}

	if err := database.SaveSchedules(db, allSchedules); err != nil {
		log.Fatal("Ошибка сохранения расписаний:", err)
	}

	go bot.StartBot(cfg.TelegramToken, db)

	select {}
}
