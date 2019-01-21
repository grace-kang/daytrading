package main




type Command string
type stockSymbolType string
var server = "server1" // need to be replaced later
var log_file = "logFile" // need to be replaced later

const (
	ADD              = Command("ADD")
	QUOTE            = Command("QUOTE")
	BUY              = Command("BUY")
	COMMIT_BUY       = Command("COMMIT_BUY")
	CANCEL_BUY       = Command("CANCEL_BUY")
	SELL             = Command("SELL")
	COMMIT_SELL      = Command("COMMIT_SELL")
	CANCEL_SELL      = Command("CANCEL_SELL")
	SET_BUY_AMOUNT   = Command("SET_BUY_AMOUNT")
	CANCEL_SET_BUY   = Command("CANCEL_SET_BUY")
	SET_BUY_TRIGGER  = Command("SET_BUY_TRIGGER")
	SET_SELL_AMOUNT  = Command("SET_SELL_AMOUNT")
	SET_SELL_TRIGGER = Command("SET_SELL_TRIGGER")
	CANCEL_SET_SELL  = Command("CANCEL_SET_SELL")
	DUMPLOG          = Command("DUMPLOG")
	DISPLAY_SUMMARY  = Command("DISPLAY_SUMMARY")
)

type LogItem struct {
	Username string
	LogData string 
}

type UserCommandType struct {
	XMLName   		  xml.Name  `xml:"userCommand"`
	Timestamp         int64   `xml:"timestamp"`
	Server            string  `xml:"server"`
	TransactionNumber int64   `xml:"transactionNum"`
	Command           Command `xml:"command"`
	Username          string  `xml:"username,omitempty"`
	StockSymbol       string  `xml:"stockSymbol,omitempty"`
	Filename          string  `xml:"filename,omitempty"`
	Funds             string  `xml:"funds,omitempty"`
}

type QuoteServerType struct {
	XMLName   		  xml.Name `xml:"quoteServer"`
	Timestamp         int64  `xml:"timestamp"`
	Server            string `xml:"server"`
	TransactionNumber int64  `xml:"transactionNum"`
	Price             string `xml:"price"`
	StockSymbol       stockSymbolType `xml:"stockSymbol"`
	Username          string `xml:"username"`
	QuoteServerTime   int64  `xml:"quoteServerTime"`
	CryptoKey         string `xml:"cryptokey"`
}

type AccountTransactionType struct {
	XMLName   		  xml.Name `xml:"accountTransaction"`
	Timestamp         int64  `xml:"timestamp"`
	Server            string `xml:"server"`
	TransactionNumber int64  `xml:"transactionNum"`
	Action            string `xml:"action"`
	Username          string `xml:"username"`
	Funds             string `xml:"funds"`
}

type SystemEventType struct {
	XMLName           xml.Name  `xml:"systemEvent"`
	Timestamp         int64   `xml:"timestamp"`
	Server            string  `xml:"server"`
	TransactionNumber int64   `xml:"transactionNum"`
	Command           Command `xml:"command"`
	Username          string  `xml:"username"`
	StockSymbol       stockSymbolType  `xml:"stockSymbol"`
	Funds             string  `xml:"funds"`
}

type ErrorEventType struct {
	XMLName           xml.Name  `xml:"errorEvent"`
	Timestamp         int64   `xml:"timestamp"`
	Server            string  `xml:"server"`
	TransactionNumber int64   `xml:"transactionNum"`
	Command           Command `xml:"command"`
	Username          string  `xml:"username,omitempty"`
	StockSymbol       stockSymbolType  `xml:"stockSymbol,omitempty"`
	Funds             string  `xml:"funds,omitempty"`
	ErrorMessage      string  `xml:"errorMessage,omitempty"`
}

type DebugType struct {
	XMLName           xml.Name  `xml:"debugEvent"`
	Timestamp         int64   `xml:"timestamp"`
	Server            string  `xml:"server"`
	TransactionNumber int64   `xml:"transactionNum"`
	Command           Command `xml:"command"`
	Username          string  `xml:"username,omitempty"`
	StockSymbol       stockSymbolType  `xml:"stockSymbol,omitempty"`
	Funds             string  `xml:"funds,omitempty"`
	debugMessage      string  `xml:"errorMessage,omitempty"`
}


