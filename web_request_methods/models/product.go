package models

type Product struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	AvailableStock int   `json:"available_stock"`
	Rating       float64 `json:"rating"`
}