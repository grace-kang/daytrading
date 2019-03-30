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
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const (
	connHost = "localhost"
	connPort = "1600"
	connType = "http"
	address  = "http://reverse-proxy:80"
	server   = "webserver"
)

//var wg sync.WaitGroup
var transNum = 0

func main() {
	r := mux.NewRouter()

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
	//defer wg.Done()
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
			time.Sleep(1 * time.Millisecond)
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
	//numUsers := getNumUsers(file)
	file = "workload_files/" + file + ".txt"

	start := time.Now()

	// initAuditServer()
	// client.Cmd("FLUSHALL")
	lines, err := readLines(file)
	if err != nil {
		log.Fatalf("readLines: %s", err)
	}

	User := make(map[string]int)
	//webServeNum := 2
	counter := 0
	fmt.Println(counter)
	for i, line := range lines {
		s := strings.Split(line, ",")
		x := strings.Split(s[0], " ")
		for k := 0; k < len(s); k++ {
			s[k] = strings.TrimSpace(s[k])
		}
		data := make([]string, 2)
		data[1] = strings.TrimSpace(x[1])
		data = append(data, s[1:]...)
		if User[data[2]] == 0 {
			User[data[2]] = 1
			counter += 1
			//if i%webServeNum == 0 {
			fmt.Println("Num:", i, "-----------------------User:", data[2])
			fmt.Println("------------------counter-----------------", counter)

			time.Sleep(100 * time.Millisecond)

			if data[2] != "./testLOG" && data[2] != "" {
				//wg.Add(1)
				concurrencyLogic(address, lines, data[2])
			}

		}
	}

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

	//wg.Wait()
	dumpLogFile("http://transaction:80", strconv.Itoa(count), nil, "./testLOG")

	//print stats for the workload file
	fmt.Println("--------------------")
	fmt.Println("^^^^^^^^^^^^^^^^^^^^")
	fmt.Println("-----STATISTICS-----")
	end := time.Now()
	difference := end.Sub(start)
	difference_seconds := float64(difference) / float64(time.Second)
	fmt.Println("Total time: ", difference)
	fmt.Println("Average time for each transaction: ", difference_seconds/float64(count))
	fmt.Println("Transactions per second: ", float64(count)/difference_seconds)
}
