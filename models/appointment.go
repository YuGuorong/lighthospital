package models

import (
	"time"
)

type Appointment struct {
	ID          int       `json:"id" db:"id"`
	PatientID   int       `json:"patient_id" db:"patient_id"`
	DoctorID    int       `json:"doctor_id" db:"doctor_id"`
	AppointmentTime time.Time `json:"appointment_time" db:"appointment_time"`
	Duration    int       `json:"duration" db:"duration"` // 分钟
	Status      string    `json:"status" db:"status"` // scheduled, completed, cancelled, no_show
	Notes       string    `json:"notes" db:"notes"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	
	// 关联数据
	Patient     *Patient `json:"patient,omitempty"`
	Doctor      *User    `json:"doctor,omitempty"`
}

type AppointmentSearch struct {
	PatientName string    `json:"patient_name"`
	DoctorName  string    `json:"doctor_name"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Status      string    `json:"status"`
} 