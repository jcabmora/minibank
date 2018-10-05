package main

import (
	"minibank/handlers"
	"minibank/models"
	"net/http"
	"os"
)

func main() {
	// Connect to the database

	dbDoneCh := make(chan bool)
	dbDone := false

	go models.InitDB(dbConn(), dbDoneCh)
	defer models.Database.Close()

	http.HandleFunc("/api/account/register", validateDBConn(handlers.RegisterHandler, &dbDone))
	http.HandleFunc("/api/account/login", validateDBConn(handlers.LoginHandler, &dbDone))
	http.HandleFunc("/api/account/token", validateDBConn(handlers.TokenHandler, &dbDone))
	http.HandleFunc("/api/account/sessions", handlers.AuthValidationMiddleware(handlers.SessionListHandler))

	go updateDBDone(&dbDone, dbDoneCh)
	http.ListenAndServe(port(), nil)
}

func validateDBConn(next http.HandlerFunc, dbDone *bool) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if *dbDone {
			next(w, r)
		} else {
			handlers.ServerUnavailableHandler(w, r)
		}
	})
}

func updateDBDone(dbdone *bool, dbDoneCh <-chan bool) {
	*dbdone = <-dbDoneCh
}

// port looks up service listening port
func port() string {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}
	return ":" + port
}

// dbConn looks up database connection string on environment
func dbConn() string {
	connectString := os.Getenv("DB_CONNECTION_STRING")
	if len(connectString) == 0 {
		connectString = "minibank:minibank@tcp(mysql)/minibank"
	}
	return connectString
}
