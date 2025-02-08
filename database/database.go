package database

import (
	"schedule-bot/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

func InitDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
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

		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&schedules).Error; err != nil {
			return err
		}

		return nil
	})
}
