package service

import (
	"errors"
	"time"

	"MovingCompanyGo/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Using repository.Token instead of local Token type

type JWTTokenService struct {
	secretKey []byte
	repo      *repository.SupabaseTokenRepository
}

type JWTClaims struct {
	TokenID string `json:"tid"`
	jwt.RegisteredClaims
}

func NewJWTTokenService(secretKey string, repo *repository.SupabaseTokenRepository) *JWTTokenService {
	return &JWTTokenService{
		secretKey: []byte(secretKey),
		repo:      repo,
	}
}

// GenerateToken creates a JWT token with unique ID for one-time use
func (s *JWTTokenService) GenerateToken(ttl time.Duration) (string, string, error) {
	tokenID := uuid.New().String()
	expiresAt := time.Now().Add(ttl)

	// Create claims
	claims := JWTClaims{
		TokenID: tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "moving-booking-system",
			Subject:   "one-time-form-access",
			ID:        tokenID,
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", "", err
	}

	// Store token record in database
	tokenRecord := &repository.Token{
		ID:        tokenID,
		TokenHash: tokenString, // JWT doesn't need hash storage as it's verified by signature
		Used:      false,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Store(tokenRecord); err != nil {
		return "", "", err
	}

	return tokenID, tokenString, nil
}

// ValidateToken verifies JWT and checks if it has been used
func (s *JWTTokenService) ValidateToken(tokenString string) (string, bool, error) {
	// Parse the JWT
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.secretKey, nil
	})

	if err != nil {
		return "", false, err
	}

	if !token.Valid {
		return "", false, errors.New("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return "", false, errors.New("invalid token claims")
	}

	tokenID := claims.TokenID

	// Check if token exists and hasn't been used
	tokenRecord, err := s.repo.FindByID(tokenID)
	if err != nil {
		return "", false, err
	}

	if tokenRecord.Used {
		return tokenID, false, errors.New("token already used")
	}

	return tokenID, true, nil
}

// ConsumeToken marks a token as used
func (s *JWTTokenService) ConsumeToken(tokenString string) (bool, error) {
	tokenID, valid, err := s.ValidateToken(tokenString)
	if err != nil || !valid {
		return false, err
	}

	// Mark token as used
	if err := s.repo.MarkAsUsed(tokenID); err != nil {
		return false, err
	}

	return true, nil
}
