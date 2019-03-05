package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/mediocregopher/radix.v2/pool"
)

const (
	connHost = "localhost"
	connPort = "1300"
	connType = "http"
)

var display bool
var db *pool.Pool
var transNum = 10000
var server = "server1"

func init() {
	var err error
	// Establish a pool of 10 connections to the Redis server listening on
	// port 6379 of the local machine.
	db, err = pool.New("tcp", "redis:6379", 10)
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

	client := dialRedis()
	flushRedis(client)

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
	http.HandleFunc("/dumpLog", dumpLogHandler)

	err := http.ListenAndServe(":1300", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func quote(transNumber int, username string, stock string) {
	stringQ := stock + ":QUOTE"
	client, _ := db.Get()
	defer db.Put(client)
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
		quoteTimestamp := strings.TrimSpace(split[3])
		crytpoKey := split[4]

		quoteServerTime := ParseUint(quoteTimestamp, 10, 64)
		LogQuoteServerCommand(server, transNumber, strings.TrimSpace(split[0]), stock, username, quoteServerTime, crytpoKey)

		stringQ := stock + ":QUOTE"
		client.Cmd("HSET", stringQ, stringQ, price)
	} else {
		stringQ := stock + ":QUOTE"
		currentprice, _ := client.Cmd("HGET", stringQ, stringQ).Float64()
		LogSystemEventCommand(server, transNumber, "QUOTE", username, fmt.Sprintf("%f", currentprice), stock, nil)
	}
}

func checkUserExists(transNum int, username string, command string) {
	client, _ := db.Get()
	defer db.Put(client)
	exists := exists(client, username)
	if exists == false {
		message := "Account" + username + " does not exist"
		LogErrorEventCommand(server, transNum, command, username, nil, nil, nil, message)
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
		go LogUserCommand(server, transNum, "DUMPLOG", username, nil, nil, filename)
		go DumpLog(filename, username)
	} else {
		go LogSystemEventCommand(server, transNum, "DUMPLOG", nil, nil, nil, filename)
		DumpLog(filename, nil)
	}
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db.Get()
	defer db.Put(client)
	user := r.Form.Get("user")
	amount, _ := strconv.ParseFloat(r.Form.Get("amount"), 64)

	if display == false {
		redisADD(client, user, amount)
	} else {
		displayADD(client, user, amount)
	}

	LogUserCommand(server, transNum, "ADD", user, r.Form.Get("amount"), nil, nil)
	checkUserExists(transNum, user, "ADD")
	LogAccountTransactionCommand(server, transNum, "add", user, r.Form.Get("amount"))
	//w.Write([]byte("ADD complete"))
}

func buyHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db.Get()
	defer db.Put(client)
	user := r.Form.Get("user")
	symbol := r.Form.Get("symbol")
	amount, _ := strconv.ParseFloat(r.Form.Get("amount"), 64)

	LogUserCommand(server, transNum, "BUY", user, r.Form.Get("amount"), symbol, nil)
	checkUserExists(transNum, user, "BUY")

	if display == false {
		redisBUY(client, user, symbol, amount)
	} else {
		displayBUY(client, user, symbol, amount)
	}

	//w.Write([]byte("BUY complete"))
}

func sellHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db.Get()
	defer db.Put(client)
	user := r.Form.Get("user")
	symbol := r.Form.Get("symbol")
	amount, _ := strconv.ParseFloat(r.Form.Get("amount"), 64)

	LogUserCommand(server, transNum, "SELL", user, r.Form.Get("amount"), symbol, nil)
	/*check if user exists or not*/
	checkUserExists(transNum, user, "SELL")

	if display == false {
		redisSELL(client, user, symbol, amount)
	} else {
		displaySELL(client, user, symbol, amount)
	}

	//w.Write([]byte("SELL complete"))
}

func quoteHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db.Get()
	defer db.Put(client)
	transNum, _ := strconv.Atoi(r.Form.Get("transNum"))
	user := r.Form.Get("user")
	symbol := r.Form.Get("symbol")

	quote(transNum, user, symbol)
	if display == true {
		displayQUOTE(client, user, symbol)
	}

	//w.Write([]byte("QUOTE complete"))
}

func commitBuyHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db.Get()
	defer db.Put(client)
	user := r.Form.Get("user")

	LogUserCommand(server, transNum, "COMMIT_BUY", user, nil, nil, nil)
	checkUserExists(transNum, user, "COMMIT_BUY")

	if display == false {
		redisCOMMIT_BUY(client, user)
	} else {
		displayCOMMIT_BUY(client, user)
	}

	//w.Write([]byte("COMMIT BUY complete"))
}

func commitSellHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db.Get()
	defer db.Put(client)
	user := r.Form.Get("user")

	LogUserCommand(server, transNum, "COMMIT_SELL", user, nil, nil, nil)

	if display == false {
		redisCOMMIT_SELL(client, user)
	} else {
		displayCOMMIT_SELL(client, user)
	}

	//w.Write([]byte("COMMIT SELL complete"))
}

func cancelBuyHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db.Get()
	defer db.Put(client)
	user := r.Form.Get("user")

	LogUserCommand(server, transNum, "CANCEL_BUY", user, nil, nil, nil)

	if display == false {
		redisCANCEL_BUY(client, user)
	} else {
		displayCANCEL_BUY(client, user)
	}

	//w.Write([]byte("CANCEL BUY complete"))
}

func cancelSellHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db.Get()
	defer db.Put(client)
	user := r.Form.Get("user")

	LogUserCommand(server, transNum, "CANCEL_SELL", user, nil, nil, nil)

	if display == false {
		redisCANCEL_SELL(client, user)
	} else {
		displayCANCEL_SELL(client, user)
	}

	//w.Write([]byte("CANCEL SELL complete"))
}

func setBuyAmountHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db.Get()
	defer db.Put(client)
	user := r.Form.Get("user")
	symbol := r.Form.Get("symbol")
	amount, _ := strconv.ParseFloat(r.Form.Get("amount"), 64)

	LogUserCommand(server, transNum, "SET_BUY_AMOUNT", user, r.Form.Get("amount"), symbol, nil)

	if display == false {
		redisSET_BUY_AMOUNT(client, user, symbol, amount)
	} else {
		displaySET_BUY_AMOUNT(client, user, symbol, amount)
	}

	//w.Write([]byte("SET BUY AMOUNT complete"))
}

func setBuyTriggerHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client, _ := db.Get()
	defer db.Put(client)
	user := r.Form.Get("user")
	symbol := r.Form.Get("symbol")
	amount, _ := strconv.ParseFloat(r.Form.Get("amount"), 64)

	LogUserCommand(server, transNum, "SET_BUY_TRIGGER", user, r.Form.Get("amount"), symbol, nil)

	if display == false {
		redisSET_BUY_TRIGGER(client, user, symbol, amount)
	} else {
		displaySET_BUY_TRIGGER(client, user, symbol, amount)
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
	symbol := r.Form.Get("symbol")

	LogUserCommand(server, transNum, "CANCEL_SET_BUY", user, nil, symbol, nil)

	if display == false {
		redisCANCEL_SET_BUY(client, user, symbol)
	} else {
		displayCANCEL_SET_BUY(client, user, symbol)
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
	symbol := r.Form.Get("symbol")
	amount, _ := strconv.ParseFloat(r.Form.Get("amount"), 64)

	LogUserCommand(server, transNum, "SET_SELL_AMOUNT", user, r.Form.Get("amount"), symbol, nil)

	if display == false {
		redisSET_SELL_AMOUNT(client, user, symbol, amount)
	} else {
		displaySET_SELL_AMOUNT(client, user, symbol, amount)
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
	symbol := r.Form.Get("symbol")
	amount, _ := strconv.ParseFloat(r.Form.Get("amount"), 64)

	LogUserCommand(server, transNum, "SET_SELL_TRIGGER", user, r.Form.Get("amount"), symbol, nil)

	if display == false {
		redisSET_SELL_TRIGGER(client, user, symbol, amount)
	} else {
		displaySET_SELL_TRIGGER(client, user, symbol, amount)
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
	symbol := r.Form.Get("symbol")

	LogUserCommand(server, transNum, "CANCEL_SET_SELL", user, nil, symbol, nil)

	if display == false {
		redisCANCEL_SET_SELL(client, user, symbol)
	} else {
		displayCANCEL_SET_SELL(client, user, symbol)
	}

	//w.Write([]byte("CANCEL SET SELL complete"))
}

func displaySummaryHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	user := "user1"
	LogUserCommand(server, transNum, "DISPLAY_SUMMARY", user, nil, nil, nil)

	if display == true {
		client, _ := db.Get()
		defer db.Put(client)
		user := r.Form.Get("user")
		redisDISPLAY_SUMMARY(client, user)
	}

	//w.Write([]byte("DISPLAY SUMMARY complete"))
}
