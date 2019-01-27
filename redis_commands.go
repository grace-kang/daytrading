package main

import (
	"fmt"

	"github.com/mediocregopher/radix.v2/redis"
)

func dialRedis() *redis.Client {
	cli, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		// handle err
	}
	return cli
}

func exists(client *redis.Client, username string) bool {
	//client := dialRedis()
	exists, _ := client.Cmd("HGETALL", username).Map()
	if len(exists) == 0 {
		return false
	} else {
		return true
	}
}

func redisADD(client *redis.Client, username string, amount float64) {
	fmt.Println("newADD", username, amount)
	exists := exists(client, username)
	if exists == false {
		client.Cmd("HMSET", username, "User", username, "Balance", amount)
	} else {
		client.Cmd("HINCRBYFLOAT", username, "Balance", amount)
	}
	/* Display - get new balance (HGET) */
	fmt.Print("ADD:   ", amount)
	x, _ := client.Cmd("HGET", username, "Balance").Float64()
	fmt.Println(" Balance: ", x)
}

func redisBUY(client *redis.Client, username string, symbol string, amount float64) {
	fmt.Println("newBUY", username, symbol, amount)
	string2 := symbol + ":BUY"
	client.Cmd("HSET", username, string2, amount)
	fmt.Println("BUY: ", amount)
	currentBalance, _ := client.Cmd("HGET", username, "Balance").Float64()
	fmt.Println(" Balance: ", currentBalance)
}
func redisSELL(client *redis.Client, username string, symbol string, amount float64) {
	fmt.Println("newSELL", username, symbol, amount)
	string2 := symbol + ":SELL"
	/* HSET: set the sell amount in dollars for the chosen stock
	(still needs to be committed to make sale) */
	client.Cmd("HSET", username, string2, amount)
	fmt.Println("SELL:  ", amount)
}

