package parser

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"schedule-bot/models" // Убедитесь, что вы импортируете правильный пакет

	"github.com/PuerkitoBio/goquery"
)

const baseURL = "https://www.polessu.by/ruz/term2ng/"

// ParseGroups парсит группы из веб-страницы
func ParseGroups() ([]string, error) {
	resp, err := http.Get(baseURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var groups []string
	doc.Find("table.iksweb2 tr").Each(func(_ int, row *goquery.Selection) {
		row.Find("td:not(:first-child) a").Each(func(_ int, link *goquery.Selection) {
			group := strings.TrimSpace(link.Text())
			if group != "" {
				groups = append(groups, group)
			}
		})
	})

	return groups, nil
}

// ParseSchedules парсит расписания для всех групп
func ParseSchedules(groups []string) ([]models.Schedule, error) {
	resp, err := http.Get(baseURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var schedules []models.Schedule

	for _, group := range groups {
		schedule, err := parseScheduleForGroup(doc, group)
		if err != nil {
			log.Printf("Ошибка при парсинге расписания для группы %s: %v", group, err)
			continue
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

// parseScheduleForGroup парсит расписание для конкретной группы
func parseScheduleForGroup(doc *goquery.Document, group string) (models.Schedule, error) {
	var schedule models.Schedule
	schedule.Group = group

	// Найдите таблицу для группы
	table := findGroupTable(doc, group)
	if table == nil {
		return schedule, fmt.Errorf("таблица для группы %s не найдена", group)
	}

	// Парсинг расписания
	table.Find("thead tr:nth-child(2) th.xAxis").Each(func(i int, s *goquery.Selection) {
		date := strings.TrimSpace(s.Text())

		table.Find("tbody tr").Not(".foot").Each(func(j int, row *goquery.Selection) {
			timeCell := row.Find("th.yAxis")
			if timeCell.Length() == 0 {
				return
			}

			timeStr := strings.TrimSpace(timeCell.Text())
			timeStr = regexp.MustCompile(`\s+`).ReplaceAllString(timeStr, "")
			timeParts := strings.SplitN(timeStr, "-", 2)
			if len(timeParts) != 2 {
				return
			}
			time := fmt.Sprintf("%s-%s", timeParts[0], timeParts[1])

			row.Find("td").Each(func(k int, cell *goquery.Selection) {
				if cell.HasClass("empty") {
					return
				}

				subgroup := strings.TrimSpace(cell.Find(".studentsset").Text())
				subject := strings.TrimSpace(cell.Find(".subject").Text())
				teacher := strings.TrimSpace(cell.Find(".teacher").Text())
				room := strings.TrimSpace(cell.Find(".room").Text())

				// Добавление расписания
				if subject != "" {
					schedule.Date = date
					schedule.Time = time
					schedule.Subject = subject
					schedule.Teacher = teacher
					schedule.Room = room
					schedule.Subgroup = subgroup
				}
			})
		})
	})

	return schedule, nil
}

// findGroupTable находит таблицу для заданной группы
func findGroupTable(doc *goquery.Document, groupName string) *goquery.Selection {
	var table *goquery.Selection
	doc.Find("table").Each(func(i int, s *goquery.Selection) {
		header := s.Find("thead tr:first-child th:contains('" + groupName + "')")
		if header.Length() > 0 {
			table = s
			return
		}
	})
	return table
}
