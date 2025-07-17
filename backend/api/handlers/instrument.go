package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"m-data-storage/api/dto"
	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
	"m-data-storage/internal/infrastructure/logger"
)

// InstrumentHandler handles instrument-related HTTP requests
type InstrumentHandler struct {
	instrumentManager interfaces.InstrumentManager
	logger           *logger.Logger
}

// NewInstrumentHandler creates a new InstrumentHandler instance
func NewInstrumentHandler(instrumentManager interfaces.InstrumentManager, logger *logger.Logger) *InstrumentHandler {
	return &InstrumentHandler{
		instrumentManager: instrumentManager,
		logger:           logger,
	}
}

// CreateInstrument handles POST /api/v1/instruments
func (h *InstrumentHandler) CreateInstrument(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateInstrumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Failed to decode request body", err.Error())
		return
	}

	// Convert DTO to domain entity
	instrument := entities.InstrumentInfo{
		Symbol:            req.Symbol,
		BaseAsset:         req.BaseAsset,
		QuoteAsset:        req.QuoteAsset,
		Type:              req.Type,
		Market:            req.Market,
		IsActive:          req.IsActive,
		MinPrice:          req.MinPrice,
		MaxPrice:          req.MaxPrice,
		MinQuantity:       req.MinQuantity,
		MaxQuantity:       req.MaxQuantity,
		PricePrecision:    req.PricePrecision,
		QuantityPrecision: req.QuantityPrecision,
	}

	// Validate instrument
	if !instrument.IsValid() {
		h.writeError(w, http.StatusBadRequest, "invalid_instrument", "Instrument validation failed", "")
		return
	}

	// Add instrument
	if err := h.instrumentManager.AddInstrument(r.Context(), instrument); err != nil {
		h.logger.WithFields(map[string]interface{}{
			"symbol": req.Symbol,
			"error":  err.Error(),
		}).Error("Failed to add instrument")
		h.writeError(w, http.StatusInternalServerError, "instrument_error", "Failed to add instrument", err.Error())
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"symbol": req.Symbol,
		"type":   req.Type,
		"market": req.Market,
	}).Info("Instrument created successfully")

	h.writeSuccess(w, http.StatusCreated, instrument)
}

// GetInstrument handles GET /api/v1/instruments/{symbol}
func (h *InstrumentHandler) GetInstrument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]

	if symbol == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Symbol is required", "")
		return
	}

	instrument, err := h.instrumentManager.GetInstrument(r.Context(), symbol)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"symbol": symbol,
			"error":  err.Error(),
		}).Error("Failed to get instrument")
		h.writeError(w, http.StatusNotFound, "instrument_not_found", "Instrument not found", err.Error())
		return
	}

	h.writeSuccess(w, http.StatusOK, instrument)
}

// ListInstruments handles GET /api/v1/instruments
func (h *InstrumentHandler) ListInstruments(w http.ResponseWriter, r *http.Request) {
	instruments, err := h.instrumentManager.ListInstruments(r.Context())
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("Failed to list instruments")
		h.writeError(w, http.StatusInternalServerError, "instrument_error", "Failed to list instruments", err.Error())
		return
	}

	h.writeSuccess(w, http.StatusOK, instruments)
}

// UpdateInstrument handles PUT /api/v1/instruments/{symbol}
func (h *InstrumentHandler) UpdateInstrument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]

	if symbol == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Symbol is required", "")
		return
	}

	var req dto.UpdateInstrumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Failed to decode request body", err.Error())
		return
	}

	// Get existing instrument
	existing, err := h.instrumentManager.GetInstrument(r.Context(), symbol)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"symbol": symbol,
			"error":  err.Error(),
		}).Error("Failed to get existing instrument")
		h.writeError(w, http.StatusNotFound, "instrument_not_found", "Instrument not found", err.Error())
		return
	}

	// Apply updates
	updated := *existing
	if req.BaseAsset != nil {
		updated.BaseAsset = *req.BaseAsset
	}
	if req.QuoteAsset != nil {
		updated.QuoteAsset = *req.QuoteAsset
	}
	if req.Type != nil {
		updated.Type = *req.Type
	}
	if req.Market != nil {
		updated.Market = *req.Market
	}
	if req.IsActive != nil {
		updated.IsActive = *req.IsActive
	}
	if req.MinPrice != nil {
		updated.MinPrice = *req.MinPrice
	}
	if req.MaxPrice != nil {
		updated.MaxPrice = *req.MaxPrice
	}
	if req.MinQuantity != nil {
		updated.MinQuantity = *req.MinQuantity
	}
	if req.MaxQuantity != nil {
		updated.MaxQuantity = *req.MaxQuantity
	}
	if req.PricePrecision != nil {
		updated.PricePrecision = *req.PricePrecision
	}
	if req.QuantityPrecision != nil {
		updated.QuantityPrecision = *req.QuantityPrecision
	}

	// Validate updated instrument
	if !updated.IsValid() {
		h.writeError(w, http.StatusBadRequest, "invalid_instrument", "Updated instrument validation failed", "")
		return
	}

	// Update instrument
	if err := h.instrumentManager.AddInstrument(r.Context(), updated); err != nil {
		h.logger.WithFields(map[string]interface{}{
			"symbol": symbol,
			"error":  err.Error(),
		}).Error("Failed to update instrument")
		h.writeError(w, http.StatusInternalServerError, "instrument_error", "Failed to update instrument", err.Error())
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"symbol": symbol,
	}).Info("Instrument updated successfully")

	h.writeSuccess(w, http.StatusOK, updated)
}

// DeleteInstrument handles DELETE /api/v1/instruments/{symbol}
func (h *InstrumentHandler) DeleteInstrument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]

	if symbol == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Symbol is required", "")
		return
	}

	// Check if instrument exists
	_, err := h.instrumentManager.GetInstrument(r.Context(), symbol)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"symbol": symbol,
			"error":  err.Error(),
		}).Error("Instrument not found for deletion")
		h.writeError(w, http.StatusNotFound, "instrument_not_found", "Instrument not found", err.Error())
		return
	}

	// Note: We need to implement DeleteInstrument in the InstrumentManager interface
	// For now, we'll return a not implemented error
	h.writeError(w, http.StatusNotImplemented, "not_implemented", "Delete instrument not yet implemented", "")
}

// CreateSubscription handles POST /api/v1/instruments/{symbol}/subscriptions
func (h *InstrumentHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]

	if symbol == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Symbol is required", "")
		return
	}

	var req dto.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Failed to decode request body", err.Error())
		return
	}

	// Ensure symbol matches URL parameter
	if req.Symbol != symbol {
		h.writeError(w, http.StatusBadRequest, "symbol_mismatch", "Symbol in URL and body must match", "")
		return
	}

	// Convert DTO to domain entity
	subscription := entities.InstrumentSubscription{
		ID:        generateSubscriptionID(),
		Symbol:    req.Symbol,
		Type:      req.Type,
		Market:    req.Market,
		DataTypes: req.DataTypes,
		StartDate: req.StartDate,
		Settings:  req.Settings,
		BrokerID:  req.BrokerID,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Validate subscription
	if !subscription.IsValid() {
		h.writeError(w, http.StatusBadRequest, "invalid_subscription", "Subscription validation failed", "")
		return
	}

	// Add subscription
	if err := h.instrumentManager.AddSubscription(r.Context(), subscription); err != nil {
		h.logger.WithFields(map[string]interface{}{
			"symbol":    req.Symbol,
			"broker_id": req.BrokerID,
			"error":     err.Error(),
		}).Error("Failed to add subscription")
		h.writeError(w, http.StatusInternalServerError, "subscription_error", "Failed to add subscription", err.Error())
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"symbol":         req.Symbol,
		"broker_id":      req.BrokerID,
		"subscription_id": subscription.ID,
	}).Info("Subscription created successfully")

	h.writeSuccess(w, http.StatusCreated, subscription)
}

// Helper functions
func (h *InstrumentHandler) writeSuccess(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(dto.Response{
		Success: true,
		Data:    data,
	})
}

func (h *InstrumentHandler) writeError(w http.ResponseWriter, statusCode int, code, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(dto.Response{
		Success: false,
		Error: &dto.Error{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// generateSubscriptionID generates a unique subscription ID
func generateSubscriptionID() string {
	return "sub_" + time.Now().Format("20060102150405") + "_" + randomString(8)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
