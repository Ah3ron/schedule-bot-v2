package main

import (
	"log"
	"time"

	"schedule-bot/bot"
	"schedule-bot/config"
	"schedule-bot/database"
	"schedule-bot/parser"

	"gorm.io/gorm"
)

func updateSchedules(db *gorm.DB) error {
	schedules, err := parser.ParseAllSchedules()
	if err != nil {
		log.Println("Ошибка парсинга расписаний:", err)
		return err
	}
	if err := database.SaveSchedules(db, schedules); err != nil {
		log.Println("Ошибка сохранения расписаний:", err)
		return err
	}
	return nil
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Ошибка загрузки конфигурации:", err)
	}

	db, err := database.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	if err := updateSchedules(db); err != nil {
		log.Fatal("Первичное обновление расписаний не выполнено:", err)
	}
	log.Println("Первичное расписание успешно обновлено.")

	go bot.StartBot(cfg.TelegramToken, db)

	ticker := time.NewTicker(30 * time.Minute)
	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println("Обновление расписаний...")
				if err := updateSchedules(db); err != nil {
					continue
				}
				log.Println("Расписание успешно обновлено.")
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	select {}
}
