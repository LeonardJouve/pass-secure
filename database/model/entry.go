package model

import "gorm.io/gorm"

type Entry struct {
	gorm.Model
	Name     string
	Password string
}
