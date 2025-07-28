package db

import (
	"gorm.io/gorm"
)

type Customer struct {
	gorm.Model
	PhoneNumber    string `gorm:"unique"`
	HistoryRecords []ChatHistory
}
