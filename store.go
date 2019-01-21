package main

// The sql go library is needed to interact with the database
import (
	"database/sql"
	"github.com/jinzhu/gorm"
)

type Store interface {
	CreateUser(user *User) error
	GetUsers() ([]*User, error)
}

type dbStore struct {
	db *sql.DB
}

func (store *dbStore) CreateUser(user *User) error {
	_, err := store.db.Query("INSERT INTO users(username, balance) VALUES ($1,$2)", user.username, user.balance)
	return err
}

func (store *dbStore) GetUsers() ([]*User, error) {
	
	rows, err := store.db.Query("SELECT username, balance from users")
	
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
		users = append(users, user)
	}
	return users, nil
}



var store Store

func InitStore(s Store) {
	store = s
}