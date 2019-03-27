package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	connHost = "localhost"
	connPort = "1600"
	connType = "http"
	address  = "http://reverse-proxy:80"
	server   = "webserver"
)

var wg sync.WaitGroup
var transNum = 0
var currentUser User

type Stock struct {
	Symbol string `json:"symbol"`
	Num    int64  `json:"num"`
}

type User struct {
	Username string  `json:"username"` // username
	Balance  float64 `json:"balance"`  // balance
	Stocks   []Stock `json:"stocks"`   // stocks owned
}

func set(w http.ResponseWriter, r *http.Request) {
	fmt.Println("in set fun")
	fm := []byte("This is a flashed message!")
	SetFlash(w, "message", fm)
}

func get(w http.ResponseWriter, r *http.Request) {
	fmt.Println("in get fun")
	fm, err := GetFlash(w, r, "message")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if fm == nil {
		fmt.Fprint(w, "No flash messages")
		return
	}
	fmt.Fprintf(w, "%s", fm)
}

func main() {
	r := http.NewServeMux()
	fs := http.FileServer(http.Dir("tmp"))
	r.Handle("/tmp/", http.StripPrefix("/tmp/", fs))
	r.HandleFunc("/tmp", serveTemplate)
	// r.PathPrefix("/tmp/").Handler(http.StripPrefix("/tmp/", http.FileServer(http.Dir("./tmp"))))
	r.HandleFunc("/userCommands.js", SendJqueryJs)
	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/login", loginHandler)
	r.HandleFunc("/sendCommand", sendCommandHandle)
	r.HandleFunc("/runWorkload/{file}", runWorkload)

	err := http.ListenAndServe(":"+connPort, r)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
		return
	}

	fmt.Println("Web server is listening on port " + connPort)
}

func serveTemplate(w http.ResponseWriter, r *http.Request) {
	lp := filepath.Join("tmp", "layout.html")
	fp := filepath.Join("tmp", filepath.Clean(r.URL.Path))

	// Return a 404 if the template doesn't exist
	info, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
	}

	// Return a 404 if the request is for a directory
	if info.IsDir() {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFiles(lp, fp)
	if err != nil {
		// Log the detailed error
		log.Println(err.Error())
		// Return a generic "Internal Server Error" message
		http.Error(w, http.StatusText(500), 500)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout", nil); err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

func sendCommandHandle(w http.ResponseWriter, r *http.Request) {
	fmt.Println("in sendhandler")
	transNum = transNum + 1
	username := currentUser.Username

	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	command := r.Form.Get("command")
	amountInput := r.Form.Get("amount")
	stringInput := r.Form.Get("string")
	client := &http.Client{}
	fmt.Println("zmount is ", amountInput, "string is ", stringInput)
	v := url.Values{
		"server":         {server},
		"command":        {command},
		"user":           {username},
		"transactionNum": {strconv.Itoa(transNum)},
	}

	switch command {
	case "ADD":
		amount := amountInput
		fmt.Println(amount)
		addr := address + "/add"
		v.Add("amount", amount)
		req, err := http.NewRequest("POST", addr, strings.NewReader(v.Encode()))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		req.Host = "transaction"
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resp.Body.Close()

		// write the response
		w.Write([]byte("added balance " + amount + " successfully\n"))

	case "QUOTE":

		// var cookie, err = r.Cookie("Symbol")
		// fmt.Println("coolie is ", cookie)

		response := getQuote(strings.ToLower(stringInput), username)
		fmt.Println("response is ", response)
		split := strings.Split(response, ",")
		expire := time.Now().Add(1 * time.Minute)
		stockCookie := http.Cookie{Name: "Symbol:" + stringInput, Value: split[0], Path: "/login", Expires: expire, MaxAge: 90000}
		http.SetCookie(w, &stockCookie)
		stockCookie = http.Cookie{Name: "successMessage", Value: "quote price is " + split[0]}
		http.SetCookie(w, &stockCookie)
		w.Write([]byte("get quote price of stock " + stringInput + " : " + split[0] + " \n"))

	case "BUY":
		amount := amountInput
		symbol := stringInput
		addr := address + "/buy"
		v.Add("symbol", symbol)
		v.Add("amount", amount)
		req, err := http.NewRequest("POST", addr, strings.NewReader(v.Encode()))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		req.Host = "transaction"
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resp.Body.Close()
		// write the response
		w.Write([]byte("buy amount " + amount + " of stock " + symbol + " successfully\n"))

	case "COMMIT_BUY":

		addr := address + "/commit_buy"
		req, err := http.NewRequest("POST", addr, strings.NewReader(v.Encode()))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		req.Host = "transaction"
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resp.Body.Close()
		// write the response
		w.Write([]byte("commit buy successfully\n"))

	case "CANCEL_BUY":

		addr := address + "/cancel_buy"
		req, err := http.NewRequest("POST", addr, strings.NewReader(v.Encode()))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		req.Host = "transaction"
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resp.Body.Close()
		// write the response
		w.Write([]byte("cancel buy successfully\n"))

	case "SELL":

		amount := amountInput
		symbol := stringInput
		addr := address + "/sell"
		v.Add("symbol", symbol)
		v.Add("amount", amount)
		req, err := http.NewRequest("POST", addr, strings.NewReader(v.Encode()))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		req.Host = "transaction"
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resp.Body.Close()
		// write the response
		w.Write([]byte("buy amount " + amount + " of stock " + symbol + " successfully\n"))

	case "COMMIT_SELL":

		addr := address + "/commit_sell"
		req, err := http.NewRequest("POST", addr, strings.NewReader(v.Encode()))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		req.Host = "transaction"
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resp.Body.Close()
		// write the response
		w.Write([]byte("commit buy successfully\n"))

	case "CANCEL_SELL":

		addr := address + "/cancel_sell"
		req, err := http.NewRequest("POST", addr, strings.NewReader(v.Encode()))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		req.Host = "transaction"
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resp.Body.Close()
		// write the response
		w.Write([]byte("cancel buy successfully\n"))

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
	//
	// quotey := getQuote("abc", "123")
	// fmt.Println(string(quotey))
	tpl, _ := ioutil.ReadFile("tmp/home.html")
	tplParsed, _ := template.New("test").Parse(string(tpl))
	// templateData := map[string]interface{}{"Quote": quotey}
	tplParsed.Execute(w, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// tmpl := template.Must(template.ParseFiles("tmp/userCommands.html"))
	// if r.Method != http.MethodPost {
	// 	tmpl.Execute(w, nil)
	// 	return
	// }

	// tmpl.Execute(w, struct{ Success bool }{true})
	c, err := r.Cookie("username")
	if err != nil {
		switch err {
		case http.ErrNoCookie:
			fmt.Println("ttp.ErrNoCookie")
			// return nil, nil
		default:
			fmt.Println("default error")
			// return nil, err
		}
	}

	value := c.Value
	if err != nil {
		// return nil, err
		fmt.Println("derror in decoding")
	}
	fmt.Println("cookir username is ", value, "c.value is ", c.Value)

	currentUser = User{Username: value}
	// fmt.Println("current user is ", currentUser)

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

func add(username string) string {
	amount := "10000"
	//TRANSACTION_URL := os.Getenv("TRANSACTION_URL")
	addr := address + "/add"
	transNum_str := "1"
	resp, err := http.PostForm(addr, url.Values{
		"transNum": {transNum_str},
		"user":     {username},
		"amount":   {amount}})
	if err != nil {
		return "error"
	}
	fmt.Println("resp:", resp)
	resp.Body.Close()

	return "response"
}

func concurrencyLogic(address string, lines []string, username string) {
	defer wg.Done()
	// httpclient := http.Client{}
	// client := dialRedis()
	for i, line := range lines {
		s := strings.Split(line, ",")
		x := strings.Split(s[0], " ")
		command := x[1]

		for ij := 0; ij < len(s); ij++ {
			s[ij] = strings.TrimSpace(s[ij])
		}

		data := make([]string, 2)

		data[1] = strings.TrimSpace(x[1])
		data = append(data, s[1:]...)

		if username == data[2] {
			transNum := i + 1
			transNum_str := strconv.Itoa(transNum)
			fmt.Println(transNum)
			//time.Sleep(5 * time.Millisecond)
			client := &http.Client{}
			switch command {

			case "ADD":
				amount := data[3]
				fmt.Println(amount)
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
				req.Host = "transaction"
				resp, err := client.Do(req)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				resp.Body.Close()

			case "BUY":
				symbol := data[3]
				amount := data[4]
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
				req.Host = "transaction"
				resp, err := client.Do(req)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				resp.Body.Close()

			case "SELL":
				symbol := data[3]
				amount := data[4]
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
				req.Host = "transaction"
				resp, err := client.Do(req)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				resp.Body.Close()

			case "QUOTE":
				symbol := data[3]
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
				req.Host = "transaction"
				resp, err := client.Do(req)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				resp.Body.Close()

			case "COMMIT_BUY":
				addr := address + "/commit_buy"
				form := url.Values{
					"transNum": {transNum_str},
					"user":     {username}}
				req, err := http.NewRequest("POST", addr, strings.NewReader(form.Encode()))
				if err != nil {
					fmt.Println(err)
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
				req.Host = "transaction"
				resp, err := client.Do(req)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				resp.Body.Close()

			case "COMMIT_SELL":
				addr := address + "/commit_sell"
				form := url.Values{
					"transNum": {transNum_str},
					"user":     {username}}
				req, err := http.NewRequest("POST", addr, strings.NewReader(form.Encode()))
				if err != nil {
					fmt.Println(err)
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
				req.Host = "transaction"
				resp, err := client.Do(req)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				resp.Body.Close()

			case "CANCEL_BUY":
				addr := address + "/cancel_buy"
				form := url.Values{
					"transNum": {transNum_str},
					"user":     {username}}
				req, err := http.NewRequest("POST", addr, strings.NewReader(form.Encode()))
				if err != nil {
					fmt.Println(err)
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
				req.Host = "transaction"
				resp, err := client.Do(req)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				resp.Body.Close()

			case "CANCEL_SELL":
				addr := address + "/cancel_sell"
				form := url.Values{
					"transNum": {transNum_str},
					"user":     {username}}
				req, err := http.NewRequest("POST", addr, strings.NewReader(form.Encode()))
				if err != nil {
					fmt.Println(err)
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
				req.Host = "transaction"
				resp, err := client.Do(req)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				resp.Body.Close()

			case "SET_BUY_AMOUNT":
				symbol := data[3]
				amount := data[4]
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
				req.Host = "transaction"
				resp, err := client.Do(req)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				resp.Body.Close()

			case "SET_BUY_TRIGGER":
				symbol := data[3]
				amount := data[4]
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
				req.Host = "transaction"
				resp, err := client.Do(req)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				resp.Body.Close()

			case "CANCEL_SET_BUY":
				symbol := data[3]
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
				req.Host = "transaction"
				resp, err := client.Do(req)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				resp.Body.Close()

			case "DISPLAY_SUMMARY":
				addr := address + "/display_summary"
				form := url.Values{
					"transNum": {transNum_str},
					"user":     {username}}
				req, err := http.NewRequest("POST", addr, strings.NewReader(form.Encode()))
				if err != nil {
					fmt.Println(err)
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
				req.Host = "transaction"
				resp, err := client.Do(req)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				resp.Body.Close()

			case "SET_SELL_AMOUNT":
				symbol := data[3]
				amount := data[4]
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
				req.Host = "transaction"
				resp, err := client.Do(req)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				resp.Body.Close()

			case "SET_SELL_TRIGGER":
				symbol := data[3]
				amount := data[4]
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
				req.Host = "transaction"
				resp, err := client.Do(req)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				resp.Body.Close()

			case "CANCEL_SET_SELL":
				symbol := data[3]
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
				req.Host = "transaction"
				resp, err := client.Do(req)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				resp.Body.Close()
			}
		}
	}
}

/* readLines reads a whole file into memory
and returns a slice of its lines. */
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

/* won't return error because readlines worked */
func getTransactionCount(file string) int {
	var count int
	if file == "workload1" {
		count = 100
	} else if file == "workload2" || file == "workload3" || file == "workload4" {
		count = 10000
	} else if file == "workload5" {
		count = 100000
	} else if file == "workload6" {
		count = 1000000
	} else if file == "2018" {
		count = 1200000
	} else {
		fmt.Println("invalid workload file. exiting.")
		os.Exit(1)
	}
	return count
}

func getNumUsers(file string) int {
	var count int
	if file == "workload1" {
		count = 1
	} else if file == "workload2" {
		count = 2
	} else if file == "workload3" {
		count = 10
	} else if file == "workload4" {
		count = 45
	} else if file == "workload5" {
		count = 100
	} else if file == "workload6" {
		count = 1000
	} else if file == "2018" {
		count = 10000
	} else {
		fmt.Println("invalid workload file. exiting.")
		os.Exit(1)
	}
	return count
}

func dumpLogFile(address string, transNum string, username interface{}, filename string) {
	addr := address + "/dumpLog"
	v := url.Values{}
	v.Set("transNum", transNum)
	v.Set("filename", filename)
	if username != nil {
		v.Set("username", username.(string))
	} else {
		v.Set("username", "")
	}
	resp, err := http.PostForm(addr, v)
	if err != nil {
		fmt.Println(err)
	}
	resp.Body.Close()
}

func clearSystemLogs() {
	client := &http.Client{}
	addr := "http://transaction:80" + "/clearSystemLogs"
	req, err := http.NewRequest("POST", addr, nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Host = "transaction"
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	resp.Body.Close()
}

func runWorkload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("runWorkload:hello")
	file := r.URL.Path[len("/runWorkload/"):]
	count := getTransactionCount(file)
	numUsers := getNumUsers(file)
	file = "workload_files/" + file + ".txt"

	start := time.Now()

	// initAuditServer()
	// client.Cmd("FLUSHALL")
	lines, err := readLines(file)
	if err != nil {
		log.Fatalf("readLines: %s", err)
	}

	User := make(map[string]int)

	for i, line := range lines {
		s := strings.Split(line, ",")
		x := strings.Split(s[0], " ")
		for i = 0; i < len(s); i++ {
			s[i] = strings.TrimSpace(s[i])
		}
		data := make([]string, 2)
		data[1] = strings.TrimSpace(x[1])
		data = append(data, s[1:]...)
		User[data[2]] = 1
	}

	p := 0
	userS := make([]string, numUsers+10)
	for key, value := range User {
		if value == 1 {
			userS[p] = key
			//userCount += 1
			fmt.Println(userS[p])
			p = p + 1
			fmt.Println("Key:", key, "Value:", value)
		}
	}
	for u := 0; u < (numUsers + 1); u++ {

		if userS[u] != "./testLOG" && userS[u] != "" {
			//fmt.Println(u, ":", userS[u])
			time.Sleep(130 * time.Millisecond)
			//go concurrencyLogic("http://transaction:1300", lines, userS[u])
			/*
				threads := 7
				if u%threads == 0 {
					go concurrencyLogic("http://transaction2:1300", lines, userS[u])
				} else if u%threads == 1 {
					go concurrencyLogic("http://transaction3:1300", lines, userS[u])
				} else if u%threads == 2 {
					go concurrencyLogic("http://transaction4:1300", lines, userS[u])
				} else if u%threads == 3 {
					go concurrencyLogic("http://transaction5:1300", lines, userS[u])
				} else if u%threads == 4 {
					go concurrencyLogic("http://transaction:1300", lines, userS[u])
				} else if u%threads == 5 {
					go concurrencyLogic("http://transaction6:1300", lines, userS[u])
				} else if u%threads == 6 {
					go concurrencyLogic("http://transaction7:1300", lines, userS[u])
				}
			*/
			webServeNum := 2
			if u%webServeNum == 0 {
				wg.Add(1)
				go concurrencyLogic(address, lines, userS[u])
			}

		}
	}
	wg.Wait()
	//wg.Add(1)
	dumpLogFile("http://transaction:80", strconv.Itoa(count), nil, "./testLOG")
	clearSystemLogs()
	//wg.Wait()
	//print stats for the workload file
	fmt.Println("\n\n")
	fmt.Println("-----STATISTICS-----")
	end := time.Now()
	difference := end.Sub(start)
	difference_seconds := float64(difference) / float64(time.Second)
	fmt.Println("Total time: ", difference)
	fmt.Println("Average time for each transaction: ", difference_seconds/float64(count))
	fmt.Println("Transactions per second: ", float64(count)/difference_seconds)
}
