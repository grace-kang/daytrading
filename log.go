package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"time"
)

type LogItem struct {
	Username string
	LogData  string
}

var logList []LogItem

type Command string
type stockSymbolType string

var server = "server1"        // need to be replaced later
var log_file = "log_File.xml" // need to be replaced later

const (
	// A generic XML header suitable for use with the output of Marshal.
	// This is not automatically added to any output of this package,
	// it is provided as a convenience.
	Header = `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
)

const (
	ADD              = Command("ADD")
	QUOTE            = Command("QUOTE")
	BUY              = Command("BUY")
	COMMIT_BUY       = Command("COMMIT_BUY")
	CANCEL_BUY       = Command("CANCEL_BUY")
	SELL             = Command("SELL")
	COMMIT_SELL      = Command("COMMIT_SELL")
	CANCEL_SELL      = Command("CANCEL_SELL")
	SET_BUY_AMOUNT   = Command("SET_BUY_AMOUNT")
	CANCEL_SET_BUY   = Command("CANCEL_SET_BUY")
	SET_BUY_TRIGGER  = Command("SET_BUY_TRIGGER")
	SET_SELL_AMOUNT  = Command("SET_SELL_AMOUNT")
	SET_SELL_TRIGGER = Command("SET_SELL_TRIGGER")
	CANCEL_SET_SELL  = Command("CANCEL_SET_SELL")
	DUMPLOG          = Command("DUMPLOG")
	DISPLAY_SUMMARY  = Command("DISPLAY_SUMMARY")
)

type UserCommandType struct {
	XMLName           xml.Name        `xml:"userCommand"`
	Timestamp         int64           `xml:"timestamp"`
	Server            string          `xml:"server"`
	TransactionNumber int             `xml:"transactionNum"`
	Command           Command         `xml:"command"`
	Username          string          `xml:"username,omitempty"`
	StockSymbol       stockSymbolType `xml:"stockSymbol,omitempty"`
	Filename          string          `xml:"filename,omitempty"`
	Funds             string          `xml:"funds,omitempty"`
}

type QuoteServerType struct {
	XMLName           xml.Name        `xml:"quoteServer"`
	Timestamp         int64           `xml:"timestamp"`
	Server            string          `xml:"server"`
	TransactionNumber int             `xml:"transactionNum"`
	Price             string          `xml:"price"`
	StockSymbol       stockSymbolType `xml:"stockSymbol"`
	Username          string          `xml:"username"`
	QuoteServerTime   int64           `xml:"quoteServerTime"`
	CryptoKey         string          `xml:"cryptokey"`
}

type AccountTransactionType struct {
	XMLName           xml.Name `xml:"accountTransaction"`
	Timestamp         int64    `xml:"timestamp"`
	Server            string   `xml:"server"`
	TransactionNumber int      `xml:"transactionNum"`
	Action            string   `xml:"action"`
	Username          string   `xml:"username"`
	Funds             string   `xml:"funds"`
}

type SystemEventType struct {
	XMLName           xml.Name        `xml:"systemEvent"`
	Timestamp         int64           `xml:"timestamp"`
	Server            string          `xml:"server"`
	TransactionNumber int             `xml:"transactionNum"`
	Command           Command         `xml:"command"`
	Username          string          `xml:"username"`
	StockSymbol       stockSymbolType `xml:"stockSymbol"`
	Funds             string          `xml:"funds"`
}

type ErrorEventType struct {
	XMLName           xml.Name        `xml:"errorEvent"`
	Timestamp         int64           `xml:"timestamp"`
	Server            string          `xml:"server"`
	TransactionNumber int             `xml:"transactionNum"`
	Command           Command         `xml:"command"`
	Username          string          `xml:"username,omitempty"`
	StockSymbol       stockSymbolType `xml:"stockSymbol,omitempty"`
	Funds             string          `xml:"funds,omitempty"`
	ErrorMessage      string          `xml:"errorMessage,omitempty"`
}

type DebugType struct {
	XMLName           xml.Name        `xml:"debugEvent"`
	Timestamp         int64           `xml:"timestamp"`
	Server            string          `xml:"server"`
	TransactionNumber int             `xml:"transactionNum"`
	Command           Command         `xml:"command"`
	Username          string          `xml:"username,omitempty"`
	StockSymbol       stockSymbolType `xml:"stockSymbol,omitempty"`
	Funds             string          `xml:"funds,omitempty"`
	DebugMessage      string          `xml:"errorMessage,omitempty"`
}

func getUnixTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func GetKwds(kwds []interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for i := 0; i < len(kwds); i += 2 {
		result[kwds[i].(string)] = kwds[i+1]
	}

	return result
}

func logUserCommand(kwds ...interface{}) {
	time := getUnixTimestamp()
	args := GetKwds(kwds)
	userCommandData := &UserCommandType{Timestamp: time, Server: server}
	if value, ok := args["transNum"]; ok {
		userCommandData.TransactionNumber = value.(int)
	}
	if value, ok := args["command"]; ok {
		userCommandData.Command = Command(value.(string))
	}
	if value, ok := args["username"]; ok {
		userCommandData.Username = value.(string)
	}
	if value, ok := args["amount"]; ok {
		amount := value.(float64)
		amount2f := fmt.Sprintf("%.2f", amount)
		userCommandData.Funds = amount2f
	}
	if value, ok := args["symbol"]; ok {
		symbol := value.(string)
		userCommandData.StockSymbol = stockSymbolType(symbol)
	}
	if value, ok := args["fileName"]; ok {
		userCommandData.Filename = value.(string)
	}

	out, err := xml.MarshalIndent(userCommandData, "", "   ")

	if err != nil {
		panic(err)
	}

	logList = append(logList, LogItem{Username: userCommandData.Username, LogData: xml.Header + string(out)})
}

func logAccountTransactionCommand(transNum int, action string, username string, amount float64) {
	time := getUnixTimestamp()
	amountF := fmt.Sprintf("%.2f", amount)
	transCommandData := &AccountTransactionType{Timestamp: time, Server: server, TransactionNumber: transNum, Action: action, Username: username, Funds: amountF}
	out, err := xml.MarshalIndent(transCommandData, "", "   ")

	if err != nil {
		panic(err)
	}

	logList = append(logList, LogItem{Username: username, LogData: xml.Header + string(out)})

}

func logSystemEventCommand(transNum int, command string, username string, stock string, amount float64) {
	time := getUnixTimestamp()
	amountF := fmt.Sprintf("%.2f", amount)
	systemEventCommandData := &SystemEventType{Timestamp: time, Server: server, TransactionNumber: transNum, Command: Command(command), Username: username, StockSymbol: stockSymbolType(stock), Funds: amountF}
	out, err := xml.MarshalIndent(systemEventCommandData, "", "   ")

	if err != nil {
		panic(err)
	}

	logList = append(logList, LogItem{Username: username, LogData: xml.Header + string(out)})

}

func logQuoteServerCommand(transNum int, price string, stock string, username string, quoteServerTime int64, cryptoKey string) {
	time := getUnixTimestamp()
	quoteEventCommandData := &QuoteServerType{Timestamp: time, Server: server, TransactionNumber: transNum, Price: price, StockSymbol: stockSymbolType(stock), Username: username, QuoteServerTime: quoteServerTime, CryptoKey: cryptoKey}
	out, err := xml.MarshalIndent(quoteEventCommandData, "", "   ")

	if err != nil {
		panic(err)
	}

	logList = append(logList, LogItem{Username: username, LogData: xml.Header + string(out)})

}

func logErrorEventCommand(kwds ...interface{}) {
	time := getUnixTimestamp()
	args := GetKwds(kwds)
	errorEvent := &ErrorEventType{Timestamp: time, Server: server}
	if value, ok := args["transNum"]; ok {
		errorEvent.TransactionNumber = value.(int)
	}
	if value, ok := args["command"]; ok {
		errorEvent.Command = Command(value.(string))
	}
	if value, ok := args["username"]; ok {
		errorEvent.Username = value.(string)
	}
	if value, ok := args["amount"]; ok {
		amount := value.(float64)
		amount2f := fmt.Sprintf("%.2f", amount)
		errorEvent.Funds = amount2f
	}
	if value, ok := args["symbol"]; ok {
		symbol := value.(string)
		errorEvent.StockSymbol = stockSymbolType(symbol)
	}
	if value, ok := args["errorMessage"]; ok {
		errorEvent.ErrorMessage = value.(string)
	}

	out, err := xml.MarshalIndent(errorEvent, "", "   ")

	if err != nil {
		panic(err)
	}

	logList = append(logList, LogItem{Username: errorEvent.Username, LogData: xml.Header + string(out)})
}

func logDebugMessageCommand(kwds ...interface{}) {
	time := getUnixTimestamp()
	args := GetKwds(kwds)
	debugMessage := &DebugType{Timestamp: time, Server: server}
	if value, ok := args["transNum"]; ok {
		debugMessage.TransactionNumber = value.(int)
	}
	if value, ok := args["command"]; ok {
		debugMessage.Command = Command(value.(string))
	}
	if value, ok := args["username"]; ok {
		debugMessage.Username = value.(string)
	}
	if value, ok := args["amount"]; ok {
		amount := value.(float64)
		amount2f := fmt.Sprintf("%.2f", amount)
		debugMessage.Funds = amount2f
	}
	if value, ok := args["symbol"]; ok {
		symbol := value.(string)
		debugMessage.StockSymbol = stockSymbolType(symbol)
	}
	if value, ok := args["debugMessage"]; ok {
		debugMessage.DebugMessage = value.(string)
	}

	out, err := xml.MarshalIndent(debugMessage, "", "   ")

	if err != nil {
		panic(err)
	}

	logList = append(logList, LogItem{Username: debugMessage.Username, LogData: xml.Header + string(out)})
}

func dumpLog(username string) {

	var logS = ""
	for _, s := range logList {
		if s.Username == username {
			logS = logS + s.LogData + "\n"
		}
	}

	f, err := os.OpenFile(log_file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	_, err = f.Write([]byte(logS))
	if err != nil {
		log.Fatal(err)
	}

	f.Close()
}

func deleteFile() {
	// delete file
	var err = os.Remove(log_file)
	if isError(err) {
		return
	}

	fmt.Println("File Deleted")
}

func isError(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
	}

	return (err != nil)
}
