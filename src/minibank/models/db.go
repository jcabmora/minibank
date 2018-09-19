package models

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var Database *sql.DB

func InitDB(dataSourceName string) {
	var err error
	Database, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Panic(err)
	}
	err = Database.Ping()
	if err != nil {
		log.Panic(err)
	}
}
