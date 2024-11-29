package model

import "github.com/LeonardJouve/pass-secure/database"

func Migrate() {
	database.Database.AutoMigrate(&Entry{})
	database.Database.AutoMigrate(&Folder{})
	database.Database.AutoMigrate(&User{})
}
