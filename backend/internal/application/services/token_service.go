package services

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/errors"
	"m-data-storage/internal/domain/interfaces"
)

// TokenService implements JWT token operations
type TokenService struct {
	secretKey       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	issuer          string
}

// NewTokenService creates a new token service
func NewTokenService(secretKey string, accessTokenTTL, refreshTokenTTL time.Duration, issuer string) interfaces.TokenService {
	return &TokenService{
		secretKey:       []byte(secretKey),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		issuer:          issuer,
	}
}

// Claims represents JWT claims
type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// GenerateAccessToken generates a new access token for the user
func (s *TokenService) GenerateAccessToken(ctx context.Context, user *entities.User) (string, time.Time, error) {
	if user == nil {
		return "", time.Time{}, errors.NewAuthError(errors.CodeInvalidInput, "user cannot be nil", nil)
	}

	expiresAt := time.Now().Add(s.accessTokenTTL)

	// Extract role name (user has single role)
	var roles []string
	if user.Role != nil {
		roles = []string{user.Role.Name}
	}

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   user.ID,
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", time.Time{}, errors.NewAuthError("TOKEN_GENERATION_ERROR", "failed to sign access token", err)
	}

	return tokenString, expiresAt, nil
}

// GenerateRefreshToken generates a new refresh token for the user
func (s *TokenService) GenerateRefreshToken(ctx context.Context, user *entities.User) (string, time.Time, error) {
	if user == nil {
		return "", time.Time{}, errors.NewAuthError(errors.CodeInvalidInput, "user cannot be nil", nil)
	}

	expiresAt := time.Now().Add(s.refreshTokenTTL)

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   user.ID,
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", time.Time{}, errors.NewAuthError("TOKEN_GENERATION_ERROR", "failed to sign refresh token", err)
	}

	return tokenString, expiresAt, nil
}

// ValidateAccessToken validates an access token and returns the user
func (s *TokenService) ValidateAccessToken(ctx context.Context, tokenString string) (*entities.User, error) {
	claims, err := s.parseToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Create user from claims
	user := &entities.User{
		ID:       claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
	}

	return user, nil
}

// ValidateRefreshToken validates a refresh token and returns the user
func (s *TokenService) ValidateRefreshToken(ctx context.Context, tokenString string) (*entities.User, error) {
	claims, err := s.parseToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Create user from claims
	user := &entities.User{
		ID:       claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
	}

	return user, nil
}

// RevokeToken revokes a token (implementation depends on token blacklist strategy)
func (s *TokenService) RevokeToken(ctx context.Context, token string) error {
	// For now, we'll just validate the token format
	_, err := s.parseToken(token)
	if err != nil {
		return errors.NewAuthError(errors.CodeInvalidToken, "invalid token format", err)
	}

	// TODO: Implement token blacklist storage
	// This could be implemented using Redis or database storage
	// For now, we'll just return success
	return nil
}

// ExtractUserFromToken extracts user information from a token
func (s *TokenService) ExtractUserFromToken(ctx context.Context, tokenString string) (*entities.User, error) {
	return s.ValidateAccessToken(ctx, tokenString)
}

// GetTokenExpiration returns the expiration time of a token
func (s *TokenService) GetTokenExpiration(tokenString string) (time.Time, error) {
	claims, err := s.parseToken(tokenString)
	if err != nil {
		return time.Time{}, err
	}

	if claims.ExpiresAt == nil {
		return time.Time{}, errors.NewAuthError(errors.CodeInvalidToken, "token has no expiration", nil)
	}

	return claims.ExpiresAt.Time, nil
}

// IsTokenExpired checks if a token is expired
func (s *TokenService) IsTokenExpired(tokenString string) bool {
	expiresAt, err := s.GetTokenExpiration(tokenString)
	if err != nil {
		return true
	}

	return time.Now().After(expiresAt)
}

// parseToken parses and validates a JWT token
func (s *TokenService) parseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, errors.NewAuthError(errors.CodeInvalidToken, "failed to parse token", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.NewAuthError(errors.CodeInvalidToken, "invalid token claims", nil)
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		return nil, errors.NewAuthError(errors.CodeTokenExpired, "token is expired", nil)
	}

	return claims, nil
}
