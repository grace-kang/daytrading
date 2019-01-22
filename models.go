package main

type TransType string

type User struct {
	username     string `json:"username"`
	balance int `json:"balance"`
}

const (
	Trans_BUY = TransType("BUY")
	Trans_SELL = TransType("SELL")
)

type Transaction struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	StockSymbol  string    `json:"stock"`
	Type         TransType `json:"type"`
	Amount       int       `json:"amount"`
	Cost 		 int       `json:"cost"`
	Time         int64     `json:"time"`
}

type Trigger struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	StockSymbol  string    `json:"stock"`
	Order        TransType `json:"type"`
	Amount       int       `json:"amount"`
	TriggerPrice int       `json:"triggerprice"`
	Executable   bool      `json:"executable"`
	Time         int64     `json:"time"`
}

