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
	LogType           string  `xml:"userCommand"`
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
	LogType           string `xml:"quoteServer"`
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
	LogType           string `xml:"accountTransaction"`
	Timestamp         int64  `xml:"timestamp"`
	Server            string `xml:"server"`
	TransactionNumber int64  `xml:"transactionNum"`
	Action            string `xml:"action"`
	Username          string `xml:"username"`
	Funds             string `xml:"funds"`
}

type SystemEventType struct {
	LogType           string  `xml:"systemEvent"`
	Timestamp         int64   `xml:"timestamp"`
	Server            string  `xml:"server"`
	TransactionNumber int64   `xml:"transactionNum"`
	Command           Command `xml:"command"`
	Username          string  `xml:"username"`
	StockSymbol       stockSymbolType  `xml:"stockSymbol"`
	Funds             string  `xml:"funds"`
}

type ErrorEventType struct {
	LogType           string  `xml:"errorEvent"`
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
	LogType           string  `xml:"debugEvent"`
	Timestamp         int64   `xml:"timestamp"`
	Server            string  `xml:"server"`
	TransactionNumber int64   `xml:"transactionNum"`
	Command           Command `xml:"command"`
	Username          string  `xml:"username,omitempty"`
	StockSymbol       stockSymbolType  `xml:"stockSymbol,omitempty"`
	Funds             string  `xml:"funds,omitempty"`
	debugMessage      string  `xml:"errorMessage,omitempty"`
}


func (logconn *LogConnection) LogUserCommand(command Command, vars map[string]string) {
	if _, exist := validCommands[command]; exist {
		timestamp := getUnixTimestamp()
		userCommand := UserCommandType{Timestamp: timestamp, Server: server, Command: command}

		if val, exist := vars["trans"]; exist {
			userCommand.TransactionNumber = parseTransactionNumber(val)
		}
		if val, exist := vars["username"]; exist {
			userCommand.Username = val
		}
		if val, exist := vars["symbol"]; exist {
			userCommand.Symbol = val
		}
		if val, exist := vars["filename"]; exist {
			userCommand.Filename = val
		}
		if val, exist := vars["amount"]; exist {
			var err error
			userCommand.Funds, err = formatStrAmount(val)
			if err != nil {
				utils.LogErr(err, "Failed to format amount")
				return
			}
		}

		msg := Message{UserCommand: &userCommand}
		logconn.publishMessage(msg)
	}
}

func (logconn *LogConnection) LogQuoteServ(stockQuote *models.StockQuote, trans string) {
	timestamp := getUnixTimestamp()
	quoteTimeInt, err := strconv.ParseInt(stockQuote.QuoteTimestamp, 10, 64)
	if err != nil {
		utils.LogErr(err, "Failed to parse quote server timestamp")
	}
	tnum := parseTransactionNumber(trans)
	price, err := formatStrAmount(stockQuote.Value)
	if err != nil{
		utils.LogErr(err, "Failed to parse str amount")
	}

	quoteServer := QuoteServerType{Timestamp: timestamp,
		Server:            server,
		QuoteServerTime:   quoteTimeInt,
		Username:          stockQuote.Username,
		Symbol:            stockQuote.Symbol,
		Price:             price,
		CryptoKey:         stockQuote.CrytpoKey,
		TransactionNumber: tnum}

	msg := Message{QuoteServer: &quoteServer}
	logconn.publishMessage(msg)
}

func (logconn *LogConnection) LogTransaction(action string, username string, amount int, trans string) {
	timestamp := getUnixTimestamp()
	tnum := parseTransactionNumber(trans)

	accountTransaction := AccountTransactionType{
		Timestamp:         timestamp,
		Server:            server,
		TransactionNumber: tnum,
		Username:          username,
		Action:            action,
		Funds:             formatAmount(amount),
	}

	msg := Message{AccountTransaction: &accountTransaction}
	logconn.publishMessage(msg)
}

func (logconn *LogConnection) SendDumpLog(filename string, username string) {
	dumpLog := DumpLogType{Filename: filename, Username: username}
	msg := Message{DumpLog: &dumpLog}
	logconn.publishMessage(msg)

}

func (logconn *LogConnection) LogSystemEvent(command Command, username string, stocksymbol string, funds string, trans string) {
	timestamp := getUnixTimestamp()
	tnum := parseTransactionNumber(trans)
	systemEvent := SystemEventType{
		Timestamp:         timestamp,
		Server:            server,
		TransactionNumber: tnum,
		Command:           command,
		Username:          username,
		Symbol:            stocksymbol,
		Funds:             funds,
	}
	msg := Message{SystemEvent: &systemEvent}
	logconn.publishMessage(msg)
}

func (logconn *LogConnection) LogErrorEvent(command Command, vars map[string]string, emessage string) {
	timestamp := getUnixTimestamp()

	errorEvent := ErrorEventType{
		Timestamp:    timestamp,
		Server:       server,
		Command:      command,
		ErrorMessage: emessage}

	if val, exist := vars["trans"]; exist {
		errorEvent.TransactionNumber = parseTransactionNumber(val)
	}
	if val, exist := vars["username"]; exist {
		errorEvent.Username = val
	}
	if val, exist := vars["symbol"]; exist {
		errorEvent.Symbol = val
	}
	if val, exist := vars["amount"]; exist {
		var err error
		errorEvent.Funds, err = formatStrAmount(val)
		if err != nil {
			utils.LogErr(err, "Failed to format amount")
			return
		}
	}

	msg := Message{ErrorEvent: &errorEvent}
	logconn.publishMessage(msg)
}

