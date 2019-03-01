package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

// var stockPrices = map[string]float64{}
// var stocksAmount = map[string]int{}

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
	} else if file == "2018" {
		count = 10000
	} else {
		fmt.Println("invalid workload file. exiting.")
		os.Exit(1)
	}
	return count
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Error please specify the workload file (eg. workload2)")
		os.Exit(1)
	}
	file := os.Args[1]
	count := getTransactionCount(file)
	numUsers := getNumUsers(file)
	file = "workload_files/" + file + ".txt"

	start := time.Now()
	wg.Add(numUsers)
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
	// go linearLogic(lines)
	// go linearLogic2(lines)
	// go linearLogic3(lines)
	p := 0
	userS := make([]string, 100)
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
			time.Sleep(50 * time.Millisecond)
			if u%5 == 0 {
				go concurrencyLogic("http://transaction:1300", lines, userS[u])
			} else if u%5 == 1 {
				go concurrencyLogic("http://transaction:1300", lines, userS[u])
			} else if u%5 == 2 {
				go concurrencyLogic("http://transaction:1300", lines, userS[u])
			} else if u%5 == 3 {
				go concurrencyLogic("http://transaction:1300", lines, userS[u])
			} else if u%5 == 4 {
				go concurrencyLogic("http://transaction:1300", lines, userS[u])
			}
		}

	}
	wg.Wait()

	dumpLogFile("http://transaction:1300", "100000", nil, "./testLog")

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

// func linearLogic(lines []string) {
// 	client := dialRedis()
// 	defer wg.Done()
// 	for i, line := range lines {
// 		s := strings.Split(line, ",")
// 		x := strings.Split(s[0], " ")
// 		command := x[1]
//
// 		for ik := 0; ik < len(s); ik++ {
// 			s[ik] = strings.TrimSpace(s[ik])
// 		}
//
// 		data := make([]string, 2)
//
// 		data[1] = strings.TrimSpace(x[1])
// 		data = append(data, s[1:]...)
//
// 		transNum := i + 1
// 		fmt.Println(transNum)
// 		switch command {
// 		case "ADD":
// 			amount, _ := strconv.ParseFloat(data[3], 64)
// 			username := data[2]
// 			add(transNum, username, amount, client)
//
// 		case "BUY":
// 			username := data[2]
// 			symbol := data[3]
// 			amount, _ := strconv.ParseFloat(data[4], 64)
// 			buy(transNum, username, symbol, amount, client)
//
// 		case "SELL":
// 			username := data[2]
// 			symbol := data[3]
// 			amount, _ := strconv.ParseFloat(data[4], 64)
// 			sell(transNum, username, symbol, amount, client)
//
// 		}
//
// 	}
// }
//
// func linearLogic2(lines []string) {
// 	client := dialRedis()
// 	defer wg.Done()
// 	for i, line := range lines {
// 		s := strings.Split(line, ",")
// 		x := strings.Split(s[0], " ")
// 		command := x[1]
//
// 		for ik := 0; ik < len(s); ik++ {
// 			s[ik] = strings.TrimSpace(s[ik])
// 		}
//
// 		data := make([]string, 2)
//
// 		data[1] = strings.TrimSpace(x[1])
// 		data = append(data, s[1:]...)
//
// 		transNum := i + 1
// 		fmt.Println(transNum)
// 		switch command {
//
// 		case "SET_BUY_AMOUNT":
// 			username := data[2]
// 			symbol := data[3]
// 			amount, _ := strconv.ParseFloat(data[4], 64)
// 			set_buy_amount(transNum, username, symbol, amount, client)
//
// 		case "SET_BUY_TRIGGER":
// 			username := data[2]
// 			symbol := data[3]
// 			amount, _ := strconv.ParseFloat(data[4], 64)
// 			set_buy_trigger(transNum, username, symbol, amount, client)
//
// 		case "CANCEL_SET_BUY":
// 			username := data[2]
// 			symbol := data[3]
// 			cancel_set_buy(transNum, username, symbol, client)
//
// 		case "SET_SELL_AMOUNT":
// 			username := data[2]
// 			symbol := data[3]
// 			amount, _ := strconv.ParseFloat(data[4], 64)
// 			set_sell_amount(transNum, username, symbol, amount, client)
//
// 		case "DUMPLOG":
// 			if len(data) == 3 {
// 				filename := data[2]
// 				dumplog(transNum, filename)
// 			} else if len(data) == 4 {
// 				username := data[2]
// 				filename := data[3]
// 				dumplog(transNum, username, filename)
// 			}
//
// 		case "SET_SELL_TRIGGER":
// 			username := data[2]
// 			symbol := data[3]
// 			amount, _ := strconv.ParseFloat(data[4], 64)
// 			set_sell_trigger(transNum, username, symbol, amount, client)
//
// 		case "CANCEL_SET_SELL":
// 			username := data[2]
// 			symbol := data[3]
// 			cancel_set_sell(transNum, username, symbol, client)
// 		}
//
// 	}
// }
//
// func linearLogic3(lines []string) {
// 	client := dialRedis()
// 	defer wg.Done()
// 	for i, line := range lines {
// 		s := strings.Split(line, ",")
// 		x := strings.Split(s[0], " ")
// 		command := x[1]
//
// 		for ik := 0; ik < len(s); ik++ {
// 			s[ik] = strings.TrimSpace(s[ik])
// 		}
//
// 		data := make([]string, 2)
//
// 		data[1] = strings.TrimSpace(x[1])
// 		data = append(data, s[1:]...)
//
// 		transNum := i + 1
// 		fmt.Println(transNum)
// 		switch command {
//
// 		case "QUOTE":
// 			//username := data[2]
// 			//stock := data[3]
// 			//quote(transNum, username, stock, client)
//
// 		case "COMMIT_BUY":
// 			username := data[2]
// 			commit_buy(transNum, username, client)
//
// 		case "COMMIT_SELL":
// 			username := data[2]
// 			commit_sell(transNum, username, client)
// 			/*
// 		case "DISPLAY_SUMMARY":
// 			username := data[2]
// 			display_summary(transNum, username, client)
// 			*/
// 		case "CANCEL_BUY":
// 			username := data[2]
// 			cancel_buy(transNum, username, client)
//
// 		case "CANCEL_SELL":
// 			username := data[2]
// 			cancel_sell(transNum, username, client)
//
// 		}
//
// 	}
// }
//

func dumpLogFile(address string, transNum string, username interface{}, filename string) {
	addr := address + "/dumpLog"
	v := url.Values{}
	v.Set("transNum", transNum)
	v.Set("filename", filename)
	if username != nil {
		v.Set("filename", username.(string))
	}
	resp, err := http.PostForm(addr, v)
	if err != nil {
		fmt.Println(err)
	}
	resp.Body.Close()
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
			fmt.Println(transNum)
			transNum_str := strconv.Itoa(transNum)
			//time.Sleep(5 * time.Millisecond)
			switch command {

			case "ADD":
				amount := data[3]
				addr := address + "/add"
				resp, err := http.PostForm(addr, url.Values{
					"transNum": {transNum_str},
					"user":     {username},
					"amount":   {amount}})
				if err != nil {
					fmt.Println(err)
				}
				resp.Body.Close()

			case "BUY":
				symbol := data[3]
				amount := data[4]
				addr := address + "/buy"
				resp, err := http.PostForm(addr, url.Values{
					"transNum": {transNum_str},
					"user":     {username},
					"symbol":   {symbol},
					"amount":   {amount}})
				if err != nil {
					fmt.Println(err)
					//os.Exit(1)
				}
				resp.Body.Close()

			case "SELL":
				symbol := data[3]
				amount := data[4]
				addr := address + "/sell"
				resp, err := http.PostForm(addr, url.Values{
					"transNum": {transNum_str},
					"user":     {username},
					"symbol":   {symbol},
					"amount":   {amount}})
				if err != nil {
					fmt.Println(err)
					//os.Exit(1)
				}
				resp.Body.Close()

			case "QUOTE":
				symbol := data[3]
				addr := address + "/quote"
				transNum_str := strconv.Itoa(transNum)
				resp, err := http.PostForm(addr, url.Values{
					"transNum": {transNum_str},
					"user":     {username},
					"symbol":   {symbol}})
				if err != nil {
					fmt.Println(err)
					//os.Exit(1)
				}
				resp.Body.Close()

			case "COMMIT_BUY":
				addr := address + "/commit_buy"
				resp, err := http.PostForm(addr, url.Values{
					"transNum": {transNum_str},
					"user":     {username}})
				if err != nil {
					fmt.Println(err)
					//os.Exit(1)
				}
				resp.Body.Close()

			case "COMMIT_SELL":
				addr := address + "/commit_sell"
				resp, err := http.PostForm(addr, url.Values{
					"transNum": {transNum_str},
					"user":     {username}})
				if err != nil {
					fmt.Println(err)
					//os.Exit(1)
				}
				resp.Body.Close()

			case "CANCEL_BUY":
				addr := address + "/cancel_buy"
				resp, err := http.PostForm(addr, url.Values{
					"transNum": {transNum_str},
					"user":     {username}})
				if err != nil {
					fmt.Println(err)
					//os.Exit(1)
				}
				resp.Body.Close()

			case "CANCEL_SELL":
				addr := address + "/cancel_sell"
				resp, err := http.PostForm(addr, url.Values{
					"transNum": {transNum_str},
					"user":     {username}})
				if err != nil {
					fmt.Println(err)
					//os.Exit(1)
				}
				resp.Body.Close()

			case "SET_BUY_AMOUNT":
				symbol := data[3]
				amount := data[4]
				addr := address + "/set_buy_amount"
				resp, err := http.PostForm(addr, url.Values{
					"transNum": {transNum_str},
					"user":     {username},
					"symbol":   {symbol},
					"amount":   {amount}})
				if err != nil {
					fmt.Println(err)
					//os.Exit(1)
				}
				resp.Body.Close()

			case "SET_BUY_TRIGGER":
				symbol := data[3]
				amount := data[4]
				addr := address + "/set_buy_trigger"
				resp, err := http.PostForm(addr, url.Values{
					"transNum": {transNum_str},
					"user":     {username},
					"symbol":   {symbol},
					"amount":   {amount}})
				if err != nil {
					fmt.Println(err)
					//os.Exit(1)
				}
				resp.Body.Close()

			case "CANCEL_SET_BUY":
				symbol := data[3]
				addr := address + "/cancel_set_buy"
				resp, err := http.PostForm(addr, url.Values{
					"transNum": {transNum_str},
					"user":     {username},
					"symbol":   {symbol}})
				if err != nil {
					fmt.Println(err)
					//os.Exit(1)
				}
				resp.Body.Close()

			case "DISPLAY_SUMMARY":
				addr := address + "/display_summary"
				resp, err := http.PostForm(addr, url.Values{
					"transNum": {transNum_str},
					"user":     {username}})
				if err != nil {
					fmt.Println(err)
					//os.Exit(1)
				}
				resp.Body.Close()

			case "SET_SELL_AMOUNT":
				symbol := data[3]
				amount := data[4]
				addr := address + "/set_sell_amount"
				resp, err := http.PostForm(addr, url.Values{
					"transNum": {transNum_str},
					"user":     {username},
					"symbol":   {symbol},
					"amount":   {amount}})
				if err != nil {
					fmt.Println(err)
					//os.Exit(1)
				}
				resp.Body.Close()

			case "SET_SELL_TRIGGER":
				symbol := data[3]
				amount := data[4]
				addr := address + "/set_sell_trigger"
				resp, err := http.PostForm(addr, url.Values{
					"transNum": {transNum_str},
					"user":     {username},
					"symbol":   {symbol},
					"amount":   {amount}})
				if err != nil {
					fmt.Println(err)
					//os.Exit(1)
				}
				resp.Body.Close()

			case "CANCEL_SET_SELL":
				symbol := data[3]
				addr := address + "/cancel_set_sell"
				resp, err := http.PostForm(addr, url.Values{
					"transNum": {transNum_str},
					"user":     {username},
					"symbol":   {symbol}})
				if err != nil {
					fmt.Println(err)
					//os.Exit(1)
				}
				resp.Body.Close()

			}
		}
	}
}
