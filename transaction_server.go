//https://www.alexedwards.net/blog/working-with-redis
/*
macowner$ redis-cli
127.0.0.1:6379> flushall
127.0.0.1:6379> hgetall oY01WVirLr
*/

package main

import (
	//"github.com/gomodule/redigo/redis"
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/mediocregopher/radix.v2/redis"
)

var client *redis.Client
var stockPrices = map[string]float64{}
var stocksAmount = map[string]int{}

func dialRedis() *redis.Client {
	cli, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		// handle err
	}
	return cli
}

func main() {
	client := dialRedis()

	lines, err := readLines("workload_files/workload1.txt")
	if err != nil {
		log.Fatalf("readLines: %s", err)
	}
	for i, line := range lines {
		s := strings.Split(line, ",")
		x := strings.Split(s[0], " ")
		transNum := strconv.Itoa(i + 1)
		for i = 0; i < len(s); i++ {
			s[i] = strings.TrimSpace(s[i])
		}

		data := make([]string, 2)
		fmt.Println(transNum)
		data[0] = transNum
		data[1] = strings.TrimSpace(x[1])
		data = append(data, s[1:]...)
		// ParseCommandData(data)

		switch x[1] {
		case "ADD":
			fmt.Println("-----ADD-----")

			/* Check to see if User already exists. Add User if not
			else just increase an existing User's balance.
			s[1] is the user id
			s[2] is the amount they wish to add. */

			exists, _ := client.Cmd("HGETALL", s[1]).Map()

			string1 := s[2]
			string1 = strings.TrimSpace(string1)
			dollar, _ := strconv.ParseFloat(string1, 64)

			if len(exists) == 0 {
				client.Cmd("HMSET", s[1], "User", s[1], "Balance", dollar)
			} else {
				client.Cmd("HINCRBYFLOAT", s[1], "Balance", dollar)
			}
			/* ------------------------------------*/

			/* Display - get new balance (HGET) */
			fmt.Print("ADD:	  ", dollar)
			x, _ := client.Cmd("HGET", s[1], "Balance").Float64()
			fmt.Println("	Balance: ", x)

		case "BUY":
			fmt.Println("-----BUY-----")
			string2 := s[2] + ":BUY"
			string3 := strings.TrimSpace(s[3])
			dollar, _ := strconv.ParseFloat(string3, 64)

			/* HSET: set the buy amount in dollars for the chosen stock
			(still needs to be committed to purchase) */
			client.Cmd("HSET", s[1], string2, dollar)
			fmt.Println("BUY:	", dollar)

		case "SELL":
			fmt.Println("-----SELL-----")
			string2 := s[2] + ":SELL"

			string3 := strings.TrimSpace(s[3])
			dollar, _ := strconv.ParseFloat(string3, 64)

			/* HSET: set the sell amount in dollars for the chosen stock
			(still needs to be committed to make sale) */
			client.Cmd("HSET", s[1], string2, dollar)
			fmt.Println("SELL:	", dollar)

		case "QUOTE":
			fmt.Println("-----QUOTE-----")
			req, err := http.NewRequest("GET", "http://localhost:1200", nil)
			req.Header.Add("If-None-Match", `W/"wyzzy"`)

			q := req.URL.Query()
			q.Add("user", s[1])
			q.Add("stock", s[2])
			q.Add("transNum", transNum)
			req.URL.RawQuery = q.Encode()

			httpclient := http.Client{}

			var resp *http.Response
			for {
				resp, err = httpclient.Do(req)

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

			// fmt.Println(string(body))
			split := strings.Split(string(body), ",")[0]
			price, _ := strconv.ParseFloat(split, 64)
			fmt.Println(price)
			resp.Body.Close()

			/* HINCRBYFLOAT: change a float value. Quote costs a User
			$0.50 */
			client.Cmd("HINCRBYFLOAT", s[1], "Balance", -0.50)

			/* Display - HGET new balance for display */
			fmt.Print("QUOTE:	", s[2])
			x, _ := client.Cmd("HGET", s[1], "Balance").Float64()
			fmt.Println("	Balance: ", x)

		case "COMMIT_BUY":
			fmt.Println("-----COMMIT_BUY-----")

			/* HGET dollar amount from stock BUY action. */
			string3 := strings.TrimSpace(s[1])
			x, _ := client.Cmd("HGET", string3, "S:BUY").Float64()

			/* Stock price is hardcoded to $22.00 for testing.
			Calculate how many units User can buy with the amount
			set aside for purchase. */
			price := 22.0
			total := x / price
			units := int(total)
			final := float64(units) * price

			/* Decrease balance by price */
			client.Cmd("HINCRBYFLOAT", string3, "Balance", -final)

			/* get new balance for Display, error checking */
			y, _ := client.Cmd("HGET", string3, "Balance").Float64()
			fmt.Println("COMMIT_BUY: ", final, "Balance: ", y)

			//relevant := s[2] + ":Number"

			/* HINCBRY: Increase the number of stocks a User owns
			HGET: the number for display
			Display... */
			client.Cmd("HINCRBY", string3, "S:Number", units)
			a, _ := client.Cmd("HGET", string3, "S:Number").Float64()
			fmt.Println("STOCK(S): ", units, "TOTAL(S): ", a)

		case "COMMIT_SELL":
			fmt.Println("-----COMMIT_SELL-----")
			string5 := strings.TrimSpace(s[1])

			/* HGET: get dollar amount stock SELL action */
			be, _ := client.Cmd("HGET", string5, "S:SELL").Float64()

			/* Calculate how many stocks User can sell */
			numU := be / 22.0
			numUnits := int(numU)
			fmt.Println("COMMIT_SELL: ", numUnits)
			cost := numUnits * 22
			fmt.Println("AT COST: ", cost)

			/* HINCRBY: Decrease User's stocks and then Display # */
			client.Cmd("HINCRBY", string5, "S:Number", -numUnits)
			ab, _ := client.Cmd("HGET", string5, "S:Number").Float64()
			fmt.Println("STOCK(S): ", ab)

			/* HGET: Decrease User's balance and then display new balance */
			client.Cmd("HINCRBYFLOAT", string5, "Balance", cost)
			za, _ := client.Cmd("HGET", string5, "Balance").Float64()
			fmt.Println("Balance: ", za)

		case "DISPLAY_SUMMARY":
			/* TODO: Not implemented yet, Display User's transaction history */
			fmt.Println("-----DISPLAY_SUMMARY-----")

		case "CANCEL_BUY":
			fmt.Println("-----CANCEL_BUY-----")

			/* HSET: Cancel stock BUY amount
			Display new value ex. S:BUY should equal 0 now */
			string7 := strings.TrimSpace(s[1])
			client.Cmd("HSET", string7, "S:BUY", 0)
			zas, _ := client.Cmd("HGET", string7, "S:BUY").Float64()
			fmt.Println("BUY: ", zas)

		case "CANCEL_SELL":
			fmt.Println("-----CANCEL_SELL-----")

			/* HSET: Cancel stock SELL amount
			Display new value ex. S:SELL should equal 0 now */
			string9 := strings.TrimSpace(s[1])
			client.Cmd("HSET", string9, "S:SELL", 0)
			zps, _ := client.Cmd("HGET", string9, "S:SELL").Float64()
			fmt.Println("SELL: ", zps)

		case "SET_BUY_AMOUNT":
			fmt.Println("-----SET_BUY_AMOUNT-----")
			string10 := strings.TrimSpace(s[1])
			string11 := s[2] + ":TBUYAMOUNT"
			string13 := strings.TrimSpace(s[3])
			dollar2, _ := strconv.ParseFloat(string13, 64)

			/* HSET: Amount of money set aside for Buy Trigger to be activated */
			client.Cmd("HSET", string10, string11, dollar2)
			fmt.Println("TBUYAMOUNT:	", dollar2)

			/* HINCRBYFLOAT: Decrease User's Balance by amount set aside, Display */
			client.Cmd("HINCRBYFLOAT", string10, "Balance", -dollar2)
			zazz, _ := client.Cmd("HGET", string10, "Balance").Float64()
			fmt.Println("Balance: ", zazz)

		case "SET_BUY_TRIGGER":
			fmt.Println("-----SET_BUY_TRIGGER-----")
			string14 := strings.TrimSpace(s[1])
			string15 := s[2] + ":TBUYTRIG"

			/* HSET: Set Stock price for when the Buy Trigger will be activated */
			string16 := strings.TrimSpace(s[3])
			dollar3, _ := strconv.ParseFloat(string16, 64)
			client.Cmd("HSET", string14, string15, dollar3)
			fmt.Println("TBUYTRIG:	", dollar3)

		case "CANCEL_SET_BUY":
			fmt.Println("-----CANCEL_SET_BUY-----")
			string1000 := strings.TrimSpace(s[1])
			stringx00 := strings.TrimSpace(s[2])
			string115 := stringx00 + ":TBUYAMOUNT"

			/* HGET: Get amount stored in reserve in STOCK:TBUYAMOUNT */
			zzz, _ := client.Cmd("HGET", string1000, string115).Float64()
			fmt.Println("Refund: ", zzz)

			/* TODO: Refund balance by reserve stored from above */

		case "SET_SELL_AMOUNT":
			fmt.Println("-----SET_SELL_AMOUNT-----")
			string100 := strings.TrimSpace(s[1])
			string110 := s[2] + ":TSELLAMOUNT"

			string130 := strings.TrimSpace(s[3])
			dollar200, _ := strconv.ParseFloat(string130, 64)
			client.Cmd("HSET", string100, string110, dollar200)
			fmt.Println("TSELLAMOUNT:	", dollar200)

		case "SET_SELL_TRIGGER":
			/* TODO */

		case "CANCEL_SET_SELL":
			/* TODO */
		}
	}
	/* How to put a map straight into Redis
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	err = client.Cmd("HMSET", "user:4", "user", "bob", "balance", 5000, m).Err */

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
