package model

import "github.com/LeonardJouve/pass-secure/database"

func Migrate() {
	database.Database.AutoMigrate(&Entry{}, &Folder{}, &User{})
}
