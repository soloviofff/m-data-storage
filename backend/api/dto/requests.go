package dto

import (
	"m-data-storage/pkg/broker"
	"time"
)

// AddBrokerRequest represents request to add a new broker
type AddBrokerRequest struct {
	Config broker.BrokerConfig `json:"config" validate:"required"`
}

// SubscriptionRequest represents request to subscribe to market data
type SubscriptionRequest struct {
	BrokerID      string                          `json:"broker_id" validate:"required"`
	Subscriptions []broker.InstrumentSubscription `json:"subscriptions" validate:"required,min=1"`
}

// MarketDataRequest represents request to get market data
type MarketDataRequest struct {
	BrokerID string     `json:"broker_id" validate:"required"`
	Symbol   string     `json:"symbol" validate:"required"`
	From     *time.Time `json:"from,omitempty"`
	To       *time.Time `json:"to,omitempty"`
	Limit    int        `json:"limit,omitempty" validate:"omitempty,min=1,max=1000"`
}

// UpdateConfigRequest represents request to update system configuration
type UpdateConfigRequest struct {
	Config SystemConfig `json:"config" validate:"required"`
}

// LoginRequest represents authentication request
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}
