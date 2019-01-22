package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func getUserHandler(w http.ResponseWriter, r *http.Request) {

	users, err := store.GetUsers()

	// Everything else is the same as before
	userListBytes, err := json.Marshal(users)

	if err != nil {
		fmt.Println("getUserHandler: ", fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(userListBytes)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {

	user := User{}

	err := r.ParseForm()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	user.username = r.Form.Get("username")
	intBalance, err := strconv.Atoi(r.Form.Get("balance"))
	if err != nil {
		//handle
	}
	user.balance = intBalance
	err = store.CreateUser(&user)
	if err != nil {
		fmt.Println("createUserHandler 2: ", err)
	}
	logUserCommand(1, user)
	dumpLog(user)

	http.Redirect(w, r, "/assets/", http.StatusFound)
}
