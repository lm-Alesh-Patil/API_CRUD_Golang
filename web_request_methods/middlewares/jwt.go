package middlewares

import (
	"context"
	"net/http"
	"strings"
	"web_request_methods/handlers"

	"github.com/dgrijalva/jwt-go"
)

// JWTAuth middleware to protect routes with token validation
func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the token from the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// Return error if token is missing
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Split the Bearer token
		tokenString := strings.Split(authHeader, "Bearer ")[1]
		if tokenString == "" {
			// Return error if the token is malformed
			http.Error(w, "Malformed Authorization header", http.StatusUnauthorized)
			return
		}

		// Validate the token
		token, err := validateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Attach the claims to the context for use in the next handler
		claims, ok := token.Claims.(*handlers.Claims)
		if !ok || !token.Valid {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		// Set the claims in the context (you can also pass the claims into your handler functions)
		ctx := context.WithValue(r.Context(), "claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateToken function to validate the JWT token
func validateToken(tokenString string) (*jwt.Token, error) {
	// Specify the JWT secret key used for signing
	// Replace "your_secret_key" with the actual secret key
	return jwt.ParseWithClaims(tokenString, &handlers.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("alesh123"), nil
	})
}
