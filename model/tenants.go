package model

import (
	"time"
)

var TenantStatus = newTenantStatus()

func newTenantStatus() *tenantStatus {
	return &tenantStatus{
		ENABLED:  "ENABLED",
		DISABLED: "DISABLED",
	}
}

type tenantStatus struct {
	ENABLED  string
	DISABLED string
}

type Property struct {
	Brokers string `json:"brokers"`
	Topic   string `json:"topic"`
}

type Tenant struct {
	ID         int32     `json:"id" gorm:"primaryKey"`
	Name       string    `json:"name"  gorm:"not null;type:varchar(255)" sql:"unique_index"`
	Mode       string    `json:"mode"  gorm:"not null;type:varchar(50)"`
	Properties string    `json:"properties" gorm:"type:jsonb"`
	Status     string    `json:"status" gorm:"default:ENABLED"`
	UpdatedBy  string    `json:"updated_by" gorm:"default:null"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"default:null;type:timestamp"`
}
