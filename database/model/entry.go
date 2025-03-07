package model

import "gorm.io/gorm"

type Entry struct {
	gorm.Model
	Name     string
	Password string
	FolderID uint
	Folder   Folder `gorm:"foreignKey:FolderID"`
}

type SanitizedEntry struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
	FolderID uint   `json:"folderId"`
}

func (entry *Entry) Sanitize() *SanitizedEntry {
	return &SanitizedEntry{
		ID:       entry.ID,
		Name:     entry.Name,
		Password: entry.Password,
		FolderID: entry.FolderID,
	}
}
