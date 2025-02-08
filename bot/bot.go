package bot

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"schedule-bot/models"

	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
)

var pendingGroupSelection = make(map[int64]bool)

func StartBot(token string, db *gorm.DB) {
	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	mainMenu := &telebot.ReplyMarkup{}
	btnToday := mainMenu.Data("Расписание на сегодня", "today_schedule")
	btnSettings := mainMenu.Data("Настройки", "settings_menu")
	mainMenu.Inline(
		mainMenu.Row(btnToday, btnSettings),
	)

	navMenu := &telebot.ReplyMarkup{}
	btnPrev := navMenu.Data("◀ Назад", "prev_day")
	btnNext := navMenu.Data("Вперёд ▶", "next_day")
	btnMain := navMenu.Data("Меню", "main_menu")
	navMenu.Inline(
		navMenu.Row(btnPrev, btnMain, btnNext),
	)

	settingsMenu := &telebot.ReplyMarkup{}
	btnSetGroup := settingsMenu.Data("Выбрать группу", "set_group")
	settingsMenu.Inline(
		settingsMenu.Row(btnSetGroup),
	)

	b.Handle("/start", func(c telebot.Context) error {
		pendingGroupSelection[c.Sender().ID] = false
		welcome := "Добро пожаловать!\nВыберите действие:"
		return c.Send(welcome, mainMenu)
	})

	b.Handle(&btnToday, func(c telebot.Context) error {
		today := time.Now()
		dateStr := fmt.Sprintf("%02d.%02d", today.Day(), int(today.Month()))
		return showScheduleForDate(c, db, dateStr, navMenu)
	})

	b.Handle(&btnPrev, func(c telebot.Context) error {
		dateStr := parseDateFromMessage(c.Message().Text)
		prevDate := shiftDate(dateStr, -1)
		return showScheduleForDate(c, db, prevDate, navMenu)
	})

	b.Handle(&btnNext, func(c telebot.Context) error {
		dateStr := parseDateFromMessage(c.Message().Text)
		nextDate := shiftDate(dateStr, +1)
		return showScheduleForDate(c, db, nextDate, navMenu)
	})

	b.Handle(&btnMain, func(c telebot.Context) error {
		return c.Edit("Меню:", mainMenu)
	})

	b.Handle(&btnSettings, func(c telebot.Context) error {
		text := "Настройки:\nЧтобы установить группу, нажмите 'Выбрать группу'."
		return c.Edit(text, settingsMenu)
	})

	b.Handle(&btnSetGroup, func(c telebot.Context) error {
		pendingGroupSelection[c.Sender().ID] = true
		msg := "Пожалуйста, введите название группы (например: 22ИТ-1):"
		return c.Edit(msg)
	})

	b.Handle(telebot.OnText, func(c telebot.Context) error {
		senderID := c.Sender().ID
		if pendingGroupSelection[senderID] {
			group := c.Text()
			if group == "" {
				return c.Send("Название группы не может быть пустым.")
			}
			user := models.User{ID: senderID}
			if err := db.First(&user).Error; err != nil {
				user.GroupName = group
				if err := db.Create(&user).Error; err != nil {
					log.Printf("Ошибка сохранения пользователя: %v", err)
					return c.Send("Ошибка сохранения группы. Попробуйте позже.")
				}
			} else {
				user.GroupName = group
				if err := db.Save(&user).Error; err != nil {
					log.Printf("Ошибка обновления пользователя: %v", err)
					return c.Send("Ошибка обновления группы. Попробуйте позже.")
				}
			}
			pendingGroupSelection[senderID] = false
			return c.Send(fmt.Sprintf("Ваша группа установлена: %s", group))
		}
		return nil
	})

	b.Start()
}

func showScheduleForDate(c telebot.Context, db *gorm.DB, dateStr string, navMenu *telebot.ReplyMarkup) error {
	senderID := c.Sender().ID
	var user models.User
	if err := db.First(&user, "id = ?", senderID).Error; err != nil || user.GroupName == "" {
		return c.Send("Сначала установите группу через настройки (/start -> Настройки).")
	}

	var schedules []models.Schedule
	if err := db.Where(&models.Schedule{GroupName: user.GroupName, Date: dateStr}).Find(&schedules).Error; err != nil {
		return c.Edit("Не удалось получить расписание. Попробуйте позже.")
	}

	if len(schedules) == 0 {
		return c.Edit(fmt.Sprintf("Расписание на %s для группы %s не найдено.", dateStr, user.GroupName), navMenu)
	}

	response := fmt.Sprintf("📅 Расписание на %s для группы %s:\n\n", dateStr, user.GroupName)
	for _, sched := range schedules {
		response += fmt.Sprintf("⏰ %s\n📚 %s\n👨‍🏫 %s\n🏫 %s\n🔢 %s\n\n",
			sched.Time, sched.Subject, sched.Teacher, sched.Room, sched.Subgroup)
	}

	return c.Edit(response, navMenu)
}

func parseDateFromMessage(text string) string {
	re := regexp.MustCompile(`Расписание на\s+(\d{2}\.\d{2})`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}
	now := time.Now()
	return fmt.Sprintf("%02d.%02d", now.Day(), int(now.Month()))
}

func shiftDate(dateStr string, delta int) string {
	var day, month int
	_, err := fmt.Sscanf(dateStr, "%02d.%02d", &day, &month)
	if err != nil {
		now := time.Now()
		day = now.Day()
		month = int(now.Month())
	}
	currentYear := time.Now().Year()
	orig := time.Date(currentYear, time.Month(month), day, 0, 0, 0, 0, time.Local)
	shifted := orig.AddDate(0, 0, delta)
	return fmt.Sprintf("%02d.%02d", shifted.Day(), int(shifted.Month()))
}
