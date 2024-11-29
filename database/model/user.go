package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email    string `gorm:"unique"`
	Password string
	Folders  []Folder `gorm:"many2many:user_folders;constraint:OnDelete:CASCADE"`
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
