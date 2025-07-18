package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"m-data-storage/api/dto"
	"m-data-storage/internal/domain/interfaces"
	"m-data-storage/internal/infrastructure/logger"
)

// DataHandler handles data retrieval endpoints
type DataHandler struct {
	dataQuery interfaces.DataQuery
	logger    *logger.Logger
}

// NewDataHandler creates a new data handler
func NewDataHandler(dataQuery interfaces.DataQuery, logger *logger.Logger) *DataHandler {
	return &DataHandler{
		dataQuery: dataQuery,
		logger:    logger,
	}
}

// GetTickers handles GET /api/v1/data/tickers
func (h *DataHandler) GetTickers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		h.writeError(w, http.StatusBadRequest, "missing_parameter", "Symbol parameter is required", "")
		return
	}

	// Parse time range
	var from, to *time.Time
	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		if parsedFrom, err := time.Parse(time.RFC3339, fromStr); err != nil {
			h.writeError(w, http.StatusBadRequest, "invalid_parameter", "Invalid from time format", err.Error())
			return
		} else {
			from = &parsedFrom
		}
	}

	if toStr := r.URL.Query().Get("to"); toStr != "" {
		if parsedTo, err := time.Parse(time.RFC3339, toStr); err != nil {
			h.writeError(w, http.StatusBadRequest, "invalid_parameter", "Invalid to time format", err.Error())
			return
		} else {
			to = &parsedTo
		}
	}

	// Parse limit
	limit := 100 // default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err != nil {
			h.writeError(w, http.StatusBadRequest, "invalid_parameter", "Invalid limit format", err.Error())
			return
		} else if parsedLimit < 1 || parsedLimit > 1000 {
			h.writeError(w, http.StatusBadRequest, "invalid_parameter", "Limit must be between 1 and 1000", "")
			return
		} else {
			limit = parsedLimit
		}
	}

	// Create filter
	filter := interfaces.TickerFilter{
		Symbols:   []string{symbol},
		StartTime: from,
		EndTime:   to,
		Limit:     limit,
	}

	// Get tickers from data query service
	tickers, err := h.dataQuery.GetTickers(r.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get tickers")
		h.writeError(w, http.StatusInternalServerError, "data_error", "Failed to retrieve ticker data", err.Error())
		return
	}

	// Convert to response DTOs
	tickerResponses := make([]dto.TickerResponse, len(tickers))
	for i, ticker := range tickers {
		tickerResponses[i] = dto.TickerResponse{
			Symbol:    ticker.Symbol,
			Price:     ticker.Price,
			BidPrice:  ticker.BidPrice,
			AskPrice:  ticker.AskPrice,
			Volume:    ticker.Volume,
			Change:    ticker.Change,
			ChangeP:   ticker.ChangePercent,
			High:      ticker.High24h,
			Low:       ticker.Low24h,
			Open:      ticker.PrevClose24h, // Use previous close as open approximation
			Close:     ticker.Price,        // Current price as close
			Timestamp: ticker.Timestamp,
		}
	}

	response := dto.TickerListResponse{
		Tickers: tickerResponses,
		Total:   len(tickerResponses),
		From:    from,
		To:      to,
	}

	h.writeSuccess(w, http.StatusOK, response)
}

// GetCandles handles GET /api/v1/data/candles
func (h *DataHandler) GetCandles(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		h.writeError(w, http.StatusBadRequest, "missing_parameter", "Symbol parameter is required", "")
		return
	}

	timeframe := r.URL.Query().Get("timeframe")
	if timeframe == "" {
		timeframe = "1m" // default timeframe
	}

	// Parse time range
	var from, to *time.Time
	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		if parsedFrom, err := time.Parse(time.RFC3339, fromStr); err != nil {
			h.writeError(w, http.StatusBadRequest, "invalid_parameter", "Invalid from time format", err.Error())
			return
		} else {
			from = &parsedFrom
		}
	}

	if toStr := r.URL.Query().Get("to"); toStr != "" {
		if parsedTo, err := time.Parse(time.RFC3339, toStr); err != nil {
			h.writeError(w, http.StatusBadRequest, "invalid_parameter", "Invalid to time format", err.Error())
			return
		} else {
			to = &parsedTo
		}
	}

	// Parse limit
	limit := 100 // default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err != nil {
			h.writeError(w, http.StatusBadRequest, "invalid_parameter", "Invalid limit format", err.Error())
			return
		} else if parsedLimit < 1 || parsedLimit > 1000 {
			h.writeError(w, http.StatusBadRequest, "invalid_parameter", "Limit must be between 1 and 1000", "")
			return
		} else {
			limit = parsedLimit
		}
	}

	// Create filter
	filter := interfaces.CandleFilter{
		Symbols:    []string{symbol},
		Timeframes: []string{timeframe},
		StartTime:  from,
		EndTime:    to,
		Limit:      limit,
	}

	// Get candles from data query service
	candles, err := h.dataQuery.GetCandles(r.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get candles")
		h.writeError(w, http.StatusInternalServerError, "data_error", "Failed to retrieve candle data", err.Error())
		return
	}

	// Convert to response DTOs
	candleResponses := make([]dto.CandleResponse, len(candles))
	for i, candle := range candles {
		candleResponses[i] = dto.CandleResponse{
			Symbol:    candle.Symbol,
			Timeframe: candle.Timeframe,
			Open:      candle.Open,
			High:      candle.High,
			Low:       candle.Low,
			Close:     candle.Close,
			Volume:    candle.Volume,
			Timestamp: candle.Timestamp,
		}
	}

	response := dto.CandleListResponse{
		Candles: candleResponses,
		Total:   len(candleResponses),
		From:    from,
		To:      to,
	}

	h.writeSuccess(w, http.StatusOK, response)
}

// GetOrderBooks handles GET /api/v1/data/orderbooks
func (h *DataHandler) GetOrderBooks(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		h.writeError(w, http.StatusBadRequest, "missing_parameter", "Symbol parameter is required", "")
		return
	}

	// Parse depth
	depth := 20 // default depth
	if depthStr := r.URL.Query().Get("depth"); depthStr != "" {
		if parsedDepth, err := strconv.Atoi(depthStr); err != nil {
			h.writeError(w, http.StatusBadRequest, "invalid_parameter", "Invalid depth format", err.Error())
			return
		} else if parsedDepth < 1 || parsedDepth > 100 {
			h.writeError(w, http.StatusBadRequest, "invalid_parameter", "Depth must be between 1 and 100", "")
			return
		} else {
			depth = parsedDepth
		}
	}

	// Create filter
	filter := interfaces.OrderBookFilter{
		Symbols: []string{symbol},
		Limit:   depth,
	}

	// Get order books from data query service
	orderBooks, err := h.dataQuery.GetOrderBooks(r.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get order books")
		h.writeError(w, http.StatusInternalServerError, "data_error", "Failed to retrieve order book data", err.Error())
		return
	}

	// Convert to response DTOs
	if len(orderBooks) == 0 {
		h.writeError(w, http.StatusNotFound, "not_found", "No order book data found for symbol", symbol)
		return
	}

	// Get the latest order book
	orderBook := orderBooks[0]

	// Convert bids and asks
	bids := make([]dto.OrderBookLevel, len(orderBook.Bids))
	for i, bid := range orderBook.Bids {
		bids[i] = dto.OrderBookLevel{
			Price:    bid.Price,
			Quantity: bid.Quantity,
		}
	}

	asks := make([]dto.OrderBookLevel, len(orderBook.Asks))
	for i, ask := range orderBook.Asks {
		asks[i] = dto.OrderBookLevel{
			Price:    ask.Price,
			Quantity: ask.Quantity,
		}
	}

	response := dto.OrderBookResponse{
		Symbol:    orderBook.Symbol,
		Bids:      bids,
		Asks:      asks,
		Timestamp: orderBook.Timestamp,
	}

	h.writeSuccess(w, http.StatusOK, response)
}

// Helper functions
func (h *DataHandler) writeSuccess(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(dto.Response{
		Success: true,
		Data:    data,
	})
}

func (h *DataHandler) writeError(w http.ResponseWriter, statusCode int, code, message, details string) {
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
