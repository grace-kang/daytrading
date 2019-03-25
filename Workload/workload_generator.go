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
	userly := 0
	for u := 0; u < (numUsers + 1); u++ {

		if userS[u] != "./testLOG" && userS[u] != "" {
			//fmt.Println(u, ":", userS[u])
			time.Sleep(150 * time.Millisecond)
			//go concurrencyLogic("http://transaction:1300", lines, userS[u])
/*
			threads := 11
			if u%threads == 0 {
				go concurrencyLogic("http://localhost:1301", lines, userS[u])
			} else if u%threads == 1 {
				go concurrencyLogic("http://localhost:1302", lines, userS[u])
			} else if u%threads == 2 {
				go concurrencyLogic("http://localhost:1303", lines, userS[u])
			} else if u%threads == 3 {
				go concurrencyLogic("http://localhost:1304", lines, userS[u])
			} else if u%threads == 4 {
				go concurrencyLogic("http://localhost:1305", lines, userS[u])
			} else if u%threads == 5 {
				go concurrencyLogic("http://localhost:1306", lines, userS[u])
			} else if u%threads == 6 {
				go concurrencyLogic("http://localhost:1300", lines, userS[u])
			} else if u%threads == 7 {
				go concurrencyLogic("http://localhost:1307", lines, userS[u])
			} else if u%threads == 8 {
				go concurrencyLogic("http://localhost:1308", lines, userS[u])
			} else if u%threads == 9 {
				go concurrencyLogic("http://localhost:1310", lines, userS[u])
			} else if u%threads == 10 {
				go concurrencyLogic("http://localhost:1311", lines, userS[u])
			}
*/
			go concurrencyLogic("http://localhost:80", lines, userS[u])
			userly += 1
			fmt.Println("numUsers: ", userly)

		}
	}
	wg.Wait()
	//wg.Add(1)
	dumpLogFile("http://localhost:1330", "120000", nil, "./testLOG")
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
func concurrencyLogic(address string, lines []string, username string) {
	defer wg.Done()
	// httpclient := http.Client{}
	// client := dialRedis()
	for i, line := range lines {

		s := strings.Split(line, ",")
		x := strings.Split(s[0], " ")
		command := x[1]

		params := strings.Split(line, command+",")[1]

		for ij := 0; ij < len(s); ij++ {
			s[ij] = strings.TrimSpace(s[ij])
		}

		data := make([]string, 2)

		data[1] = strings.TrimSpace(x[1])
		data = append(data, s[1:]...)

		if username == data[2] {
			//time.Sleep(10 * time.Millisecond)

			transNum := i + 1
			transNum_str := strconv.Itoa(transNum)
			fmt.Println(transNum)
			time.Sleep(5 * time.Millisecond)

			client := &http.Client{}
			form := url.Values{
				"transNum": {transNum_str},
				"command": {command},
				"params": {params}}
			req, err := http.NewRequest("POST", address + "/workloadTransaction", strings.NewReader(form.Encode()))
			if err != nil {
				fmt.Println(err)
			}
			req.Host = "web"
			resp, err := client.Do(req)

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			resp.Body.Close()
		}
	}
	fmt.Println("user fin:", username)
}
