package models

import (
	"time"
)

type Patient struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Gender      string    `json:"gender" db:"gender"`
	Age         int       `json:"age" db:"age"`
	Phone       string    `json:"phone" db:"phone"`
	Address     string    `json:"address" db:"address"`
	IDCard      string    `json:"id_card" db:"id_card"`
	MedicalHistory string `json:"medical_history" db:"medical_history"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type PatientSearch struct {
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	IDCard string `json:"id_card"`
} 