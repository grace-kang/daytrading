package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/mediocregopher/radix.v2/redis"
)

func dialRedis() *redis.Client {
	cli, err := redis.Dial("tcp", "redis:6379")
	if err != nil {
		// handle err
	}
	return cli
}

func flushRedis(client *redis.Client) {
	client.Cmd("FLUSHALL")
}

func getBalance(client *redis.Client, username string) float64 {
	getBAL, _ := client.Cmd("HGET", username, "Balance").Float64()
	return getBAL
}

func addBalance(client *redis.Client, username string, amount float64) {
	client.Cmd("HINCRBYFLOAT", username, "Balance", amount)
}

func stockOwned(client *redis.Client, username string, stock string) int {
	command := stock + ":OWNED"
	stocksOwned, _ := client.Cmd("HGET", username, command).Int()
	return stocksOwned
}

func addStock(client *redis.Client, username string, symbol string, num_stocks int) {
	command := symbol + ":OWNED"
	client.Cmd("HINCRBY", username, command, num_stocks)
	// add stock to OWNED:[username] for ease of access in display summary
	client.Cmd("HINCRBY", "OWNED:"+username, symbol, num_stocks)
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

func qExists(client *redis.Client, stock string) bool {
	//client := dialRedis()
	exists, _ := client.Cmd("EXISTS", stock).Int()
	fmt.Println("Exists: ", exists, " Stock:", stock)
	if exists == 0 {
		return false
	} else {
		return true
	}
}

func listStack(client *redis.Client, stackName string) []string {
	stack, _ := client.Cmd("LRANGE", stackName, 0, -1).List()
	return stack
}

func listNotEmpty(client *redis.Client, stackName string) bool {
	len, _ := client.Cmd("LLEN", stackName).Int()
	return len != 0
}

func clearStack(client *redis.Client, stackName string) {
	for listNotEmpty(client, stackName) == true {
		client.Cmd("LPOP", stackName)
	}
}

func getTransaction(client *redis.Client, username string) bytes.Buffer {
	fmt.Println("in getTransactioin")
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 0, '\t', 0)
	// getBalance(client, username

	if username == "root123" {
		fmt.Fprintf(writer, "\t|%s\t|%s\t|%s\t|%s\t|%s\n", "Time", "User", "Command", "Transaction Description", "Current Balance")
		client.Cmd("SORT", "HISTORY")
		history, _ := client.Cmd("ZRANGE", "HISTORY", 0, -1).List()
		fmt.Println("history iis ", history)
		for _, val := range history {
			time := strings.Split(val, ":")[0]
			user := strings.Split(val, ":")[1]
			transList := strings.Split(strings.Split(val, ":")[2], ";")
			fmt.Println("in trans hit, time is ", time, "user is ", user, "command is ", transList[0], "trans hist is ", transList[2], "current balacne is ", transList[1])
			fmt.Fprintf(writer, "\t\t|%s\t|%s\t|%s\t|%s\t|%s\n", time, user, transList[0], transList[2], transList[1])
		}
	} else {
		fmt.Fprintf(writer, "\t\t|%s\t|%s\t|%s\t|%s\n", "Time", "Command", "Transaction Description", "Current Balance")
		client.Cmd("SORT", "HISTORY")
		history, _ := client.Cmd("ZRANGE", "HISTORY", 0, -1, true).Map()
		for _, val := range history {
			user := strings.Split(val, ":")[1]
			if user == username {
				time := strings.Split(val, ":")[0]
				transList := strings.Split(strings.Split(val, ":")[2], ";")
				fmt.Fprintf(writer, "\t\t|%s\t|%s\t|%s\t|%s\t|%s\n", time, user, transList[0], transList[2], transList[1])
			}
		}
	}
	writer.Flush()
	fmt.Println(b.String())

	return b

}

func saveTransactionString(client *redis.Client, username string, command string, comment string, new_balance float64) {

	timestamp := time.Now().Local()
	timestamp_str := timestamp.String()
	amountS := fmt.Sprintf("%.2f", new_balance)
	transaction_string := command + ";" + amountS + ";" + comment
	client.Cmd("ZADD", "HISTORY", timestamp.UnixNano(), timestamp_str+":"+username+":"+transaction_string)
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

	exists := exists(client, username)
	if exists == false {
		client.Cmd("HMSET", username, "User", username, "Balance", amount)
	} else {
		addBalance(client, username, amount)
	}
	client.Cmd("HMSET", "usersList", username, amount)
	newBalance := getBalance(client, username)
	comment := fmt.Sprintf("add %.2f into balance", amount)
	saveTransactionString(client, username, "ADD", comment, newBalance)
}

func displayADD(client *redis.Client, username string, amount float64) {
	fmt.Println("-----ADD-----")
	fmt.Println("User: ", username)
	oldBalance := getBalance(client, username)
	fmt.Println("Old Balance: ", oldBalance)
	redisADD(client, username, amount)
	fmt.Println("ADD: ", amount)
	newBalance := getBalance(client, username)
	fmt.Println("New Balance: ", newBalance, "\n")
}

func redisQUOTE(client *redis.Client, transNum int, username string, stock string) {
	stringQ := stock + ":QUOTE"
	ex := qExists(client, stringQ)
	if ex == false {
		go goQuote(client, transNum, username, stock)
	} else {
		stringQ := stock + ":QUOTE"
		currentprice, _ := client.Cmd("GET", stringQ).Float64()
		LogSystemEventCommand(server, transNum, "QUOTE", username, fmt.Sprintf("%f", currentprice), stock, nil)
	}
	newBalance := getBalance(client, username)
	comment := fmt.Sprintf("quote stock %s", stock)
	saveTransactionString(client, username, "QUOTE", comment, newBalance)
}

func getQUOTE(client *redis.Client, transNum int, username string, stock string, force bool) float64 {
	if force == true {
		return goQuote(client, transNum, username, stock)
	}
	stringQ := stock + ":QUOTE"
	ex := qExists(client, stringQ)
	if ex == false {
		return goQuote(client, transNum, username, stock)
	}
	currentprice, _ := client.Cmd("GET", stringQ).Float64()
	LogSystemEventCommand(server, transNum, "QUOTE", username, fmt.Sprintf("%f", currentprice), stock, nil)
	return currentprice
}

func goQuote(client *redis.Client, transNum int, username string, stock string) float64 {
	stringQ := stock + ":QUOTE"

	QUOTE_URL := os.Getenv("QUOTE_URL")
	// fmt.Println("quoye url is " + QUOTE_URL)
	conn, _ := net.Dial("tcp", QUOTE_URL)

	conn.Write([]byte((stock + "," + username + "\n")))
	respBuf := make([]byte, 2048)
	_, err := conn.Read(respBuf)
	conn.Close()

	if err != nil {
		fmt.Printf("Error reading body: %s", err.Error())
	}
	respBuf = bytes.Trim(respBuf, "\x00")
	message := bytes.NewBuffer(respBuf).String()
	message = strings.TrimSpace(message)

	fmt.Println(string(message))

	split := strings.Split(message, ",")
	price, error := strconv.ParseFloat(strings.TrimSpace(split[0]), 64)
	if error != nil {
		LogErrorEventCommand(server, transNum, "QUOTE", username, strconv.FormatFloat(price, 'f', 2, 64), stock, nil, "failed in parsing quote stock price into float number")
	}
	quoteTimestamp := strings.TrimSpace(split[3])
	crytpoKey := split[4]

	quoteServerTime := ParseUint(quoteTimestamp, 10, 64)

	LogQuoteServerCommand(server, transNum, strings.TrimSpace(split[0]), stock, username, quoteServerTime, crytpoKey)
	client.Cmd("SET", stringQ, price)
	client.Cmd("EXPIRE", stringQ, 60)

	return price
}

func displayQUOTE(client *redis.Client, transNum int, username string, symbol string) {
	fmt.Println("-----QUOTE-----")
	redisQUOTE(client, transNum, username, symbol)
	stringQ := symbol + ":QUOTE"
	stockPrice, _ := client.Cmd("GET", stringQ).Float64()
	fmt.Println("QUOTE:", stockPrice, "\n")
}

func redisBUY(client *redis.Client, username string, symbol string, totalCost float64, stockBuy int) {
	/*
	  check to see buy stack in redis cli
	  LRANGE userBUY:oY01WVirLr 0 -1
	*/
	string3 := "userBUY:" + username
	client.Cmd("LPUSH", string3, symbol)
	client.Cmd("LPUSH", string3, totalCost)
	client.Cmd("LPUSH", string3, stockBuy)
	t := time.Now().Unix()
	fmt.Println("t is ", t)
	client.Cmd("LPUSH", string3, t)

	newBalance := getBalance(client, username)
	comment := fmt.Sprintf("buy %d of stocks %s. Total cost is %.2f", stockBuy, symbol, totalCost)
	saveTransactionString(client, username, "BUY", comment, newBalance)

}

func displayBUY(client *redis.Client, username string, symbol string, totalCost float64, stockAmount int) {
	fmt.Println("-----BUY-----")
	redisBUY(client, username, symbol, totalCost, stockAmount)
	string3 := "userBUY:" + username
	stack := listStack(client, string3)
	fmt.Println("User:", username, " BUYStack:", stack, "\n")
}

func redisSELL(client *redis.Client, username string, symbol string, totalEarned float64, stockNeeded int) {
	//fmt.Println("newSELL", username, symbol, amount)
	string3 := "userSELL:" + username
	client.Cmd("LPUSH", string3, symbol)
	client.Cmd("LPUSH", string3, totalEarned)
	client.Cmd("LPUSH", string3, stockNeeded)
	t := time.Now().Unix()
	fmt.Println("t is ", t)
	client.Cmd("LPUSH", string3, t)

	newBalance := getBalance(client, username)
	comment := fmt.Sprintf("sell %d of stocks %s. Total earned will be %.2f", stockNeeded, symbol, totalEarned)
	saveTransactionString(client, username, "SELL", comment, newBalance)

}

func displaySELL(client *redis.Client, username string, symbol string, totalEarned float64, stockNeeded int) {
	fmt.Println("-----SELL-----")
	redisSELL(client, username, symbol, totalEarned, stockNeeded)
	string3 := "userSELL:" + username
	stack := listStack(client, string3)
	fmt.Println("User: ", username, "SELLStack: ", stack, "\n")
}

func redisCOMMIT_BUY(client *redis.Client, username string, transNum int) string {

	/*
	  1. LPOP - stock symbol
	  2. LPOP - amount
	  3. check user balance
	  4. calculate num. stocks to buy
	  5. decrease balance
	  6. increase stocks
	*/

	string3 := "userBUY:" + username
	old_time, _ := client.Cmd("LPOP", string3).Int64()
	t := time.Now().Unix()
	diff := t - old_time
	fmt.Println("in commit_buy, old time is ", old_time, " time now is ", t, " diff is ", diff)
	if diff > 60 { //buy expire
		LogErrorEventCommand(server, transNum, "COMMIT_BUY", username, nil, nil, nil, "there is no buy to commit")
		clearStack(client, string3)
		fmt.Println("expiration!")
		stack := listStack(client, string3)
		fmt.Println("After clear the stack, User:", username, " BUYStack:", stack, "\n")
		return "there is no buy to commit"
	}
	stockBuy, _ := client.Cmd("LPOP", string3).Int()
	totalCost, _ := client.Cmd("LPOP", string3).Float64()
	stock, _ := client.Cmd("LPOP", string3).Str()

	getBAL := getBalance(client, username)

	if getBAL < totalCost {
		LogErrorEventCommand(server, transNum, "COMMIT_BUY", username, strconv.FormatFloat(totalCost, 'f', 2, 64), nil, nil, "user "+username+" doesn not have enough balance to buy stock "+stock)
		return "balace is not enough to commit buy"
	}

	/* 5 decrease the balance by totalCost*/
	addBalance(client, username, -totalCost)
	LogAccountTransactionCommand(server, transNum, "COMMIT_BUY", username, strconv.FormatFloat(totalCost, 'f', 2, 64))

	/* 6 add stock to the account*/
	addStock(client, username, stock, stockBuy)
	stockUnitPrice := totalCost / float64(stockBuy)

	/* 7 save to transaction list*/
	newBalance := getBalance(client, username)
	comment := fmt.Sprintf("Commit buy %d of stocks %s. Unit price is %.2f. Total cost is %.2f.", stockBuy, stock, stockUnitPrice, totalCost)
	saveTransactionString(client, username, "COMMIT_BUY", comment, newBalance)
	return ""
}

func displayCOMMIT_BUY(client *redis.Client, username string, transNum int) string {
	//string3 := "userBUY:" + username
	//fmt.Println("SYMBOL:", stock, "AMOUNT:", amount)
	fmt.Println("-----COMMIT_BUY-----")
	fmt.Println("User: ", username)
	/* 3 */
	getBAL := getBalance(client, username)
	fmt.Println("Old Balance:", getBAL)

	message := redisCOMMIT_BUY(client, username, transNum)

	getBAL2 := getBalance(client, username)
	fmt.Println("New Balance:", getBAL2, "\n")

	string3 := "userBUY:" + username
	stack := listStack(client, string3)
	fmt.Println("User:", username, " BUYStack:", stack, "\n")
	return message
}

func redisCOMMIT_SELL(client *redis.Client, username string, transNum int) string {

	/*
	  1. LPOP - stock symbol
	  2. LPOP - amount
	  4. calculate num. stocks to buy
	  5. increase balance
	  6. decrease stocks
	*/

	/* 1, 2 */
	string3 := "userSELL:" + username
	old_time, _ := client.Cmd("LPOP", string3).Int64()
	t := time.Now().Unix()
	diff := t - old_time
	fmt.Println("in commit_sell, old time is ", old_time, " time now is ", t, " diff is ", diff)
	if diff > 60 {
		LogErrorEventCommand(server, transNum, "COMMIT_SELL", username, nil, nil, nil, "there is no sell to commit")
		clearStack(client, string3)
		fmt.Println("expiration! Going to clear the stack")
		return "there is no sell to commit"
	}
	stockNeeded, _ := client.Cmd("LPOP", string3).Int()
	totalEarned, _ := client.Cmd("LPOP", string3).Float64()
	stock, _ := client.Cmd("LPOP", string3).Str()

	/* 4 */
	stocksOwned := stockOwned(client, username, stock)

	/* 5 */

	if stocksOwned < stockNeeded {
		LogErrorEventCommand(server, transNum, "COMMIT_SELL", username, strconv.FormatFloat(totalEarned, 'f', 2, 64), nil, nil, "user "+username+" doesn't have enough stock "+stock+" to sell ")
		return "stock owned is not enough to sell"
	}

	addBalance(client, username, totalEarned)
	LogAccountTransactionCommand(server, transNum, "COMMIT_SELL", username, strconv.FormatFloat(totalEarned, 'f', 2, 64))

	/* 6 */
	addStock(client, username, stock, -stockNeeded)

	newBalance := getBalance(client, username)
	comment := fmt.Sprintf("commit sell %d of stocks %s. Total earned is %.2f", stockNeeded, stock, totalEarned)
	saveTransactionString(client, username, "COMMIT_SELL", comment, newBalance)

	return ""

}

func displayCOMMIT_SELL(client *redis.Client, username string, transNum int) string {
	fmt.Println("-----COMMIT_SELL-----")
	fmt.Println("User: ", username)
	getBAL := getBalance(client, username)
	fmt.Println("Old Balance:", getBAL)

	message := redisCOMMIT_SELL(client, username, transNum)

	getBAL2 := getBalance(client, username)
	fmt.Println("New Balance:", getBAL2, "\n")

	string3 := "userSELL:" + username
	stack := listStack(client, string3)
	fmt.Println("User:", username, " SELLStack:", stack, "\n")
	return message
}

func redisCANCEL_BUY(client *redis.Client, username string, transNum int) string {

	/* Pop off 2 items from buy stack: stock name, and $ amount */
	string3 := "userBUY:" + username
	old_time, _ := client.Cmd("LPOP", string3).Int64()
	t := time.Now().Unix()
	diff := t - old_time
	fmt.Println("in cancel_buy, time is ", old_time, " time now is ", t, " diff is ", diff)
	if diff > 60 {
		LogErrorEventCommand(server, transNum, "CANCEL_BUY", username, nil, nil, nil, "there is no buy to cancel")
		clearStack(client, string3)
		fmt.Println("expiration!")
		stack := listStack(client, string3)
		fmt.Println("After clear the stack, User:", username, " BUYStack:", stack, "\n")
		return "there is no buy to cancel"
	}

	stockBuy, _ := client.Cmd("LPOP", string3).Int()
	totalCost, _ := client.Cmd("LPOP", string3).Float64()
	stock, _ := client.Cmd("LPOP", string3).Str()

	//fmt.Println("Stock:", stock, "Amount:", amount)
	newBalance := getBalance(client, username)
	comment := fmt.Sprintf("cancel buy %d of stocks %s. Total earned is %.2f", stockBuy, stock, totalCost)
	saveTransactionString(client, username, "CANCEL_BUY", comment, newBalance)

	return ""

}
func displayCANCEL_BUY(client *redis.Client, username string, transNum int) string {
	fmt.Println("-----CANCEL_BUY-----")
	message := redisCANCEL_BUY(client, username, transNum)
	string3 := "userBUY:" + username
	stack := listStack(client, string3)
	fmt.Println("User:", username, " BUYStack:", stack, "\n")
	return message
}

func redisCANCEL_SELL(client *redis.Client, username string, transNum int) string {

	/* Pop off 2 items from buy stack: stock name, and $ amount */
	string3 := "userSELL:" + username
	old_time, _ := client.Cmd("LPOP", string3).Int64()
	t := time.Now().Unix()
	diff := t - old_time
	fmt.Println("in cancel_sell, old time is ", old_time, " time now is ", t, " diff is ", diff)
	if diff > 60 {
		LogErrorEventCommand(server, transNum, "CANCEL_SELL", username, nil, nil, nil, "there is no sell to cancel")
		clearStack(client, string3)
		fmt.Println("expiration!")
		stack := listStack(client, string3)
		fmt.Println("After clear the stack, User:", username, " BUYStack:", stack, "\n")
		return "there is no sell to cancel"
	}

	stockNeeded, _ := client.Cmd("LPOP", string3).Int()
	totalEarned, _ := client.Cmd("LPOP", string3).Float64()
	stock, _ := client.Cmd("LPOP", string3).Str()

	newBalance := getBalance(client, username)
	comment := fmt.Sprintf("cancel buy %d of stocks %s. Total earned is %.2f", stockNeeded, stock, totalEarned)
	saveTransactionString(client, username, "CANCEL_BUY", comment, newBalance)
	//fmt.Println("Stock:", stock, "Amount:", amount)
	return ""
}

func displayCANCEL_SELL(client *redis.Client, username string, transNum int) string {
	fmt.Println("-----CANCEL_SELL-----")
	message := redisCANCEL_SELL(client, username, transNum)
	string3 := "userSELL:" + username
	stack := listStack(client, string3)
	fmt.Println("User:", username, " SELLStack:", stack, "\n")
	return message
}

func addSetBuyAmount(client *redis.Client, username string, symbol string, amount float64) {
	string3 := symbol + ":BUY:" + username
	client.Cmd("LPUSH", string3, amount)
	client.Cmd("HSET", username+":SETBUY", symbol, 0) //store all the symbols of setbuy for a given user
}

func hasSetBuyBefore(client *redis.Client, username string, symbol string) bool {
	exist, _ := client.Cmd("HEXISTS", username+":SETBUY", symbol).Int()
	return exist == 1
}

func hasSetSellBefore(client *redis.Client, username string, symbol string) bool {
	exist, _ := client.Cmd("HEXISTS", username+":SETSELL", symbol).Int()
	return exist == 1
}

func redisSET_BUY_AMOUNT(client *redis.Client, username string, symbol string, amount float64, transNum int) string {

	balance := getBalance(client, username)
	if balance < amount {
		LogErrorEventCommand(server, transNum, "SET_BUY_AMOUNT", username, nil, nil, nil, "user "+username+" does not have any enough balance to set buy amount")
		return "balance is not enough to set buy amount"
	}

	// push dollarAmount into stack
	addSetBuyAmount(client, username, symbol, amount)

	// decrease balance put in reserve
	addBalance(client, username, -amount)
	LogAccountTransactionCommand(server, transNum, "SET_BUY_AMOUNT", username, strconv.FormatFloat(amount, 'f', 2, 64))

	newBalance := getBalance(client, username)
	comment := fmt.Sprintf("add set buy amount for stock %s. Total cost is %.2f.", symbol, amount)
	saveTransactionString(client, username, "SET_BUY_AMOUNT", comment, newBalance)

	return ""
}

func displaySET_BUY_AMOUNT(client *redis.Client, username string, symbol string, amount float64, transNum int) string {
	fmt.Println("-----SET_BUY_AMOUNT-----")
	message := redisSET_BUY_AMOUNT(client, username, symbol, amount, transNum)
	fmt.Println("Username: ", username, " old balance: ", getBalance(client, username))

	string3 := symbol + ":BUY:" + username
	stack, _ := client.Cmd("LRANGE", string3, 0, -1).List()
	fmt.Println("SETBUYAMOUNTstack for ", symbol, ":", stack)

	getBAL2 := getBalance(client, username)
	fmt.Println("NEWBalance:", getBAL2, "\n")
	return message
}

func addSetBuyTrigger(client *redis.Client, username string, symbol string, totalCost float64, unitPricePoint float64) {
	string3 := symbol + ":BUYTRIG"
	client.Cmd("HSET", username, string3, unitPricePoint)
	//save for transaction
	t := time.Now().Format("20060102150405")
	client.Cmd("HSET", "BUYTRIGGERS:"+username+":UNIT", symbol+":"+t, unitPricePoint)
	client.Cmd("HSET", "BUYTRIGGERS:"+username+":TOTAL", symbol+":"+t, totalCost)
}

func clearSetBuyTriggers(client *redis.Client, username string, symbol string) {
	string4 := symbol + ":BUYTRIG"
	client.Cmd("HDEL", username, string4)

	unitTriggers, _ := client.Cmd("HGETALL", "BUYTRIGGERS:"+username+":UNIT").Map()
	totalTriggers, _ := client.Cmd("HGETALL", "BUYTRIGGERS:"+username+":TOTAL").Map()
	for symbolTime, unitTriggerPoint := range unitTriggers {
		total := totalTriggers[symbolTime]
		totalCost, _ := strconv.ParseFloat(total, 64)
		s := strings.Split(symbolTime, ":")[0]
		if s == symbol {
			client.Cmd("HDEL", "BUYTRIGGERS:"+username+":UNIT", symbolTime)
			addBalance(client, username, totalCost)
			client.Cmd("HDEL", "BUYTRIGGERS:"+username+":TOTAL", symbolTime)
			newBalance := getBalance(client, username)
			comment := fmt.Sprintf("Cancel set buy trigger for stock %s. Trigger point is %.2f. Refund is %.2f.", symbol, unitTriggerPoint, totalCost)
			saveTransactionString(client, username, "CANCEL_SET_BUY", comment, newBalance)
		}
	}
}

func redisSET_BUY_TRIGGER(client *redis.Client, username string, symbol string, amount float64, transNum int) string {

	setBuy_string3 := symbol + ":BUY:" + username
	if listNotEmpty(client, setBuy_string3) == false {
		LogErrorEventCommand(server, transNum, "SET_BUY_TRIGGER", username, nil, symbol, nil, "user "+username+" does not have any set buy to trigger")
		return "there is no set buy to trigger"
	}
	totalCost, _ := client.Cmd("LPOP", setBuy_string3).Float64()
	addSetBuyTrigger(client, username, symbol, totalCost, amount)

	go goCheckSellTriggers(username, symbol)

	newBalance := getBalance(client, username)
	comment := fmt.Sprintf("Set buy trigger for stock %s. Total cost is %.2f. Unit trigger point is %.2f", symbol, totalCost, amount)
	saveTransactionString(client, username, "SET_BUY_TRIGGER", comment, newBalance)

	return ""
}

func displaySET_BUY_TRIGGER(client *redis.Client, username string, symbol string, amount float64, transNum int) string {
	fmt.Println("-----SET_BUY_TRIGGER-----")
	fmt.Println("Username: ", username)

	message := redisSET_BUY_TRIGGER(client, username, symbol, amount, transNum)

	string3 := "BUYTRIGGERS:" + username + ":UNIT"
	triggers, _ := client.Cmd("HGETALL", string3).List()
	fmt.Println("BUYTRIGGERS: ", triggers, "\n")
	return message
}

func redisCANCEL_SET_BUY(client *redis.Client, username string, symbol string, transNum int) string {

	if hasSetBuyBefore(client, username, symbol) == false {
		LogErrorEventCommand(server, transNum, "CANCEL_SET_BUY", username, nil, symbol, nil, "user "+username+" does not have any set buy to cancel")
		return "there is no set buy to cancel"
	}

	/* get length of stack */
	string3 := symbol + ":BUY:" + username

	stackLength, _ := client.Cmd("LLEN", string3).Int()
	//fmt.Println("Stack length:", stackLength)

	for i := 0; i < stackLength; i++ {
		refund, _ := client.Cmd("LPOP", string3).Float64()
		addBalance(client, username, refund)
		newBalance := getBalance(client, username)
		comment := fmt.Sprintf("Cancel set buy for stock %s. Refund is %.2f. Unit trigger point is %.2f", symbol, refund)
		saveTransactionString(client, username, "CANCEL_SET_BUY", comment, newBalance)
		//getBAL2 := getBalance(client, username)
		//fmt.Println("Refund:", refund)
		//fmt.Println("New Balance:", getBAL2)
	}

	clearSetBuyTriggers(client, username, symbol)
	return ""
}

func displayCANCEL_SET_BUY(client *redis.Client, username string, symbol string, transNum int) string {
	fmt.Println("-----CANCEL_SET_BUY-----")
	fmt.Println("Username: ", username)
	fmt.Println("before cancel, balance is ", getBalance(client, username), '\n')

	string3 := symbol + ":BUY:" + username
	stack, _ := client.Cmd("LRANGE", string3, 0, -1).List()
	fmt.Println("Before cancel, SETBUYAMOUNTstack for ", symbol, ":", stack)

	message := redisCANCEL_SET_BUY(client, username, symbol, transNum)

	stack, _ = client.Cmd("LRANGE", string3, 0, -1).List()
	fmt.Println("After cancel, SETBUYAMOUNTstack for ", symbol, ":", stack)

	// string4 := "BUYTRIGGERS:" + username
	// triggers, _ := client.Cmd("HGETALL", string4).List()
	// fmt.Println("BUYTRIGGERS: ", triggers)

	string4 := "BUYTRIGGERS:" + symbol + ":UNIT"
	triggers, _ := client.Cmd("HGETALL", string4).List()
	fmt.Println("BUYTRIGGERS:"+symbol+":UNIT", triggers)
	string4 = "BUYTRIGGERS:" + symbol + ":TOTAL"
	triggers, _ = client.Cmd("HGETALL", string4).List()
	fmt.Println("BUYTRIGGERS:"+symbol+":TOTAL", triggers)

	getBAL2 := getBalance(client, username)
	fmt.Println("NEWBalance:", getBAL2, "\n")
	return message
}

func addSetSellAmount(client *redis.Client, username string, symbol string, amount float64) {
	string3 := symbol + ":SELL:" + username
	client.Cmd("LPUSH", string3, amount)
	client.Cmd("HSET", username+":SETSELL", symbol, 0) //store all the symbols of setbuy for a given user
}

func redisSET_SELL_AMOUNT(client *redis.Client, username string, symbol string, amount float64, transNum int) string {

	stockOwned := stockOwned(client, username, symbol)
	getPrice := getQUOTE(client, transNum, username, symbol, true)
	stockNeeded := int(amount / getPrice)
	fmt.Println("stock owned: ", stockOwned, "stockneeded: ", stockNeeded)
	if stockOwned == 0 {
		LogErrorEventCommand(server, transNum, "SET_SELL_AMOUNT", username, strconv.FormatFloat(amount, 'f', 2, 64), symbol, nil, "user "+username+" does not have stock "+symbol+" to set sell")
		return "user doesn't have any stock " + symbol + " in the account to set sell amount"
	}
	if stockOwned < stockNeeded {
		LogErrorEventCommand(server, transNum, "SET_SELL_AMOUNT", username, strconv.FormatFloat(amount, 'f', 2, 64), symbol, nil, "user "+username+" does not have enough stock "+symbol+" to set sell")
		return "stack owned is not enough to set sell"
	}

	addSetSellAmount(client, username, symbol, amount)

	newBalance := getBalance(client, username)
	comment := fmt.Sprintf("Add set sell amount for stock %s. Total benefit will be %.2f", symbol, amount)
	saveTransactionString(client, username, "SET_SELL_AMOUNT", comment, newBalance)

	return ""
}

func displaySET_SELL_AMOUNT(client *redis.Client, username string, symbol string, amount float64, transNum int) string {
	fmt.Println("-----SET_SELL_AMOUNT-----")
	message := redisSET_SELL_AMOUNT(client, username, symbol, amount, transNum)

	fmt.Println("Username: ", username)
	string3 := symbol + ":SELL:" + username
	stack := listStack(client, string3)
	fmt.Println("SETSELLAMOUNTStack for ", symbol, ": ", stack, "\n")
	return message
}

func addSetSellTrigger(client *redis.Client, username string, symbol string, totalEarn float64, unitPrice float64, maxStock int) {
	string3 := symbol + ":SELLTRIG"
	client.Cmd("HSET", username, string3, unitPrice)

	t := time.Now().Format("20060102150405")
	client.Cmd("HSET", "SELLTRIGGERS:"+username+":UNIT", symbol+":"+t, unitPrice)
	client.Cmd("HSET", "SELLTRIGGERS:"+username+":TOTAL", symbol+":"+t, totalEarn)
	client.Cmd("HSET", "SELLTRIGGERS:"+username+":STOCKS", symbol+":"+t, maxStock)

}

func redisSET_SELL_TRIGGER(client *redis.Client, username string, symbol string, amount float64, transNum int) string {

	setBuy_string3 := symbol + ":SELL:" + username
	if listNotEmpty(client, setBuy_string3) == false {
		LogErrorEventCommand(server, transNum, "SET_SELL_TRIGGER", username, nil, symbol, nil, "user "+username+" does not have any set sell to trigger")
		return "there is no set sell to trigger"
	}

	unitPrice := amount

	totalEarn, _ := client.Cmd("LPOP", setBuy_string3).Float64()
	maxStock := int(totalEarn / unitPrice)
	stockOwned := stockOwned(client, username, symbol)

	if stockOwned < maxStock {
		LogErrorEventCommand(server, transNum, "SET_SELL_TRIGGER", username, nil, symbol, nil, "user "+username+" does not have any stock to set sell trigger")
		return "stock owned is not enough to set sell trigger "
	}

	addStock(client, username, symbol, -maxStock)
	addSetSellTrigger(client, username, symbol, totalEarn, unitPrice, maxStock)

	go goCheckSellTriggers(username, symbol)

	newBalance := getBalance(client, username)
	comment := fmt.Sprintf("Add set sell trigger for stock %s. Total earn will be %.2f. Unit trigger point is %.2f. Maximum stocks to sell is ", symbol, totalEarn, unitPrice, maxStock)
	saveTransactionString(client, username, "SET_SELL_TRIGGER", comment, newBalance)
	return ""
}

func displaySET_SELL_TRIGGER(client *redis.Client, username string, symbol string, amount float64, transNum int) string {
	fmt.Println("-----SET_SELL_TRIGGER-----")
	fmt.Println("Username: ", username)
	message := redisSET_SELL_TRIGGER(client, username, symbol, amount, transNum)

	triggers, _ := client.Cmd("HGETALL", "SELLTRIGGERS:"+username+":UNIT").List()
	fmt.Println("SELLTRIGGERS: ", triggers, "\n")
	return message
}

func clearSetSellTriggers(client *redis.Client, username string, symbol string) {

	string3 := symbol + ":SELLTRIG"
	client.Cmd("HDEL", username, string3)

	UNIT_triggers, _ := client.Cmd("HGETALL", "SELLTRIGGERS:"+username+":UNIT").Map()
	TOTAL_triggers, _ := client.Cmd("HGETALL", "SELLTRIGGERS:"+username+":TOTAL").Map()
	MAXSTOCKS_triggers, _ := client.Cmd("HGETALL", "SELLTRIGGERS:"+username+":STOCKS").Map()
	for symbolTime, triggerPrice := range UNIT_triggers {
		s := strings.Split(symbolTime, ":")[0]
		if s == symbol {
			client.Cmd("HDEL", "SELLTRIGGERS:"+username+":UNIT", symbolTime)
			totalAmount, _ := TOTAL_triggers[symbolTime]
			client.Cmd("HDEL", "SELLTRIGGERS:"+username+":TOTAL", symbolTime)
			numStockInt, _ := strconv.Atoi(MAXSTOCKS_triggers[symbolTime])
			addStock(client, username, symbol, numStockInt) // add stock back
			client.Cmd("HDEL", "SELLTRIGGERS:"+username+":STOCKS", symbolTime)

			newBalance := getBalance(client, username)
			comment := fmt.Sprintf("Cancel set sell trigger for stock %s. Refund is %.2f. Unit trigger point is %.2f. Maximum stocks to sell is ", symbol, totalAmount, triggerPrice, numStockInt)
			saveTransactionString(client, username, "CANCEL_SET_SELL", comment, newBalance)
		}

	}
}

func redisCANCEL_SET_SELL(client *redis.Client, username string, symbol string, transNum int) string {

	if hasSetSellBefore(client, username, symbol) == false {
		LogErrorEventCommand(server, transNum, "CANCEL_SELL_AMOUNT", username, nil, symbol, nil, "user "+username+" does not have any set sell to cancel")
		return "there is no set sell to cancel"
	}

	string3 := symbol + ":SELL:" + username

	stackLength, _ := client.Cmd("LLEN", string3).Int()
	for i := 0; i < stackLength; i++ {
		amount, _ := client.Cmd("LPOP", string3).Float64()
		newBalance := getBalance(client, username)
		comment := fmt.Sprintf("Cancel set sell amount for stock %s. Amount for this set_sell is %.2f. ", symbol, amount)
		saveTransactionString(client, username, "CANCEL_SET_SELL", comment, newBalance)
	}

	clearSetSellTriggers(client, username, symbol)

	return ""

}

func displayCANCEL_SET_SELL(client *redis.Client, username string, symbol string, transNum int) string {
	fmt.Println("-----CANCEL_SET_SELL-----")
	fmt.Println("Username: ", username)
	message := redisCANCEL_SET_SELL(client, username, symbol, transNum)

	string3 := symbol + ":SELL:" + username
	stack, _ := client.Cmd("LRANGE", string3, 0, -1).List()
	fmt.Println("SETSELLAMOUNTstack for ", symbol, ":", stack)

	string4 := "SELLTRIGGERS:" + username + ":UNIT"
	triggers, _ := client.Cmd("HGETALL", string4).List()
	fmt.Println("SELLTRIGGERS: ", triggers)

	getBAL2 := getBalance(client, username)
	fmt.Println("NEWBalance:", getBAL2, "\n")
	return message
}

func displayAllAccountInfo(client *redis.Client) bytes.Buffer {
	userMaps, _ := client.Cmd("HGETALL", "usersList").Map()
	var b bytes.Buffer
	for user, _ := range userMaps {
		b.WriteString("\n--------------------User:" + user + "-----------------------\n")
		s := redisDISPLAY_SUMMARY(client, user)
		b.Write(s.Bytes())
	}
	return b
}

func redisDISPLAY_SUMMARY(client *redis.Client, username string) bytes.Buffer {
	fmt.Println("-----DISPLAY_SUMMARY-----")

	// writer.Init(os.Stdout, 0, 8, 0, '\t', 0)
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 0, '\t', 0)
	fmt.Fprintf(writer, "Username: %s\nBalance: %.2f\n", username, getBalance(client, username))
	// fmt.Println("Username: ", username)
	// fmt.Println("Balance: ", getBalance(client, username))
	stocks_owned, _ := client.Cmd("HGETALL", "OWNED:"+username).Map()
	fmt.Fprintf(writer, "Stocks owned:\n")
	fmt.Fprintf(writer, "\t\t|%s\t|%s\n", "stock", "num_stocks")
	for key, val := range stocks_owned {
		fmt.Fprintf(writer, "\t\t%s\t%s\n", key, val)
	}
	fmt.Fprintf(writer, "\nPending buys in the stack:\n")
	fmt.Fprintf(writer, "\t\t|%s\t|%s\t|%s\t|%s\n", "stock", "num_stocks", "totalCost", "timestamp")
	buys := listStack(client, "userBUY:"+username)
	// fmt.Println("buys are: ", buys)
	for i := 0; i < len(buys); i = i + 4 {
		time := buys[i]
		num_stocks, _ := strconv.Atoi(buys[i+1])
		totalCost, _ := strconv.ParseFloat(buys[i+2], 64)
		symbol := buys[i+3]
		fmt.Fprintf(writer, "\t\t|%s\t|%d\t|%.2f\t|%s\n", symbol, num_stocks, totalCost, time)
	}

	// client.Cmd("LPUSH", string3, symbol)
	// client.Cmd("LPUSH", string3, totalEarned)
	// client.Cmd("LPUSH", string3, stockNeeded)
	// t := time.Now().Unix()
	// fmt.Println("t is ", t)
	// client.Cmd("LPUSH", string3, t)

	fmt.Fprintf(writer, "\nPending sells in the stack:\n")
	fmt.Fprintf(writer, "\t\t|%s\t|%s\t|%s\t|%s\n", "stock", "num_stocks", "totalEarn", "timestamp")
	sells := listStack(client, "userSELL:"+username)
	// fmt.Println("sells are: ", sells)
	for i := 0; i < len(sells); i = i + 4 {
		time := sells[i]
		num_stocks := sells[i+1]
		totalEarned, _ := strconv.ParseFloat(sells[i+2], 64)
		symbol := sells[i+3]
		fmt.Fprintf(writer, "\t\t|%s\t|%s\t|%.2f\t|%s\n", symbol, num_stocks, totalEarned, time)
	}
	fmt.Fprintf(writer, "\nSet Buy Amount in the stack:\n")
	fmt.Fprintf(writer, "\t\t|%s\t|%s\n", "stock", "totalCost")

	set_buy_stocks, _ := client.Cmd("HGETALL", username+":SETBUY").Map()
	// fmt.Println("set buys are: ", set_buy_stocks)
	for symbol, _ := range set_buy_stocks {
		// fmt.Println("setbuysymbol is ", symbol)
		set_buys := listStack(client, symbol+":BUY:"+username)
		if len(set_buys) == 0 {
			continue
		}
		// fmt.Println("set_buys is ", set_buys)
		for j := 0; j < len(set_buys); j = j + 1 {
			amount, _ := strconv.ParseFloat(set_buys[j], 64)
			fmt.Fprintf(writer, "\t\t|%s\t|%.2f\n", symbol, amount)
		}
		// fmt.Fprintf(writer, "\t\t%s\t%s\n", "----", "----")
	}

	fmt.Fprintf(writer, "\nSet Buy Triggers:\n")
	fmt.Fprintf(writer, "\t\t|%s\t|%s\t|%s\n", "stock", "triggerPoint", "totalCost")

	buy_triggers, _ := client.Cmd("HGETALL", "BUYTRIGGERS:"+username+":UNIT").Map()
	totalCostMap, _ := client.Cmd("HGETALL", "BUYTRIGGERS:"+username+":TOTAL").Map()
	fmt.Println("printing out unit buy_triggers")
	for symbolTime, unitPrice := range buy_triggers {
		fmt.Println("symbol is ", symbolTime, ", unit price is ", unitPrice)
		symbol := strings.Split(symbolTime, ":")[0]
		totalCost := totalCostMap[symbolTime]
		totalCostF, _ := strconv.ParseFloat(totalCost, 64)
		unitPriceF, _ := strconv.ParseFloat(unitPrice, 64)
		fmt.Fprintf(writer, "\t\t|%s\t|%.2f\t|%.2f\n", symbol, unitPriceF, totalCostF)
		// fmt.Fprintf(writer, "\t\t%s\t%s\t%s\n", "----", "----", "----")
	}

	fmt.Fprintf(writer, "\nSet Sell Amount in the stack:\n")
	fmt.Fprintf(writer, "\t\t|%s\t|%s\n", "stock", "totalEarn")

	set_sell_stocks, _ := client.Cmd("HGETALL", username+":SETSELL").Map()
	for symbol, _ := range set_sell_stocks {
		set_sells := listStack(client, symbol+":SELL:"+username)
		if len(set_sells) == 0 {
			continue
		}
		// fmt.Println("set_sells is ", set_buys)
		for j := 0; j < len(set_sells); j = j + 1 {
			amount, _ := strconv.ParseFloat(set_sells[j], 64)
			fmt.Fprintf(writer, "\t\t%s\t%.2f\n", symbol, amount)
		}
		// fmt.Fprintf(writer, "\t\t%s\t%s\n", "----", "----")
	}

	fmt.Fprintf(writer, "\t\t|%s\t|%s\t|%s\t|%s\n", "stock", "triggerPoint", "totalEarn", "maxStock")
	sellTriggers, _ := client.Cmd("HGETALL", "SELLTRIGGERS:"+username+":UNIT").Map()
	totalEarnMap, _ := client.Cmd("HGETALL", "SELLTRIGGERS:"+username+":TOTAL").Map()
	maxStockMap, _ := client.Cmd("HGETALL", "SELLTRIGGERS:"+username+":STOCKS").Map()
	for symbolTime, unitPrice := range sellTriggers {
		symbol := strings.Split(symbolTime, ":")[0]
		totalEarn := totalEarnMap[symbolTime]
		maxStock := maxStockMap[symbolTime]
		totalEarnF, _ := strconv.ParseFloat(totalEarn, 64)
		unitPriceF, _ := strconv.ParseFloat(unitPrice, 64)
		// fmt.Println("has sell triggers: symbol:", symbol, " unitPricd:", unitPriceF, " total earn: ", totalEarnF, " max stock:", maxStock)
		fmt.Fprintf(writer, "\t\t|%s\t|%.2f\t|%.2f\t|%s\n", symbol, unitPriceF, totalEarnF, maxStock)
	}
	writer.Flush()

	// fmt.Println("Transaction history: ")
	// client.Cmd("SORT", "HISTORY:"+username)
	// history, _ := client.Cmd("ZRANGE", "HISTORY:"+username, 0, -1).List()
	// for _, val := range history {
	// 	fmt.Println("    " + val)
	// }

	fmt.Println(b.String())
	return b

}
