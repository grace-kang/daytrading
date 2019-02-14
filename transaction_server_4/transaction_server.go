package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
)

const (
	connHost = "localhost"
	connPort = "1304"
	connType = "http"
)

var display bool

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

	err := http.ListenAndServe(":1304", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client := dialRedis()
	user := r.Form.Get("user")
	amount, _ := strconv.ParseFloat(r.Form.Get("amount"), 64)

	if display == false {
		redisADD(client, user, amount)
	} else {
		displayADD(client, user, amount)
	}

	//w.Write([]byte("ADD complete"))
}

func buyHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	client := dialRedis()
	user := r.Form.Get("user")
	symbol := r.Form.Get("symbol")
	amount, _ := strconv.ParseFloat(r.Form.Get("amount"), 64)

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

	client := dialRedis()
	user := r.Form.Get("user")
	symbol := r.Form.Get("symbol")
	amount, _ := strconv.ParseFloat(r.Form.Get("amount"), 64)

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

	client := dialRedis()
	transNum, _ := strconv.Atoi(r.Form.Get("transNum"))
	user := r.Form.Get("user")
	symbol := r.Form.Get("symbol")

	quote(transNum, user, symbol, client)
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

	client := dialRedis()
	user := r.Form.Get("user")

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

	client := dialRedis()
	user := r.Form.Get("user")

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

	client := dialRedis()
	user := r.Form.Get("user")

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

	client := dialRedis()
	user := r.Form.Get("user")

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

	client := dialRedis()
	user := r.Form.Get("user")
	symbol := r.Form.Get("symbol")
	amount, _ := strconv.ParseFloat(r.Form.Get("amount"), 64)

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

	client := dialRedis()
	user := r.Form.Get("user")
	symbol := r.Form.Get("symbol")
	amount, _ := strconv.ParseFloat(r.Form.Get("amount"), 64)

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

	client := dialRedis()
	user := r.Form.Get("user")
	symbol := r.Form.Get("symbol")

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

	client := dialRedis()
	user := r.Form.Get("user")
	symbol := r.Form.Get("symbol")
	amount, _ := strconv.ParseFloat(r.Form.Get("amount"), 64)

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

	client := dialRedis()
	user := r.Form.Get("user")
	symbol := r.Form.Get("symbol")
	amount, _ := strconv.ParseFloat(r.Form.Get("amount"), 64)

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

	client := dialRedis()
	user := r.Form.Get("user")
	symbol := r.Form.Get("symbol")

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

	if display == true {
		client := dialRedis()
		user := r.Form.Get("user")
		redisDISPLAY_SUMMARY(client, user)
	}

	//w.Write([]byte("DISPLAY SUMMARY complete"))
}
