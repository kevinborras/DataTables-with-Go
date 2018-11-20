package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"github.com/op/go-logging"
)

//PaginateDataStruct fields for the AJAX response to paginate
type PaginateDataStruct struct {
	Draw            string `json:"draw"`
	RecordsTotal    int    `json:"recordsTotal"`
	RecordsFiltered int    `json:"recordsFiltered"`
	BookList        []Book `json:"data"`
}

// Book Struct contains useful information
type Book struct {
	Title   string `json:"Title"`
	Country string `json:"Country"`
	Date    string `json:"Date"`
	Author  string `json:"Author"`
	ID      string `json:"identifier"`
}

//Database path
var Database = `data/largeDB.db`
var db *sql.DB

//Here we store the recordsTotal and recordsFiltered value
var final int

//Format
var log = logging.MustGetLogger("example")
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

//search function returns the result of the query
func search(query, start, end string) (dataList []Book) {
	var book Book
	rows, err := db.Query(query, start, end)
	if err != nil {
		log.Critical("QueryRow: %v\n", err)
	}
	columns, err := rows.Columns()
	if err != nil {
		log.Critical(err.Error())
	}
	values := make([]sql.RawBytes, len(columns))
	// rows.Scan wants '[]interface{}' as an argument, so we must copy the
	// references into such a slice
	// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			log.Critical(err.Error())
		}
		var value string

		for i, col := range values {
			// Here we can check if the value is nil (NULL value)
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			if columns[i] == "id" {
				book.ID = value
			}
			if columns[i] == "Author" {
				book.Author = value
			}
			if columns[i] == "Date" {
				book.Date = value
			}
			if columns[i] == "Country" {
				book.Country = value
			}
			if columns[i] == "Title" {
				book.Title = value
			}
		}

		dataList = append(dataList, book)
	}

	return dataList
}

//paging function is used to implement
func paging(w http.ResponseWriter, r *http.Request) {
	log.Info(" -  Method:", r.Method, " -  /populateDataTable")
	var paging PaginateDataStruct
	var count string

	if r.Method == "POST" {
		//look if we are in the initial setup iteration
		r.ParseForm()
		count = `SELECT count(*) as frequency FROM Book`
		start := r.FormValue("start")
		end := r.FormValue("length")
		draw := r.FormValue("draw")
		searchValue := r.FormValue("search[value]")
		log.Info("Start: ", start+" Length: "+end+" Draw: "+draw)
		log.Info("Search Value: " + searchValue)

		if draw == "1" {
			rows, err := db.Query(count)
			if err != nil {
				fmt.Printf("QueryRow: %v\n", err)
			}
			defer rows.Close()
			for rows.Next() {
				err = rows.Scan(&final)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
		log.Infof("Records Filtered: %d", final)
		query := `SELECT id,Title,Country,Date,Author FROM Book
					ORDER BY Title
					Limit ? , ?;`
		result := search(query, start, end)

		paging.BookList = result
		paging.Draw = draw
		paging.RecordsFiltered = final
		paging.RecordsFiltered = final
		e, err := json.Marshal(paging)
		if err != nil {
			fmt.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(e)

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func mainpage(w http.ResponseWriter, r *http.Request) {
	log.Info(" -  Method:", r.Method, " - /")

	if r.Method == "GET" {
		t, _ := template.ParseFiles("html/index.html")
		t.Execute(w, nil)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func init() {
	log.Info(" -  Setup our Ctrl+C handler...")
	SetupCloseHandler()

	log.Info(" -  Opening the database...")
	var err error

	db, err = sql.Open("sqlite3", Database)
	if err != nil {
		fmt.Println(err)
	}
	db.SetMaxOpenConns(1)
	db.Exec("PRAGMA journal_mode=MEMORY;")
	db.Exec("PRAGMA _synchronous=OFF;")

	statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS Book (id INTEGER PRIMARY KEY, Title TEXT, Country TEXT, Date TEXT, Author TEXT)")
	if err != nil {
		fmt.Println(err)
	}
	statement.Exec()

}

// SetupCloseHandler https://golangcode.com/handle-ctrl-c-exit-in-terminal/
func SetupCloseHandler() {

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Info("\r- Ctrl+C pressed in Terminal, closing the DB")
		if err := db.Close(); err != nil {
			log.Critical(err)
		}
		os.Exit(0)
	}()
}

func main() {
	http.HandleFunc("/", mainpage)
	http.HandleFunc("/populateDataTable", paging)
	fileServer := http.FileServer(http.Dir("html/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))

	log.Info(" -  Listening on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(" -  ListenAndServe: ", err)
	}
}
