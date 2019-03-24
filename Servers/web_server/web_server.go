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
	"strings"
)

const (
	connHost = "localhost"
	connPort = "1600"
	connType = "http"
)

func main() {

	r := http.NewServeMux()

	r.HandleFunc("/userCommands.js", SendJqueryJs)

	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/login", loginHandler)
	r.HandleFunc("/sendCommand", sendCommandHandle)

	err := http.ListenAndServe(":"+connPort, r)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
		return
	}

	fmt.Println("Web server is listening on port " + connPort)
}

func sendCommandHandle(w http.ResponseWriter, r *http.Request) {
	fmt.Println("in sendhandler")

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

	quotey := getQuote("abc", "123")
	fmt.Println(string(quotey))
	addy := add("abc")
	fmt.Println(string(addy))
	myvar := map[string]interface{}{"Quote": quotey, "Add": addy}
	outputHTML(w, "tmp/home.html", myvar)

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

func add(username string) string {
	amount := "10000"
	//TRANSACTION_URL := os.Getenv("TRANSACTION_URL")
	address := "http://transaction:1300"
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
