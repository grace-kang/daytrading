package main

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/mediocregopher/radix.v2/redis"
)

/* Logger: log obj from loggingService.LoggingService */
var logger Logger
var mutex = &sync.Mutex{}

/* Server: server name for transaction server
Address: address for audit server*/
const (
	server  = "trans1"
	address = "http://audit:1400"
)

func init() {
	logger = Logger{Address: address}
}

func FloatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}

func IntToString(input_num int) string {
	// to convert a float number to a string
	return strconv.Itoa(input_num)
}

func ParseUint(s string, base int, bitSize int) uint64 {
	unit_, _ := strconv.ParseUint(s, base, bitSize)
	return unit_
}

func add(transNum int, username string, amount string, client *redis.Client) {

	logger.LogUserCommand(server, transNum, "ADD", username, amount, nil, nil)
	logger.LogAccountTransactionCommand(server, transNum, "add", username, amount)
}

func quote(transNum int, username string, stock string, client *redis.Client) {
	stringQ := stock + ":QUOTE"
	ex := exists(client, stringQ)
	if ex == false {
		conn, _ := net.Dial("tcp", "quote:1200")
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
		priceStr := strings.Replace(strings.TrimSpace(split[0]), ".", "", 1)
		price, _ := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			return
		}
		//quoteTimestamp := strings.TrimSpace(split[3])
		crytpoKey := split[4]

		quoteServerTime := ParseUint(split[3], 10, 64)
		logger.LogQuoteServerCommand(server, transNum, strings.TrimSpace(split[0]), stock, username, quoteServerTime, crytpoKey)

		stringQ := stock + ":QUOTE"
		client.Cmd("HSET", stringQ, stringQ, price)
	} else {
		//stringQ := stock + ":QUOTE"
		//currentprice, _ := client.Cmd("HGET", stringQ, stringQ).Float64()
		//logSystemEventCommand(transNum, "QUOTE", username, stock, currentprice)
	}
}

func buy(transNum int, username string, symbol string, amount float64, client *redis.Client) {
	logger.LogUserCommand(server, transNum, "BUY", username, amount, symbol, nil)

	exists := exists(client, username)
	if exists == false {
		//message := "Account" + username + " does not exist"
		//logErrorEventCommand("transNum", transNum, "command", "BUY", "username", username, "amount", amount, "symbol", symbol, "errorMessage", message)
	}

	currentBalance, _ := client.Cmd("HGET", username, "Balance").Float64()
	hasBalance := currentBalance >= amount
	if !hasBalance {
		//message := "Balance of " + username + " is not enough"
		//logErrorEventCommand("transNum", transNum, "command", "BUY", "username", username, "amount", amount, "symbol", symbol, "errorMessage", message)
	}
	//logSystemEventCommand(transNum, "BUY", username, symbol, amount)
}

func sell(transNum int, username string, symbol string, amount float64, client *redis.Client) {
	logger.LogUserCommand(server, transNum, "SELL", username, amount, symbol, nil)
	/*check if user exists or not*/
	exists := exists(client, username)
	if exists == false {
		//message := "Account" + username + " does not exist"
		//logErrorEventCommand("transNum", transNum, "command", "SELL", "username", username, "amount", amount, "symbol", symbol, "errorMessage", message)
	}
}

func commit_buy(transNum int, username string, client *redis.Client) {

	//symbol := "S"
	/* HGET dollar amount from stock BUY action. */
	x, _ := client.Cmd("HGET", username, "S:BUY").Float64()

	logger.LogUserCommand(server, transNum, "COMMIT_BUY", username, fmt.Sprintf("%f", x), nil, nil)
}

func commit_sell(transNum int, username string, client *redis.Client) {

	//symbol := "S"
	/* HGET: get dollar amount stock SELL action */
	be, _ := client.Cmd("HGET", username, "S:SELL").Float64()

	logger.LogUserCommand(server, transNum, "COMMIT_SELL", username, fmt.Sprintf("%f", be), nil, nil)
}

func cancel_buy(transNum int, username string, client *redis.Client) {
	logger.LogUserCommand(server, transNum, "CANCEL_BUY", username, nil, nil, nil)
}

func cancel_sell(transNum int, username string, client *redis.Client) {
	logger.LogUserCommand(server, transNum, "CANCEL_SELL", username, nil, nil, nil)
}

func set_buy_amount(transNum int, username string, symbol string, amount float64, client *redis.Client) {

	logger.LogUserCommand(server, transNum, "SET_BUY_AMOUNT", username, amount, symbol, nil)
}

func set_buy_trigger(transNum int, username string, symbol string, amount float64, client *redis.Client) {

	logger.LogUserCommand(server, transNum, "SET_BUY_TRIGGER", username, amount, symbol, nil)

}

func cancel_set_buy(transNum int, username string, symbol string, client *redis.Client) {

	logger.LogUserCommand(server, transNum, "CANCEL_SET_BUY", username, nil, symbol, nil)
}

func set_sell_amount(transNum int, username string, symbol string, amount float64, client *redis.Client) {

	logger.LogUserCommand(server, transNum, "SET_SELL_AMOUNT", username, amount, symbol, nil)
}

func set_sell_trigger(transNum int, username string, symbol string, amount float64, client *redis.Client) {

	logger.LogUserCommand(server, transNum, "SET_SELL_TRIGGER", username, amount, symbol, nil)
}

func cancel_set_sell(transNum int, username string, symbol string, client *redis.Client) {

	logger.LogUserCommand(server, transNum, "CANCEL_SET_SELL", username, nil, symbol, nil)
}

func dumplog(transNum int, params ...string) {
	fmt.Println("-----DUMPLOG-----")
	if (len(params)) == 1 {
		go logger.LogSystemEventCommand(server, transNum, "DUMPLOG", nil, nil, nil, params[0])
		fmt.Println("in dumplog param 0 is " + params[0])
		logger.DumpLog(params[0], nil)
	}
	if (len(params)) == 2 {
		go logger.LogUserCommand(server, transNum, "DUMPLOG", params[0], nil, nil, params[1])
		logger.DumpLog(params[1], params[0])
	}

	fmt.Println("-----DUMPLOG-----")
}

func display_summary(transNum int, username string, client *redis.Client) {
	fmt.Println("-----DISPLAY_SUMMARY-----")
	/* TODO: Not implemented yet, Display User's transaction history */
	redisDISPLAY_SUMMARY(client, username)
	logger.LogUserCommand(server, transNum, "DISPLAY_SUMMARY", username, nil, nil, nil)
}
