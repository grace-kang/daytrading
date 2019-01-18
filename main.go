package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"fmt"
db  "./db"
	pg "github.com/go-pg/pg"
)

// type User struct {
// 	id   string
//     balance     int64
// }

// func (u User) String() string {
//     return fmt.Sprintf("User<%s %d>", u.id, u.balance)
// }

const (
    host     = "localhost"
    port     = 5432
    user     = "postgres"
    password = ""
    dbname   = "postgres"
)

func newRouter() *mux.Router {
	r := mux.NewRouter()

	staticFileDirectory := http.Dir("./assets/")
	staticFileHandler := http.StripPrefix("/assets/", http.FileServer(staticFileDirectory))
	r.PathPrefix("/assets/").Handler(staticFileHandler).Methods("GET")

	// r.HandleFunc("/user", getUserHandler).Methods("GET")
	// r.HandleFunc("/user", createUserHandler).Methods("POST")
	return r
}

func main() {
	
	pg_db := db.Connect()
	fmt.Println("after InitStore")
	saveUser(pg_db)
	r := newRouter()
	http.ListenAndServe(":8080", r)
}

func saveUser(dbRef *pg.DB) {
	newAPI := db.User {
		Id: "adcw",
		Balance: 300.0,
	}
	newAPI.Save(dbRef)
}
