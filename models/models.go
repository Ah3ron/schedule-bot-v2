package models

import "gorm.io/gorm"

type Schedule struct {
	gorm.Model
	Group    string `gorm:"uniqueIndex"` // Название группы
	Date     string // Дата занятия
	Time     string // Время занятия
	Subject  string // Название предмета
	Teacher  string // Имя преподавателя
	Room     string // Номер аудитории
	Subgroup string // Подгруппа (если есть)
}
