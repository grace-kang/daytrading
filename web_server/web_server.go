package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	connHost = "localhost"
	connPort = "1600"
	connType = "http"
)

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/userCommands.js", SendJqueryJs)

	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/login", loginHandler)

	err := http.ListenAndServe(":"+connPort, r)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
		return
	}

	fmt.Println("Web server is listening on port " + connPort)
}

func SendJqueryJs(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile("tmp/userCommands.js")
	if err != nil {
		http.Error(w, "Couldn't read file", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Write(data)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// p := "test"
	// t := template.Must(template.ParseFiles("tmp/home.html"))
	// t.Execute(w, p)
	tpl, _ := ioutil.ReadFile("tmp/home.html")
	tplParsed, _ := template.New("test").Parse(string(tpl))
	tplParsed.Execute(w, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// tmpl := template.Must(template.ParseFiles("tmp/userCommands.html"))
	// if r.Method != http.MethodPost {
	// 	tmpl.Execute(w, nil)
	// 	return
	// }

	// tmpl.Execute(w, struct{ Success bool }{true})
	fmt.Println("in loginhandler")
	tpl, _ := ioutil.ReadFile("tmp/userCommands.html")
	tplParsed, _ := template.New("test").Parse(string(tpl))
	tplParsed.Execute(w, nil)
}
