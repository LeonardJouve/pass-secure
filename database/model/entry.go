package model

import "gorm.io/gorm"

type Entry struct {
	gorm.Model
	Parent   *Folder
	Name     string
	Password string
}

type SanitizedEntry struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (entry *Entry) SanitizeEntry() *SanitizedEntry {
	return &SanitizedEntry{
		ID:       entry.ID,
		Name:     entry.Name,
		Password: entry.Password,
	}
}
