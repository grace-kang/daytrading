package main

import (
	"fmt"
	"net/http"
	"html/template"

	"github.com/gorilla/mux"
)

const (
	connHost = "localhost"
	connPort = "1400"
	connType = "http"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/login", loginHandler)

	err := http.ListenAndServe(":" + connPort, r)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
		return
	}

	fmt.Println("Web server is listening on port " + connPort)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	p := "test"
	t := template.Must(template.ParseFiles("assets/home.html"))
	t.Execute(w, p)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("login.html"))
	if r.Method != http.MethodPost {
		tmpl.Execute(w, nil)
		return
	}

	tmpl.Execute(w, struct{ Success bool }{true})
}
