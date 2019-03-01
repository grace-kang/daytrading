package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

type LoggingService interface {
	LogUserCommand(server string, transNum int, command string, username interface{}, funds interface{}, stockSymbol interface{}, filename interface{})
	LogAccountTransactionCommand(server string, transNum int, action string, username string, funds string)
	LogSystemEventCommand(server string, transNum int, command string, username interface{}, funds interface{}, stockSymbol interface{}, filename interface{})
	LogQuoteServerCommand(server string, transNum int, price string, stock string, username string, quoteServerTime uint64, cryptoKey string)
	LogErrorEventCommand(server string, transNum int, command string, username interface{}, funds interface{}, stockSymbol interface{}, filename interface{}, errorMessage interface{})
	LogDebugEventCommand(server string, transNum int, command string, username interface{}, funds interface{}, stockSymbol interface{}, filename interface{}, debugMessage interface{})
	DumpLog(filename string, username interface{})
}

type Logger struct {
	Address string
}

func convertStringToDecimal(value string) string {
	amount, err := strconv.ParseFloat(value, 64)
	if err == nil {
		/** displaying the string variable into the console */
		fmt.Println("Value:", amount)
	}
	amount2f := fmt.Sprintf("%.2f", amount)
	return amount2f
}

func (logger Logger) LogUserCommand(server string, transNum int, command string, username interface{}, funds interface{}, stockSymbol interface{}, filename interface{}) {
	params := map[string]string{
		"server":         server,
		"command":        command,
		"transactionNum": strconv.Itoa(transNum),
	}

	if username != nil {
		params["username"] = username.(string)
	}
	if funds != nil {
		params["funds"] = convertStringToDecimal(funds.(string))
	}
	if stockSymbol != nil {
		params["stockSymbol"] = stockSymbol.(string)
	}
	if filename != nil {
		params["filename"] = filename.(string)
	}

	logger.SendRequest("/userCommand", params)

}

func (logger Logger) LogAccountTransactionCommand(server string, transNum int, action string, username string, funds string) {
	params := map[string]string{
		"server":         server,
		"transactionNum": strconv.Itoa(transNum),
		"action":         action,
		"username":       username,
		"funds":          convertStringToDecimal(funds),
	}
	logger.SendRequest("/accountTransactionCommand", params)
}

func (logger Logger) LogSystemEventCommand(server string, transNum int, command string, username interface{}, funds interface{}, stockSymbol interface{}, filename interface{}) {
	params := map[string]string{
		"server":         server,
		"command":        command,
		"transactionNum": strconv.Itoa(transNum),
	}

	if username != nil {
		params["username"] = username.(string)
	}
	if funds != nil {
		params["funds"] = convertStringToDecimal(funds.(string))
	}
	if stockSymbol != nil {
		params["stockSymbol"] = stockSymbol.(string)
	}
	if filename != nil {
		params["filename"] = filename.(string)
	}

	logger.SendRequest("/systemEventCommand", params)
}

func (logger Logger) LogQuoteServerCommand(server string, transNum int, price string, stock string, username string, quoteServerTime uint64, cryptoKey string) {
	params := map[string]string{
		"server":          server,
		"transactionNum":  strconv.Itoa(transNum),
		"price":           price,
		"stockSymbol":     stock,
		"username":        username,
		"quoteServerTime": strconv.FormatUint(quoteServerTime, 10),
		"cryptokey":       cryptoKey,
	}
	logger.SendRequest("/quoteServerCommand", params)
}

func (logger Logger) LogErrorEventCommand(server string, transNum int, command string, username interface{}, funds interface{}, stockSymbol interface{}, filename interface{}, errorMessage interface{}) {
	params := map[string]string{
		"server":         server,
		"transactionNum": strconv.Itoa(transNum),
		"command":        command,
	}
	if username != nil {
		params["username"] = username.(string)
	}
	if funds != nil {
		params["funds"] = convertStringToDecimal(funds.(string))
	}
	if stockSymbol != nil {
		params["stockSymbol"] = stockSymbol.(string)
	}
	if filename != nil {
		params["filename"] = filename.(string)
	}
	if errorMessage != nil {
		params["errorMessage"] = errorMessage.(string)
	}
	logger.SendRequest("/errorEventCommand", params)
}

func (logger Logger) LogDebugEventCommand(server string, transNum int, command string, username interface{}, funds interface{}, stockSymbol interface{}, filename interface{}, debugMessage interface{}) {
	params := map[string]string{
		"server":         server,
		"transactionNum": strconv.Itoa(transNum),
		"command":        command,
	}
	if username != nil {
		params["username"] = username.(string)
	}
	if funds != nil {
		params["funds"] = convertStringToDecimal(funds.(string))
	}
	if stockSymbol != nil {
		params["stockSymbol"] = stockSymbol.(string)
	}
	if filename != nil {
		params["filename"] = filename.(string)
	}
	if debugMessage != nil {
		params["debugMessage"] = debugMessage.(string)
	}
	logger.SendRequest("/debugEventCommand", params)
}

func (logger Logger) DumpLog(filename string, username interface{}) {
	fmt.Println("in DumpLog")
	params := map[string]string{
		"filename": filename,
	}
	if username != nil {
		params["username"] = username.(string)
	}
	fmt.Println("in DumpLog, ready to send request")
	logger.SendRequest("/dumpLog", params)
}

func (logger Logger) SendRequestTest() {
	resp, err := http.Get("http://127.0.0.1:1400/test")
	fmt.Println("in SendRequestTest, after init resp")
	if err != nil {
		fmt.Println("getting response failed: " + err.Error())
		os.Exit(1)
	} else {
		fmt.Println("getting response succedd: ")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading all body: %s", err.Error())
	}
	fmt.Println("body is " + string(body))
}

func (logger Logger) SendRequest(subaddress string, params map[string]string) {
	req, err := http.NewRequest("GET", logger.Address+subaddress, nil)
	req.Close = true
	if err != nil {
		fmt.Println("connecting to audit server failed: " + err.Error())
		return
	}

	url := req.URL.Query()
	for k, v := range params {
		url.Add(k, v)
	}
	req.URL.RawQuery = url.Encode()
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   time.Second * 15,
			KeepAlive: 0,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	client := &http.Client{Transport: transport}
	var resp *http.Response
	for {
		resp, err = client.Do(req)

		if err != nil { // trans server down? retry
			fmt.Println("getting audit response error: " + err.Error())
		} else {
			fmt.Println("gettting response from audit server")
			break
		}
	}
	defer resp.Body.Close()

}
