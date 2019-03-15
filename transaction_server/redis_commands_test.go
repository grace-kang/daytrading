package main

import (
	"testing"
)

/*
/ These tests assume that the functions
/ dialRedis() and flushRedis() are
/ working properly. If they are not,
/ no tests in this test suite will
/ pass.
*/

func TestGetBalance(t *testing.T) {
	client := dialRedis()
	flushRedis(client)

	username := "user"
	balance := 1200.00

	client.Cmd("HSET", username, "Balance", balance)
	result := getBalance(client, username)
	if result != balance {
		t.Errorf("getBalance was incorrect, got: %f, want: %f.", result, balance)
	}
}

func TestAddBalance(t *testing.T) {
	client := dialRedis()
	flushRedis(client)

	username := "user"
	add := 300.00

	addBalance(client, username, add)
	result := getBalance(client, username)
	if result != add {
		t.Errorf("addBalance was incorrect, got: %f, want %f.", result, add)
	}
}

func TestStockOwned(t *testing.T) {
	client := dialRedis()
	flushRedis(client)

	username := "user"
	stock := "ABC"

	result := stockOwned(client, username, stock)
	if result != 0 {
		t.Errorf("stockOwned was incorrect, got %d, want %d.", result, 0)
	}

	amount := 31

	client.Cmd("HSET", username, stock, amount)
	result2 := stockOwned(client, username, stock)
	if result2 != amount {
		t.Errorf("stockOwned was incorrect, got %d, want %d.", result, amount)
	}
}

func TestsExists(t *testing.T) {
	client := dialRedis()
	flushRedis(client)

	username := "user"

	result := exists(client, username)
	if result != false {
		t.Errorf("exists was incorrect, got %t, want %t.", result, false)
	}

	client.Cmd("HMSET", username, "Balance", 0.00)

	result = exists(client, username)
	if result != true {
		t.Errorf("exists was incorrect, got %t, want %t.", result, true)
	}
}

