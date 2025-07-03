package broker

import "time"

// SpotMarket represents a spot market
type SpotMarket struct {
	Symbol      string  `json:"symbol"`
	BaseAsset   string  `json:"base_asset"`
	QuoteAsset  string  `json:"quote_asset"`
	Status      string  `json:"status"`
	MinPrice    float64 `json:"min_price"`
	MaxPrice    float64 `json:"max_price"`
	TickSize    float64 `json:"tick_size"`
	MinQuantity float64 `json:"min_quantity"`
	MaxQuantity float64 `json:"max_quantity"`
	StepSize    float64 `json:"step_size"`
	MinNotional float64 `json:"min_notional"`
}

// FuturesMarket represents a futures market
type FuturesMarket struct {
	Symbol                string    `json:"symbol"`
	BaseAsset             string    `json:"base_asset"`
	QuoteAsset            string    `json:"quote_asset"`
	ContractType          string    `json:"contract_type"` // PERPETUAL, CURRENT_QUARTER, etc.
	DeliveryDate          time.Time `json:"delivery_date"`
	OnboardDate           time.Time `json:"onboard_date"`
	Status                string    `json:"status"`
	MaintMarginPercent    float64   `json:"maint_margin_percent"`
	RequiredMarginPercent float64   `json:"required_margin_percent"`
	BaseAssetPrecision    int       `json:"base_asset_precision"`
	QuotePrecision        int       `json:"quote_precision"`
	UnderlyingType        string    `json:"underlying_type"`
	UnderlyingSubType     []string  `json:"underlying_sub_type"`
	SettlePlan            int       `json:"settle_plan"`
}

// ContractInfo represents detailed information about a futures contract
type ContractInfo struct {
	Symbol                string   `json:"symbol"`
	Status                string   `json:"status"`
	MaintMarginPercent    float64  `json:"maint_margin_percent"`
	RequiredMarginPercent float64  `json:"required_margin_percent"`
	BaseAssetPrecision    int      `json:"base_asset_precision"`
	QuotePrecision        int      `json:"quote_precision"`
	TriggerProtect        float64  `json:"trigger_protect"`
	UnderlyingType        string   `json:"underlying_type"`
	UnderlyingSubType     []string `json:"underlying_sub_type"`
	SettlePlan            int      `json:"settle_plan"`
}

// StockMarket represents a stock market
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

// Sector represents a market sector
type Sector struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CompanyInfo represents detailed information about a company
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

// Dividend represents dividend information
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

// CorporateAction represents a corporate action
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

// Earnings represents an earnings report
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
