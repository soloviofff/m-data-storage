package services

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/argon2"
	"github.com/google/uuid"

	"m-data-storage/internal/domain/errors"
	"m-data-storage/internal/domain/interfaces"
)

// PasswordService implements password operations
type PasswordService struct {
	// Argon2 parameters
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
	
	// Password policy
	minLength    int
	maxLength    int
	requireUpper bool
	requireLower bool
	requireDigit bool
	requireSpecial bool
	
	// Reset token settings
	resetTokenTTL time.Duration
}

// NewPasswordService creates a new password service
func NewPasswordService() interfaces.PasswordService {
	return &PasswordService{
		// Argon2 parameters (recommended values)
		memory:      64 * 1024, // 64 MB
		iterations:  3,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32,
		
		// Password policy
		minLength:    8,
		maxLength:    128,
		requireUpper: true,
		requireLower: true,
		requireDigit: true,
		requireSpecial: true,
		
		// Reset token TTL
		resetTokenTTL: 1 * time.Hour,
	}
}

// HashPassword hashes a password using Argon2
func (s *PasswordService) HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.NewAuthError(errors.CodeInvalidInput, "password cannot be empty", nil)
	}

	// Generate a random salt
	salt := make([]byte, s.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", errors.NewAuthError("INTERNAL_ERROR", "failed to generate salt", err)
	}

	// Hash the password
	hash := argon2.IDKey([]byte(password), salt, s.iterations, s.memory, s.parallelism, s.keyLength)

	// Encode the hash in the format: $argon2id$v=19$m=memory,t=iterations,p=parallelism$salt$hash
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		s.memory,
		s.iterations,
		s.parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encodedHash, nil
}

// ValidatePassword validates a password against its hash
func (s *PasswordService) ValidatePassword(password, encodedHash string) bool {
	if password == "" || encodedHash == "" {
		return false
	}

	// Parse the encoded hash
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false
	}

	if parts[1] != "argon2id" {
		return false
	}

	// Parse parameters
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false
	}

	var memory, iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false
	}

	// Decode salt and hash
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	// Hash the provided password with the same parameters
	providedHash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(hash)))

	// Compare hashes using constant-time comparison
	return subtle.ConstantTimeCompare(hash, providedHash) == 1
}

// ValidatePasswordStrength validates password strength according to policy
func (s *PasswordService) ValidatePasswordStrength(password string) error {
	if len(password) < s.minLength {
		return errors.NewAuthError(errors.CodePasswordTooWeak, 
			fmt.Sprintf("password must be at least %d characters long", s.minLength), nil)
	}

	if len(password) > s.maxLength {
		return errors.NewAuthError(errors.CodePasswordTooWeak, 
			fmt.Sprintf("password must be at most %d characters long", s.maxLength), nil)
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	var missing []string

	if s.requireUpper && !hasUpper {
		missing = append(missing, "uppercase letter")
	}
	if s.requireLower && !hasLower {
		missing = append(missing, "lowercase letter")
	}
	if s.requireDigit && !hasDigit {
		missing = append(missing, "digit")
	}
	if s.requireSpecial && !hasSpecial {
		missing = append(missing, "special character")
	}

	if len(missing) > 0 {
		return errors.NewAuthError(errors.CodePasswordTooWeak, 
			fmt.Sprintf("password must contain at least one: %s", strings.Join(missing, ", ")), nil)
	}

	// Check for common weak patterns
	if err := s.checkCommonPatterns(password); err != nil {
		return err
	}

	return nil
}

// GenerateRandomPassword generates a random password of specified length
func (s *PasswordService) GenerateRandomPassword(length int) string {
	if length < s.minLength {
		length = s.minLength
	}
	if length > s.maxLength {
		length = s.maxLength
	}

	const (
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digits    = "0123456789"
		special   = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	)

	var charset string
	var required []string

	if s.requireLower {
		charset += lowercase
		required = append(required, lowercase)
	}
	if s.requireUpper {
		charset += uppercase
		required = append(required, uppercase)
	}
	if s.requireDigit {
		charset += digits
		required = append(required, digits)
	}
	if s.requireSpecial {
		charset += special
		required = append(required, special)
	}

	if charset == "" {
		charset = lowercase + uppercase + digits + special
	}

	password := make([]byte, length)

	// Ensure at least one character from each required set
	for i, req := range required {
		if i < length {
			char, _ := rand.Int(rand.Reader, big.NewInt(int64(len(req))))
			password[i] = req[char.Int64()]
		}
	}

	// Fill the rest with random characters from the full charset
	for i := len(required); i < length; i++ {
		char, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		password[i] = charset[char.Int64()]
	}

	// Shuffle the password
	for i := length - 1; i > 0; i-- {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		password[i], password[j.Int64()] = password[j.Int64()], password[i]
	}

	return string(password)
}

// GenerateResetToken generates a password reset token
func (s *PasswordService) GenerateResetToken(userID string) (string, time.Time, error) {
	if userID == "" {
		return "", time.Time{}, errors.NewAuthError(errors.CodeInvalidInput, "user ID cannot be empty", nil)
	}

	// Generate a secure random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", time.Time{}, errors.NewAuthError("INTERNAL_ERROR", "failed to generate reset token", err)
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)
	expiresAt := time.Now().Add(s.resetTokenTTL)

	return token, expiresAt, nil
}

// ValidateResetToken validates a password reset token and returns the user ID
func (s *PasswordService) ValidateResetToken(token string) (string, error) {
	if token == "" {
		return "", errors.NewAuthError(errors.CodeInvalidInput, "reset token cannot be empty", nil)
	}

	// Decode the token to validate format
	_, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return "", errors.NewAuthError(errors.CodeInvalidToken, "invalid reset token format", err)
	}

	// TODO: In a real implementation, you would:
	// 1. Look up the token in the database
	// 2. Check if it's expired
	// 3. Return the associated user ID
	// For now, we'll return a placeholder
	return uuid.New().String(), nil
}

// checkCommonPatterns checks for common weak password patterns
func (s *PasswordService) checkCommonPatterns(password string) error {
	lower := strings.ToLower(password)

	// Check for common weak passwords
	commonPasswords := []string{
		"password", "123456", "123456789", "qwerty", "abc123",
		"password123", "admin", "letmein", "welcome", "monkey",
	}

	for _, common := range commonPasswords {
		if lower == common {
			return errors.NewAuthError(errors.CodePasswordTooWeak, "password is too common", nil)
		}
	}

	// Check for sequential patterns
	if matched, _ := regexp.MatchString(`(012|123|234|345|456|567|678|789|890|abc|bcd|cde|def|efg|fgh|ghi|hij|ijk|jkl|klm|lmn|mno|nop|opq|pqr|qrs|rst|stu|tuv|uvw|vwx|wxy|xyz)`, lower); matched {
		return errors.NewAuthError(errors.CodePasswordTooWeak, "password contains sequential characters", nil)
	}

	// Check for repeated patterns
	if matched, _ := regexp.MatchString(`(.)\1{2,}`, password); matched {
		return errors.NewAuthError(errors.CodePasswordTooWeak, "password contains repeated characters", nil)
	}

	return nil
}
