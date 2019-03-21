package main

import (
	"os"
	"testing"
	"math"
	"strconv"

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

func TestRedisQUOTE(t *testing.T) {
	os.Setenv("QUOTE_URL", ":1200")
	os.Setenv("AUDIT_URL", "localhost:1400")

	client := dialTestRedis()

	stock := "ABC"
	username := "user"
	transNum := 0 //no transaction number

	redisQUOTE(client, transNum, username, stock)

	quote_key := stock + ":QUOTE"
	price, _ := client.Cmd("GET", quote_key).Float64()

	if price <= 0 {
		t.Errorf("redisQUOTE is incorrect, got %f for price.", price)
	}

	// QUOTE again to see if we get the existing price
	redisQUOTE(client, transNum, username, stock)

	new_price, _ := client.Cmd("GET", quote_key).Float64()

	if new_price != price {
		t.Errorf("redisQUOTE is incorrect, got %f, want %f for quote price", new_price, price)
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
	expected_balance := float64(balance - total_cost)

	formated_new_bal := strconv.FormatFloat(new_balance, 'f', 2, 64)
	formated_expected_bal := strconv.FormatFloat(expected_balance, 'f', 2, 64)

	if formated_new_bal != formated_expected_bal {
		t.Errorf("redisCOMMIT_BUY is incorrect, got %s, want %s balance.", formated_new_bal, formated_expected_bal)
	}
}

func TestRedisCOMMIT_SELL(t *testing.T) {
	os.Setenv("QUOTE_URL", ":1200")
	os.Setenv("AUDIT_URL", "localhost:1400")

	client := dialTestRedis()

	username := "user"
	stock := "ABC"
	amount := 1000.00
	balance := 5000.00
	stack_name := "userSELL:" + username

	// let user own 5000 of that stock
	client.Cmd("HSET", username, stock + ":OWNED", 5000)

	redisADD(client, username, balance)
	redisQUOTE(client, 0, username, stock)
	redisSELL(client, username, stock, amount)
	redisCOMMIT_SELL(client, username)

	stack_len, _ := client.Cmd("LLEN", stack_name).Int()

	if stack_len != 0 {
		t.Errorf("redisCOMMIT_SELL is incorrect, got %d, want %d.", stack_len, 0)
	}

	new_balance := getBalance(client, username)
	quote_price, _ := client.Cmd("GET", stock + ":QUOTE").Float64()
	stock2sell := int(math.Floor(amount / quote_price))
	total_cost := quote_price * float64(stock2sell)
	expected_balance := float64(balance + total_cost)

	formated_new_bal := strconv.FormatFloat(new_balance, 'f', 2, 64)
	formated_expected_bal := strconv.FormatFloat(expected_balance, 'f', 2, 64)

	if formated_new_bal != formated_expected_bal {
		t.Errorf("redisCOMMIT_BUY is incorrect, got %s, want %s balance.", formated_new_bal, formated_expected_bal)
	}
}

func TestRedisCANCEL_BUY (t *testing.T) {
	client := dialTestRedis()

	username := "user"
	amount := 123.00
	stock := "ABC"
	stack_name := "userBUY:" + username

	redisBUY(client, username, stock, amount)
	redisCANCEL_BUY(client, username)

	stack_len, _ := client.Cmd("LLEN", stack_name).Int()

	if stack_len != 0 {
		t.Errorf("redisCANCEL_BUY didn't cancel the BUY")
	}
}

func TestRedisCANCEL_SELL (t *testing.T) {
	client := dialTestRedis()

	username := "user"
	amount := 123.00
	stock := "ABC"
	stack_name := "userSELL:" + username

	redisSELL(client, username, stock, amount)
	redisCANCEL_SELL(client, username)

	stack_len, _ := client.Cmd("LLEN", stack_name).Int()

	if stack_len != 0 {
		t.Errorf("redisCANCEL_SELL didn't cancel the SELL")
	}
}

func TestRedisSET_BUY_AMOUNT(t *testing.T) {
	client := dialTestRedis()

	username := "user"
	stock := "ABC"
	amount := 123.00
	stack_name := stock + ":BUY:" + username

	redisSET_BUY_AMOUNT(client, username, stock, amount)

	set_buy_len, _ := client.Cmd("LLEN", stack_name).Int()

	if set_buy_len != 1 {
		t.Errorf("redisSET_BUY_AMOUNT is incorrect, got %d, want %d.", set_buy_len, 1)
	}
}

func TestRedisSET_BUY_TRIGGER(t *testing.T) {
	client := dialTestRedis()

	username := "user"
	stock := "ABC"
	amount := 123.00
	key_name := stock + ":BUYTRIG"
	hash_name := "BUYTRIGGERS:" + username

	redisSET_BUY_TRIGGER(client, username, stock, amount)

	result, _ := client.Cmd("HGET", username, key_name).Float64()
	if result != amount {
		t.Errorf("redisSET_BUY_TRIGGER is incorrect, got %f, want %f.", result, amount)
	}

	result, _ = client.Cmd("HGET", hash_name, stock).Float64()

	if result != amount {
		t.Errorf("redisSET_BUY_TRIGGER is incorrect, got %f, want %f.", result, amount)
	}
}

func TestRedisSET_SELL_TRIGGER(t *testing.T) {
	client := dialTestRedis()

	username := "user"
	stock := "ABC"
	amount := 123.00
	key_name := stock + ":SELLTRIG"
	hash_name := "SELLTRIGGERS:" + username

	redisSET_SELL_TRIGGER(client, username, stock, amount)

	result, _ := client.Cmd("HGET", username, key_name).Float64()
	if result != amount {
		t.Errorf("redisSET_SELL_TRIGGER is incorrect, got %f, want %f.", result, amount)
	}

	result, _ = client.Cmd("HGET", hash_name, stock).Float64()

	if result != amount {
		t.Errorf("redisSET_SELL_TRIGGER is incorrect, got %f, want %f.", result, amount)
	}
}

