package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mediocregopher/radix.v2/redis"
)

func dialRedis() *redis.Client {
	cli, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		// handle err
	}
	return cli
}

func getBalance(client *redis.Client, username string) float64 {
	getBAL, _ := client.Cmd("HGET", username, "Balance").Float64()
	return getBAL
}

func addBalance(client *redis.Client, username string, amount float64) {
	client.Cmd("HINCRBYFLOAT", username, "Balance", amount)
}

func stockOwned(client *redis.Client, username string, id string) int {
	stocksOwned, _ := client.Cmd("HGET", username, id).Int()
	return stocksOwned
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

func listStack(client *redis.Client, stackName string) []string {
	stack, _ := client.Cmd("LRANGE", stackName, 0, -1).List()
	return stack
}

func saveTransaction(client *redis.Client, username string, command string, params ...string) {
	// save transaction to HISTORY:username hash set
	timestamp := time.Now().Local()
	timestamp_str := timestamp.String()
	var transaction_string string
	var save bool

	if len(params) == 2 {
		// ADD
		save = true
		amount := params[0]
		newBalance := params[1]

		transaction_string = "[" + string(timestamp_str) + "] " + command + "- Amount: " + amount + ", New Balance: " + newBalance

	} else if len(params) == 5 {
		// COMMIT BUY OR SELL
		save = true
		stock := params[0]
		number := params[1]
		number_int, _ := strconv.Atoi(number)
		if number_int < 1 {
			save = false
		}
		price := params[2]
		totalCost := params[3]
		newBalance := params[4]

		transaction_string = "[" + string(timestamp_str) + "] " + command + "- Stock: " + stock + ", Number: " + number + ", Price: " + price + ", Total Cost: " + totalCost + ", New Balance: " + newBalance

	} else {
		fmt.Println("Error: Wrong number of arguments for saveTransaction()")
		os.Exit(1)
	}

	if save {
		client.Cmd("ZADD", "HISTORY:"+username, timestamp.UnixNano(), transaction_string)
	}
}

func redisADD(client *redis.Client, username string, amount float64) {
	oldBalance := getBalance(client, username)
	fmt.Println("Old Balance: ", oldBalance)
	exists := exists(client, username)
	if exists == false {
		client.Cmd("HMSET", username, "User", username, "Balance", amount)
	} else {
		addBalance(client, username, amount)
	}

	fmt.Println("ADD: ", amount)
	newBalance := getBalance(client, username)
	fmt.Println("New Balance: ", newBalance)

	//save to transaction history
	saveTransaction(client, username, "ADD", strconv.FormatFloat(amount, 'f', 2, 64), strconv.FormatFloat(newBalance, 'f', 2, 64))
}

func redisQUOTE(client *redis.Client, username string, symbol string) {
	fmt.Println("-----QUOTE-----")
	stringQ := symbol + ":QUOTE"
	stockPrice, _ := client.Cmd("HGET", username, stringQ).Float64()
	fmt.Println("QUOTE:", stockPrice)
	//client.Cmd("HINCRBYFLOAT", username, "Balance", -0.50)
	amount := -0.50
	addBalance(client, username, amount)
}

func redisBUY(client *redis.Client, username string, symbol string, amount float64) {
	//fmt.Println("newBUY", username, symbol, amount)
	/*
		check to see buy stack in redis cli
		LRANGE userBUY:oY01WVirLr 0 -1
	*/
	string3 := "userBUY:" + username
	client.Cmd("LPUSH", string3, amount)
	client.Cmd("LPUSH", string3, symbol)

	stack := listStack(client, string3)
	fmt.Println("User:", username, " BUYStack:", stack)
}

func redisSELL(client *redis.Client, username string, symbol string, amount float64) {
	//fmt.Println("newSELL", username, symbol, amount)
	string3 := "userSELL:" + username

	client.Cmd("LPUSH", string3, amount)
	client.Cmd("LPUSH", string3, symbol)

	stack := listStack(client, string3)
	fmt.Println("SELLStack: ", stack)
}
func redisCOMMIT_BUY(client *redis.Client, username string) {
	fmt.Println("-----COMMIT_BUY-----")

	/*
		1. LPOP - stock symbol
		2. LPOP - amount
		3. check user balance
		4. calculate num. stocks to buy
		5. decrease balance
		6. increase stocks
	*/

	/* 1, 2 */
	string3 := "userBUY:" + username
	stock, _ := client.Cmd("LPOP", string3).Str()
	amount, _ := client.Cmd("LPOP", string3).Float64()
	fmt.Println("SYMBOL:", stock, "AMOUNT:", amount)

	/* 3 */
	getBAL := getBalance(client, username)
	fmt.Println("Old Balance:", getBAL)

	/* 4 */
	stockQ := stock + ":QUOTE"
	stockPrice, _ := client.Cmd("HGET", username, stockQ).Float64()
	stock2BUY := int(math.Floor(amount / stockPrice))
	totalCOST := stockPrice * float64(stock2BUY)
	fmt.Println("Price:", stockPrice, "BUYAmount:", stock2BUY)
	fmt.Println("TotalCost:", totalCOST)

	/* 5 */
	addBalance(client, username, -totalCOST)
	getBAL2 := getBalance(client, username)
	fmt.Println("New Balance:", getBAL2)

	/* 6 */
	id := stock + ":OWNED"

	if stock2BUY > 0 {
		client.Cmd("HINCRBY", username, id, stock2BUY)
		// add stock to OWNED:[username] for ease of access in display summary
		client.Cmd("HINCRBY", "OWNED:"+username, id, stock2BUY)
	}

	stockOWNS := stockOwned(client, username, id)
	fmt.Println("Stock:", stock, "TOTAL:", stockOWNS)

	// save to transaction history
	saveTransaction(client, username, "COMMIT_BUY", stock, strconv.Itoa(stock2BUY), strconv.FormatFloat(stockPrice, 'f', 2, 64), strconv.FormatFloat(totalCOST, 'f', 2, 64), strconv.FormatFloat(getBAL2, 'f', 2, 64))
}

func redisCOMMIT_SELL(client *redis.Client, username string) {
	fmt.Println("-----COMMIT_SELL-----")

	/*
		1. LPOP - stock symbol
		2. LPOP - amount
		4. calculate num. stocks to buy
		5. increase balance
		6. decrease stocks
	*/

	/* 1, 2 */
	string3 := "userSELL:" + username
	stock, _ := client.Cmd("LPOP", string3).Str()
	amount, _ := client.Cmd("LPOP", string3).Float64()
	fmt.Println("SYMBOL:", stock, "AMOUNT:", amount)

	/* 4 */
	stockPrice, _ := client.Cmd("HGET", username, "QUOTE").Float64()
	fmt.Println("QUOTE:", stockPrice)
	stock2SELL := int(math.Floor(amount / stockPrice))
	totalCOST := stockPrice * float64(stock2SELL)
	fmt.Println("Price:", stockPrice, "SELLAmount:", stock2SELL)
	fmt.Println("TotalCost:", totalCOST)
	id := stock + ":OWNED"

	stocksOwned := stockOwned(client, username, id)
	/* 5 */
	if stocksOwned >= stock2SELL {
		client.Cmd("HINCRBYFLOAT", username, "Balance", totalCOST)
		getBAL3 := getBalance(client, username)
		fmt.Println("NEWBalance:", getBAL3)

		/* 6 */

		if stock2SELL > 0 {
			client.Cmd("HINCRBY", username, id, -stock2SELL)
			saveTransaction(client, username, "COMMIT_SELL", stock, strconv.Itoa(stock2SELL), strconv.FormatFloat(stockPrice, 'f', 2, 64), strconv.FormatFloat(totalCOST, 'f', 2, 64), strconv.FormatFloat(getBAL3, 'f', 2, 64))
		}
	}

	stockOWNS := stockOwned(client, username, id)
	fmt.Println("Stock:", stock, "TOTAL:", stockOWNS)

	// save to transaction history

}

func redisCANCEL_BUY(client *redis.Client, username string) {

	fmt.Println("-----CANCEL_BUY-----")
	/* Pop off 2 items from buy stack: stock name, and $ amount */
	string3 := "userBUY:" + username
	stock, _ := client.Cmd("LPOP", string3).Str()
	amount, _ := client.Cmd("LPOP", string3).Float64()
	fmt.Println("Stock:", stock, "Amount:", amount)

}

func redisCANCEL_SELL(client *redis.Client, username string) {

	fmt.Println("-----CANCEL_SELL-----")
	/* Pop off 2 items from buy stack: stock name, and $ amount */
	string3 := "userSELL:" + username
	stock, _ := client.Cmd("LPOP", string3).Str()
	amount, _ := client.Cmd("LPOP", string3).Float64()
	fmt.Println("Stock:", stock, "Amount:", amount)

}

func redisSET_BUY_AMOUNT(client *redis.Client, username string, symbol string, amount float64) {
	fmt.Println("-----SET_BUY_AMOUNT-----")
	string3 := symbol + ":BUY:" + username
	client.Cmd("LPUSH", string3, amount)

	/*
		push amount then trigger price in SET_BUY_TRIGGER
		these two operate in pairs
	*/

	stack, _ := client.Cmd("LRANGE", string3, 0, -1).List()
	fmt.Println("SETBUYTRIGGERStack: ", stack)

	/*
		decrease balance put in reserve
	*/

	client.Cmd("HINCRBYFLOAT", username, "Balance", -amount)
	getBAL2 := getBalance(client, username)
	fmt.Println("NEWBalance:", getBAL2)
}
func redisSET_BUY_TRIGGER(client *redis.Client, username string, symbol string, amount float64) {
	fmt.Println("-----SET_BUY_TRIGGER-----")
	string3 := symbol + ":BUYTRIG"
	client.Cmd("HSET", username, string3, amount)
	client.Cmd("HSET", "BUYTRIGGERS:"+username, symbol, amount)
}
func redisCANCEL_SET_BUY(client *redis.Client, username string, symbol string) {
	fmt.Println("-----CANCEL_SET_BUY-----")
	/* get length of stack */
	string3 := symbol + ":BUY:" + username

	stackLength, _ := client.Cmd("LLEN", string3).Int()
	fmt.Println("Stack length:", stackLength)

	for i := 0; i < stackLength; i++ {
		refund, _ := client.Cmd("LPOP", string3).Float64()
		addBalance(client, username, refund)
		getBAL2 := getBalance(client, username)
		fmt.Println("Refund:", refund)
		fmt.Println("New Balance:", getBAL2)
	}

	string4 := symbol + ":BUYTRIG"
	client.Cmd("HSET", username, string4, 0.00)

}

func redisSET_SELL_AMOUNT(client *redis.Client, username string, symbol string, amount float64) {
	fmt.Println("-----SET_SELL_AMOUNT-----")
	string3 := symbol + ":SELL:" + username
	client.Cmd("LPUSH", string3, amount)

	/*
		push amount then trigger price in SET_BUY_TRIGGER
		these two operate in pairs
	*/
	stack := listStack(client, string3)
	fmt.Println("SETSELLTRIGGERStack: ", stack)
}
func redisSET_SELL_TRIGGER(client *redis.Client, username string, symbol string, amount float64) {
	fmt.Println("-----SET_SELL_TRIGGER-----")
	string3 := symbol + ":SELLTRIG"
	client.Cmd("HSET", username, string3, amount)

	client.Cmd("HSET", "SELLTRIGGERS:"+username, symbol, amount)
}

func redisCANCEL_SET_SELL(client *redis.Client, username string, symbol string) {
	fmt.Println("-----CANCEL_SET_SELL-----")
	/* get length of stack */
	string3 := symbol + ":SELL:" + username
	stackLength, _ := client.Cmd("LLEN", string3).Int()
	fmt.Println("Stack length:", stackLength)

	for i := 0; i < stackLength; i++ {
		client.Cmd("LPOP", string3).Float64()
	}
	string4 := symbol + ":SELLTRIG"
	client.Cmd("HSET", username, string4, 0.00)
}

func redisDISPLAY_SUMMARY(client *redis.Client, username string) {
	fmt.Println("Username: ", username)
	fmt.Println("Balance: ", getBalance(client, username))
	stocks_owned, _ := client.Cmd("HGETALL", "OWNED:"+username).Map()
	fmt.Println("Stocks owned:")
	for key, val := range stocks_owned {
		fmt.Println(strings.Split("    "+key, ":")[0] + ": " + val)
	}
	fmt.Println("Buy triggers: ")
	buy_triggers, _ := client.Cmd("HGETALL", "BUYTRIGGERS:"+username).Map()
	sell_triggers, _ := client.Cmd("HGETALL", "SELLTRIGGERS:"+username).Map()
	for key, val := range buy_triggers {
		fmt.Println(strings.Split("    "+key, ":")[0] + ": " + val)
	}
	fmt.Println("Sell triggers: ")
	for key, val := range sell_triggers {
		fmt.Println(strings.Split("    "+key, ":")[0] + ": " + val)
	}
	fmt.Println("Transaction history: ")
	client.Cmd("SORT", "HISTORY:"+username)
	history, _ := client.Cmd("ZRANGE", "HISTORY:"+username, 0, -1).List()
	for _, val := range history {
		fmt.Println("    " + val)
	}
}
