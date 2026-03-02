package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	supabase "github.com/supabase-community/supabase-go"
)

var SupabaseClient *supabase.Client
var SupabaseAdminClient *supabase.Client

func InitSupabase() error {
	if err := godotenv.Load(".env"); err != nil {
		return err
	}

	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	if supabaseUrl == "" || supabaseKey == "" {
		panic("Missing Supabase credentials.")
	}

	// Log the key type for debugging (first 20 chars only for security)
	if len(supabaseKey) > 20 {
		log.Printf("Using Supabase key starting with: %s...", supabaseKey[:20])
	}

	// Regular client for user operations (subject to RLS)
	client, err := supabase.NewClient(supabaseUrl, supabaseKey,
		&supabase.ClientOptions{
			Schema: "api",
			Headers: map[string]string{
				"Content-Profile": "api",
				"Accept-Profile":  "api",
			},
		})
	if err != nil {
		return err
	}

	SupabaseClient = client

	// Admin client with service role key (bypasses RLS)
	if supabaseServiceKey != "" {
		adminClient, err := supabase.NewClient(supabaseUrl, supabaseServiceKey,
			&supabase.ClientOptions{
				Schema: "api",
				Headers: map[string]string{
					"Content-Profile": "api",
					"Accept-Profile":  "api",
				},
			})
		if err != nil {
			return err
		}
		SupabaseAdminClient = adminClient
		log.Println("Supabase admin client initialized with service role key")
	} else {
		log.Println("Warning: SUPABASE_SERVICE_ROLE_KEY not set, admin operations may fail")
	}

	return nil
}
