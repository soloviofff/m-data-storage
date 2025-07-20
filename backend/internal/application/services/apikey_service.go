package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/errors"
	"m-data-storage/internal/domain/interfaces"
)

// APIKeyService implements API key management operations
type APIKeyService struct {
	userStorage interfaces.UserStorage
	roleStorage interfaces.RoleStorage
}

// NewAPIKeyService creates a new API key service
func NewAPIKeyService(userStorage interfaces.UserStorage, roleStorage interfaces.RoleStorage) interfaces.APIKeyService {
	return &APIKeyService{
		userStorage: userStorage,
		roleStorage: roleStorage,
	}
}

// CreateAPIKey creates a new API key for a user
func (s *APIKeyService) CreateAPIKey(ctx context.Context, userID string, req *entities.CreateAPIKeyRequest) (*entities.APIKey, string, error) {
	if userID == "" {
		return nil, "", errors.NewAuthError(errors.CodeInvalidInput, "user ID cannot be empty", nil)
	}

	if req == nil {
		return nil, "", errors.NewAuthError(errors.CodeInvalidInput, "request cannot be nil", nil)
	}

	// Validate request
	if err := s.validateCreateAPIKeyRequest(req); err != nil {
		return nil, "", err
	}

	// Check if user exists
	user, err := s.userStorage.GetUserByID(ctx, userID)
	if err != nil {
		return nil, "", errors.NewAuthError(errors.CodeUserNotFound, "user not found", err)
	}

	if user.Status != entities.UserStatusActive {
		return nil, "", errors.NewAuthError(errors.CodeUserInactive, "user is not active", nil)
	}

	// Generate API key
	rawKey, hashedKey, err := s.GenerateAPIKey()
	if err != nil {
		return nil, "", err
	}

	// Create API key entity
	apiKey := &entities.APIKey{
		ID:          uuid.New().String(),
		UserID:      userID,
		Name:        req.Name,
		KeyHash:     hashedKey,
		Prefix:      rawKey[:12], // First 12 chars for identification
		Permissions: req.Permissions,
		ExpiresAt:   req.ExpiresAt,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}

	// Save to storage
	err = s.userStorage.CreateAPIKey(ctx, apiKey)
	if err != nil {
		return nil, "", errors.NewAuthError("INTERNAL_ERROR", "failed to create API key", err)
	}

	return apiKey, rawKey, nil
}

// GetAPIKeyByID retrieves an API key by ID
func (s *APIKeyService) GetAPIKeyByID(ctx context.Context, id string) (*entities.APIKey, error) {
	if id == "" {
		return nil, errors.NewAuthError(errors.CodeInvalidInput, "API key ID cannot be empty", nil)
	}

	apiKey, err := s.userStorage.GetAPIKeyByID(ctx, id)
	if err != nil {
		return nil, errors.NewAuthError(errors.CodeAPIKeyNotFound, "API key not found", err)
	}

	return apiKey, nil
}

// GetAPIKeysByUserID retrieves all API keys for a user
func (s *APIKeyService) GetAPIKeysByUserID(ctx context.Context, userID string) ([]*entities.APIKey, error) {
	if userID == "" {
		return nil, errors.NewAuthError(errors.CodeInvalidInput, "user ID cannot be empty", nil)
	}

	apiKeys, err := s.userStorage.GetAPIKeysByUserID(ctx, userID)
	if err != nil {
		return nil, errors.NewAuthError("INTERNAL_ERROR", "failed to get API keys", err)
	}

	return apiKeys, nil
}

// GetAPIKeys retrieves API keys with filtering
func (s *APIKeyService) GetAPIKeys(ctx context.Context, filter entities.APIKeyFilter) ([]*entities.APIKey, error) {
	apiKeys, err := s.userStorage.GetAPIKeys(ctx, filter)
	if err != nil {
		return nil, errors.NewAuthError("INTERNAL_ERROR", "failed to get API keys", err)
	}

	return apiKeys, nil
}

// UpdateAPIKey updates an API key
func (s *APIKeyService) UpdateAPIKey(ctx context.Context, id string, req *entities.CreateAPIKeyRequest) (*entities.APIKey, error) {
	if id == "" {
		return nil, errors.NewAuthError(errors.CodeInvalidInput, "API key ID cannot be empty", nil)
	}

	if req == nil {
		return nil, errors.NewAuthError(errors.CodeInvalidInput, "request cannot be nil", nil)
	}

	// Get existing API key
	existingKey, err := s.userStorage.GetAPIKeyByID(ctx, id)
	if err != nil {
		return nil, errors.NewAuthError(errors.CodeAPIKeyNotFound, "API key not found", err)
	}

	// Validate request
	if err := s.validateCreateAPIKeyRequest(req); err != nil {
		return nil, err
	}

	// Update fields
	existingKey.Name = req.Name
	existingKey.Permissions = req.Permissions
	existingKey.ExpiresAt = req.ExpiresAt

	// Save to storage
	err = s.userStorage.UpdateAPIKey(ctx, existingKey)
	if err != nil {
		return nil, errors.NewAuthError("INTERNAL_ERROR", "failed to update API key", err)
	}

	return existingKey, nil
}

// DeleteAPIKey deletes an API key
func (s *APIKeyService) DeleteAPIKey(ctx context.Context, id string) error {
	if id == "" {
		return errors.NewAuthError(errors.CodeInvalidInput, "API key ID cannot be empty", nil)
	}

	// Check if API key exists
	_, err := s.userStorage.GetAPIKeyByID(ctx, id)
	if err != nil {
		return errors.NewAuthError(errors.CodeAPIKeyNotFound, "API key not found", err)
	}

	// Delete from storage
	if err := s.userStorage.DeleteAPIKey(ctx, id); err != nil {
		return errors.NewAuthError("INTERNAL_ERROR", "failed to delete API key", err)
	}

	return nil
}

// ValidateAPIKey validates an API key and returns the key and user
func (s *APIKeyService) ValidateAPIKey(ctx context.Context, key string) (*entities.APIKey, *entities.User, error) {
	if key == "" {
		return nil, nil, errors.NewAuthError(errors.CodeInvalidInput, "API key cannot be empty", nil)
	}

	// Validate key format
	if !s.ValidateAPIKeyFormat(key) {
		return nil, nil, errors.NewAuthError(errors.CodeInvalidAPIKey, "invalid API key format", nil)
	}

	// Hash the provided key
	hashedKey, err := s.HashAPIKey(key)
	if err != nil {
		return nil, nil, err
	}

	// Find API key by hash - we need to search through all keys
	// TODO: This is inefficient, should add GetAPIKeyByHash method to storage
	allKeys, err := s.userStorage.GetAPIKeys(ctx, entities.APIKeyFilter{})
	if err != nil {
		return nil, nil, errors.NewAuthError("INTERNAL_ERROR", "failed to search API keys", err)
	}

	var apiKey *entities.APIKey
	for _, key := range allKeys {
		if key.KeyHash == hashedKey {
			apiKey = key
			break
		}
	}

	if apiKey == nil {
		return nil, nil, errors.NewAuthError(errors.CodeInvalidAPIKey, "API key not found", nil)
	}

	// Check if API key is active
	if !apiKey.IsActive {
		return nil, nil, errors.NewAuthError("API_KEY_REVOKED", "API key is revoked", nil)
	}

	// Check if API key is expired
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, nil, errors.NewAuthError(errors.CodeAPIKeyExpired, "API key is expired", nil)
	}

	// Get user
	user, err := s.userStorage.GetUserByID(ctx, apiKey.UserID)
	if err != nil {
		return nil, nil, errors.NewAuthError(errors.CodeUserNotFound, "user not found", err)
	}

	// Check if user is active
	if user.Status != entities.UserStatusActive {
		return nil, nil, errors.NewAuthError(errors.CodeUserInactive, "user is not active", nil)
	}

	// Update last used timestamp
	if err := s.UpdateAPIKeyLastUsed(ctx, apiKey.ID); err != nil {
		// Log error but don't fail the validation
		// TODO: Add proper logging
	}

	return apiKey, user, nil
}

// RevokeAPIKey revokes an API key
func (s *APIKeyService) RevokeAPIKey(ctx context.Context, id string) error {
	if id == "" {
		return errors.NewAuthError(errors.CodeInvalidInput, "API key ID cannot be empty", nil)
	}

	// Get API key
	apiKey, err := s.userStorage.GetAPIKeyByID(ctx, id)
	if err != nil {
		return errors.NewAuthError(errors.CodeAPIKeyNotFound, "API key not found", err)
	}

	// Revoke the key
	apiKey.IsActive = false

	// Save to storage
	err = s.userStorage.UpdateAPIKey(ctx, apiKey)
	if err != nil {
		return errors.NewAuthError("INTERNAL_ERROR", "failed to revoke API key", err)
	}

	return nil
}

// UpdateAPIKeyLastUsed updates the last used timestamp of an API key
func (s *APIKeyService) UpdateAPIKeyLastUsed(ctx context.Context, id string) error {
	if id == "" {
		return errors.NewAuthError(errors.CodeInvalidInput, "API key ID cannot be empty", nil)
	}

	if err := s.userStorage.UpdateAPIKeyLastUsed(ctx, id); err != nil {
		return errors.NewAuthError("INTERNAL_ERROR", "failed to update API key last used", err)
	}

	return nil
}

// GenerateAPIKey generates a new API key and its hash
func (s *APIKeyService) GenerateAPIKey() (string, string, error) {
	// Generate 32 random bytes
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", "", errors.NewAuthError("INTERNAL_ERROR", "failed to generate API key", err)
	}

	// Encode as base64 with prefix
	rawKey := fmt.Sprintf("mds_%s", base64.URLEncoding.EncodeToString(keyBytes))

	// Hash the key
	hashedKey, err := s.HashAPIKey(rawKey)
	if err != nil {
		return "", "", err
	}

	return rawKey, hashedKey, nil
}

// HashAPIKey hashes an API key using SHA-256
func (s *APIKeyService) HashAPIKey(key string) (string, error) {
	if key == "" {
		return "", errors.NewAuthError(errors.CodeInvalidInput, "API key cannot be empty", nil)
	}

	hash := sha256.Sum256([]byte(key))
	return base64.URLEncoding.EncodeToString(hash[:]), nil
}

// ValidateAPIKeyFormat validates the format of an API key
func (s *APIKeyService) ValidateAPIKeyFormat(key string) bool {
	if key == "" {
		return false
	}

	// Check if key starts with "mds_" prefix
	if !strings.HasPrefix(key, "mds_") {
		return false
	}

	// Remove prefix and validate base64 encoding
	keyPart := strings.TrimPrefix(key, "mds_")
	if len(keyPart) == 0 {
		return false
	}

	// Check if it's valid base64
	_, err := base64.URLEncoding.DecodeString(keyPart)
	return err == nil
}

// validateCreateAPIKeyRequest validates a create API key request
func (s *APIKeyService) validateCreateAPIKeyRequest(req *entities.CreateAPIKeyRequest) error {
	if req.Name == "" {
		return errors.NewAuthError(errors.CodeInvalidInput, "API key name cannot be empty", nil)
	}

	if len(req.Name) > 100 {
		return errors.NewAuthError(errors.CodeInvalidInput, "API key name too long", nil)
	}

	// Validate name format (alphanumeric, spaces, hyphens, underscores)
	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9\s\-_]+$`)
	if !nameRegex.MatchString(req.Name) {
		return errors.NewAuthError(errors.CodeInvalidInput, "API key name contains invalid characters", nil)
	}

	// Description field doesn't exist in CreateAPIKeyRequest, skip validation

	// Validate expiration date
	if req.ExpiresAt != nil && req.ExpiresAt.Before(time.Now()) {
		return errors.NewAuthError(errors.CodeInvalidInput, "API key expiration date cannot be in the past", nil)
	}

	// Validate permissions format
	for _, permission := range req.Permissions {
		if permission == "" {
			return errors.NewAuthError(errors.CodeInvalidInput, "permission cannot be empty", nil)
		}

		// Basic permission format validation (resource:action)
		parts := strings.Split(permission, ":")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return errors.NewAuthError(errors.CodeInvalidInput,
				fmt.Sprintf("invalid permission format: %s (expected resource:action)", permission), nil)
		}
	}

	return nil
}
