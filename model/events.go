package model

import (
	"time"
)

var Status = NewEventStatus()

func NewEventStatus() *eventStatus {
	return &eventStatus{
		REQUESTED:  "REQUESTED",
		SCHEDULED:  "SCHEDULED",
		CANCELLED:  "CANCELLED",
		PROCESSING: "PROCESSING",
		FAILED:     "FAILED",
		COMPLETED:  "COMPLETED",
	}
}

type eventStatus struct {
	REQUESTED  string
	SCHEDULED  string
	CANCELLED  string
	PROCESSING string
	FAILED     string
	COMPLETED  string
}

type Event struct {
	ID           string    `json:"id" valid:"required" gorm:"primaryKey;autoIncrement:false"`
	TenantID     int32     `json:"tenant_id" valid:"required" gorm:"primaryKey;autoIncrement:false;not null"`
	Type         string    `json:"type" valid:"required" gorm:"not null;type:varchar(25);default:DELAY"`
	ProcessAt    time.Time `json:"process_at" valid:"required" gorm:"not null;type:timestamp" sql:"index"`
	CurrentQueue string    `json:"current_queue" gorm:"type:varchar(25)"`
	Payload      string    `json:"payload" gorm:"type:jsonb"`
	Status       string    `json:"status" valid:"required" gorm:"not null;default:REQUESTED"`
	UpdatedBy    string    `json:"updated_by" valid:"required" gorm:"null"`
	UpdatedAt    time.Time `json:"updated_at" valid:"required" gorm:"null"`
}

type QueuedEventsInfo struct {
	TenantID int32  `json:"tenant_id"`
	Count    int    `json:"count"`
	Mode     string `json:"mode"`
}
type UnProcessedEventsCount struct {
	Status string
	Count  int32
}
