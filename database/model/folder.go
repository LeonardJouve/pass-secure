package model

import (
	"github.com/LeonardJouve/pass-secure/database"
	"gorm.io/gorm"
)

type Folder struct {
	gorm.Model
	Parent *Folder
	Users  []User
	Owner  *User
	Name   string
}

type SanitizedFolder struct {
	ID      uint   `json:"id"`
	UserIds []uint `json:"userIds"`
	OwnerID uint   `json:"ownerId"`
	Name    string `json:"name"`
}

func (folder *Folder) SanitizeFolder() *SanitizedFolder {
	database.Database.Model(&Folder{}).Preload("Users").Find(&folder)

	userIds := []uint{}
	for _, user := range folder.Users {
		userIds = append(userIds, user.ID)
	}

	return &SanitizedFolder{
		ID:      folder.ID,
		UserIds: userIds,
		OwnerID: folder.Owner.ID,
		Name:    folder.Name,
	}
}
