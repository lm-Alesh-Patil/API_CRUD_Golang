package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"web_request_methods/models"
	"github.com/go-chi/chi/v5"
)

type ProductHandler struct {
    db *sql.DB
}

func NewProductHandler(db *sql.DB) *ProductHandler {
    return &ProductHandler{db: db}
}



func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Insert product into DB
	result, err := h.db.Exec("INSERT INTO products (name, price, available_stock, rating) VALUES (?, ?, ?, ?)",
		product.Name, product.Price, product.AvailableStock, product.Rating)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	product.ID = int(id)
	h.respondWithJSON(w, http.StatusCreated, product)
}

func (h *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	products := []models.Product{}
	rows, err := h.db.Query("SELECT id, name, price, available_stock, rating FROM products")
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.AvailableStock, &product.Rating); err != nil {
			h.respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		products = append(products, product)
	}
	h.respondWithJSON(w, http.StatusOK, products)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var product models.Product
	err := h.db.QueryRow("SELECT id, name, price, available_stock, rating FROM products WHERE id = ?", id).
		Scan(&product.ID, &product.Name, &product.Price, &product.AvailableStock, &product.Rating)
	if err != nil {
		if err == sql.ErrNoRows {
			h.respondWithError(w, http.StatusNotFound, "Product not found")
		} else {
			h.respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	h.respondWithJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var product models.Product

	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	_, err := h.db.Exec("UPDATE products SET name = ?, price = ?, available_stock = ?, rating = ? WHERE id = ?",
		product.Name, product.Price, product.AvailableStock, product.Rating, id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	product.ID = id
	h.respondWithJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	result, err := h.db.Exec("DELETE FROM products WHERE id = ?", id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		h.respondWithError(w, http.StatusNotFound, "Product not found")
		return
	}

	h.respondWithJSON(w, http.StatusNoContent, nil)
}

// Existing methods and utilities for users and authentication omitted for brevity

func (h *ProductHandler) respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func (h *ProductHandler) respondWithError(w http.ResponseWriter, status int, message string) {
	h.respondWithJSON(w, status, map[string]string{
		"error": message,
	})
}