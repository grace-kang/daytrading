package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"strings"
	"strconv"
	"math"

	"github.com/mediocregopher/radix.v2/redis"
)

func add(transNum int, username string, amount float64, client *redis.Client) {
	fmt.Println("-----ADD-----")

	redisADD(client, username, amount)
	logUserCommand("transNum", transNum, "command", "ADD", "username", username, "amount", amount)
	logAccountTransactionCommand(transNum, "add", username, amount)
}

func quote(transNum int, username string, stock string, client *redis.Client) {
	fmt.Println("-----QUOTE-----")

	req, err := http.NewRequest("GET", "http://localhost:1200", nil)
	req.Header.Add("If-None-Match", `W/"wyzzy"`)

	q := req.URL.Query()
	q.Add("user", username)
	q.Add("stock", stock)
	q.Add("transNum", strconv.Itoa(transNum))
	req.URL.RawQuery = q.Encode()

	httpclient := http.Client{}

	var resp *http.Response
	for {
		resp, err = httpclient.Do(req)

		if err != nil { // trans server down? retry
			fmt.Println(err)
		} else {
			break
		}
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("Error reading body: %s", err.Error())
	}

	split := strings.Split(string(body), ",")[0]
	price, _ := strconv.ParseFloat(split, 64)
	stockPrices[stock] = price
	fmt.Println(price)
	resp.Body.Close()

	/* HINCRBYFLOAT: change a float value. Quote costs a User
	$0.50 */
	client.Cmd("HINCRBYFLOAT", username, "Balance", -0.50)

	/* Display - HGET new balance for display */
	fmt.Print("QUOTE: ", stock)
	x, _ := client.Cmd("HGET", username, "Balance").Float64()
	fmt.Println(" Balance: ", x)
}

func buy(transNum int, username string, symbol string, amount float64, client *redis.Client) {
	fmt.Println("-----BUY-----")

	redisBUY(client, username, symbol, amount)
	logUserCommand("transNum", transNum, "command", "BUY", "username", username, "amount", amount, "symbol", symbol)

	exists := exists(client, username)
	if exists == false {
		message := "Account" + username + " does not exist"
		logErrorEventCommand("transNum", transNum, "command", "BUY", "username", username, "amount", amount, "symbol", symbol, "errorMessage", message)
	}

	currentBalance, _ := client.Cmd("HGET", username, "Balance").Float64()
	hasBalance := currentBalance >= amount
	if !hasBalance {
		message := "Balance of " + username + " is not enough"
		logErrorEventCommand("transNum", transNum, "command", "BUY", "username", username, "amount", amount, "symbol", symbol, "errorMessage", message)
	}
	logSystemEventCommand(transNum, "BUY", username, symbol, amount)
}

func sell(transNum int, username string, symbol string, amount float64, client *redis.Client) {
	fmt.Println("-----SELL-----")

	redisSELL(client, username, symbol, amount)

	logUserCommand("transNum", transNum, "command", "SELL", "username", username, "amount", amount, "symbol", symbol)
	/*check if user exists or not*/
	exists := exists(client, username)
	if exists == false {
		message := "Account" + username + " does not exist"
		logErrorEventCommand("transNum", transNum, "command", "SELL", "username", username, "amount", amount, "symbol", symbol, "errorMessage", message)
	}
	/*check if cache has stock. if not, senf request to quote server*/
	if _, ok := stockPrices[symbol]; ok {
		logSystemEventCommand(transNum, "SELL", username, symbol, amount)
	} else {
		quote(transNum, username, symbol, client)
	}
	stockPrice := stockPrices[symbol]
	amountSell := int(math.Ceil(amount / stockPrice))
	fmt.Println("in buy, amount sell is ", strconv.Itoa(amountSell))
	// TODO: check if the amount of stocks user hold is smaller than amount. if yes, call logErrorEventCommand and exit the function
	if amountSell > stocksAmount[symbol] {
		message := "Account" + username + " does not have enough stock amount for " + symbol
		logErrorEventCommand("transNum", transNum, "command", "SELL", "username", username, "amount", amount, "symbol", symbol, "errorMessage", message)
	} else {
		// logAccountTransactionCommand(transNumInt, "add", username, amount)
	}
}

func commit_buy(transNum int, username string, client *redis.Client) {
	fmt.Println("-----COMMIT_BUY-----")

	symbol := "S"
	/* HGET dollar amount from stock BUY action. */
	x, _ := client.Cmd("HGET", username, "S:BUY").Float64()

	// TODO: need to check if last buy command is made within 60 seconds. If not, log errorEvent

	logUserCommand("transNum", transNum, "command", "COMMIT_BUY", "username", username, "amount", x)

	/*check if cache has stock. if not, senf request to quote server*/
	if _, ok := stockPrices[symbol]; ok {
		logSystemEventCommand(transNum, "COMMIT_BUY", username, symbol, x)
	} else {
		quote(transNum, username, symbol, client)
	}
	stockPrice := stockPrices[symbol]
	amountBuy := int(math.Ceil(x / stockPrice))
	final := float64(amountBuy) * stockPrice

	/* Decrease balance by price */
	client.Cmd("HINCRBYFLOAT", username, "Balance", -final)
	logAccountTransactionCommand(transNum, "remove", username, final)

	/* get new balance for Display, error checking */
	y, _ := client.Cmd("HGET", username, "Balance").Float64()
	fmt.Println("COMMIT_BUY: ", final, "Balance: ", y)

	//relevant := s[2] + ":Number"

	/* HINCBRY: Increase the number of stocks a User owns
	HGET: the number for display
	Display... */
	client.Cmd("HINCRBY", username, "S:Number", amountBuy)
	a, _ := client.Cmd("HGET", username, "S:Number").Float64()
	fmt.Println("STOCK(S): ", amountBuy, "TOTAL(S): ", a)
}

func commit_sell(transNum int, username string, client *redis.Client) {
	fmt.Println("-----COMMIT_SELL-----")

	symbol := "S"
	/* HGET: get dollar amount stock SELL action */
	be, _ := client.Cmd("HGET", username, "S:SELL").Float64()

	logUserCommand("transNum", transNum, "command", "COMMIT_SELL", "username", username, "amount", be)

	if _, ok := stockPrices[symbol]; ok {
		logSystemEventCommand(transNum, "COMMIT_SELL", username, symbol, be)
	} else {
		quote(transNum, username, symbol, client)
	}
	stockPrice := stockPrices[symbol]
	amountSell := int(math.Ceil(be / stockPrice))
	finalCost := float64(amountSell) * stockPrice
	/* Calculate how many stocks User can sell */

	fmt.Println("COMMIT_SELL: ", amountSell)
	fmt.Println("AT COST: ", finalCost)

	/* HINCRBY: Decrease User's stocks and then Display # */
	client.Cmd("HINCRBY", username, "S:Number", -amountSell)
	ab, _ := client.Cmd("HGET", username, "S:Number").Float64()
	fmt.Println("STOCK(S): ", ab)

	/* HGET: Decrease User's balance and then display new balance */
	client.Cmd("HINCRBYFLOAT", username, "Balance", finalCost)
	logAccountTransactionCommand(transNum, "add", username, finalCost)
	za, _ := client.Cmd("HGET", username, "Balance").Float64()
	fmt.Println("Balance: ", za)
}

func cancel_buy(transNum int, username string, client *redis.Client) {
	fmt.Println("-----CANCEL_BUY-----")

	/* HSET: Cancel stock BUY amount
	Display new value ex. S:BUY should equal 0 now */
	client.Cmd("HSET", username, "S:BUY", 0)
	zas, _ := client.Cmd("HGET", username, "S:BUY").Float64()
	fmt.Println("BUY: ", zas)
	logUserCommand("transNum", transNum, "command", "CANCEL_BUY", "username", username)
}

func cancel_sell(transNum int, username string, client *redis.Client) {
	fmt.Println("-----CANCEL_SELL-----")

	/* HSET: Cancel stock SELL amount
	Display new value ex. S:SELL should equal 0 now */
	client.Cmd("HSET", username, "S:SELL", 0)
	logUserCommand("transNum", transNum, "command", "CANCEL_SELL", "username", username)
	zps, _ := client.Cmd("HGET", username, "S:SELL").Float64()
	fmt.Println("SELL: ", zps)
}

func set_buy_amount(transNum int, username string, symbol string, amount float64, client *redis.Client) {
	fmt.Println("-----SET_BUY_AMOUNT-----")
	cmd := symbol + ":TBUYAMOUNT"

	/* HSET: Amount of money set aside for Buy Trigger to be activated */
	client.Cmd("HSET", username, cmd, amount)
	fmt.Println("TBUYAMOUNT:  ", amount)
	logUserCommand("transNum", transNum, "command", "SET_BUY_AMOUNT", "username", username, "symbol", symbol, "amount", amount)

	/* HINCRBYFLOAT: Decrease User's Balance by amount set aside, Display */
	client.Cmd("HINCRBYFLOAT", username, "Balance", -amount)
	zazz, _ := client.Cmd("HGET", username, "Balance").Float64()
	fmt.Println("Balance: ", zazz)
}

func set_buy_trigger(transNum int, username string, symbol string, amount float64, client *redis.Client) {
	fmt.Println("-----SET_BUY_TRIGGER-----")
	cmd := symbol + ":TBUYTRIG"

	/* HSET: Set Stock price for when the Buy Trigger will be activated */
	client.Cmd("HSET", username, cmd, amount)
	logUserCommand("transNum", transNum, "command", "SET_BUY_TRIGGER", "username", username, "symbol", symbol, "amount", amount)
	fmt.Println("TBUYTRIG:  ", amount)
}

func cancel_set_buy(transNum int, username string, symbol string, client *redis.Client) {
	fmt.Println("-----CANCEL_SET_BUY-----")

	cmd := symbol + ":TBUYAMOUNT"

	/* HGET: Get amount stored in reserve in STOCK:TBUYAMOUNT */
	zzz, _ := client.Cmd("HGET", username, cmd).Float64()
	logUserCommand("transNum", transNum, "command", "CANCEL_SET_BUY", "username", username, "symbol", symbol)
	fmt.Println("Refund: ", zzz)

	/* TODO: Refund balance by reserve stored from above */
}

func set_sell_amount(transNum int, username string, symbol string, amount float64, client *redis.Client) {
	fmt.Println("-----SET_SELL_AMOUNT-----")

	cmd := symbol + ":TSELLAMOUNT"

	client.Cmd("HSET", username, cmd, amount)
	logUserCommand("transNum", transNum, "command", "SET_SELL_AMOUNT", "username", username, "symbol", symbol, "amount", amount)
	fmt.Println("TSELLAMOUNT: ", amount)
}

func set_sell_trigger(transNum int, username string, symbol string, amount float64, client *redis.Client) {
	fmt.Println("-----SET_SELL_TRIGGER-----")
	/* TODO */
	logUserCommand("transNum", transNum, "command", "SET_SELL_TRIGGER", "username", username, "symbol", symbol, "amount", amount)
}

func cancel_set_sell(transNum int, username string, symbol string, client *redis.Client) {
	fmt.Println("-----CANCEL_SET_SELL-----")
	/* TODO */
	logUserCommand("transNum", transNum, "command", "CANCEL_SET_SELL", "username", username, "symbol", symbol)
}

func dumplog(transNum int, params ...string) {
	fmt.Println("-----DUMPLOG-----")
	if len(params) == 1 {
		filename := params[0]
		logUserCommand("transNum", transNum, "command", "DUMPLOG", "filename", filename)
		dumpAllLogs(filename)
	} else if len(params) == 2 {
		username := params[0]
		filename := params[1]
		logUserCommand("transNum", transNum, "command", "DUMPLOG", "username", username, "filename", filename)
		dumpLog(username, filename)
	}
}

func display_summary(transNum int, username string) {
	fmt.Println("-----DISPLAY_SUMMARY-----")
	/* TODO: Not implemented yet, Display User's transaction history */
}
