package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

//Book contains the format of the json file
type Book struct {
	Title   string `json:"Title"`
	Country string `json:"Country"`
	Date    string `json:"Date"`
	Author  string `json:"Author"`
}

func mainpage(w http.ResponseWriter, r *http.Request) {
	log.Println(" -  Method:", r.Method, " - /")
	var myJSON []Book
	plan, err := ioutil.ReadFile(`data/data.json`)
	if err != nil {
		fmt.Println(err)
	}
	if err := json.Unmarshal(plan, &myJSON); err != nil {
		fmt.Println(err)
	}
	if r.Method == "GET" {
		t, _ := template.ParseFiles("html/index.html")
		t.Execute(w, myJSON)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/", mainpage)
	fileServer := http.FileServer(http.Dir("html/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))

	log.Println(" -  Listening on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(" -  ListenAndServe: ", err)
	}
}
