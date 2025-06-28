package models

import (
	"time"
)

type Prescription struct {
	ID           int       `json:"id" db:"id"`
	PatientID    int       `json:"patient_id" db:"patient_id"`
	DoctorID     int       `json:"doctor_id" db:"doctor_id"`
	Diagnosis    string    `json:"diagnosis" db:"diagnosis"`
	DoctorAdvice string    `json:"doctor_advice" db:"doctor_advice"`
	TotalAmount  float64   `json:"total_amount" db:"total_amount"`
	Status       string    `json:"status" db:"status"` // draft, completed, printed
	Notes        string    `json:"notes" db:"notes"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`

	// 关联数据
	Patient *Patient           `json:"patient,omitempty"`
	Doctor  *User              `json:"doctor,omitempty"`
	Items   []PrescriptionItem `json:"items,omitempty"`
}

type PrescriptionItem struct {
	ID             int     `json:"id" db:"id"`
	PrescriptionID int     `json:"prescription_id" db:"prescription_id"`
	MedicineID     int     `json:"medicine_id" db:"medicine_id"`
	MedicineName   string  `json:"medicine_name" db:"medicine_name"`
	Specification  string  `json:"specification" db:"specification"`
	Dosage         string  `json:"dosage" db:"dosage"`
	Usage          string  `json:"usage" db:"usage"`
	Frequency      string  `json:"frequency" db:"frequency"`
	Days           int     `json:"days" db:"days"`
	Quantity       int     `json:"quantity" db:"quantity"`
	UnitPrice      float64 `json:"unit_price" db:"unit_price"`
	TotalPrice     float64 `json:"total_price" db:"total_price"`

	// 关联数据
	Medicine *Medicine `json:"medicine,omitempty"`
}

type PrescriptionSearch struct {
	PatientName string    `json:"patient_name"`
	DoctorName  string    `json:"doctor_name"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Status      string    `json:"status"`
}
