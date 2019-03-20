package main

import (
	"os"
	"testing"
	"math"

	"github.com/mediocregopher/radix.v2/redis"
)


func dialTestRedis() *redis.Client {
	cli, err := redis.Dial("tcp", "redis:6380")
	if err != nil {
		panic(err)
	}
	cli.Cmd("FLUSHALL")
	return cli
}

func TestGetBalance(t *testing.T) {
	client := dialTestRedis()

	username := "user"
	balance := 1200.00

	client.Cmd("HSET", username, "Balance", balance)
	result := getBalance(client, username)
	if result != balance {
		t.Errorf("getBalance was incorrect, got: %f, want: %f.", result, balance)
	}
}

func TestAddBalance(t *testing.T) {
	client := dialTestRedis()

	username := "user"
	add := 300.00

	addBalance(client, username, add)
	result := getBalance(client, username)
	if result != add {
		t.Errorf("addBalance was incorrect, got: %f, want %f.", result, add)
	}
}

func TestStockOwned(t *testing.T) {
	client := dialTestRedis()

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
	client := dialTestRedis()

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

func TestQExists(t *testing.T) {
	client := dialTestRedis()

	stock := "ABC"

	result := qExists(client, stock)
	if result != false {
		t.Errorf("qExists was incorrect, got %t, want %t.", result, false)
	}

	client.Cmd("SET", stock, 123.00)

	result = qExists(client, stock)
	if result != true {
		t.Errorf("qExists was incorrect, got %t, want %t.", result, true)
	}
}

func TestSaveTransaction(t *testing.T) {
	client := dialTestRedis()

	username := "user"

	result, _ := client.Cmd("ZCOUNT", "HISTORY:"+username, "-inf", "+inf").Int()

	if result != 0 {
		t.Errorf("saveTransaction was incorrect, got %d, want %d.", result, 0)
	}

	command := "ADD"
	amount := "300.00"
	newBalance := "300.00"

	saveTransaction(client, username, command, amount, newBalance)

	result, _ = client.Cmd("ZCOUNT", "HISTORY:"+username, "-inf", "+inf").Int()

	if result != 1 {
		t.Errorf("saveTransaction was incorrect, got %d, want %d.", result, 1)
	}
}

func TestRedisADD(t *testing.T) {
	client := dialTestRedis()

	username := "user"
	amount := 123.00

	redisADD(client, username, amount)

	result := exists(client, username)
	if result != true{
		t.Errorf("redisADD was incorrect, got %t, want %t.", result, true)
	}

	newBalance := getBalance(client, username)
	if newBalance != amount {
		t.Errorf("redisADD was incorrect, got %f, want %f.", newBalance, amount)
	}

	// check that the transaction was saved in HISTORY:username
	count, _ := client.Cmd("ZCOUNT", "HISTORY:"+username, "-inf", "+inf").Int()

	if count != 1 {
		t.Errorf("saveTransaction was incorrect, got %d, want %d.", count, 1)
	}
}

func TestRedisBUY(t *testing.T) {
	client := dialTestRedis()

	username := "user"
	stock := "ABC"
	amount := 123.00
	stack_name := "userBUY:" + username

	redisBUY(client, username, stock, amount)

	count, _ := client.Cmd("LLEN", stack_name).Int()

	if count != 2 {
		t.Errorf("redisBUY is incorrect, got %d, want %d.", count, 2)
	}

	pop1, _ := client.Cmd("LPOP", stack_name).Str()
	if pop1 != stock {
		t.Errorf("redisBUY is incorrect, got %s, want %s.", pop1, stock)
	}

	pop2, _ := client.Cmd("LPOP", stack_name).Float64()
	if pop2 != amount {
		t.Errorf("redisBUY is incorrect, got %f, want %f.", pop2, amount)
	}
}

func TestRedisSELL(t *testing.T) {
	client := dialTestRedis()

	username := "user"
	stock := "ABC"
	amount := 123.00
	stack_name:= "userSELL:" + username

	redisSELL(client, username, stock, amount)

	count, _ := client.Cmd("LLEN", stack_name).Int()

	if count != 2 {
		t.Errorf("redisSELL is incorrect, got %d, want %d.", count, 2)
	}

	pop1, _ := client.Cmd("LPOP", stack_name).Str()
	if pop1 != stock {
		t.Errorf("redisSELL is incorrect, got %s, want %s.", pop1, stock)
	}

	pop2, _ := client.Cmd("LPOP", stack_name).Float64()
	if pop2 != amount {
		t.Errorf("redisSELL is incorrect, got %f, want %f.", pop2, amount)
	}
}

func TestRedisCOMMIT_BUY(t *testing.T) {
	os.Setenv("QUOTE_URL", ":1200")
	os.Setenv("AUDIT_URL", "localhost:1400")

	client := dialTestRedis()

	username := "user"
	stock := "ABC"
	amount := 1000.00
	balance := 5000.00
	stack_name := "userBUY:" + username

	redisADD(client, username, balance)
	redisQUOTE(client, 0, username, stock)
	redisBUY(client, username, stock, amount)
	redisCOMMIT_BUY(client, username)

	stack_len, _ := client.Cmd("LLEN", stack_name).Int()

	if stack_len != 0 {
		t.Errorf("redisCOMMIT_BUY is incorrect, got %d, want %d.", stack_len, 0)
	}

	new_balance := getBalance(client, username)
	quote_price, _ := client.Cmd("GET", stock + ":QUOTE").Float64()
	stock2buy := int(math.Floor(amount / quote_price))
	total_cost := quote_price * float64(stock2buy)

	if new_balance != (balance - total_cost) {
		t.Errorf("redisCOMMIT_BUY is incorrect, got %f, want %f balance.", new_balance, balance - total_cost)
	}
}
