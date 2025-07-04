package interfaces

import (
	"context"
	"time"
)

// StockMarket - фондовый рынок
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

// Sector - сектор
type Sector struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CompanyInfo - информация о компании
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

// Dividend - дивиденд
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

// CorporateAction - корпоративное действие
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

// Earnings - отчет о прибылях
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

// StockBroker - интерфейс для фондовых брокеров
type StockBroker interface {
	Broker

	// Специфичные методы для фондового рынка
	GetStockMarkets() ([]StockMarket, error)
	GetSectors() ([]Sector, error)

	// Подписка на акции
	SubscribeStocks(ctx context.Context, symbols []string) error

	// Получение информации о компаниях
	GetCompanyInfo(symbol string) (*CompanyInfo, error)

	// Специфичные каналы для фондового рынка
	GetDividendChannel() <-chan Dividend
	GetCorporateActionChannel() <-chan CorporateAction
	GetEarningsChannel() <-chan Earnings
}

// StockBrokerFactory - фабрика для фондовых брокеров
type StockBrokerFactory interface {
	Create(config BrokerConfig) (StockBroker, error)
	GetSupportedBrokers() []string
}
