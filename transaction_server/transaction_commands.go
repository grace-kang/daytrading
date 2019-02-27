package main

import (
  "fmt"
  "math"
  "net"
  "strconv"
  "strings"
  "bytes"

  "github.com/mediocregopher/radix.v2/redis"
)

var stockPrices = map[string]float64{}
var stocksAmount = map[string]int{}

func add(transNum int, username string, amount float64, client *redis.Client) {
  logUserCommand("transNum", transNum, "command", "ADD", "username", username, "amount", amount)
  logAccountTransactionCommand(transNum, "add", username, amount)
}

func quote(transNum int, username string, stock string, client *redis.Client) {
  stringQ := stock + ":QUOTE"
  ex := exists(client, stringQ)
  if ex == false {
    conn, _ := net.Dial("tcp", "quote:1200")
    conn.Write([]byte((stock+","+username+"\n")))
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
    quoteTimestamp := strings.TrimSpace(split[3])
    crytpoKey := split[4]

    logQuoteServerCommand(transNum, price, stock, username, quoteTimestamp, crytpoKey)

    stringQ := stock + ":QUOTE"
    client.Cmd("HSET", stringQ, stringQ, price)
  } else {
    stringQ := stock + ":QUOTE"
    currentprice, _ := client.Cmd("HGET", stringQ, stringQ).Float64()
    logSystemEventCommand(transNum, "QUOTE", username, stock, currentprice)
  }
}

func buy(transNum int, username string, symbol string, amount float64, client *redis.Client) {
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

  logAccountTransactionCommand(transNum, "remove", username, final)
}

func commit_sell(transNum int, username string, client *redis.Client) {

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

  logAccountTransactionCommand(transNum, "add", username, finalCost)
}

func cancel_buy(transNum int, username string, client *redis.Client) {
  logUserCommand("transNum", transNum, "command", "CANCEL_BUY", "username", username)
}

func cancel_sell(transNum int, username string, client *redis.Client) {
  logUserCommand("transNum", transNum, "command", "CANCEL_SELL", "username", username)
}

func set_buy_amount(transNum int, username string, symbol string, amount float64, client *redis.Client) {

  logUserCommand("transNum", transNum, "command", "SET_BUY_AMOUNT", "username", username, "symbol", symbol, "amount", amount)
}

func set_buy_trigger(transNum int, username string, symbol string, amount float64, client *redis.Client) {

  logUserCommand("transNum", transNum, "command", "SET_BUY_TRIGGER", "username", username, "symbol", symbol, "amount", amount)

}

func cancel_set_buy(transNum int, username string, symbol string, client *redis.Client) {

  logUserCommand("transNum", transNum, "command", "CANCEL_SET_BUY", "username", username, "symbol", symbol)
}

func set_sell_amount(transNum int, username string, symbol string, amount float64, client *redis.Client) {

  logUserCommand("transNum", transNum, "command", "SET_SELL_AMOUNT", "username", username, "symbol", symbol, "amount", amount)

}

func set_sell_trigger(transNum int, username string, symbol string, amount float64, client *redis.Client) {

  logUserCommand("transNum", transNum, "command", "SET_SELL_TRIGGER", "username", username, "symbol", symbol, "amount", amount)
}

func cancel_set_sell(transNum int, username string, symbol string, client *redis.Client) {

  logUserCommand("transNum", transNum, "command", "CANCEL_SET_SELL", "username", username, "symbol", symbol)
}

func dumplog(transNum int, params ...string) {
  fmt.Println("-----DUMPLOG-----")
  if len(params) == 1 {
    filename := params[0]
    logUserCommand("transNum", transNum, "command", "DUMPLOG", "filename", filename)
    dumpAllLogs(transNum, filename)
  } else if len(params) == 2 {
    username := params[0]
    filename := params[1]
    logUserCommand("transNum", transNum, "command", "DUMPLOG", "username", username, "filename", filename)
    dumpLog(transNum, username, filename)
  }
}

func display_summary(transNum int, username string, client *redis.Client) {
  fmt.Println("-----DISPLAY_SUMMARY-----")
  /* TODO: Not implemented yet, Display User's transaction history */
  redisDISPLAY_SUMMARY(client, username)
}
