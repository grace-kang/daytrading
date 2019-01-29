//https://www.alexedwards.net/blog/working-with-redis
/*
macowner$ redis-cli
127.0.0.1:6379> flushall
127.0.0.1:6379> hgetall oY01WVirLr
*/

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var stockPrices = map[string]float64{}
var stocksAmount = map[string]int{}

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

/* writeLines writes the lines to the given file. */
func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func main() {
	start := time.Now()
	count := 0
	client := dialRedis()
	client.Cmd("FLUSHALL")
	lines, err := readLines("workload_files/workload2.txt")
	if err != nil {
		log.Fatalf("readLines: %s", err)
	}
	for i, line := range lines {
		count = i + 1
		s := strings.Split(line, ",")
		x := strings.Split(s[0], " ")
		command := x[1]
		transNum := i + 1
		for i = 0; i < len(s); i++ {
			s[i] = strings.TrimSpace(s[i])
		}

		data := make([]string, 2)
		fmt.Println(transNum)
		// data[0] = transNum
		data[1] = strings.TrimSpace(x[1])
		data = append(data, s[1:]...)

		switch command {
		case "ADD":
			amount, _ := strconv.ParseFloat(data[3], 64)
			username := data[2]
			add(transNum, username, amount, client)

		case "BUY":
			username := data[2]
			symbol := data[3]
			amount, _ := strconv.ParseFloat(data[4], 64)
			buy(transNum, username, symbol, amount, client)

		case "SELL":
			username := data[2]
			symbol := data[3]
			amount, _ := strconv.ParseFloat(data[4], 64)
			sell(transNum, username, symbol, amount, client)

		case "QUOTE":
			username := data[2]
			stock := data[3]
			quote(transNum, username, stock, client)

		case "COMMIT_BUY":
			username := data[2]
			commit_buy(transNum, username, client)

		case "COMMIT_SELL":
			username := data[2]
			commit_sell(transNum, username, client)

		case "DISPLAY_SUMMARY":
			username := data[1]
			display_summary(transNum, username)

		case "CANCEL_BUY":
			username := data[2]
			cancel_buy(transNum, username, client)

		case "CANCEL_SELL":
			username := data[2]
			cancel_sell(transNum, username, client)

		case "SET_BUY_AMOUNT":
			username := data[2]
			symbol := data[3]
			amount, _ := strconv.ParseFloat(data[4], 64)
			set_buy_amount(transNum, username, symbol, amount, client)

		case "SET_BUY_TRIGGER":
			username := data[2]
			symbol := data[3]
			amount, _ := strconv.ParseFloat(data[4], 64)
			set_buy_trigger(transNum, username, symbol, amount, client)

		case "CANCEL_SET_BUY":
			username := data[2]
			symbol := data[3]
			cancel_set_buy(transNum, username, symbol, client)

		case "SET_SELL_AMOUNT":
			username := data[2]
			symbol := data[3]
			amount, _ := strconv.ParseFloat(data[4], 64)
			set_sell_amount(transNum, username, symbol, amount, client)

		case "DUMPLOG":
			if len(data) == 3 {
				filename := data[2]
				dumplog(transNum, filename)

			} else if len(data) == 4 {
				username := data[2]
				filename := data[3]
				dumplog(transNum, username, filename)
			}

		case "SET_SELL_TRIGGER":
			username := data[2]
			symbol := data[3]
			amount, _ := strconv.ParseFloat(data[4], 64)
			set_sell_trigger(transNum, username, symbol, amount, client)

		case "CANCEL_SET_SELL":
			username := data[2]
			symbol := data[3]
			cancel_set_sell(transNum, username, symbol, client)
		}
	}

	//print stats for the workload file
	fmt.Println("\n\n")
	fmt.Println("-----STATISTICS-----")
	end := time.Now()
	difference := end.Sub(start)
	difference_seconds := float64(difference) / float64(time.Second)
	fmt.Println("Total time: ", difference)
	fmt.Println("Average time for each transaction: ", difference_seconds/float64(count))
	fmt.Println("Transactions per second: ", float64(count)/difference_seconds)
	/* How to put a map straight into Redis
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	err = client.Cmd("HMSET", "user:4", "user", "bob", "balance", 5000, m).Err */
}
