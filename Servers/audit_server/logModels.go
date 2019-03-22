package main

import "encoding/xml"

type Log struct {
	XMLName xml.Name `xml:"log"`
	LogData []LogType
}

type Command string
type stockSymbolType string

func (log *Log) append(logType LogType) {
	log.LogData = append(log.LogData, logType)
}

type LogType struct {
	UserCommand        *UserCommandType        `xml:"userCommand"`
	QuoteServer        *QuoteServerType        `xml:"quoteServer"`
	AccountTransaction *AccountTransactionType `xml:"accountTransaction"`
	SystemEvent        *SystemEventType        `xml:"systemEvent"`
	ErrorEvent         *ErrorEventType         `xml:"errorEvent"`
	DebugEvent         *DebugType              `xml:"debugEvent"`
}

type UserCommandType struct {
	XMLName           xml.Name        `xml:"userCommand"`
	Timestamp         int64           `xml:"timestamp"`
	Server            string          `xml:"server"`
	TransactionNumber int             `xml:"transactionNum"`
	Command           Command         `xml:"command"`
	Username          string          `xml:"username,omitempty"`
	StockSymbol       stockSymbolType `xml:"stockSymbol,omitempty"`
	Filename          string          `xml:"filename,omitempty"`
	Funds             string          `xml:"funds,omitempty"`
}

type QuoteServerType struct {
	XMLName           xml.Name        `xml:"quoteServer"`
	Timestamp         int64           `xml:"timestamp"`
	Server            string          `xml:"server"`
	TransactionNumber int             `xml:"transactionNum"`
	Price             string          `xml:"price"`
	StockSymbol       stockSymbolType `xml:"stockSymbol"`
	Username          string          `xml:"username"`
	QuoteServerTime   int64           `xml:"quoteServerTime"`
	CryptoKey         string          `xml:"cryptokey"`
}

type AccountTransactionType struct {
	XMLName           xml.Name `xml:"accountTransaction"`
	Timestamp         int64    `xml:"timestamp"`
	Server            string   `xml:"server"`
	TransactionNumber int      `xml:"transactionNum"`
	Action            string   `xml:"action"`
	Username          string   `xml:"username"`
	Funds             string   `xml:"funds"`
}

type SystemEventType struct {
	XMLName           xml.Name        `xml:"systemEvent"`
	Timestamp         int64           `xml:"timestamp"`
	Server            string          `xml:"server"`
	TransactionNumber int             `xml:"transactionNum"`
	Command           Command         `xml:"command"`
	Username          string          `xml:"username,omitempty"`
	StockSymbol       stockSymbolType `xml:"stockSymbol,omitempty"`
	Funds             string          `xml:"funds,omitempty"`
	Filename          string          `xml:"filename,omitempty"`
}

type ErrorEventType struct {
	XMLName           xml.Name        `xml:"errorEvent"`
	Timestamp         int64           `xml:"timestamp"`
	Server            string          `xml:"server"`
	TransactionNumber int             `xml:"transactionNum"`
	Command           Command         `xml:"command"`
	Username          string          `xml:"username,omitempty"`
	StockSymbol       stockSymbolType `xml:"stockSymbol,omitempty"`
	Filename          string          `xml:"filename,omitempty"`
	Funds             string          `xml:"funds,omitempty"`
	ErrorMessage      string          `xml:"errorMessage,omitempty"`
}

type DebugType struct {
	XMLName           xml.Name        `xml:"debugEvent"`
	Timestamp         int64           `xml:"timestamp"`
	Server            string          `xml:"server"`
	TransactionNumber int             `xml:"transactionNum"`
	Command           Command         `xml:"command"`
	Username          string          `xml:"username,omitempty"`
	StockSymbol       stockSymbolType `xml:"stockSymbol,omitempty"`
	Filename          string          `xml:"filename,omitempty"`
	Funds             string          `xml:"funds,omitempty"`
	DebugMessage      string          `xml:"errorMessage,omitempty"`
}

func (cd *LogType) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if err := e.EncodeElement(cd.UserCommand, xml.StartElement{Name: xml.Name{Local: "userCommand"}}); err != nil {
		return err
	}
	if err := e.EncodeElement(cd.QuoteServer, xml.StartElement{Name: xml.Name{Local: "quoteServer"}}); err != nil {
		return err
	}
	if err := e.EncodeElement(cd.AccountTransaction, xml.StartElement{Name: xml.Name{Local: "accountTransaction"}}); err != nil {
		return err
	}
	if err := e.EncodeElement(cd.SystemEvent, xml.StartElement{Name: xml.Name{Local: "systemEvent"}}); err != nil {
		return err
	}
	if err := e.EncodeElement(cd.ErrorEvent, xml.StartElement{Name: xml.Name{Local: "errorEvent"}}); err != nil {
		return err
	}
	if err := e.EncodeElement(cd.DebugEvent, xml.StartElement{Name: xml.Name{Local: "debugEvent"}}); err != nil {
		return err
	}
	return nil
}
