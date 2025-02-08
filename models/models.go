package models

import "gorm.io/gorm"

type Schedule struct {
	gorm.Model
	GroupName string // Название группы
	Date      string // Дата занятия
	Time      string // Время занятия
	Subject   string // Название предмета
	Teacher   string // Имя преподавателя
	Room      string // Номер аудитории
	Subgroup  string // Подгруппа (если есть)
}
