package models

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var Database *sql.DB

func InitDB(dataSourceName string, dbDone_chan chan<- bool) {
	var err error
	Database, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Panic(err)
	}
	for true {
		err = Database.Ping()
		if err != nil {
			log.Print(err)
			time.Sleep(1 * time.Second)
		} else {
			log.Print("Successfully connected to Database!")
			dbDone_chan <- true
			break
		}
	}
}
