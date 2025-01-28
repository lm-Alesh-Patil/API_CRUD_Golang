package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"web_request_methods/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
)

var jwtKey = []byte("alesh123") // Replace with a secure key

type UserHandler struct {
	db *sql.DB
}

func NewUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Insert user into DB
	result, err := h.db.Exec("INSERT INTO users (name, email, password, mobile_no) VALUES (?, ?, ?, ?)",
		user.Name, user.Email, user.Password, user.MobileNo)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	user.ID = int(id)
	h.respondWithJSON(w, http.StatusCreated, user)
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	rows, err := h.db.Query("SELECT id, name, email, password, mobile_no, created_at FROM users")
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.MobileNo, &user.CreatedAt); err != nil {
			h.respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		users = append(users, user)
	}
	h.respondWithJSON(w, http.StatusOK, users)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var user models.User
	err := h.db.QueryRow("SELECT id, name, email, password, mobile_no, created_at FROM users WHERE id = ?", id).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.MobileNo, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.respondWithError(w, http.StatusNotFound, "User not found")
		} else {
			h.respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	h.respondWithJSON(w, http.StatusOK, user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var user models.User

	// Decode input and check for errors
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Check if user exists in DB
	var existingUser models.User
	err := h.db.QueryRow("SELECT id, name, email, password, mobile_no FROM users WHERE id = ?", id).
		Scan(&existingUser.ID, &existingUser.Name, &existingUser.Email, &existingUser.Password, &existingUser.MobileNo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.respondWithError(w, http.StatusNotFound, "User not found")
		} else {
			h.respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Assign default values if not provided in request
	if user.Name == "" {
		user.Name = existingUser.Name
	}
	if user.Email == "" {
		user.Email = existingUser.Email
	}
	if user.Password == "" {
		user.Password = existingUser.Password
	}
	if user.MobileNo == "" {
		user.MobileNo = existingUser.MobileNo
	}

	// Update the user in DB
	_, err = h.db.Exec("UPDATE users SET name = ?, email = ?, password = ?, mobile_no = ? WHERE id = ?",
		user.Name, user.Email, user.Password, user.MobileNo, id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	user.ID = id
	h.respondWithJSON(w, http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	result, err := h.db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		h.respondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	h.respondWithJSON(w, http.StatusNoContent, nil)
}

type Claims struct {
	Email     string `json:"email"`
	Timestamp int64  `json:"timestamp"`
	jwt.StandardClaims
}

func generateSSOToken(email string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token valid for 24 hours
	timestamp := time.Now().Unix()
	claims := &Claims{
		Email:     email,
		Timestamp: timestamp,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Issuer:    "Alesh123",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	var user models.User
	err := h.db.QueryRow("SELECT id, name, email, mobile_no FROM users WHERE email = ? AND password = ?",
		input.Email, input.Password).Scan(&user.ID, &user.Name, &user.Email, &user.MobileNo)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		} else {
			h.respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	tokenString, err := generateSSOToken(user.Email)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error generating SSO token")
		return
	}

	w.Header().Set("Authorization", "Bearer "+tokenString)
	h.respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Login successful",
		"token":   tokenString,
	})
}

func (h *UserHandler) authenticateSSO(w http.ResponseWriter, r *http.Request) (*models.User, error) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		h.respondWithError(w, http.StatusUnauthorized, "Missing token")
		return nil, fmt.Errorf("missing token")
	}

	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		h.respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		h.respondWithError(w, http.StatusUnauthorized, "Could not parse claims")
		return nil, fmt.Errorf("could not parse claims")
	}

	var user models.User
	err = h.db.QueryRow("SELECT id, name, email, mobile_no FROM users WHERE email = ?", claims.Email).
		Scan(&user.ID, &user.Name, &user.Email, &user.MobileNo)
	if err != nil {
		h.respondWithError(w, http.StatusUnauthorized, "User not found")
		return nil, err
	}

	return &user, nil
}

func (h *UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	user, err := h.authenticateSSO(w, r)
	if err != nil {
		return
	}
	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"id":        user.ID,
		"name":      user.Name,
		"email":     user.Email,
		"mobile_no": user.MobileNo,
		"timestamp": time.Now().Unix(),
	})
}

func (h *UserHandler) respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func (h *UserHandler) respondWithError(w http.ResponseWriter, status int, message string) {
	h.respondWithJSON(w, status, map[string]string{
		"error": message,
	})
}
