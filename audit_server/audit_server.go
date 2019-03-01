package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var server = "server1" // need to be replaced later

var localLog Log

var channel = make(chan LogType, 20000)

var mutex = &sync.Mutex{} // used to lock localUserLogs map

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

func isError(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
	}

	return (err != nil)
}

func initAuditServer() {
	//client = dialRedis()
	localLog = Log{LogData: make([]LogType, 500000)}
}

func getUnixTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func fileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func deleteFile(logFilePath string) {
	// delete file
	if fileExists(logFilePath) == false {
		return
	}
	var err = os.Remove(logFilePath)
	if isError(err) {
		return
	}

	fmt.Println("File Deleted")
}

func worker() {
	for {
		// receive from channel, or be blocked
		command := <-channel
		localLog.append(command)
	}
}

func UserCommandHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	transNum, _ := strconv.Atoi(query.Get("transactionNum"))
	data := &UserCommandType{
		Timestamp:         getUnixTimestamp(),
		Server:            query.Get("server"),
		TransactionNumber: transNum,
		Command:           Command(query.Get("command")),
		Username:          query.Get("username"),
		StockSymbol:       stockSymbolType(query.Get("stockSymbol")),
		Filename:          query.Get("filename"),
		Funds:             query.Get("funds"),
	}
	channel <- LogType{UserCommand: data}

	w.Write([]byte("OK"))
}

func quoteServerHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	transNum, _ := strconv.Atoi(query.Get("transactionNum"))
	QuoteServerTime, _ := strconv.ParseInt(query.Get("quoteServerTime"), 10, 64)
	data := &QuoteServerType{
		Timestamp:         getUnixTimestamp(),
		Server:            query.Get("server"),
		TransactionNumber: transNum,
		Username:          query.Get("username"),
		StockSymbol:       stockSymbolType(query.Get("stockSymbol")),
		Price:             query.Get("price"),
		QuoteServerTime:   QuoteServerTime,
		CryptoKey:         query.Get("cryptokey"),
	}
	channel <- LogType{QuoteServer: data}

	w.Write([]byte("OK"))

}

func accountTransactionHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	transNum, _ := strconv.Atoi(query.Get("transactionNum"))
	data := &AccountTransactionType{
		Timestamp:         getUnixTimestamp(),
		Server:            query.Get("server"),
		TransactionNumber: transNum,
		Action:            query.Get("action"),
		Username:          query.Get("username"),
		Funds:             query.Get("funds"),
	}
	channel <- LogType{AccountTransaction: data}

	w.Write([]byte("OK"))

}

func systemEventHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	transNum, _ := strconv.Atoi(query.Get("transactionNum"))
	data := &SystemEventType{
		Timestamp:         getUnixTimestamp(),
		Server:            query.Get("server"),
		TransactionNumber: transNum,
		Command:           Command(query.Get("command")),
		Username:          query.Get("username"),
		StockSymbol:       stockSymbolType(query.Get("stockSymbol")),
		Filename:          query.Get("filename"),
		Funds:             query.Get("funds"),
	}
	channel <- LogType{SystemEvent: data}

	w.Write([]byte("OK"))

}

func errorEventHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	transNum, _ := strconv.Atoi(query.Get("transactionNum"))
	data := &ErrorEventType{
		Timestamp:         getUnixTimestamp(),
		Server:            query.Get("server"),
		TransactionNumber: transNum,
		Command:           Command(query.Get("command")),
		Username:          query.Get("username"),
		StockSymbol:       stockSymbolType(query.Get("stockSymbol")),
		Filename:          query.Get("filename"),
		Funds:             query.Get("funds"),
		ErrorMessage:      query.Get("errorMessage"),
	}
	channel <- LogType{ErrorEvent: data}

	w.Write([]byte("OK"))

}

func debugEventHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	transNum, _ := strconv.Atoi(query.Get("transactionNum"))
	data := &DebugType{
		Timestamp:         getUnixTimestamp(),
		Server:            query.Get("server"),
		TransactionNumber: transNum,
		Command:           Command(query.Get("command")),
		Username:          query.Get("username"),
		StockSymbol:       stockSymbolType(query.Get("stockSymbol")),
		Filename:          query.Get("filename"),
		Funds:             query.Get("funds"),
		DebugMessage:      query.Get("debugMessage"),
	}

	channel <- LogType{DebugEvent: data}

	w.Write([]byte("OK"))

}

func dumpLogHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("dumpLogHandler")
	query := r.URL.Query()
	// username := query.Get("username")
	filename := query.Get("filename")
	filePath := os.Getenv("log_dir") + "/" + filename + ".xml"
	fmt.Println("filepath is " + filePath)
	mutex.Lock()
	out, err := xml.MarshalIndent(localLog, "", "   ")

	if err != nil {
		panic(err)
	}

	var logS = Header
	logS += string(out)

	fmt.Println("log is " + string(logS))
	deleteFile(filePath)

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("cannot cretae the file, error is " + err.Error())
	}

	_, err = f.Write([]byte(logS))
	if err != nil {
		panic(err)
	}
	f.Close()
	fmt.Println("after close file in dumploghandler")
	mutex.Unlock()
}

func main() {
	mux := http.NewServeMux()
	initAuditServer()
	// mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Println(w, "Hello, %q", html.EscapeString(r.URL.Path))
	// })
	mux.HandleFunc("/userCommand", UserCommandHandler)
	mux.HandleFunc("/quoteServerCommand", quoteServerHandler)
	mux.HandleFunc("/accountTransactionCommand", accountTransactionHandler)
	mux.HandleFunc("/systemEventCommand", systemEventHandler)
	mux.HandleFunc("/errorEventCommand", errorEventHandler)
	mux.HandleFunc("/debugEventCommand", debugEventHandler)
	mux.HandleFunc("/dumpLog", dumpLogHandler)
	fmt.Printf("Audit server listening on %s:%s\n", "http://audit", "1400")
	go worker()
	if err := http.ListenAndServe(":1400", mux); err != nil {
		panic(err)
	}
}
