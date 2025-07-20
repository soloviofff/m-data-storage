package routes

import (
	"github.com/gorilla/mux"

	"m-data-storage/api/handlers"
	"m-data-storage/api/middleware"
)

// RegisterAuthRoutes registers authentication routes
func RegisterAuthRoutes(router *mux.Router, authHandler *handlers.AuthHandler, authMiddleware *middleware.AuthMiddleware) {
	// Public routes (no authentication required)
	auth := router.PathPrefix("/auth").Subrouter()
	
	// Authentication endpoints
	auth.HandleFunc("/login", authHandler.Login).Methods("POST")
	auth.HandleFunc("/register", authHandler.Register).Methods("POST")
	auth.HandleFunc("/refresh", authHandler.RefreshToken).Methods("POST")
	
	// Protected routes (authentication required)
	protected := auth.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.RequireAuth)
	
	// User profile endpoints
	protected.HandleFunc("/profile", authHandler.GetProfile).Methods("GET")
	protected.HandleFunc("/profile", authHandler.UpdateProfile).Methods("PUT")
	protected.HandleFunc("/logout", authHandler.Logout).Methods("POST")
	protected.HandleFunc("/change-password", authHandler.ChangePassword).Methods("POST")
	
	// API key management endpoints
	apiKeys := protected.PathPrefix("/api-keys").Subrouter()
	apiKeys.HandleFunc("", authHandler.ListAPIKeys).Methods("GET")
	apiKeys.HandleFunc("", authHandler.CreateAPIKey).Methods("POST")
	apiKeys.HandleFunc("/{id}", authHandler.RevokeAPIKey).Methods("DELETE")
}
