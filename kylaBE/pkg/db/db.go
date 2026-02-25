package db

import (
	"fmt"
	"kyla-be/config"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB
var err error

func InitDB(configs *config.PostgresConfig) {
	host, port, dbName, dbUser, password := configs.PostgresHost, configs.PostgresPort, configs.PostgresDB, configs.PostgresUser, configs.PostgresPass
	// Open DB connection
	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=",
		host,
		port,
		dbUser,
		dbName,
		password,
	)
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("DB connection failed %v", err)
	}
	fmt.Print("Database connected successfully...")

}
