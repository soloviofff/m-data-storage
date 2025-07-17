package interfaces

import (
	"context"
	"time"
)

// StockMarket - stock market
type StockMarket struct {
	Symbol            string    `json:"symbol"`
	CompanyName       string    `json:"company_name"`
	Exchange          string    `json:"exchange"`
	Currency          string    `json:"currency"`
	Country           string    `json:"country"`
	Sector            string    `json:"sector"`
	Industry          string    `json:"industry"`
	MarketCap         float64   `json:"market_cap"`
	SharesOutstanding float64   `json:"shares_outstanding"`
	Status            string    `json:"status"`
	ListingDate       time.Time `json:"listing_date"`
}

// Sector - sector
type Sector struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CompanyInfo - company information
type CompanyInfo struct {
	Symbol            string    `json:"symbol"`
	Name              string    `json:"name"`
	Exchange          string    `json:"exchange"`
	Currency          string    `json:"currency"`
	Country           string    `json:"country"`
	Sector            string    `json:"sector"`
	Industry          string    `json:"industry"`
	MarketCap         float64   `json:"market_cap"`
	SharesOutstanding float64   `json:"shares_outstanding"`
	PERatio           float64   `json:"pe_ratio"`
	DividendYield     float64   `json:"dividend_yield"`
	Beta              float64   `json:"beta"`
	EPS               float64   `json:"eps"`
	BookValue         float64   `json:"book_value"`
	Description       string    `json:"description"`
	Website           string    `json:"website"`
	CEO               string    `json:"ceo"`
	Employees         int       `json:"employees"`
	Founded           time.Time `json:"founded"`
	IPODate           time.Time `json:"ipo_date"`
}

// Dividend - dividend
type Dividend struct {
	Symbol      string    `json:"symbol"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	ExDate      time.Time `json:"ex_date"`
	PayDate     time.Time `json:"pay_date"`
	RecordDate  time.Time `json:"record_date"`
	DeclareDate time.Time `json:"declare_date"`
	Frequency   string    `json:"frequency"`
	Type        string    `json:"type"`
	BrokerID    string    `json:"broker_id"`
}

// CorporateAction - corporate action
type CorporateAction struct {
	Symbol      string    `json:"symbol"`
	Type        string    `json:"type"` // SPLIT, MERGER, SPINOFF, etc.
	Ratio       string    `json:"ratio"`
	ExDate      time.Time `json:"ex_date"`
	PayDate     time.Time `json:"pay_date"`
	RecordDate  time.Time `json:"record_date"`
	Description string    `json:"description"`
	BrokerID    string    `json:"broker_id"`
}

// Earnings - earnings report
type Earnings struct {
	Symbol          string    `json:"symbol"`
	Period          string    `json:"period"`
	EPS             float64   `json:"eps"`
	EPSEstimate     float64   `json:"eps_estimate"`
	Revenue         float64   `json:"revenue"`
	RevenueEstimate float64   `json:"revenue_estimate"`
	ReportDate      time.Time `json:"report_date"`
	Time            string    `json:"time"` // BMO, AMC
	BrokerID        string    `json:"broker_id"`
}

// StockBroker - interface for stock brokers
type StockBroker interface {
	Broker

	// Specific methods for stock market
	GetStockMarkets() ([]StockMarket, error)
	GetSectors() ([]Sector, error)

	// Stock subscription
	SubscribeStocks(ctx context.Context, symbols []string) error

	// Company information retrieval
	GetCompanyInfo(symbol string) (*CompanyInfo, error)

	// Specific channels for stock market
	GetDividendChannel() <-chan Dividend
	GetCorporateActionChannel() <-chan CorporateAction
	GetEarningsChannel() <-chan Earnings
}

// StockBrokerFactory - factory for stock brokers
type StockBrokerFactory interface {
	Create(config BrokerConfig) (StockBroker, error)
	GetSupportedBrokers() []string
}
