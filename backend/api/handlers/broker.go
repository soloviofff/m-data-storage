package handlers

import (
	"encoding/json"
	"net/http"

	"m-data-storage/api/dto"
	"m-data-storage/internal/service"
)

// BrokerHandler handles broker-related HTTP requests
type BrokerHandler struct {
	brokerService service.BrokerService
}

// NewBrokerHandler creates a new BrokerHandler instance
func NewBrokerHandler(brokerService service.BrokerService) *BrokerHandler {
	return &BrokerHandler{
		brokerService: brokerService,
	}
}

// AddBroker handles broker addition
func (h *BrokerHandler) AddBroker(w http.ResponseWriter, r *http.Request) {
	var req dto.AddBrokerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid_request", "Failed to decode request body", err.Error())
		return
	}

	if err := h.brokerService.AddBroker(r.Context(), req.Config); err != nil {
		writeError(w, "broker_error", "Failed to add broker", err.Error())
		return
	}

	writeSuccess(w, nil)
}

// ListBrokers handles broker listing
func (h *BrokerHandler) ListBrokers(w http.ResponseWriter, r *http.Request) {
	brokers, err := h.brokerService.ListBrokers(r.Context())
	if err != nil {
		writeError(w, "broker_error", "Failed to list brokers", err.Error())
		return
	}

	writeSuccess(w, brokers)
}

// Subscribe handles market data subscription
func (h *BrokerHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	var req dto.SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid_request", "Failed to decode request body", err.Error())
		return
	}

	if err := h.brokerService.Subscribe(r.Context(), req.BrokerID, req.Subscriptions); err != nil {
		writeError(w, "subscription_error", "Failed to subscribe", err.Error())
		return
	}

	writeSuccess(w, nil)
}

// GetMarketData handles market data retrieval
func (h *BrokerHandler) GetMarketData(w http.ResponseWriter, r *http.Request) {
	var req dto.MarketDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid_request", "Failed to decode request body", err.Error())
		return
	}

	data, err := h.brokerService.GetMarketData(r.Context(), req.BrokerID, req.Symbol)
	if err != nil {
		writeError(w, "market_data_error", "Failed to get market data", err.Error())
		return
	}

	writeSuccess(w, data)
}

// Helper functions for writing responses
func writeSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.Response{
		Success: true,
		Data:    data,
	})
}

func writeError(w http.ResponseWriter, code, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(dto.Response{
		Success: false,
		Error: &dto.Error{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}
