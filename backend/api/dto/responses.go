package dto

import (
	"time"
)

// Response represents a generic API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Error represents an API error
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// BrokerResponse represents broker information
type BrokerResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	Connected bool      `json:"connected"`
	LastSeen  time.Time `json:"last_seen"`
}

// MarketDataResponse represents market data response
type MarketDataResponse struct {
	Symbol    string      `json:"symbol"`
	BrokerID  string      `json:"broker_id"`
	DataType  string      `json:"data_type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// LoginResponse represents successful authentication
type LoginResponse struct {
	Token        string   `json:"token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int      `json:"expires_in"`
	RefreshToken string   `json:"refresh_token"`
	Scope        []string `json:"scope"`
}
