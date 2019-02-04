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
	"sync"
	"time"
)

var wg sync.WaitGroup

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
	if len(os.Args) != 4 {
		fmt.Println("Error: Argument format is file, numUsers, numTransactions")
		os.Exit(1)
	}
	file := os.Args[1]
	numUsers, _ := strconv.Atoi(os.Args[2])
	count, _ := strconv.Atoi(os.Args[3])
	start := time.Now()
	wg.Add(numUsers + 3)
	initAuditServer()
	client.Cmd("FLUSHALL")
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

	go linearLogic(lines)
	go linearLogic2(lines)
	go linearLogic3(lines)
	p := 0
	userS := make([]string, 100)
	for key, value := range User {
		userS[p] = key
		fmt.Println(userS[p])
		p = p + 1
		fmt.Println("Key:", key, "Value:", value)
	}
	for u := 0; u < (numUsers + 1); u++ {

		if userS[u] != "./testLOG" && userS[u] != "" {
			//fmt.Println(u, ":", userS[u])
			go concurrencyLogic(lines, userS[u])
		}
	}
	wg.Wait()

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
func linearLogic(lines []string) {
	client := dialRedis()
	defer wg.Done()
	for i, line := range lines {
		s := strings.Split(line, ",")
		x := strings.Split(s[0], " ")
		command := x[1]

		for ik := 0; ik < len(s); ik++ {
			s[ik] = strings.TrimSpace(s[ik])
		}

		data := make([]string, 2)

		data[1] = strings.TrimSpace(x[1])
		data = append(data, s[1:]...)

		transNum := i + 1
		fmt.Println(transNum)
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

		}

	}
}

func linearLogic2(lines []string) {
	client := dialRedis()
	defer wg.Done()
	for i, line := range lines {
		s := strings.Split(line, ",")
		x := strings.Split(s[0], " ")
		command := x[1]

		for ik := 0; ik < len(s); ik++ {
			s[ik] = strings.TrimSpace(s[ik])
		}

		data := make([]string, 2)

		data[1] = strings.TrimSpace(x[1])
		data = append(data, s[1:]...)

		transNum := i + 1
		fmt.Println(transNum)
		switch command {

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
}

func linearLogic3(lines []string) {
	client := dialRedis()
	defer wg.Done()
	for i, line := range lines {
		s := strings.Split(line, ",")
		x := strings.Split(s[0], " ")
		command := x[1]

		for ik := 0; ik < len(s); ik++ {
			s[ik] = strings.TrimSpace(s[ik])
		}

		data := make([]string, 2)

		data[1] = strings.TrimSpace(x[1])
		data = append(data, s[1:]...)

		transNum := i + 1
		fmt.Println(transNum)
		switch command {

		case "QUOTE":
			//username := data[2]
			//stock := data[3]
			//quote(transNum, username, stock, client)

		case "COMMIT_BUY":
			username := data[2]
			commit_buy(transNum, username, client)

		case "COMMIT_SELL":
			username := data[2]
			commit_sell(transNum, username, client)
			/*
				case "DISPLAY_SUMMARY":
					username := data[2]
					display_summary(transNum, username, client)
			*/
		case "CANCEL_BUY":
			username := data[2]
			cancel_buy(transNum, username, client)

		case "CANCEL_SELL":
			username := data[2]
			cancel_sell(transNum, username, client)

		}

	}
}

func concurrencyLogic(lines []string, username string) {
	defer wg.Done()
	client := dialRedis()
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
			switch command {
			case "ADD":
				amount, _ := strconv.ParseFloat(data[3], 64)
				redisADD(client, username, amount)

			case "BUY":
				symbol := data[3]
				amount, _ := strconv.ParseFloat(data[4], 64)
				redisBUY(client, username, symbol, amount)

			case "SELL":
				symbol := data[3]
				amount, _ := strconv.ParseFloat(data[4], 64)
				redisSELL(client, username, symbol, amount)

			case "QUOTE":
				stock := data[3]
				quote(transNum, username, stock, client)
				//redisQUOTE(client, username, stock)

			case "COMMIT_BUY":
				redisCOMMIT_BUY(client, username)

			case "COMMIT_SELL":
				redisCOMMIT_SELL(client, username)

			case "CANCEL_BUY":
				redisCANCEL_BUY(client, username)

			case "CANCEL_SELL":
				redisCANCEL_SELL(client, username)

			case "SET_BUY_AMOUNT":
				symbol := data[3]
				amount, _ := strconv.ParseFloat(data[4], 64)
				redisSET_BUY_AMOUNT(client, username, symbol, amount)

			case "SET_BUY_TRIGGER":
				symbol := data[3]
				amount, _ := strconv.ParseFloat(data[4], 64)
				redisSET_BUY_TRIGGER(client, username, symbol, amount)

			case "CANCEL_SET_BUY":
				symbol := data[3]
				redisCANCEL_SET_BUY(client, username, symbol)

			case "DISPLAY_SUMMARY":
				//username := data[2]
				//display_summary(transNum, username, client)

			case "SET_SELL_AMOUNT":
				symbol := data[3]
				amount, _ := strconv.ParseFloat(data[4], 64)
				redisSET_SELL_AMOUNT(client, username, symbol, amount)

			case "SET_SELL_TRIGGER":
				symbol := data[3]
				amount, _ := strconv.ParseFloat(data[4], 64)
				redisSET_SELL_TRIGGER(client, username, symbol, amount)

			case "CANCEL_SET_SELL":
				symbol := data[3]
				redisCANCEL_SET_SELL(client, username, symbol)
			}
		}
	}

}
