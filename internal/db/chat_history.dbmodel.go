package db

import "gorm.io/gorm"

type ChatHistory struct {
	gorm.Model
	CustomerId uint
}
