package main

import (
	"encoding/xml"
	"log"
	"os"
	"strconv"
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
	XMLName           xml.Name `xml:"userCommand"`
	Timestamp         int64    `xml:"timestamp"`
	Server            string   `xml:"server"`
	TransactionNumber int64    `xml:"transactionNum"`
	Command           Command  `xml:"command"`
	Username          string   `xml:"username,omitempty"`
	StockSymbol       string   `xml:"stockSymbol,omitempty"`
	Filename          string   `xml:"filename,omitempty"`
	Funds             string   `xml:"funds,omitempty"`
}

type QuoteServerType struct {
	XMLName           xml.Name        `xml:"quoteServer"`
	Timestamp         int64           `xml:"timestamp"`
	Server            string          `xml:"server"`
	TransactionNumber int64           `xml:"transactionNum"`
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
	TransactionNumber int64    `xml:"transactionNum"`
	Action            string   `xml:"action"`
	Username          string   `xml:"username"`
	Funds             string   `xml:"funds"`
}

type SystemEventType struct {
	XMLName           xml.Name        `xml:"systemEvent"`
	Timestamp         int64           `xml:"timestamp"`
	Server            string          `xml:"server"`
	TransactionNumber int64           `xml:"transactionNum"`
	Command           Command         `xml:"command"`
	Username          string          `xml:"username"`
	StockSymbol       stockSymbolType `xml:"stockSymbol"`
	Funds             string          `xml:"funds"`
}

type ErrorEventType struct {
	XMLName           xml.Name        `xml:"errorEvent"`
	Timestamp         int64           `xml:"timestamp"`
	Server            string          `xml:"server"`
	TransactionNumber int64           `xml:"transactionNum"`
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
	TransactionNumber int64           `xml:"transactionNum"`
	Command           Command         `xml:"command"`
	Username          string          `xml:"username,omitempty"`
	StockSymbol       stockSymbolType `xml:"stockSymbol,omitempty"`
	Funds             string          `xml:"funds,omitempty"`
	DebugMessage      string          `xml:"errorMessage,omitempty"`
}

func getUnixTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func logUserCommand(user User) {
	time := getUnixTimestamp()
	userCommandData := &UserCommandType{Timestamp: time, Server: server, Command: ADD, Username: user.username, Funds: strconv.Itoa(user.balance)}
	out, err := xml.MarshalIndent(userCommandData, "", "   ")

	if err != nil {
		panic(err)
	}

	logList = append(logList, LogItem{Username: user.username, LogData: xml.Header + string(out)})

}

func dumpLog(user User) {

	var logS = ""
	for _, s := range logList {
		if s.Username == user.username {
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
