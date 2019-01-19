package main

// The sql go library is needed to interact with the database
import (
	"database/sql"
	"fmt"
)

type Store interface {
	CreateUser(user *User) error
	GetUsers() ([]*User, error)
}

type dbStore struct {
	db *sql.DB
}

func (store *dbStore) CreateUser(user *User) error {
	fmt.Println("in CreateUser now")
	// id_char_arr := []rune(user.id)
	_, err := store.db.Query("INSERT INTO users(username, balance) VALUES ($1,$2)", user.username, user.balance)
	return err
}

func (store *dbStore) GetUsers() ([]*User, error) {
	
	rows, err := store.db.Query("SELECT username, balance from users")

	fmt.Println("rowa are: ", rows)
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*User{}
	for rows.Next() {
		
		user := &User{}
		
		if err := rows.Scan(&user.username, &user.balance); err != nil {
			return nil, err
		}
		fmt.Println("in getuser, user is ", user)
		users = append(users, user)
	}
	return users, nil
}

var store Store

func InitStore(s Store) {
	store = s
}