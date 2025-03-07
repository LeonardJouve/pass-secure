package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/LeonardJouve/pass-secure/api"
	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/database/model"
	"github.com/LeonardJouve/pass-secure/schema"
	"gorm.io/driver/mysql"
)

func main() {
	schema.Init()

	connectionURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_PASSWORD"), os.Getenv("MYSQL_HOST"), os.Getenv("MYSQL_PORT"), os.Getenv("MYSQL_DATABASE"))
	if err := database.Init(mysql.Open(connectionURL)); err != nil {
		panic("Could not initialize database")
	}

	model.Migrate()

	port, err := strconv.ParseUint(os.Getenv("PORT"), 10, 16)
	if err != nil {
		panic("Could not get port")
	}

	shutdown := api.Start(uint16(port))
	defer shutdown()
}
