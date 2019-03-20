package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
  "os"
)

var address = "http://" + os.Getenv("AUDIT_URL")

func convertStringToDecimal(value string) string {
	amount, err := strconv.ParseFloat(value, 64)
	if err == nil {
		/** displaying the string variable into the console */
		fmt.Println("Value:", amount)
	}
	amount2f := fmt.Sprintf("%.2f", amount)
	return amount2f
}

func LogUserCommand(server string, transNum int, command string, username interface{}, funds interface{}, stockSymbol interface{}, filename interface{}) {

	addr := address + "/userCommand"
	v := url.Values{
		"server":         {server},
		"command":        {command},
		"transactionNum": {strconv.Itoa(transNum)},
	}

	if username != nil {
		v.Set("username", username.(string))
	}
	if funds != nil {
		v.Set("funds", convertStringToDecimal(funds.(string)))
	}
	if stockSymbol != nil {
		v.Set("stockSymbol", stockSymbol.(string))
	}
	if filename != nil {
		v.Set("filename", filename.(string))
	}
	resp, err := http.PostForm(addr, v)
	if err != nil {
		fmt.Println(err)
	}
	resp.Body.Close()

}

func LogAccountTransactionCommand(server string, transNum int, action string, username string, funds string) {

	addr := address + "/accountTransactionCommand"
	v := url.Values{
		"server":         {server},
		"transactionNum": {strconv.Itoa(transNum)},
		"action":         {action},
		"username":       {username},
		"funds":          {convertStringToDecimal(funds)},
	}
	resp, err := http.PostForm(addr, v)
	if err != nil {
		fmt.Println(err)
	}
	resp.Body.Close()
}

func LogSystemEventCommand(server string, transNum int, command string, username interface{}, funds interface{}, stockSymbol interface{}, filename interface{}) {

  address = "http://" + os.Getenv("AUDIT_URL")
	addr := address + "/systemEventCommand"
	v := url.Values{
		"server":         {server},
		"command":        {command},
		"transactionNum": {strconv.Itoa(transNum)},
	}

	if username != nil {
		v.Set("username", username.(string))
	}
	if funds != nil {
		v.Set("funds", convertStringToDecimal(funds.(string)))
	}
	if stockSymbol != nil {
		v.Set("stockSymbol", stockSymbol.(string))
	}
	if filename != nil {
		v.Set("filename", filename.(string))
	}
	resp, err := http.PostForm(addr, v)
	if err != nil {
		fmt.Println(err)
	}
	resp.Body.Close()
}

func LogQuoteServerCommand(server string, transNum int, price string, stock string, username string, quoteServerTime uint64, cryptoKey string) {

  address = "http://" + os.Getenv("AUDIT_URL")
	addr := address + "/quoteServerCommand"
	v := url.Values{
		"server":          {server},
		"transactionNum":  {strconv.Itoa(transNum)},
		"price":           {price},
		"stockSymbol":     {stock},
		"username":        {username},
		"quoteServerTime": {strconv.FormatUint(quoteServerTime, 10)},
		"cryptokey":       {cryptoKey},
	}
	resp, err := http.PostForm(addr, v)
	if err != nil {
		fmt.Println(err)
	}
	resp.Body.Close()

}

func LogErrorEventCommand(server string, transNum int, command string, username interface{}, funds interface{}, stockSymbol interface{}, filename interface{}, errorMessage interface{}) {

	addr := address + "/errorEventCommand"
	v := url.Values{
		"server":         {server},
		"command":        {command},
		"transactionNum": {strconv.Itoa(transNum)},
	}

	if username != nil {
		v.Set("username", username.(string))
	}
	if funds != nil {
		v.Set("funds", convertStringToDecimal(funds.(string)))
	}
	if stockSymbol != nil {
		v.Set("stockSymbol", stockSymbol.(string))
	}
	if filename != nil {
		v.Set("filename", filename.(string))
	}
	if errorMessage != nil {
		v.Set("errorMessage", errorMessage.(string))
	}
	resp, err := http.PostForm(addr, v)
	if err != nil {
		fmt.Println(err)
	}
	resp.Body.Close()
}

func LogDebugEventCommand(server string, transNum int, command string, username interface{}, funds interface{}, stockSymbol interface{}, filename interface{}, debugMessage interface{}) {

	addr := address + "/debugEventCommand"
	v := url.Values{
		"server":         {server},
		"command":        {command},
		"transactionNum": {strconv.Itoa(transNum)},
	}

	if username != nil {
		v.Set("username", username.(string))
	}
	if funds != nil {
		v.Set("funds", convertStringToDecimal(funds.(string)))
	}
	if stockSymbol != nil {
		v.Set("stockSymbol", stockSymbol.(string))
	}
	if filename != nil {
		v.Set("filename", filename.(string))
	}
	if debugMessage != nil {
		v.Set("debugMessage", debugMessage.(string))
	}
	resp, err := http.PostForm(addr, v)
	if err != nil {
		fmt.Println(err)
	}
	resp.Body.Close()
}

func DumpLog(filename string, username interface{}) {

	addr := address + "/dumpLog"
	v := url.Values{
		"filename": {filename},
	}

	if username != nil {
		v.Set("username", username.(string))
	}

	resp, err := http.PostForm(addr, v)
	if err != nil {
		fmt.Println(err)
	}
	resp.Body.Close()
}
