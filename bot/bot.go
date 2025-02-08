package bot

import (
	"fmt"
	"log"

	"schedule-bot/models"

	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
)

func StartBot(token string, db *gorm.DB) {
	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	// Handler for the /start command.
	b.Handle("/start", func(c telebot.Context) error {
		return c.Send("Добро пожаловать! Используйте команду /schedule <группа> для получения расписания.")
	})

	// Handler for the /schedule command with safety checks for arguments.
	b.Handle("/schedule", func(c telebot.Context) error {
		args := c.Args()
		// Check if any argument is provided.
		if len(args) == 0 {
			return c.Send("Пожалуйста, укажите группу. Пример: /schedule группа1")
		}

		group := args[0]
		var schedules []models.Schedule

		if err := db.Where(&models.Schedule{Group: group}).Find(&schedules).Error; err != nil {
			return c.Send("Группа не найдена или произошла ошибка при получении расписания.")
		}

		if len(schedules) == 0 {
			return c.Send("Расписание для группы не найдено.")
		}

		response := fmt.Sprintf("📅 Расписание для группы %s:\n", group)
		for _, schedule := range schedules {
			response += fmt.Sprintf("📆 Дата: %s\n⏰ Время: %s\n📚 Предмет: %s\n👨‍🏫 Преподаватель: %s\n🏫 Аудитория: %s\n🔢 Подгруппа: %s\n\n",
				schedule.Date, schedule.Time, schedule.Subject, schedule.Teacher, schedule.Room, schedule.Subgroup)
		}
		return c.Send(response)
	})

	b.Start()
}
