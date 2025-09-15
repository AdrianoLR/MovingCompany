package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	supabase "github.com/supabase-community/supabase-go"
)

var SupabaseClient *supabase.Client

func InitSupabase() error {
	if err := godotenv.Load(".env"); err != nil {
		return err
	}

	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	if supabaseUrl == "" || supabaseKey == "" {
		panic("Missing Supabase credentials.")
	}

	// Log the key type for debugging (first 20 chars only for security)
	if len(supabaseKey) > 20 {
		log.Printf("Using Supabase key starting with: %s...", supabaseKey[:20])
	}

	client, err := supabase.NewClient(supabaseUrl, supabaseKey,
		&supabase.ClientOptions{
			Schema: "api",
		})
	if err != nil {
		return err
	}

	SupabaseClient = client
	return nil
}
