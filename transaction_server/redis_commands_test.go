package main

import (
	"testing"
)

func TestGetBalance(t *testing.T) {
	client := dialRedis()
	username := "user"
	balance := 1200.00

	client.Cmd("HSET", username, "Balance", balance)
	result := getBalance(client, username)
	if result != balance {
		t.Errorf("getBalance was incorrect, got: %f, want: %f.", result, balance)
	}
}

