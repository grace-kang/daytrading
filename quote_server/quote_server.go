package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

const (
	connHost = "localhost" // Run on the local machine
	connPort = "1200"       // Same port as on the regular system
	connType = "tcp"        // NOTE: not HTPP
)

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":1200", nil)
}

// Handles incoming requests.
func handler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	w.Write([]byte(makeResponse(params["stock"][0], params["user"][0])))
}

func makeResponse(stock string, username string) string {
	now := time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
	amount := fmt.Sprintf("%d.%d", rand.Intn(500)+1, rand.Intn(100))
	crypto := randSeq(25)
	// (?P<quote>.+),(?P<stock>.+),(?P<user>.+),(?P<time>.+),(?P<key>.+)
	output := fmt.Sprintf("%s,%s,%s,%d,%s\n",
	amount,
	stock,
	username,
	now,
	crypto)
	fmt.Println(output)
	return output
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
