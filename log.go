package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/mediocregopher/radix.v2/redis"
)

type Command string
type stockSymbolType string

var server = "server1" // need to be replaced later
var client *redis.Client

var localLog Log
var localUserLogs = map[string]Log{}

const (
	// A generic XML header suitable for use with the output of Marshal.
	// This is not automatically added to any output of this package,
	// it is provided as a convenience.
	Header = `<?xml version="1.0"?>` + "\n"
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

// type RESPONSE struct {
// 	compactData
// }

type Log struct {
	LogData []LogType
}

func (cd *LogType) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if err := e.EncodeElement(cd.UserCommand, xml.StartElement{Name: xml.Name{Local: "userCommand"}}); err != nil {
		return err
	}
	if err := e.EncodeElement(cd.QuoteServer, xml.StartElement{Name: xml.Name{Local: "quoteServer"}}); err != nil {
		return err
	}
	if err := e.EncodeElement(cd.AccountTransaction, xml.StartElement{Name: xml.Name{Local: "accountTransaction"}}); err != nil {
		return err
	}
	if err := e.EncodeElement(cd.SystemEvent, xml.StartElement{Name: xml.Name{Local: "systemEvent"}}); err != nil {
		return err
	}
	if err := e.EncodeElement(cd.ErrorEvent, xml.StartElement{Name: xml.Name{Local: "errorEvent"}}); err != nil {
		return err
	}
	if err := e.EncodeElement(cd.DebugEvent, xml.StartElement{Name: xml.Name{Local: "debugEvent"}}); err != nil {
		return err
	}
	return nil
}

type LogType struct {
	UserCommand        *UserCommandType        `xml:"userCommand"`
	QuoteServer        *QuoteServerType        `xml:"quoteServer"`
	AccountTransaction *AccountTransactionType `xml:"accountTransaction"`
	SystemEvent        *SystemEventType        `xml:"systemEvent"`
	ErrorEvent         *ErrorEventType         `xml:"errorEvent"`
	DebugEvent         *DebugType              `xml:"debugEvent"`
}

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

func initAuditServer() {
	client = dialRedis()
	localLog.LogData = make([]LogType, 0)
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

	newUserCommandLog := LogType{UserCommand: userCommandData}
	localLog.LogData = append(localLog.LogData, newUserCommandLog)
	// if val, ok := localUserLogs[userCommandData.Username]; ok {
	// 	//do nothing
	// } else {
	// 	logData := make([]LogType, 0)
	// 	localUserLogs[userCommandData.Username] = Log{LogData: logData}
	// }

	// localUserLogs[userCommandData.Username].LogData = append(localUserLogs[userCommandData.Username], newUserCommandLog)
}

func logAccountTransactionCommand(transNum int, action string, username string, amount float64) {
	time := getUnixTimestamp()
	amountF := fmt.Sprintf("%.2f", amount)
	transCommandData := &AccountTransactionType{Timestamp: time, Server: server, TransactionNumber: transNum, Action: action, Username: username, Funds: amountF}

	newTransCommandLog := LogType{AccountTransaction: transCommandData}
	localLog.LogData = append(localLog.LogData, newTransCommandLog)
}

func logSystemEventCommand(transNum int, command string, username string, stock string, amount float64) {
	time := getUnixTimestamp()
	amountF := fmt.Sprintf("%.2f", amount)
	systemEventCommandData := &SystemEventType{Timestamp: time, Server: server, TransactionNumber: transNum, Command: Command(command), Username: username, StockSymbol: stockSymbolType(stock), Funds: amountF}
	newSystemCommandLog := LogType{SystemEvent: systemEventCommandData}
	localLog.LogData = append(localLog.LogData, newSystemCommandLog)

}

func logQuoteServerCommand(transNum int, price float64, stock string, username string, quoteServerTime string, cryptoKey string) {
	time := getUnixTimestamp()
	stockPrice2f := fmt.Sprintf("%.2f", price)
	quoteTime, _ := strconv.ParseInt(quoteServerTime, 10, 64)
	quoteEventCommandData := &QuoteServerType{Timestamp: time, Server: server, TransactionNumber: transNum, Price: stockPrice2f, StockSymbol: stockSymbolType(stock), Username: username, QuoteServerTime: quoteTime, CryptoKey: cryptoKey}

	quoteServerCommandLog := LogType{QuoteServer: quoteEventCommandData}
	localLog.LogData = append(localLog.LogData, quoteServerCommandLog)
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
	errorEventCommandLog := LogType{ErrorEvent: errorEvent}
	localLog.LogData = append(localLog.LogData, errorEventCommandLog)
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

	debugEventCommandLog := LogType{DebugEvent: debugMessage}
	localLog.LogData = append(localLog.LogData, debugEventCommandLog)
}

func dumpLog(username string, filename string) {
	fmt.Println("in dumpAllLogs")
	out, err := xml.MarshalIndent(localLog, "", "   ")

	if err != nil {
		panic(err)
	}

	var logS = Header
	logS += string(out)

	f, err := os.OpenFile(filename+".xml", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	_, err = f.Write([]byte(logS))
	if err != nil {
		log.Fatal(err)
	}

	f.Close()
}

func dumpAllLogs(filename string) {

	fmt.Println("in dumpAllLogs")
	out, err := xml.MarshalIndent(localLog, "", "   ")

	if err != nil {
		panic(err)
	}

	var logS = Header
	logS += string(out)

	deleteFile(filename)

	f, err := os.OpenFile(filename+".xml", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	_, err = f.Write([]byte(logS))
	if err != nil {
		log.Fatal(err)
	}

	f.Close()
}

func deleteFile(logFileName string) {
	// delete file
	if _, err := os.Stat(logFileName + ".xml"); os.IsNotExist(err) {
		return
	}
	var err = os.Remove(logFileName + ".xml")
	if isError(err) {
		return
	}

	// fmt.Println("File Deleted")
}

func isError(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
	}

	return (err != nil)
}

func addCommandLogs(username string, log string) {
	cmd := "LOGS:" + username
	client.Cmd("RPUSH", cmd, log)

	cmdAll := "ALLLOGS"
	client.Cmd("RPUSH", cmdAll, log)
}

func getCommandLogs(params ...string) []string {
	if (len(params)) == 1 {
		cmd := "ALLLOGS"
		logs, _ := client.Cmd("LRANGE", cmd, 0, -1).List()
		return logs
	}
	if (len(params)) == 2 {
		cmd := "LOGS:" + params[0]
		logs, _ := client.Cmd("LRANGE", cmd, 0, -1).List()
		return logs
	}
	return nil

}
