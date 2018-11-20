package main

import (
	"database/sql"
	"fmt"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	logging "github.com/op/go-logging"
)

var db *sql.DB

//Format
var log = logging.MustGetLogger("example")
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

func initDB() {

	log.Info(" -  Opening the database...")
	var err error

	db, err = sql.Open("sqlite3", "largeDB.db")
	if err != nil {
		fmt.Println(err)
	}
	db.SetMaxOpenConns(1)
	db.Exec("PRAGMA journal_mode=MEMORY;")
	db.Exec("PRAGMA _synchronous=OFF;")

	statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS Book (id INTEGER PRIMARY KEY, Title TEXT, Author TEXT, Country TEXT, Date TEXT)")
	if err != nil {
		fmt.Println(err)
	}
	statement.Exec()
}
func main() {

	initDB()

	tx, _ := db.Begin()
	statement, err := tx.Prepare("INSERT INTO Book (Title, Author, Country, Date) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Critical(err)
	}
	//Add 1.000.000 entries to the DB
	for i := 0; i < 1000000; i++ {
		data := strconv.Itoa(i)
		statement.Exec(data, data, data, data)
	}
	tx.Commit()

}
