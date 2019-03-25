package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	connHost = "localhost"
	connPort = "80"
	connType = "http"
	address  = "http://localhost:80"
	server   = "webserver"
)

var transNum = 0

func main() {

	r := http.NewServeMux()

	r.HandleFunc("/userCommands.js", SendJqueryJs)

	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/login", loginHandler)
	r.HandleFunc("/sendCommand", sendCommandHandle)
	r.HandleFunc("/workloadTransaction", workloadTransaction)

	err := http.ListenAndServe(":"+connPort, r)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
		return
	}

	fmt.Println("Web server is listening on port " + connPort)
}

func sendCommandHandle(w http.ResponseWriter, r *http.Request) {
	fmt.Println("in sendhandler")

	transNum = transNum + 1

	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	command := r.Form.Get("command")
	amountInput := r.Form.Get("amount")
	stringInput := r.Form.Get("string")
	fmt.Println("zmount is ", amountInput, "string is ", stringInput)
	v := url.Values{
		"server":         {server},
		"command":        {command},
		"transactionNum": {strconv.Itoa(transNum)},
	}

	switch command {
	case "ADD":

		resp, err := http.PostForm(address+"/add", v)
		if err != nil {
			//send error message back
			// myvar := map[string]interface{}{"Quote": quotey, "Add": addy}
			// outputHTML(w, "tmp/home.html", myvar)
		}
		fmt.Println("resp:", resp)
		resp.Body.Close()

	case "QUOTE":
		getQuote(stringInput, "user")

	case "BUY":

	case "COMMIT_BUY":

	case "CANCEL_BUY":

	case "SELL":

	case "COMMIT_SELL":

	case "CANCEL_SELL":

	case "SET_BUY_AMOUNT":

	case "CANCEL_SET_BUY":

	case "SET_BUY_TRIGGER":

	case "SET_SELL_AMOUNT":

	case "SET_SELL_TRIGGER":

	case "CANCEL_SET_SELL":

	case "DUMPLOG":

	case "DISPLAY_SUMMARY":

	}

}

func SendJqueryJs(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile("tmp/userCommands.js")
	if err != nil {
		http.Error(w, "Couldn't read file", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Write(data)
}

func outputHTML(w http.ResponseWriter, filename string, data interface{}) {
	t, err := template.ParseFiles(filename)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// quotey := getQuote("abc", "123")
	// fmt.Println(string(quotey))
	// addy := add("abc")
	// fmt.Println(string(addy))
	// myvar := map[string]interface{}{"Quote": quotey, "Add": addy}
	// outputHTML(w, "tmp/home.html", myvar)

	// quotey := getQuote("abc", "123")
	// fmt.Println(string(quotey))
	// tpl, _ := ioutil.ReadFile("tmp/home.html")
	// tplParsed, _ := template.New("test").Parse(string(tpl))
	// templateData := map[string]interface{}{"Quote": quotey}
	// tplParsed.Execute(w, templateData)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// tmpl := template.Must(template.ParseFiles("tmp/userCommands.html"))
	// if r.Method != http.MethodPost {
	// 	tmpl.Execute(w, nil)
	// 	return
	// }

	// tmpl.Execute(w, struct{ Success bool }{true})
	tpl, _ := ioutil.ReadFile("tmp/userCommands.html")
	tplParsed, _ := template.New("test").Parse(string(tpl))
	tplParsed.Execute(w, nil)
}

func getQuote(stock string, username string) string {

	//stringQ := stock + ":QUOTE"
	// fmt.Println("goQUOTE!!!!!")

	QUOTE_URL := os.Getenv("QUOTE_URL")
	// fmt.Println("quoye url is " + QUOTE_URL)
	conn, _ := net.Dial("tcp", QUOTE_URL)

	conn.Write([]byte((stock + "," + username + "\n")))
	respBuf := make([]byte, 2048)
	_, err := conn.Read(respBuf)
	conn.Close()

	if err != nil {
		return "error"
	}
	respBuf = bytes.Trim(respBuf, "\x00")
	message := bytes.NewBuffer(respBuf).String()
	message = strings.TrimSpace(message)
	return string(message)

	//return "hello"
}

// func add(username string) string {
// 	amount := "10000"
// 	//TRANSACTION_URL := os.Getenv("TRANSACTION_URL")
// 	addr := address + "/add"
// 	transNum_str := "1"
// 	resp, err := http.PostForm(addr, url.Values{
// 		"transNum": {transNum_str},
// 		"user":     {username},
// 		"amount":   {amount}})
// 	if err != nil {
// 		return "error"
// 	}
// 	fmt.Println("resp:", resp)
// 	resp.Body.Close()
//
// 	return "response"
// }
//
func workloadTransaction(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	fmt.Println("Hostname: " + hostname)

	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	transNum_str := r.Form.Get("transNum")
	command := r.Form.Get("command")
	params_str := r.Form.Get("params")
	params := strings.Split(params_str, ",")

	client := &http.Client{}

	switch command {

	case "ADD":
		username := params[0]
		amount := params[1]
		addr := address + "/add"
		form := url.Values{
			"transNum": {transNum_str},
			"user":     {username},
			"amount":   {amount}}
		req, err := http.NewRequest("POST", addr, strings.NewReader(form.Encode()))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		req.Host = "web"
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(resp)
		resp.Body.Close()

	case "BUY":
		username := params[0]
		symbol := params[1]
		amount := params[2]
		addr := address + "/buy"
		form := url.Values{
			"transNum": {transNum_str},
			"user":     {username},
			"symbol":   {symbol},
			"amount":   {amount}}
		req, err := http.NewRequest("POST", addr, strings.NewReader(form.Encode()))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		req.Host = "web"
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resp.Body.Close()

	case "SELL":
		username := params[0]
		symbol := params[1]
		amount := params[2]
		addr := address + "/sell"
		form := url.Values{
			"transNum": {transNum_str},
			"user":     {username},
			"symbol":   {symbol},
			"amount":   {amount}}
		req, err := http.NewRequest("POST", addr, strings.NewReader(form.Encode()))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		req.Host = "web"
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resp.Body.Close()

	case "QUOTE":
		username := params[0]
		symbol := params[1]
		addr := address + "/quote"
		transNum_str := strconv.Itoa(transNum)
		form := url.Values{
			"transNum": {transNum_str},
			"user":     {username},
			"symbol":   {symbol}}
		req, err := http.NewRequest("POST", addr, strings.NewReader(form.Encode()))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		req.Host = "web"
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resp.Body.Close()

	case "COMMIT_BUY":
		username := params[0]
		addr := address + "/commit_buy"
		form := url.Values{
			"transNum": {transNum_str},
			"user":     {username}}
		req, err := http.NewRequest("POST", addr, strings.NewReader(form.Encode()))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		req.Host = "web"
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resp.Body.Close()

	case "COMMIT_SELL":
		username := params[0]
		addr := address + "/commit_sell"
		form := url.Values{
			"transNum": {transNum_str},
			"user":     {username}}
		req, err := http.NewRequest("POST", addr, strings.NewReader(form.Encode()))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		req.Host = "web"
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resp.Body.Close()

	case "CANCEL_BUY":
		username := params[0]
		addr := address + "/cancel_buy"
		form := url.Values{
			"transNum": {transNum_str},
			"user":     {username}}
		req, err := http.NewRequest("POST", addr, strings.NewReader(form.Encode()))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		req.Host = "web"
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resp.Body.Close()

	case "CANCEL_SELL":
		username := params[0]
		addr := address + "/cancel_sell"
		form := url.Values{
			"transNum": {transNum_str},
			"user":     {username}}
		req, err := http.NewRequest("POST", addr, strings.NewReader(form.Encode()))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		req.Host = "web"
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resp.Body.Close()

	case "SET_BUY_AMOUNT":
		username := params[0]
		symbol := params[1]
		amount := params[2]
		addr := address + "/set_buy_amount"
		form := url.Values{
			"transNum": {transNum_str},
			"user":     {username},
			"symbol":   {symbol},
			"amount":   {amount}}
		req, err := http.NewRequest("POST", addr, strings.NewReader(form.Encode()))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		req.Host = "web"
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resp.Body.Close()

	case "SET_BUY_TRIGGER":
		username := params[0]
		symbol := params[1]
		amount := params[2]
		addr := address + "/set_buy_trigger"
		form := url.Values{
			"transNum": {transNum_str},
			"user":     {username},
			"symbol":   {symbol},
			"amount":   {amount}}
		req, err := http.NewRequest("POST", addr, strings.NewReader(form.Encode()))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		req.Host = "web"
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resp.Body.Close()

	case "CANCEL_SET_BUY":
		username := params[0]
		symbol := params[1]
		addr := address + "/cancel_set_buy"
		form := url.Values{
			"transNum": {transNum_str},
			"user":     {username},
			"symbol":   {symbol}}
		req, err := http.NewRequest("POST", addr, strings.NewReader(form.Encode()))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		req.Host = "web"
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resp.Body.Close()

	case "DISPLAY_SUMMARY":
		username := params[0]
		addr := address + "/display_summary"
		form := url.Values{
			"transNum": {transNum_str},
			"user":     {username}}
		req, err := http.NewRequest("POST", addr, strings.NewReader(form.Encode()))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		req.Host = "web"
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resp.Body.Close()

	case "SET_SELL_AMOUNT":
		username := params[0]
		symbol := params[1]
		amount := params[2]
		addr := address + "/set_sell_amount"
		form := url.Values{
			"transNum": {transNum_str},
			"user":     {username},
			"symbol":   {symbol},
			"amount":   {amount}}
		req, err := http.NewRequest("POST", addr, strings.NewReader(form.Encode()))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		req.Host = "web"
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resp.Body.Close()

	case "SET_SELL_TRIGGER":
		username := params[0]
		symbol := params[1]
		amount := params[2]
		addr := address + "/set_sell_trigger"
		form := url.Values{
			"transNum": {transNum_str},
			"user":     {username},
			"symbol":   {symbol},
			"amount":   {amount}}
		req, err := http.NewRequest("POST", addr, strings.NewReader(form.Encode()))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		req.Host = "web"
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resp.Body.Close()

	case "CANCEL_SET_SELL":
		username := params[0]
		symbol := params[1]
		addr := address + "/cancel_set_sell"
		form := url.Values{
			"transNum": {transNum_str},
			"user":     {username},
			"symbol":   {symbol}}
		req, err := http.NewRequest("POST", addr, strings.NewReader(form.Encode()))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		req.Host = "web"
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resp.Body.Close()
	}
}
