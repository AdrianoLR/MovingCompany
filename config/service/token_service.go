package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"MovingCompanyGo/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenService defines the interface for token operations.
type TokenService interface {
	GenerateToken(ctx context.Context, ttl time.Duration) (string, string, error)
	ValidateToken(ctx context.Context, tokenString string) (string, bool, error)
	ConsumeToken(ctx context.Context, tokenString string) (bool, error)
}

// JWTTokenService implements TokenService using HMAC-signed JWTs backed by a DB.
type JWTTokenService struct {
	secretKey []byte
	repo      repository.TokenRepository
}

// JWTClaims holds the custom and standard JWT claims embedded in every booking token.
type JWTClaims struct {
	TokenID string `json:"tid"` // links the JWT to its DB record for one-time-use enforcement
	jwt.RegisteredClaims
}

// NewJWTTokenService constructs a JWTTokenService.
func NewJWTTokenService(secretKey string, repo repository.TokenRepository) *JWTTokenService {
	return &JWTTokenService{
		secretKey: []byte(secretKey),
		repo:      repo, // store as interface — callers can pass any TokenRepository implementation
	}
}

// hashToken returns the SHA-256 hex digest of the given JWT string.
func hashToken(tokenString string) string {
	sum := sha256.Sum256([]byte(tokenString))
	return hex.EncodeToString(sum[:])
}

// GenerateToken mints a new JWT, stores only its hash in the DB, and returns both the
func (s *JWTTokenService) GenerateToken(ctx context.Context, ttl time.Duration) (string, string, error) {
	tokenID := uuid.New().String()
	expiresAt := time.Now().Add(ttl)

	claims := JWTClaims{
		TokenID: tokenID, // embed the DB record ID inside the JWT so we can look it up on validation
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "moving-booking-system",
			Subject:   "one-time-form-access",
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", "", err
	}

	tokenRecord := &repository.Token{
		ID:        tokenID,
		TokenHash: hashToken(tokenString),
		Used:      false,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Store(ctx, tokenRecord); err != nil {
		return "", "", err
	}

	return tokenID, tokenString, nil
}

// ValidateToken parses and verifies the JWT signature, then confirms it has not been used.
func (s *JWTTokenService) ValidateToken(ctx context.Context, tokenString string) (string, bool, error) {
	// Parse the token and verify its HMAC signature against our secret key
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.secretKey, nil // supply the signing secret so the library can verify the signature
	})

	if err != nil {
		return "", false, err // covers expired, tampered, or malformed tokens
	}

	if !token.Valid {
		return "", false, errors.New("invalid token") // extra guard: should not be reached if ParseWithClaims succeeded
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return "", false, errors.New("invalid token claims") // type assertion failed; payload is malformed
	}

	tokenID := claims.TokenID // extract the DB record ID that was embedded when the token was created

	// Confirm the token record exists in the DB and has not already been consumed
	tokenRecord, err := s.repo.FindByID(ctx, tokenID) // R4: pass ctx to propagate deadlines to the DB query
	if err != nil {
		return "", false, err
	}

	if tokenRecord.Used {
		return tokenID, false, errors.New("token already used") // one-time-use enforcement
	}

	return tokenID, true, nil
}

// ConsumeToken validates the token and atomically marks it as used.
func (s *JWTTokenService) ConsumeToken(ctx context.Context, tokenString string) (bool, error) {
	tokenID, valid, err := s.ValidateToken(ctx, tokenString) // validate before consuming to avoid marking bad tokens
	if err != nil || !valid {
		return false, err
	}

	if err := s.repo.MarkAsUsed(ctx, tokenID); err != nil {
		return false, err
	}

	return true, nil
}
