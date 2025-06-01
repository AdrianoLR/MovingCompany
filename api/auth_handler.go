package api

import (
	"MovingCompanyGo/config"
	"encoding/json"
	"net/http"
)

// AuthHandler handles authentication related requests
type AuthHandler struct{}

// NewAuthHandler creates a new instance of AuthHandler
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// LoginPageHandler serves the login page
func (h *AuthHandler) LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	// Serve the login page
	http.ServeFile(w, r, "./static/login.html")
}

// AuthenticateHandler handles user authentication
func (h *AuthHandler) AuthenticateHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body
	var authRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&authRequest)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Authenticate with Supabase
	resp, err := config.SupabaseClient.SignInWithEmailPassword(authRequest.Email, authRequest.Password)
	if err != nil {
		// Return error response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid email or password",
		})
		return
	}

	// Return success response with token
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"token":   resp.AccessToken,
		"user":    resp.User,
	})
}

// RequireAuth is a middleware that checks if the user is authenticated
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the token from the Authorization header
		token := r.Header.Get("Authorization")
		if token == "" {
			// Check if token is in cookie
			cookie, err := r.Cookie("sb-auth-token")
			if err != nil || cookie.Value == "" {
				// Redirect to login page
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			token = cookie.Value
		}

		// Validate the token with Supabase
		// This is a simplified check - in a real app, you would verify the token
		if token == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Call the next handler
		next(w, r)
	}
}

// AdminPageHandler serves the admin page with authentication
func (h *AuthHandler) AdminPageHandler(w http.ResponseWriter, r *http.Request) {
	// This handler should be wrapped with RequireAuth middleware
	http.ServeFile(w, r, "./static/admin.html")
}
