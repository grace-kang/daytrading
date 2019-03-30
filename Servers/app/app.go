package main

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"
)

//Create a struct that holds information to be displayed in our HTML file
type Welcome struct {
	Name     string
	Time     string
	Date     string
	Time2    string
	Quote    string
	Response string
}

/*
func handler(w http.ResponseWriter, r *http.Request) {
	var name, _ = os.Hostname()
	fmt.Fprintf(w, "<h1>This request was processed by host: %s</h1>\n", name)
}
*/
/*
func main() {
	fmt.Fprintf(os.Stdout, "Web Server started. Listening on 0.0.0.0:80\n")
	http.HandleFunc("/", handler)
	http.ListenAndServe(":80", nil)
}
*/
//Go application entrypoint
func main() {
	//Instantiate a Welcome struct object and pass in some random information.
	//We shall get the name of the user as a query parameter from the URL
	quotey := getQuote("cde", "123")
	responsey := add("abc")
	welcome := Welcome{"Anonymous", time.Now().Format(time.Stamp),
		time.Now().Format("02-01-2006"), time.Now().Format("15:04:05"),
		quotey, responsey,
	}

	//We tell Go exactly where we can find our html file. We ask Go to parse the html file (Notice
	// the relative path). We wrap it in a call to template.Must() which handles any errors and halts if there are fatal errors

	templates := template.Must(template.ParseFiles("templates/homepage.html"))

	//Our HTML comes with CSS that go needs to provide when we run the app. Here we tell go to create
	// a handle that looks in the static directory, go then uses the "/static/" as a url that our
	//html can refer to when looking for our css and other files.

	http.Handle("/static/", //final url can be anything
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static")))) //Go looks in the relative "static" directory first using http.FileServer(), then matches it to a
	//url of our choice as shown in http.Handle("/static/"). This url is what we need when referencing our css files
	//once the server begins. Our html code would therefore be <link rel="stylesheet"  href="/static/stylesheet/...">
	//It is important to note the url in http.Handle can be whatever we like, so long as we are consistent.

	//This method takes in the URL path "/" and a function that takes in a response writer, and a http request.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		//Takes the name from the URL query e.g ?name=Martin, will set welcome.Name = Martin.
		if name := r.FormValue("name"); name != "" {
			welcome.Name = name
		}
		//If errors show an internal server error message
		//I also pass the welcome struct to the welcome-template.html file.

		if err := templates.ExecuteTemplate(w, "homepage.html", welcome); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	//Start the web server, set the port to listen to 8080. Without a path it assumes localhost
	//Print any errors from starting the webserver using fmt

	fmt.Println("Listening")
	fmt.Println(http.ListenAndServe(":80", nil))
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
	//fmt.Println(string(message))

	//return "hello"
}

func add(username string) string {
	amount := "10000"
	TRANSACTION_URL := os.Getenv("TRANSACTION_URL")
	address := TRANSACTION_URL
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
