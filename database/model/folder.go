package model

import (
	"gorm.io/gorm"
)

type Folder struct {
	gorm.Model
	Name     string
	OwnerID  uint
	Owner    User `gorm:"foreignKey:OwnerID"`
	ParentID *uint
	Parent   *Folder `gorm:"foreignKey:ParentID"`
	Users    []User  `gorm:"many2many:user_folders"`
	Entries  []Entry `gorm:"foreignKey:FolderID;constraint:OnDelete:CASCADE"`
}

type SanitizedFolder struct {
	ID       uint   `json:"id"`
	UserIds  []uint `json:"userIds"`
	OwnerID  uint   `json:"ownerId"`
	Name     string `json:"name"`
	ParentID *uint  `json:"parentId"`
}

func (folder *Folder) Sanitize() *SanitizedFolder {
	userIds := []uint{}
	for _, user := range folder.Users {
		userIds = append(userIds, user.ID)
	}

	return &SanitizedFolder{
		ID:       folder.ID,
		UserIds:  userIds,
		OwnerID:  folder.Owner.ID,
		Name:     folder.Name,
		ParentID: folder.ParentID,
	}
}
