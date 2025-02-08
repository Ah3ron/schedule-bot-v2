package parser

import (
	"log"
	"net/http"
	"regexp"
	"strings"

	"schedule-bot/models"

	"github.com/PuerkitoBio/goquery"
)

const baseURL = "https://www.polessu.by/ruz/term2ng/students.html"

func ParseAllSchedules() ([]models.Schedule, error) {
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

	doc.Find("table.odd_table, table.even_table").Each(func(_ int, table *goquery.Selection) {
		group := ""
		table.Find("thead tr").First().Find("th[colspan]").EachWithBreak(func(_ int, th *goquery.Selection) bool {
			group = strings.TrimSpace(th.Text())
			return group == ""
		})
		if group == "" {
			caption := table.Find("caption").Text()
			log.Printf("Группа не найдена в таблице, пропускаем таблицу (caption: %s)", caption)
			return
		}

		var dates []string
		table.Find("thead tr").Eq(1).Find("th.xAxis").Each(func(_ int, th *goquery.Selection) {
			date := strings.TrimSpace(th.Text()[4:])
			if date != "" {
				dates = append(dates, date)
			}
		})
		if len(dates) == 0 {
			log.Printf("Не найдены даты для группы %s", group)
			return
		}

		table.Find("tbody tr").Not(".foot").Each(func(_ int, row *goquery.Selection) {
			timeCell := row.Find("th.yAxis")
			if timeCell.Length() == 0 {
				return
			}
			timeStr := strings.TrimSpace(timeCell.Text())
			timeStr = regexp.MustCompile(`\s+`).ReplaceAllString(timeStr, " ")

			cells := row.ChildrenFiltered("td")
			cells.Each(func(i int, cell *goquery.Selection) {
				if cell.HasClass("empty") {
					return
				}

				subgroup := strings.TrimSpace(cell.Find(".studentsset").First().Text())
				subject := strings.TrimSpace(cell.Find(".subject").First().Text())
				teacher := strings.TrimSpace(cell.Find(".teacher").First().Text())
				room := strings.TrimSpace(cell.Find(".room").First().Text())

				if subject == "" {
					return
				}

				if i >= len(dates) {
					log.Printf("Количество ячеек (%d) превышает число дат (%d) для группы %s", i+1, len(dates), group)
					return
				}

				schedule := models.Schedule{
					GroupName: group,
					Date:      dates[i],
					Time:      timeStr,
					Subject:   subject,
					Teacher:   teacher,
					Room:      room,
					Subgroup:  subgroup,
				}
				schedules = append(schedules, schedule)
			})
		})
	})
	return schedules, nil
}
