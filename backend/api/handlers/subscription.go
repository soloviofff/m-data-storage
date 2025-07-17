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

// SubscriptionHandler handles subscription-related HTTP requests
type SubscriptionHandler struct {
	instrumentManager interfaces.InstrumentManager
	logger            *logger.Logger
}

// NewSubscriptionHandler creates a new SubscriptionHandler instance
func NewSubscriptionHandler(instrumentManager interfaces.InstrumentManager, logger *logger.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		instrumentManager: instrumentManager,
		logger:            logger,
	}
}

// CreateSubscription handles POST /api/v1/subscriptions
func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Failed to decode request body", err.Error())
		return
	}

	// Convert DTO to domain entity
	subscription := entities.InstrumentSubscription{
		ID:        generateSubscriptionID(), // Helper function to generate unique ID
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

	// Add subscription through service
	if err := h.instrumentManager.AddSubscription(r.Context(), subscription); err != nil {
		h.logger.WithFields(map[string]interface{}{
			"subscription": subscription,
			"error":        err.Error(),
		}).Error("Failed to add subscription")
		h.writeError(w, http.StatusInternalServerError, "subscription_error", "Failed to add subscription", err.Error())
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"subscription_id": subscription.ID,
		"symbol":          subscription.Symbol,
		"broker_id":       subscription.BrokerID,
	}).Info("Subscription created successfully")

	h.writeSuccess(w, http.StatusCreated, subscription)
}

// GetSubscription handles GET /api/v1/subscriptions/{id}
func (h *SubscriptionHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subscriptionID := vars["id"]

	if subscriptionID == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Subscription ID is required", "")
		return
	}

	subscription, err := h.instrumentManager.GetSubscription(r.Context(), subscriptionID)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"subscription_id": subscriptionID,
			"error":           err.Error(),
		}).Error("Failed to get subscription")
		h.writeError(w, http.StatusNotFound, "subscription_not_found", "Subscription not found", err.Error())
		return
	}

	h.writeSuccess(w, http.StatusOK, subscription)
}

// ListSubscriptions handles GET /api/v1/subscriptions
func (h *SubscriptionHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	subscriptions, err := h.instrumentManager.ListSubscriptions(r.Context())
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("Failed to list subscriptions")
		h.writeError(w, http.StatusInternalServerError, "subscription_error", "Failed to list subscriptions", err.Error())
		return
	}

	h.writeSuccess(w, http.StatusOK, subscriptions)
}

// UpdateSubscription handles PUT /api/v1/subscriptions/{id}
func (h *SubscriptionHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subscriptionID := vars["id"]

	if subscriptionID == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Subscription ID is required", "")
		return
	}

	var req dto.UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Failed to decode request body", err.Error())
		return
	}

	// Get existing subscription
	existing, err := h.instrumentManager.GetSubscription(r.Context(), subscriptionID)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"subscription_id": subscriptionID,
			"error":           err.Error(),
		}).Error("Failed to get existing subscription")
		h.writeError(w, http.StatusNotFound, "subscription_not_found", "Subscription not found", err.Error())
		return
	}

	// Apply updates
	updated := *existing
	if req.DataTypes != nil {
		updated.DataTypes = *req.DataTypes
	}
	if req.Settings != nil {
		updated.Settings = *req.Settings
	}
	if req.IsActive != nil {
		updated.IsActive = *req.IsActive
	}
	updated.UpdatedAt = time.Now()

	// Validate updated subscription
	if !updated.IsValid() {
		h.writeError(w, http.StatusBadRequest, "invalid_subscription", "Updated subscription validation failed", "")
		return
	}

	// Update subscription
	if err := h.instrumentManager.UpdateSubscription(r.Context(), updated); err != nil {
		h.logger.WithFields(map[string]interface{}{
			"subscription_id": subscriptionID,
			"error":           err.Error(),
		}).Error("Failed to update subscription")
		h.writeError(w, http.StatusInternalServerError, "subscription_error", "Failed to update subscription", err.Error())
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"subscription_id": subscriptionID,
	}).Info("Subscription updated successfully")

	h.writeSuccess(w, http.StatusOK, updated)
}

// DeleteSubscription handles DELETE /api/v1/subscriptions/{id}
func (h *SubscriptionHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subscriptionID := vars["id"]

	if subscriptionID == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Subscription ID is required", "")
		return
	}

	// Check if subscription exists
	_, err := h.instrumentManager.GetSubscription(r.Context(), subscriptionID)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"subscription_id": subscriptionID,
			"error":           err.Error(),
		}).Error("Subscription not found for deletion")
		h.writeError(w, http.StatusNotFound, "subscription_not_found", "Subscription not found", err.Error())
		return
	}

	// Remove subscription
	if err := h.instrumentManager.RemoveSubscription(r.Context(), subscriptionID); err != nil {
		h.logger.WithFields(map[string]interface{}{
			"subscription_id": subscriptionID,
			"error":           err.Error(),
		}).Error("Failed to remove subscription")
		h.writeError(w, http.StatusInternalServerError, "subscription_error", "Failed to remove subscription", err.Error())
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"subscription_id": subscriptionID,
	}).Info("Subscription deleted successfully")

	h.writeSuccess(w, http.StatusOK, map[string]string{
		"message": "Subscription deleted successfully",
		"id":      subscriptionID,
	})
}

// StartTracking handles POST /api/v1/subscriptions/{id}/start
func (h *SubscriptionHandler) StartTracking(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subscriptionID := vars["id"]

	if subscriptionID == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Subscription ID is required", "")
		return
	}

	// Check if subscription exists
	_, err := h.instrumentManager.GetSubscription(r.Context(), subscriptionID)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"subscription_id": subscriptionID,
			"error":           err.Error(),
		}).Error("Subscription not found for tracking start")
		h.writeError(w, http.StatusNotFound, "subscription_not_found", "Subscription not found", err.Error())
		return
	}

	// Start tracking
	if err := h.instrumentManager.StartTracking(r.Context(), subscriptionID); err != nil {
		h.logger.WithFields(map[string]interface{}{
			"subscription_id": subscriptionID,
			"error":           err.Error(),
		}).Error("Failed to start tracking")
		h.writeError(w, http.StatusInternalServerError, "tracking_error", "Failed to start tracking", err.Error())
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"subscription_id": subscriptionID,
	}).Info("Tracking started successfully")

	h.writeSuccess(w, http.StatusOK, map[string]string{
		"message": "Tracking started successfully",
		"id":      subscriptionID,
	})
}

// StopTracking handles POST /api/v1/subscriptions/{id}/stop
func (h *SubscriptionHandler) StopTracking(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subscriptionID := vars["id"]

	if subscriptionID == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Subscription ID is required", "")
		return
	}

	// Check if subscription exists
	_, err := h.instrumentManager.GetSubscription(r.Context(), subscriptionID)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"subscription_id": subscriptionID,
			"error":           err.Error(),
		}).Error("Subscription not found for tracking stop")
		h.writeError(w, http.StatusNotFound, "subscription_not_found", "Subscription not found", err.Error())
		return
	}

	// Stop tracking
	if err := h.instrumentManager.StopTracking(r.Context(), subscriptionID); err != nil {
		h.logger.WithFields(map[string]interface{}{
			"subscription_id": subscriptionID,
			"error":           err.Error(),
		}).Error("Failed to stop tracking")
		h.writeError(w, http.StatusInternalServerError, "tracking_error", "Failed to stop tracking", err.Error())
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"subscription_id": subscriptionID,
	}).Info("Tracking stopped successfully")

	h.writeSuccess(w, http.StatusOK, map[string]string{
		"message": "Tracking stopped successfully",
		"id":      subscriptionID,
	})
}

// Helper functions
func (h *SubscriptionHandler) writeSuccess(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(dto.Response{
		Success: true,
		Data:    data,
	})
}

func (h *SubscriptionHandler) writeError(w http.ResponseWriter, statusCode int, code, message, details string) {
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
