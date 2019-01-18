package main

import (
	"net/http"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"database/sql"
	"fmt"
)

const (
    host     = "localhost"
    port     = 5432
    user     = "postgres"
    password = ""
    dbname   = "dayTrading"
)

func newRouter() *mux.Router {
	r := mux.NewRouter()

	staticFileDirectory := http.Dir("./assets/")
	staticFileHandler := http.StripPrefix("/assets/", http.FileServer(staticFileDirectory))
	r.PathPrefix("/assets/").Handler(staticFileHandler).Methods("GET")

	r.HandleFunc("/user", getUserHandler).Methods("GET")
	r.HandleFunc("/user", createUserHandler).Methods("POST")
	return r
}

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
        "password=%s dbname=%s sslmode=disable",
        host, port, user, password, dbname)
    db, err := sql.Open("postgres", psqlInfo)
    if err != nil {
        panic(err)
    }
    defer db.Close()

    err = db.Ping()
    if err != nil {
        panic(err)
    }

    fmt.Println("Successfully connected!")

	InitStore(&dbStore{db: db})

	fmt.Println("after InitStore")
	sqlStatement := `
INSERT INTO public.users (id, balance)  
VALUES ($1, $2)  
RETURNING id`
    id := "adcdwc"
    err = db.QueryRow(sqlStatement, 3, 19).Scan(&id)
    if err != nil {
    	fmt.Println("inside insert, after InitStore")
        panic(err)
    }
    fmt.Println("New user ID is:", id)

	r := newRouter()
	http.ListenAndServe(":8080", r)
}
