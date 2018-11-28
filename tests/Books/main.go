package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"

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

	db, err = sql.Open("sqlite3", "book.db")
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

	file, err := os.Open("top100books.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	tx, _ := db.Begin()
	statement, err := tx.Prepare("INSERT INTO Book (Title, Author, Country, Date) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Critical(err)
	}
	defer statement.Close()
	for scanner.Scan() {
		row := strings.Split(scanner.Text(), ";")
		fmt.Println(row)
		statement.Exec(row[0], row[1], row[2], row[3])
	}
	tx.Commit()

}
