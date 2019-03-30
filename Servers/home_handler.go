package main

import (
	"html/template"
	"net/http"
)

func getHomeHandler(w http.ResponseWriter, r *http.Request) {
	p := "test"
	t, _ := template.ParseFiles("assets/home.html")
	t.Execute(w, p)
}

