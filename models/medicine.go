package models

import (
	"time"
)

type Medicine struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Specification string  `json:"specification" db:"specification"`
	Unit        string    `json:"unit" db:"unit"`
	Price       float64   `json:"price" db:"price"`
	Stock       int       `json:"stock" db:"stock"`
	MinStock    int       `json:"min_stock" db:"min_stock"`
	Category    string    `json:"category" db:"category"`
	Manufacturer string   `json:"manufacturer" db:"manufacturer"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type MedicineSearch struct {
	Name         string `json:"name"`
	Category     string `json:"category"`
	Manufacturer string `json:"manufacturer"`
} 