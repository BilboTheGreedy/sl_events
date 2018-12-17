package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	host     = "db"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "postgres"
)

func insertDB(ticketid int, sum string, sub string, keyname string, created string) {

	sqlStatement := `
		INSERT INTO events (ticketid, sum, sub, keyname, created)
		VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(sqlStatement, ticketid, sum, sub, keyname, created)
	if err != nil {
		panic(err)
	}

}
