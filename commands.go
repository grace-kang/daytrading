package main

import (
	"log"
	"math"
	"strconv"
)

var currentUsers = map[string]User{}
var stocks_amount = map[string]int{}
var stock_prices = map[string]float64{}

func parseCommandAdd(data []string) {
	amount, err := strconv.ParseFloat(data[3], 64)
	transNum, err := strconv.Atoi(data[0])
	username := data[2]
	if err != nil {
		log.Fatalf("Could not parse Amount: %s \n %s", data[3], data)
	}

	currentUsers[username] = User{username: username, balance: amount}

	logUserCommand("transNum", transNum, "command", data[1], "username", username, "amount", amount)
	logAccountTransactionCommand(transNum, "add", username, amount)
}

func parseCommandBuy(data []string) {
	transNum, err := strconv.Atoi(data[0])
	if err != nil {
		panic(err)
	}
	username := data[2]
	symbol := data[3]
	amount, err := strconv.ParseFloat(data[4], 64)
	logUserCommand("transNum", transNum, "command", data[1], "username", username, "amount", amount, "symbol", symbol)
	if _, ok := currentUsers[username]; ok {
		// do nothing
	} else {
		message := "Account" + username + " does not exist"
		logErrorEventCommand("transNum", transNum, "command", data[1], "username", username, "amount", amount, "symbol", symbol, "errorMessage", message)
		return
	}

	hasBalance := currentUsers[username].balance > amount

	if hasBalance {
		logSystemEventCommand(transNum, data[1], username, symbol, amount)
	} else {
		message := "Balance of " + username + " is not enough"
		logErrorEventCommand("transNum", transNum, "command", data[1], "username", username, "amount", amount, "symbol", symbol, "errorMessage", message)
		return
	}
	// TODO: add quote transaction xml after getting reponse from quote server. getting price from quote server. get amount of stock bought
	stockPrice := stock_prices[symbol]
	logQuoteServerCommand(transNum, stockPrice, symbol, username, getUnixTimestamp(), "cryptoKey")
	stockNumBought := int(math.Ceil(amount / stockPrice))
	stocks_amount[symbol] += stockNumBought
	updatedBalance := currentUsers[username].balance - amount
	currentUsers[username] = User{username: username, balance: updatedBalance}
	logAccountTransactionCommand(transNum, "remove", username, amount)
}

func parseCommandSell(data []string) {
	transNum, err := strconv.Atoi(data[0])
	if err != nil {
		panic(err)
	}
	username := data[2]
	symbol := data[3]
	amount, err := strconv.ParseFloat(data[4], 64)
	logUserCommand("transNum", transNum, "command", data[1], "username", username, "amount", amount, "symbol", symbol)
	if _, ok := currentUsers[username]; ok {
		// do nothing
	} else {
		message := "Account" + username + " does not exist"
		logErrorEventCommand("transNum", transNum, "command", data[1], "username", username, "amount", amount, "symbol", symbol, "errorMessage", message)
		return
	}

	logSystemEventCommand(transNum, data[1], username, symbol, amount)
	// TODO: add quote transaction xml after getting reponse from quote server
	stockPrice := stock_prices[symbol]
	logQuoteServerCommand(transNum, stockPrice, symbol, username, getUnixTimestamp(), "cryptoKey")
	amountSell := int(math.Ceil(amount / stockPrice))

	// TODO: check if the amount of stocks user hold is smaller than amount. if yes, call logErrorEventCommand and exit the function
	if amountSell > stocks_amount[symbol] {
		message := "Account" + username + " does not have enough stock amount for " + symbol
		logErrorEventCommand("transNum", transNum, "command", data[1], "username", username, "amount", amount, "symbol", symbol, "errorMessage", message)
		return
	} else {
		logAccountTransactionCommand(transNum, "add", username, amount)
	}

}

func ParseCommandData(data []string) {
	if len(data) < 2 {
		log.Fatal("invalid command")
	}

	switch cmdName := data[1]; cmdName {
	case "ADD":
		parseCommandAdd(data)
	case "BUY":
		parseCommandBuy(data)
	case "SELL":
		parseCommandSell(data)
	default:
		log.Fatalf("Invalid command: %s", data[1])
	}
}

func getStockPrice(stockSymbol string) float64 {
	return stock_prices[stockSymbol]
}
