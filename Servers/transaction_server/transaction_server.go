package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mediocregopher/radix.v2/pool"
)

const (
	connHost = "localhost"
	connPort = "80"
	connType = "http"
)

var display bool
var db *pool.Pool

var db2 *pool.Pool

//var db2 *pool.Pool
var server = "server1"

func init() {
	var err error
	// Establish a pool of 10 connections to the Redis server listening on
	// port 6379 of the local machine.
	db, err = pool.New("tcp", "redis:6379", 20)
	if err != nil {
		log.Panic(err)
	}

	db2, err = pool.New("tcp", "redis2:6379", 20)
	if err != nil {
		log.Panic(err)
	}

}

func main() {
	if len(os.Args) == 1 {
		display = false
	} else if len(os.Args) == 2 {
		if os.Args[1] == "-display" {
			display = true
		} else {
			fmt.Println("Unknown argument. Include no argument or -display")
			os.Exit(1)
		}
	} else {
		fmt.Println("Wrong number of arguments. Specify none or -display")
		os.Exit(1)
	}

	//client := dialRedis()
	//flushRedis(client)

	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/buy", buyHandler)
	http.HandleFunc("/sell", sellHandler)
	http.HandleFunc("/quote", quoteHandler)
	http.HandleFunc("/commit_buy", commitBuyHandler)
	http.HandleFunc("/commit_sell", commitSellHandler)
	http.HandleFunc("/cancel_buy", cancelBuyHandler)
	http.HandleFunc("/cancel_sell", cancelSellHandler)
	http.HandleFunc("/set_buy_amount", setBuyAmountHandler)
	http.HandleFunc("/set_buy_trigger", setBuyTriggerHandler)
	http.HandleFunc("/cancel_set_buy", cancelSetBuyHandler)
	http.HandleFunc("/set_sell_amount", setSellAmountHandler)
	http.HandleFunc("/set_sell_trigger", setSellTriggerHandler)
	http.HandleFunc("/cancel_set_sell", cancelSetSellHandler)
	http.HandleFunc("/display_summary", displaySummaryHandler)
	http.HandleFunc("/transaction_history", transactionHistoryHandler)

	http.HandleFunc("/dumpLog", dumpLogHandler)
	http.HandleFunc("/clearSystemLogs", clearSystemLogHandler)

	err := http.ListenAndServe(":"+connPort, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
	fmt.Println("Transaction server listening on " + connHost + ":" + connPort)
}

func goCheckBuyTriggers(username string, symbol string) {
	for {
		runOn(func() {
			fmt.Println(time.Now())
			checkBuyTriggers(username, symbol)
		}, 60*time.Second)
	}
}

func goCheckSellTriggers(username string, symbol string) {
	for {
		runOn(func() {
			fmt.Println(time.Now())
			checkSellTriggers(username, symbol)
		}, 60*time.Second)
	}
}

func runOn(f func(), d time.Duration) {
	now := time.Now()
	d = now.Add(d).Truncate(d).Sub(now)
	time.Sleep(d)
	f()
}

func checkSellTriggers(username string, symbol string) {
	fmt.Println("in checkSellTriggers")
	//get all triggers
	client, _ := db.Get()
	defer db.Put(client)
	UNIT_triggers, _ := client.Cmd("HGETALL", "SELLTRIGGERS:"+username+":UNIT").Map()
	TOTAL_triggers, _ := client.Cmd("HGETALL", "SELLTRIGGERS:"+username+":TOTAL").Map()
	MAXSTOCKS_triggers, _ := client.Cmd("HGETALL", "SELLTRIGGERS:"+username+":STOCKS").Map()
	for symbolTime, triggerPrice := range UNIT_triggers {
		unitTriggerPointF, _ := strconv.ParseFloat(triggerPrice, 64)
		s := strings.Split(symbolTime, ":")[0]
		if s == symbol {
			symbolPrice := getQUOTE(client, 0, username, symbol, true)
			if symbolPrice >= unitTriggerPointF {
				fmt.Println("sell trigger activated!!")
				totalAmount, _ := strconv.ParseFloat(TOTAL_triggers[symbolTime], 64)
				numStockInt, _ := strconv.Atoi(MAXSTOCKS_triggers[symbolTime])
				addStock(client, username, symbol, numStockInt) // add stock back
				stockNeeded := int(totalAmount / symbolPrice)
				newBenefit := symbolPrice * float64(stockNeeded)
				if stockNeeded <= 0 {
					continue
				}

				client.Cmd("HDEL", "SELLTRIGGERS:"+username+":UNIT", symbolTime)
				client.Cmd("HDEL", "SELLTRIGGERS:"+username+":TOTAL", symbolTime)
				client.Cmd("HDEL", "SELLTRIGGERS:"+username+":STOCKS", symbolTime)

				comment := fmt.Sprintf("%s;%.2f;%.2f;%.2f;%d;%.2f;%d", symbol, unitTriggerPointF, totalAmount, symbolPrice, numStockInt, newBenefit, stockNeeded)
				client.Cmd("HSET", "SELLTRIGGERSCOMPLETE:"+username, symbol, comment)

				if display == false {
					redisSELL(client, username, symbol, newBenefit, stockNeeded)
				} else {
					displaySELL(client, username, symbol, newBenefit, stockNeeded)
				}

			}
		}
	}
}

func checkBuyTriggers(username string, symbol string) {
	fmt.Println("in checkBuyTriggers")
	client, _ := db.Get()
	defer db.Put(client)
	//check if the price is higher than trigger point
	unitTriggers, _ := client.Cmd("HGETALL", "BUYTRIGGERS:"+username+":UNIT").Map()
	totalTriggers, _ := client.Cmd("HGETALL", "BUYTRIGGERS:"+username+":TOTAL").Map()
	for symbolTime, unitTriggerPoint := range unitTriggers {
		total := totalTriggers[symbolTime]
		totalCost, _ := strconv.ParseFloat(total, 64)
		unitTriggerPointF, _ := strconv.ParseFloat(unitTriggerPoint, 64)
		s := strings.Split(symbolTime, ":")[0]
		if s == symbol {
			symbolPrice := getQUOTE(client, 0, username, symbol, true)
			if symbolPrice <= unitTriggerPointF {
				fmt.Println("buy trigger activated!!")
				addBalance(client, username, totalCost)
				stockToBuy := int(totalCost / symbolPrice)
				exactTotalPrice := float64(stockToBuy) * symbolPrice
				fmt.Println("exactTotalPrice is ", exactTotalPrice)
				if stockToBuy <= 0 {
					continue
				}
				client.Cmd("HDEL", "BUYTRIGGERS:"+username+":UNIT", symbolTime)
				client.Cmd("HDEL", "BUYTRIGGERS:"+username+":TOTAL", symbolTime)

				comment := fmt.Sprintf("%s;%.2f;%.2f;%.2f;%.2f;%d", symbol, unitTriggerPointF, totalCost, symbolPrice, exactTotalPrice, stockToBuy)
				client.Cmd("HSET", "BUYTRIGGERSCOMPLETE:"+username, symbol, comment)

				if display == false {
					redisBUY(client, username, symbol, exactTotalPrice, stockToBuy)
				} else {
					displayBUY(client, username, symbol, exactTotalPrice, stockToBuy)
				}
			}

		}
	}
}

func clearSystemLogHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("in clearSystemLogHandler")
	clearSystemLogs()
}

func checkUserExists(transNum int, username string, command string) {
	client, _ := db.Get()
	defer db.Put(client)
	exists := exists(client, username)
	if exists == false {
		client.Cmd("HMSET", username, "User", username, "Balance", 0)
	}
}

func ParseUint(s string, base int, bitSize int) uint64 {
	unit_, _ := strconv.ParseUint(s, base, bitSize)
	return unit_
}

func dumpLogHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	transNum, _ := strconv.Atoi(r.Form.Get("transNum"))
	filename := r.Form.Get("filename")
	username := r.Form.Get("username")
	if username != "" {
		LogUserCommand(server, transNum, "DUMPLOG", username, nil, nil, filename)
		DumpLog(filename, username)
	} else {
		LogSystemEventCommand(server, transNum, "DUMPLOG", nil, nil, nil, filename)
		DumpLog(filename, nil)
	}
	w.Write([]byte("dumplog successfully. \n"))
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db.Get()
	defer db.Put(client)
	user := r.Form.Get("user")
	transNum, _ := strconv.Atoi(r.Form.Get("transNum"))
	amount, err := strconv.ParseFloat(strings.TrimSpace(r.Form.Get("amount")), 64)
	if err != nil {
		fmt.Println(err)
		w.Write([]byte("cannot parse amount into floating number"))
		return
	}

	LogUserCommand(server, transNum, "ADD", user, r.Form.Get("amount"), nil, nil)

	// log error message and return if amount is negative
	if amount < 0 {
		LogErrorEventCommand(server, transNum, "add", user, r.Form.Get("amount"), nil, nil, "cannot add negative amount into account")
		w.Write([]byte("cannot add negative balance"))
		return
	}

	if display == false {
		redisADD(client, user, amount)
	} else {
		displayADD(client, user, amount)
	}

	LogAccountTransactionCommand(server, transNum, "add", user, r.Form.Get("amount"))
	//w.Write([]byte("ADD complete"))
	w.Write([]byte("added balance " + r.Form.Get("amount") + " successfully\n"))
}

func buyHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db.Get()
	defer db.Put(client)
	user := r.Form.Get("user")
	transNum, _ := strconv.Atoi(r.Form.Get("transNum"))
	symbol := r.Form.Get("symbol")
	amount, err := strconv.ParseFloat(strings.TrimSpace(r.Form.Get("amount")), 64)
	if err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		w.Write([]byte("cannot parse amount into floating number"))
		return
	}

	if amount < 0 {
		LogErrorEventCommand(server, transNum, "buy", user, r.Form.Get("amount"), nil, nil, "cannot buy negative amount of stock")
		w.Write([]byte("amount cannot be negative"))
		return
	}

	LogUserCommand(server, transNum, "BUY", user, r.Form.Get("amount"), symbol, nil)
	checkUserExists(transNum, user, "BUY")

	getBAL := getBalance(client, user)

	if getBAL < amount {
		w.Write([]byte("balance is not enough to buy"))
		LogErrorEventCommand(server, transNum, "BUY", user, strconv.FormatFloat(amount, 'f', 2, 64), nil, nil, "user "+user+" does not have enough balance to buy stock "+symbol)
		return
	}

	getPrice := getQUOTE(client, transNum, user, symbol, true)
	fmt.Println("amount is ", amount, "price is ", getPrice)
	stockToBuy := int(amount / getPrice)
	exactTotalPrice := float64(stockToBuy) * getPrice
	fmt.Println("exactTotalPrice is ", exactTotalPrice)
	if stockToBuy <= 0 {
		LogErrorEventCommand(server, transNum, "BUY", user, strconv.FormatFloat(amount, 'f', 2, 64), nil, nil, "amount is too low to buy any of the stock")
		w.Write([]byte("amount is too low to buy any of the stock. Price of stock is higher than amount."))
		return
	}

	if display == false {
		redisBUY(client, user, symbol, exactTotalPrice, stockToBuy)
	} else {
		displayBUY(client, user, symbol, exactTotalPrice, stockToBuy)
	}

	w.Write([]byte("pushed command: buy amount " + strings.TrimSpace(r.Form.Get("amount")+" of stock "+symbol+" successfully\n")))
}

func sellHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db.Get()
	defer db.Put(client)
	user := r.Form.Get("user")
	transNum, _ := strconv.Atoi(r.Form.Get("transNum"))
	symbol := r.Form.Get("symbol")
	amount, err := strconv.ParseFloat(strings.TrimSpace(r.Form.Get("amount")), 64)

	if err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		w.Write([]byte("cannot parse amount into floating number"))
		LogUserCommand(server, transNum, "SELL", user, r.Form.Get("amount"), symbol, nil)
		return
	}

	if amount < 0 {
		LogErrorEventCommand(server, transNum, "sell", user, r.Form.Get("amount"), nil, nil, "cannot sell negative amount into stock")
		w.Write([]byte("amount cannot be negative"))
		LogUserCommand(server, transNum, "SELL", user, r.Form.Get("amount"), symbol, nil)
		return
	}

	LogUserCommand(server, transNum, "SELL", user, r.Form.Get("amount"), symbol, nil)
	/*check if user exists or not*/
	checkUserExists(transNum, user, "SELL")

	stockOwned := stockOwned(client, user, symbol)
	getPrice := getQUOTE(client, transNum, user, symbol, true)
	stockNeeded := int(amount / getPrice)
	newBenefit := getPrice * float64(stockNeeded)
	if stockOwned < stockNeeded {
		LogErrorEventCommand(server, transNum, "SELL", user, strconv.FormatFloat(amount, 'f', 2, 64), symbol, nil, "user "+user+" does not have enough stock "+symbol+" to sell")
		w.Write([]byte("stack owned is not enough to sell"))
		return
	}

	if stockNeeded <= 0 {
		w.Write([]byte("amount is too low to sell any of the stock. Price of stock is higher than amount."))
		return
	}

	if display == false {
		redisSELL(client, user, symbol, newBenefit, stockNeeded)
	} else {
		displaySELL(client, user, symbol, newBenefit, stockNeeded)
	}

	w.Write([]byte("pushed command: sell amount " + r.Form.Get("amount") + " of stock " + symbol + " successfully\n"))
}

func quoteHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db2.Get()
	defer db2.Put(client)
	transNum, _ := strconv.Atoi(r.Form.Get("transNum"))
	user := r.Form.Get("user")
	symbol := r.Form.Get("symbol")

	if display == false {
		redisQUOTE(client, transNum, user, symbol)
	} else if display == true {
		displayQUOTE(client, transNum, user, symbol)
	}
}

func commitBuyHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db.Get()
	defer db.Put(client)
	user := r.Form.Get("user")
	transNum, _ := strconv.Atoi(r.Form.Get("transNum"))

	LogUserCommand(server, transNum, "COMMIT_BUY", user, nil, nil, nil)
	checkUserExists(transNum, user, "COMMIT_BUY")

	string3 := "userBUY:" + user

	if listNotEmpty(client, string3) == false {
		LogErrorEventCommand(server, transNum, "COMMIT_BUY", user, nil, nil, nil, "user "+user+" does not have any buy to commit")
		w.Write([]byte("there is no buy to commit"))
		return
	}
	var message string
	if display == false {
		message = redisCOMMIT_BUY(client, user, transNum)
	} else {
		message = displayCOMMIT_BUY(client, user, transNum)
	}
	if message == "" {
		w.Write([]byte("commit buy successfully\n"))
	} else {
		w.Write([]byte(message))
	}

}

func commitSellHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db.Get()
	defer db.Put(client)
	user := r.Form.Get("user")
	transNum, _ := strconv.Atoi(r.Form.Get("transNum"))

	LogUserCommand(server, transNum, "COMMIT_SELL", user, nil, nil, nil)

	string3 := "userSELL:" + user

	if listNotEmpty(client, string3) == false {
		LogErrorEventCommand(server, transNum, "COMMIT_SELL", user, nil, nil, nil, "user "+user+" does not have any sell to commit")
		w.Write([]byte("there is no sell to commit"))
		return
	}

	var message string
	if display == false {
		message = redisCOMMIT_SELL(client, user, transNum)
	} else {
		message = displayCOMMIT_SELL(client, user, transNum)
	}

	if message == "" {
		w.Write([]byte("commit sell successfully\n"))
	} else {
		w.Write([]byte(message))
	}
}

func cancelBuyHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db.Get()
	defer db.Put(client)
	user := r.Form.Get("user")
	transNum, _ := strconv.Atoi(r.Form.Get("transNum"))

	LogUserCommand(server, transNum, "CANCEL_BUY", user, nil, nil, nil)

	string3 := "userBUY:" + user

	if listNotEmpty(client, string3) == false {
		LogErrorEventCommand(server, transNum, "CANCEL_BUY", user, nil, nil, nil, "user "+user+" does not have any buy to cancel")
		w.Write([]byte("there is no buy to cancel"))
		return
	}
	var message string
	if display == false {
		message = redisCANCEL_BUY(client, user, transNum)
	} else {
		message = displayCANCEL_BUY(client, user, transNum)
	}

	if message == "" {
		w.Write([]byte("cancel buy successfully\n"))
	} else {
		w.Write([]byte(message))
	}
}

func cancelSellHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db.Get()
	defer db.Put(client)
	user := r.Form.Get("user")
	transNum, _ := strconv.Atoi(r.Form.Get("transNum"))

	LogUserCommand(server, transNum, "CANCEL_SELL", user, nil, nil, nil)

	string3 := "userSELL:" + user

	if listNotEmpty(client, string3) == false {
		LogErrorEventCommand(server, transNum, "CANCEL_SELL", user, nil, nil, nil, "user "+user+" does not have any sell to cancel")
		w.Write([]byte("there is no sell to cancel"))
		return
	}
	var message string
	if display == false {
		message = redisCANCEL_SELL(client, user, transNum)
	} else {
		message = displayCANCEL_SELL(client, user, transNum)
	}
	if message == "" {
		w.Write([]byte("cancel sell successfully\n"))
	} else {
		w.Write([]byte(message))
	}

}

func setBuyAmountHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db.Get()
	defer db.Put(client)
	user := r.Form.Get("user")
	transNum, _ := strconv.Atoi(r.Form.Get("transNum"))
	symbol := r.Form.Get("symbol")
	amount, err := strconv.ParseFloat(strings.TrimSpace(r.Form.Get("amount")), 64)
	if err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		w.Write([]byte("cannot parse amount into floating number"))
		LogUserCommand(server, transNum, "SET_BUY_AMOUNT", user, r.Form.Get("amount"), symbol, nil)
		return
	}

	if amount < 0 {
		LogErrorEventCommand(server, transNum, "set_buy_amount", user, r.Form.Get("amount"), nil, nil, "amount cannot be negative")
		w.Write([]byte("amount cannot be negative"))
		LogUserCommand(server, transNum, "SET_BUY_AMOUNT", user, r.Form.Get("amount"), symbol, nil)
		return
	}

	LogUserCommand(server, transNum, "SET_BUY_AMOUNT", user, r.Form.Get("amount"), symbol, nil)
	var message string
	if display == false {
		message = redisSET_BUY_AMOUNT(client, user, symbol, amount, transNum)
	} else {
		message = displaySET_BUY_AMOUNT(client, user, symbol, amount, transNum)
	}

	if message == "" {
		w.Write([]byte("SET BUY AMOUNT complete\n"))
	} else {
		w.Write([]byte(message))
	}
}

func setBuyTriggerHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db.Get()
	defer db.Put(client)
	user := r.Form.Get("user")
	transNum, _ := strconv.Atoi(r.Form.Get("transNum"))
	symbol := r.Form.Get("symbol")
	amount, err := strconv.ParseFloat(strings.TrimSpace(r.Form.Get("amount")), 64)
	if err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		w.Write([]byte("cannot parse amount into floating number"))
		LogUserCommand(server, transNum, "SET_BUY_TRIGGER", user, r.Form.Get("amount"), symbol, nil)
		return
	}

	if amount < 0 {
		LogErrorEventCommand(server, transNum, "set_buy_trigger", user, r.Form.Get("amount"), nil, nil, "amount cannot be negative")
		w.Write([]byte("amount cannot be negative"))
		LogUserCommand(server, transNum, "SET_BUY_TRIGGER", user, r.Form.Get("amount"), symbol, nil)
		return
	}

	LogUserCommand(server, transNum, "SET_BUY_TRIGGER", user, r.Form.Get("amount"), symbol, nil)
	var message string
	if display == false {
		message = redisSET_BUY_TRIGGER(client, user, symbol, amount, transNum)
	} else {
		message = displaySET_BUY_TRIGGER(client, user, symbol, amount, transNum)
	}
	if message == "" {
		w.Write([]byte("set buy trigger successfully\n"))
	} else {
		w.Write([]byte(message))
	}
	//w.Write([]byte("SET BUY TRIGGER complete"))
}

func cancelSetBuyHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db.Get()
	defer db.Put(client)
	user := r.Form.Get("user")
	transNum, _ := strconv.Atoi(r.Form.Get("transNum"))
	symbol := r.Form.Get("symbol")

	LogUserCommand(server, transNum, "CANCEL_SET_BUY", user, nil, symbol, nil)
	var message string
	if display == false {
		message = redisCANCEL_SET_BUY(client, user, symbol, transNum)
	} else {
		message = displayCANCEL_SET_BUY(client, user, symbol, transNum)
	}
	if message == "" {
		w.Write([]byte("cancel set buy successfully\n"))
	} else {
		w.Write([]byte(message))
	}
	//w.Write([]byte("CANCEL SET BUY complete"))
}

func setSellAmountHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db.Get()
	defer db.Put(client)
	user := r.Form.Get("user")
	transNum, _ := strconv.Atoi(r.Form.Get("transNum"))
	symbol := r.Form.Get("symbol")
	amount, err := strconv.ParseFloat(strings.TrimSpace(r.Form.Get("amount")), 64)
	if err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		w.Write([]byte("cannot parse amount into floating number"))
		LogUserCommand(server, transNum, "SET_SELL_AMOUNT", user, r.Form.Get("amount"), symbol, nil)
		return
	}

	if amount < 0 {
		LogErrorEventCommand(server, transNum, "set_sell_amount", user, r.Form.Get("amount"), nil, nil, "amount cannot be negative")
		w.Write([]byte("amount cannot be negative"))
		LogUserCommand(server, transNum, "SET_SELL_AMOUNT", user, r.Form.Get("amount"), symbol, nil)
		return
	}

	LogUserCommand(server, transNum, "SET_SELL_AMOUNT", user, r.Form.Get("amount"), symbol, nil)

	//check if the user has enough amount of stocks or not based on the current price
	var message string
	if display == false {
		message = redisSET_SELL_AMOUNT(client, user, symbol, amount, transNum)
	} else {
		message = displaySET_SELL_AMOUNT(client, user, symbol, amount, transNum)
	}
	if message == "" {
		w.Write([]byte("SET SELL AMOUNT complete"))
	} else {
		w.Write([]byte(message))
	}
	//w.Write([]byte("SET SELL AMOUNT complete"))
}

func setSellTriggerHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db.Get()
	defer db.Put(client)
	user := r.Form.Get("user")
	transNum, _ := strconv.Atoi(r.Form.Get("transNum"))
	symbol := r.Form.Get("symbol")
	amount, err := strconv.ParseFloat(strings.TrimSpace(r.Form.Get("amount")), 64)

	if err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		w.Write([]byte("cannot parse amount into floating number"))
		LogUserCommand(server, transNum, "SET_SELL_TRIGGER", user, r.Form.Get("amount"), symbol, nil)
		return
	}

	LogUserCommand(server, transNum, "SET_SELL_TRIGGER", user, r.Form.Get("amount"), symbol, nil)

	if amount < 0 {
		LogErrorEventCommand(server, transNum, "set_sell_trigger", user, r.Form.Get("amount"), nil, nil, "amount cannot be negative")
		w.Write([]byte("amount cannot be negative"))
		LogUserCommand(server, transNum, "SET_SELL_TRIGGER", user, r.Form.Get("amount"), symbol, nil)
		return
	}

	var message string
	if display == false {
		message = redisSET_SELL_TRIGGER(client, user, symbol, amount, transNum)
	} else {
		message = displaySET_SELL_TRIGGER(client, user, symbol, amount, transNum)
	}
	if message == "" {
		w.Write([]byte("set sell trigger successfully\n"))
	} else {
		w.Write([]byte(message))
	}
	//w.Write([]byte("SET SELL TRIGGER complete"))
}

func cancelSetSellHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db.Get()
	defer db.Put(client)
	user := r.Form.Get("user")
	transNum, _ := strconv.Atoi(r.Form.Get("transNum"))
	symbol := r.Form.Get("symbol")

	LogUserCommand(server, transNum, "CANCEL_SET_SELL", user, nil, symbol, nil)
	var message string
	if display == false {
		message = redisCANCEL_SET_SELL(client, user, symbol, transNum)
	} else {
		message = displayCANCEL_SET_SELL(client, user, symbol, transNum)
	}
	if message == "" {
		w.Write([]byte("cancel set sell successfully\n"))
	} else {
		w.Write([]byte(message))
	}
	//w.Write([]byte("CANCEL SET SELL complete"))
}

func displaySummaryHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	user := r.Form.Get("user")
	transNum, _ := strconv.Atoi(r.Form.Get("transNum"))
	LogUserCommand(server, transNum, "DISPLAY_SUMMARY", user, nil, nil, nil)

	client, _ := db.Get()
	defer db.Put(client)
	if user == "root123" {
		s := displayAllAccountInfo(client)
		w.Write(s.Bytes())
		return
	}
	s := redisDISPLAY_SUMMARY(client, user)
	w.Write(s.Bytes())
}

func transactionHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	user := r.Form.Get("user")
	transNum, _ := strconv.Atoi(r.Form.Get("transNum"))
	LogUserCommand(server, transNum, "DISPLAY_TRANSACTION_HISTORY", user, nil, nil, nil)

	client, _ := db.Get()
	defer db.Put(client)
	s := getTransaction(client, user)
	w.Write(s.Bytes())
}
