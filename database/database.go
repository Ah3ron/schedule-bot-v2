package database

import (
	"schedule-bot/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&models.Schedule{}); err != nil {
		return nil, err
	}

	return db, nil
}

func SaveSchedules(db *gorm.DB, schedules []models.Schedule) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM schedules").Error; err != nil {
			return err
		}

		if err := tx.Create(&schedules).Error; err != nil {
			return err
		}

		return nil
	})
}
