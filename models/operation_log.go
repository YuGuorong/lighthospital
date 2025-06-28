package models

import (
	"time"
)

type OperationLog struct {
	ID          int       `json:"id" db:"id"`
	UserID      int       `json:"user_id" db:"user_id"`
	Username    string    `json:"username" db:"username"`
	Action      string    `json:"action" db:"action"`
	Module      string    `json:"module" db:"module"`
	Description string    `json:"description" db:"description"`
	IP          string    `json:"ip" db:"ip"`
	UserAgent   string    `json:"user_agent" db:"user_agent"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	
	// 关联数据
	User        *User     `json:"user,omitempty"`
}

type OperationLogSearch struct {
	Username    string    `json:"username"`
	Action      string    `json:"action"`
	Module      string    `json:"module"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
} 