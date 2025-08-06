package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"aluminium-passport/internal/config"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidPassword = errors.New("invalid password")
	ErrWeakPassword    = errors.New("password does not meet security requirements")
)

// PasswordRequirements defines password complexity requirements
type PasswordRequirements struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSpecial bool
}

var DefaultPasswordRequirements = PasswordRequirements{
	MinLength:      8,
	RequireUpper:   true,
	RequireLower:   true,
	RequireDigit:   true,
	RequireSpecial: false, // Set to true in production
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	cfg := config.AppConfig
	cost := cfg.BcryptCost

	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// VerifyPassword verifies a password against its hash using bcrypt
func VerifyPassword(password, hashedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidPassword
		}
		return fmt.Errorf("failed to verify password: %w", err)
	}
	return nil
}

// ValidatePasswordStrength validates password against security requirements
func ValidatePasswordStrength(password string) error {
	req := DefaultPasswordRequirements

	if len(password) < req.MinLength {
		return fmt.Errorf("password must be at least %d characters long", req.MinLength)
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	if req.RequireUpper && !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}

	if req.RequireLower && !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}

	if req.RequireDigit && !hasDigit {
		return errors.New("password must contain at least one digit")
	}

	if req.RequireSpecial && !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

// Argon2 implementation for enhanced security (alternative to bcrypt)
type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

var DefaultArgon2Params = Argon2Params{
	Memory:      64 * 1024, // 64MB
	Iterations:  3,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

// HashPasswordArgon2 hashes password using Argon2id
func HashPasswordArgon2(password string) (string, error) {
	params := DefaultArgon2Params

	// Generate salt
	salt := make([]byte, params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash password
	hash := argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)

	// Encode to string
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, params.Memory, params.Iterations, params.Parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}

// VerifyPasswordArgon2 verifies password against Argon2id hash
func VerifyPasswordArgon2(password, encodedHash string) error {
	// Parse the encoded hash
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return ErrInvalidPassword
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return ErrInvalidPassword
	}
	if version != argon2.Version {
		return ErrInvalidPassword
	}

	var memory, iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return ErrInvalidPassword
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return ErrInvalidPassword
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return ErrInvalidPassword
	}

	// Hash the input password
	inputHash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(hash)))

	// Compare hashes using constant-time comparison
	if subtle.ConstantTimeCompare(hash, inputHash) == 1 {
		return nil
	}

	return ErrInvalidPassword
}

// GenerateSecurePassword generates a cryptographically secure random password
func GenerateSecurePassword(length int) (string, error) {
	if length < 8 {
		length = 12 // Default secure length
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"

	password := make([]byte, length)
	for i := range password {
		randomByte := make([]byte, 1)
		if _, err := rand.Read(randomByte); err != nil {
			return "", fmt.Errorf("failed to generate secure password: %w", err)
		}
		password[i] = charset[randomByte[0]%byte(len(charset))]
	}

	return string(password), nil
}

// IsPasswordCompromised checks against common password lists (placeholder)
// In production, you would check against a database of compromised passwords
func IsPasswordCompromised(password string) bool {
	commonPasswords := []string{
		"password", "123456", "password123", "admin", "qwerty",
		"letmein", "welcome", "monkey", "1234567890", "password1",
	}

	lowerPassword := strings.ToLower(password)
	for _, common := range commonPasswords {
		if lowerPassword == common {
			return true
		}
	}

	return false
}

// ValidateAndHashPassword validates password strength and returns hash
func ValidateAndHashPassword(password string) (string, error) {
	// Check if password is compromised
	if IsPasswordCompromised(password) {
		return "", errors.New("password is commonly used and not secure")
	}

	// Validate password strength
	if err := ValidatePasswordStrength(password); err != nil {
		return "", err
	}

	// Hash the password
	return HashPassword(password)
}
