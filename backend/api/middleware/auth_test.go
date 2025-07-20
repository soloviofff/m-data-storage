package middleware

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"m-data-storage/internal/domain/entities"
)

// Mock services
type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) GenerateAccessToken(ctx context.Context, user *entities.User) (string, time.Time, error) {
	args := m.Called(ctx, user)
	return args.String(0), args.Get(1).(time.Time), args.Error(2)
}

func (m *MockTokenService) GenerateRefreshToken(ctx context.Context, user *entities.User) (string, time.Time, error) {
	args := m.Called(ctx, user)
	return args.String(0), args.Get(1).(time.Time), args.Error(2)
}

func (m *MockTokenService) ValidateAccessToken(ctx context.Context, token string) (*entities.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockTokenService) ValidateRefreshToken(ctx context.Context, token string) (*entities.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockTokenService) RevokeToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockTokenService) ExtractUserFromToken(ctx context.Context, token string) (*entities.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

type MockAPIKeyService struct {
	mock.Mock
}

func (m *MockAPIKeyService) CreateAPIKey(ctx context.Context, req *entities.CreateAPIKeyRequest) (*entities.APIKey, string, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(*entities.APIKey), args.String(1), args.Error(2)
}

func (m *MockAPIKeyService) GetAPIKeyByID(ctx context.Context, id string) (*entities.APIKey, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.APIKey), args.Error(1)
}

func (m *MockAPIKeyService) GetAPIKeys(ctx context.Context, filter entities.APIKeyFilter) ([]*entities.APIKey, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.APIKey), args.Error(1)
}

func (m *MockAPIKeyService) UpdateAPIKey(ctx context.Context, id string, req *entities.CreateAPIKeyRequest) (*entities.APIKey, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.APIKey), args.Error(1)
}

func (m *MockAPIKeyService) DeleteAPIKey(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAPIKeyService) ValidateAPIKey(ctx context.Context, key string) (*entities.APIKey, *entities.User, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(*entities.APIKey), args.Get(1).(*entities.User), args.Error(2)
}

func (m *MockAPIKeyService) RevokeAPIKey(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAPIKeyService) UpdateAPIKeyLastUsed(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAPIKeyService) GenerateAPIKey() (string, string, error) {
	args := m.Called()
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAPIKeyService) HashAPIKey(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockAPIKeyService) ValidateAPIKeyFormat(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

// Mock Authorization Service
type MockAuthorizationService struct {
	mock.Mock
}

func (m *MockAuthorizationService) HasPermission(ctx context.Context, userID, permission string) (bool, error) {
	args := m.Called(ctx, userID, permission)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthorizationService) HasAnyPermission(ctx context.Context, userID string, permissions []string) (bool, error) {
	args := m.Called(ctx, userID, permissions)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthorizationService) HasAllPermissions(ctx context.Context, userID string, permissions []string) (bool, error) {
	args := m.Called(ctx, userID, permissions)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthorizationService) HasRole(ctx context.Context, userID, roleName string) (bool, error) {
	args := m.Called(ctx, userID, roleName)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthorizationService) HasAnyRole(ctx context.Context, userID string, roleNames []string) (bool, error) {
	args := m.Called(ctx, userID, roleNames)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthorizationService) CanAccessResource(ctx context.Context, userID, resource, action string) (bool, error) {
	args := m.Called(ctx, userID, resource, action)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthorizationService) CanModifyUser(ctx context.Context, actorID, targetUserID string) (bool, error) {
	args := m.Called(ctx, actorID, targetUserID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthorizationService) CanManageRole(ctx context.Context, userID, roleID string) (bool, error) {
	args := m.Called(ctx, userID, roleID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthorizationService) CanAPIKeyAccess(ctx context.Context, apiKey *entities.APIKey, resource, action string) (bool, error) {
	args := m.Called(ctx, apiKey, resource, action)
	return args.Bool(0), args.Error(1)
}

// Mock Permission Service
type MockPermissionService struct {
	mock.Mock
}

func (m *MockPermissionService) CreatePermission(ctx context.Context, req *entities.CreatePermissionRequest) (*entities.Permission, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Permission), args.Error(1)
}

func (m *MockPermissionService) GetPermissionByID(ctx context.Context, id string) (*entities.Permission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Permission), args.Error(1)
}

func (m *MockPermissionService) GetPermissionByName(ctx context.Context, name string) (*entities.Permission, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Permission), args.Error(1)
}

func (m *MockPermissionService) GetPermissions(ctx context.Context, filter entities.PermissionFilter) ([]*entities.Permission, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Permission), args.Error(1)
}

func (m *MockPermissionService) UpdatePermission(ctx context.Context, id string, req *entities.UpdatePermissionRequest) (*entities.Permission, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Permission), args.Error(1)
}

func (m *MockPermissionService) DeletePermission(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPermissionService) CheckPermission(ctx context.Context, userID, permission string) (bool, error) {
	args := m.Called(ctx, userID, permission)
	return args.Bool(0), args.Error(1)
}

func (m *MockPermissionService) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// Mock Security Service
type MockSecurityService struct {
	mock.Mock
}

func (m *MockSecurityService) CheckRateLimit(ctx context.Context, identifier string, action string) (bool, error) {
	args := m.Called(ctx, identifier, action)
	return args.Bool(0), args.Error(1)
}

func (m *MockSecurityService) IncrementRateLimit(ctx context.Context, identifier string, action string) error {
	args := m.Called(ctx, identifier, action)
	return args.Error(0)
}

func (m *MockSecurityService) LogSecurityEvent(ctx context.Context, eventType string, data map[string]interface{}) error {
	args := m.Called(ctx, eventType, data)
	return args.Error(0)
}

func (m *MockSecurityService) ValidateSession(ctx context.Context, sessionID string) (*entities.User, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockSecurityService) CreateSession(ctx context.Context, user *entities.User, metadata map[string]interface{}) (string, time.Time, error) {
	args := m.Called(ctx, user, metadata)
	return args.String(0), args.Get(1).(time.Time), args.Error(2)
}

func (m *MockSecurityService) RevokeSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSecurityService) CheckAccountLockout(ctx context.Context, userID string) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockSecurityService) RecordFailedLogin(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockSecurityService) ResetFailedLogins(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// Test functions
func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name          string
		authorization string
		expectedToken string
	}{
		{
			name:          "valid bearer token",
			authorization: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectedToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		},
		{
			name:          "bearer with different case",
			authorization: "bearer token123",
			expectedToken: "token123",
		},
		{
			name:          "no bearer prefix",
			authorization: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectedToken: "",
		},
		{
			name:          "empty authorization",
			authorization: "",
			expectedToken: "",
		},
		{
			name:          "only bearer",
			authorization: "Bearer",
			expectedToken: "",
		},
		{
			name:          "multiple spaces",
			authorization: "Bearer  token",
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authorization != "" {
				req.Header.Set("Authorization", tt.authorization)
			}

			token := extractBearerToken(req)
			assert.Equal(t, tt.expectedToken, token)
		})
	}
}

func TestGetAuthContext(t *testing.T) {
	// Test with no auth context
	req := httptest.NewRequest("GET", "/test", nil)
	authCtx := GetAuthContext(req)
	assert.Nil(t, authCtx)

	// Test with auth context
	expectedCtx := &AuthContext{
		UserID:     "user-123",
		Username:   "testuser",
		Email:      "test@example.com",
		AuthMethod: "jwt",
	}

	ctx := context.WithValue(req.Context(), "auth", expectedCtx)
	req = req.WithContext(ctx)

	authCtx = GetAuthContext(req)
	assert.NotNil(t, authCtx)
	assert.Equal(t, expectedCtx, authCtx)

	// Test with invalid auth context type
	ctx = context.WithValue(req.Context(), "auth", "invalid")
	req = req.WithContext(ctx)

	authCtx = GetAuthContext(req)
	assert.Nil(t, authCtx)
}
