package bot

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

	"schedule-bot/models"

	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
)

var allGroups []string

var groupRegex = regexp.MustCompile(`^(\d{2})([А-ЯЁа-яё]+)-.+$`)

func extractYearSpec(group string) (year, spec string) {
	matches := groupRegex.FindStringSubmatch(group)
	if len(matches) >= 3 {
		return matches[1], matches[2]
	}
	return "", ""
}

func getUniqueYears() []string {
	yearSet := make(map[string]struct{})
	for _, group := range allGroups {
		year, _ := extractYearSpec(group)
		if year != "" {
			yearSet[year] = struct{}{}
		}
	}
	var years []string
	for year := range yearSet {
		years = append(years, year)
	}
	sort.Strings(years)
	return years
}

func getSpecsForYear(year string) []string {
	specSet := make(map[string]struct{})
	for _, group := range allGroups {
		gYear, spec := extractYearSpec(group)
		if gYear == year && spec != "" {
			specSet[spec] = struct{}{}
		}
	}
	var specs []string
	for spec := range specSet {
		specs = append(specs, spec)
	}
	sort.Strings(specs)
	return specs
}

func getGroupsForYearAndSpec(year string, spec string) []string {
	var groups []string
	for _, group := range allGroups {
		gYear, gSpec := extractYearSpec(group)
		if gYear == year && gSpec == spec {
			groups = append(groups, group)
		}
	}
	sort.Strings(groups)
	return groups
}

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
	btnDay := mainMenu.Data("📆 Расписание", "day_schedule")
	btnSettings := mainMenu.Data("⚙️ Настройки", "settings_menu")
	mainMenu.Inline(
		mainMenu.Row(btnDay),
		mainMenu.Row(btnSettings),
	)

	navMenu := &telebot.ReplyMarkup{}
	btnPrevMonday := navMenu.Data("<<", "prev_monday")
	btnPrevDay := navMenu.Data("<", "prev_day")
	btnToday := navMenu.Data("•", "today")
	btnNextDay := navMenu.Data(">", "next_day")
	btnNextMonday := navMenu.Data(">>", "next_monday")
	btnMain := navMenu.Data("Меню", "main_menu")
	navMenu.Inline(
		navMenu.Row(btnPrevMonday, btnPrevDay, btnToday, btnNextDay, btnNextMonday),
		navMenu.Row(btnMain),
	)

	settingsMenu := &telebot.ReplyMarkup{}
	btnSetGroup := settingsMenu.Data("Выбрать группу", "set_group")
	settingsMenu.Inline(
		settingsMenu.Row(btnSetGroup),
	)

	b.Handle("/start", func(c telebot.Context) error {
		welcome := "Добро пожаловать!\nВыберите действие:"
		return c.Send(welcome, mainMenu)
	})

	b.Handle(&btnPrevMonday, func(c telebot.Context) error {
		dateStr := parseDateFromMessage(c.Message().Text)
		prevMonday := shiftToMonday(dateStr, -1)
		return showScheduleForDate(c, db, prevMonday, navMenu)
	})

	b.Handle(&btnPrevDay, func(c telebot.Context) error {
		dateStr := parseDateFromMessage(c.Message().Text)
		prevDate := shiftDate(dateStr, -1)
		return showScheduleForDate(c, db, prevDate, navMenu)
	})

	b.Handle(&btnDay, func(c telebot.Context) error {
		today := time.Now().Format("02.01")
		return showScheduleForDate(c, db, today, navMenu)
	})

	b.Handle(&btnToday, func(c telebot.Context) error {
		today := time.Now().Format("02.01")
		return showScheduleForDate(c, db, today, navMenu)
	})

	b.Handle(&btnNextDay, func(c telebot.Context) error {
		dateStr := parseDateFromMessage(c.Message().Text)
		nextDate := shiftDate(dateStr, +1)
		return showScheduleForDate(c, db, nextDate, navMenu)
	})

	b.Handle(&btnNextMonday, func(c telebot.Context) error {
		dateStr := parseDateFromMessage(c.Message().Text)
		nextMonday := shiftToMonday(dateStr, +1)
		return showScheduleForDate(c, db, nextMonday, navMenu)
	})

	b.Handle(&btnMain, func(c telebot.Context) error {
		return c.Edit("Добро пожаловать!\nВыберите действие:", mainMenu)
	})

	b.Handle(&btnSettings, func(c telebot.Context) error {
		text := "Настройки:\nЧтобы установить группу, нажмите 'Выбрать группу'."
		return c.Edit(text, settingsMenu)
	})

	b.Handle(&btnSetGroup, func(c telebot.Context) error {
		if err := db.Model(&models.Schedule{}).Select("group_name").Distinct().Order("group_name").Pluck("group_name", &allGroups).Error; err != nil {
			return err
		}
		return c.Edit("Выберите год поступления:", createYearMenu())
	})

	b.Handle(&telebot.Btn{Unique: "select_year"}, func(c telebot.Context) error {
		data := c.Data()
		parts := strings.SplitN(data, "_", 2)
		if len(parts) != 2 {
			return c.Respond(&telebot.CallbackResponse{Text: "Некорректные данные."})
		}
		year := parts[1]
		return c.Edit(fmt.Sprintf("Вы выбрали год %s. Теперь выберите специальность:", year),
			createSpecMenu(year))
	})

	b.Handle(&telebot.Btn{Unique: "select_spec"}, func(c telebot.Context) error {
		data := c.Data()
		parts := strings.SplitN(data, "_", 3)
		if len(parts) != 3 {
			return c.Respond(&telebot.CallbackResponse{Text: "Некорректные данные."})
		}
		year, spec := parts[1], parts[2]
		return c.Edit(fmt.Sprintf("Вы выбрали специальность %s для года %s. Выберите группу:", spec, year),
			createGroupMenu(year, spec))
	})

	b.Handle(&telebot.Btn{Unique: "select_group"}, func(c telebot.Context) error {
		data := c.Data()
		parts := strings.SplitN(data, "_", 2)
		if len(parts) != 2 {
			return c.Respond(&telebot.CallbackResponse{Text: "Некорректные данные."})
		}
		group := parts[1]
		senderID := c.Sender().ID

		var user models.User
		if err := db.First(&user, "id = ?", senderID).Error; err != nil {
			user = models.User{
				ID:        senderID,
				GroupName: group,
			}
			if err := db.Create(&user).Error; err != nil {
				log.Printf("Ошибка сохранения пользователя: %v", err)
				return c.Edit("Ошибка сохранения группы. Попробуйте позже.")
			}
		} else {
			user.GroupName = group
			if err := db.Save(&user).Error; err != nil {
				log.Printf("Ошибка обновления пользователя: %v", err)
				return c.Edit("Ошибка обновления группы. Попробуйте позже.")
			}
		}
		return c.Edit(fmt.Sprintf("Ваша группа установлена: %s", group), mainMenu)
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

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("📅 Расписание на %s для группы %s:\n\n", dateStr, user.GroupName))

	for _, sched := range schedules {
		if sched.Time != "" {
			builder.WriteString(fmt.Sprintf("*Время*: _%s_\n", sched.Time))
		}
		if sched.Subject != "" {
			builder.WriteString(fmt.Sprintf("*Пара*: _%s_\n", sched.Subject))
		}
		if sched.Teacher != "" {
			builder.WriteString(fmt.Sprintf("*Препод.*: _%s_\n", sched.Teacher))
		}
		if sched.Room != "" {
			builder.WriteString(fmt.Sprintf("*Аудит.*: _%s_\n", sched.Room))
		}
		if sched.Subgroup != "" {
			builder.WriteString(fmt.Sprintf("*Подгруппа*: _%s_\n", sched.Subgroup))
		}
		builder.WriteString("\n")
	}

	return c.Edit(builder.String(), navMenu, telebot.ModeMarkdown)
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

func shiftToMonday(dateStr string, direction int) string {
	date, err := parseDate(dateStr)
	if err != nil {
		date = time.Now()
	}

	offset := (int(date.Weekday()) + 6) % 7
	currentMonday := date.AddDate(0, 0, -offset)

	shifted := currentMonday.AddDate(0, 0, direction*7)
	return shifted.Format("02.01")
}

func parseDate(dateStr string) (time.Time, error) {
	currentYear := time.Now().Year()
	fullDateStr := fmt.Sprintf("%s.%d", dateStr, currentYear)
	parsed, err := time.Parse("02.01.2006", fullDateStr)
	if err != nil {
		return time.Time{}, err
	}
	now := time.Now()
	if parsed.Sub(now) > (6 * 30 * 24 * time.Hour) {
		parsed = parsed.AddDate(-1, 0, 0)
	} else if now.Sub(parsed) > (6 * 30 * 24 * time.Hour) {
		parsed = parsed.AddDate(1, 0, 0)
	}
	return parsed, nil
}

func shiftDate(dateStr string, delta int) string {
	date, err := parseDate(dateStr)
	if err != nil {
		date = time.Now()
	}
	shifted := date.AddDate(0, 0, delta)
	return shifted.Format("02.01")
}

func createYearMenu() *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{}
	var rows []telebot.Row
	years := getUniqueYears()
	for _, year := range years {
		btn := menu.Data(year, "select_year", fmt.Sprintf("year_%s", year))
		rows = append(rows, menu.Row(btn))
	}
	menu.Inline(rows...)
	return menu
}

func createSpecMenu(year string) *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{}
	var rows []telebot.Row
	specs := getSpecsForYear(year)
	for _, spec := range specs {
		btn := menu.Data(spec, "select_spec", fmt.Sprintf("spec_%s_%s", year, spec))
		rows = append(rows, menu.Row(btn))
	}
	menu.Inline(rows...)
	return menu
}

func createGroupMenu(year, spec string) *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{}
	var rows []telebot.Row
	groups := getGroupsForYearAndSpec(year, spec)
	for _, group := range groups {
		btn := menu.Data(group, "select_group", fmt.Sprintf("group_%s", group))
		rows = append(rows, menu.Row(btn))
	}
	menu.Inline(rows...)
	return menu
}
