package models

import (
	"database/sql"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocql/gocql"
)

var Database *sql.DB

var CassandraSession *gocql.Session
var CassandraEnabled bool

func init() {
	CassandraEnabled = cassandraEnabled()
}

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

func InitCassandra(clusterAddress string) error {
	cluster := gocql.NewCluster(clusterAddress)
	pass := gocql.PasswordAuthenticator{"minibank", "minibank"}
	cluster.Keyspace = "minibank"
	cluster.Authenticator = pass
	cluster.Consistency = gocql.One
	sess, err := cluster.CreateSession()
	CassandraSession = sess
	return err
}

func cassandraEnabled() bool {
	enableCassandra := os.Getenv("ENABLE_CASSANDRA")
	if len(enableCassandra) == 0 {
		if strings.ToLower(enableCassandra) == "true" {
			return true
		}
	}
	return false
}
