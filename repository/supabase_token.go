package repository

import (
	"context"
	"errors"
	"time"

	"github.com/supabase-community/supabase-go"
)

// TokenRepository defines the interface for token storage operations.
type TokenRepository interface {
	Store(ctx context.Context, token *Token) error
	FindByID(ctx context.Context, id string) (*Token, error)
	MarkAsUsed(ctx context.Context, id string) error
}

// SupabaseTokenRepository is the Supabase-backed implementation of TokenRepository
type SupabaseTokenRepository struct {
	client *supabase.Client
	table  string
}

// Token represents a one-time booking link token stored in the database.
type Token struct {
	ID        string    `json:"id"`         // DB column: id (primary key)
	TokenHash string    `json:"token_hash"` // DB column: token_hash (stores SHA-256 hex of the JWT — S6)
	Used      bool      `json:"used"`       // DB column: used (one-time enforcement flag)
	ExpiresAt time.Time `json:"expires_at"` // DB column: expires_at
	CreatedAt time.Time `json:"created_at"` // DB column: created_at
}

// NewSupabaseTokenRepository constructs a repository targeting the given Supabase table.
func NewSupabaseTokenRepository(client *supabase.Client, table string) *SupabaseTokenRepository {
	return &SupabaseTokenRepository{
		client: client,
		table:  table,
	}
}

// Store inserts a new token record into the database.
func (r *SupabaseTokenRepository) Store(ctx context.Context, token *Token) error {
	var result []Token // ExecuteTo requires a destination even for a single-row insert
	_, err := r.client.From(r.table).Insert(token, false, "", "", "").ExecuteTo(&result)
	return err
}

// FindByID retrieves a token by its unique ID.
func (r *SupabaseTokenRepository) FindByID(ctx context.Context, id string) (*Token, error) {
	var tokens []Token
	_, err := r.client.From(r.table).
		Select("*", "", false).
		Eq("id", id).
		ExecuteTo(&tokens)

	if err != nil {
		return nil, err
	}

	if len(tokens) == 0 {
		return nil, errors.New("token not found")
	}

	return &tokens[0], nil
}

// MarkAsUsed sets used=true for the token identified by id.
func (r *SupabaseTokenRepository) MarkAsUsed(ctx context.Context, id string) error {
	updateData := map[string]interface{}{"used": true}

	_, err := r.client.From(r.table).
		Update(updateData, "", "").
		Eq("id", id).
		ExecuteTo(nil)

	return err
}
