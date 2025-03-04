package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email        string `gorm:"unique"`
	Password     string
	OwnedFolders []Folder `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE"`
	Folders      []Folder `gorm:"many2many:user_folders"`
}

type SanitizedUser struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
}

func (user *User) Sanitize() *SanitizedUser {
	return &SanitizedUser{
		ID:    user.ID,
		Email: user.Email,
	}
}
