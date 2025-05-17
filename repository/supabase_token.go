package repository

import (
	"errors"
	"time"

	"github.com/supabase-community/supabase-go"
)

type SupabaseTokenRepository struct {
	client *supabase.Client
	table  string
}

type Token struct {
	ID        string    `json:"id"`
	TokenHash string    `json:"token_hash"`
	Used      bool      `json:"used"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func NewSupabaseTokenRepository(client *supabase.Client, table string) *SupabaseTokenRepository {
	return &SupabaseTokenRepository{
		client: client,
		table:  table,
	}
}

func (r *SupabaseTokenRepository) Store(token *Token) error {
	var result []Token
	_, err := r.client.From(r.table).Insert(token, false, "", "", "").ExecuteTo(&result)
	return err
}

func (r *SupabaseTokenRepository) FindByID(id string) (*Token, error) {
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

func (r *SupabaseTokenRepository) MarkAsUsed(id string) error {
	// For updates, we pass a map of fields to update
	updateData := map[string]interface{}{"used": true}

	_, err := r.client.From(r.table).
		Update(updateData, "", "").
		Eq("id", id).
		ExecuteTo(nil)

	return err
}
