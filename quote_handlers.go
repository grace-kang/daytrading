package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
)

func getQuoteHandler() {
	req, err := http.NewRequest("GET", "http://quoteserve.seng.uvic.ca:4444", nil)
	req.Header.Add("If-None-Match", `W/"wyzzy"`)

	q := req.URL.Query()
	q.Add("user", "user")
	q.Add("stock", "ABC")
	q.Add("transNum", "1")
	req.URL.RawQuery = q.Encode()

	client := http.Client{}

	var resp *http.Response
	for {
		resp, err = client.Do(req)

		if err != nil { // trans server down? retry
			fmt.Println(err)
		} else {
			break
		}
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("Error reading body: %s", err.Error())
	}

	fmt.Println(string(body))
	resp.Body.Close()
}
