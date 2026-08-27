// Command mysqltest is a lightweight MySQL client used by local E2E setup.
package main

import (
	"database/sql"
	"flag"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	query := flag.String("e", "", "SQL statements to execute")
	flag.Parse()
	dsn := os.Getenv("DMS_DATABASE_DSN")
	if dsn == "" {
		log.Fatal("DMS_DATABASE_DSN is required")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(*query); err != nil {
		log.Fatal(err)
	}
}
