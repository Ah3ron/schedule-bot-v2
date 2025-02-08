package models

type Schedule struct {
	ID        uint
	GroupName string
	Date      string
	Time      string
	Subject   string
	Teacher   string
	Room      string
	Subgroup  string
}

type User struct {
	ID        int64 `gorm:"primaryKey"`
	GroupName string
}
