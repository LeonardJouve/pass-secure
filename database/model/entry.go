package model

import "gorm.io/gorm"

type Entry struct {
	gorm.Model
	Name     string
	Password string
	ParentID uint
	Parent   Folder `gorm:"many2many:folder_entries;foreignKey:ParentID"`
}

type SanitizedEntry struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
	ParentID uint   `json:"parentId"`
}

func (entry *Entry) Sanitize() *SanitizedEntry {
	return &SanitizedEntry{
		ID:       entry.ID,
		Name:     entry.Name,
		Password: entry.Password,
		ParentID: entry.ParentID,
	}
}
