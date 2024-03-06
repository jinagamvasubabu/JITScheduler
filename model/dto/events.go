package dto

import "time"

type Event struct {
	ID        string    `json:"id"`
	TenantID  int32     `json:"tenant_id"`
	Type      string    `json:"type"`
	ProcessAt string    `json:"process_at"` //An RFC3339 formatted timestamp
	Status    string    `json:"status"`
	Payload   string    `json:"payload`
	UpdatedBy string    `json:"updated_by"`
	UpdatedAt time.Time `json:"updated_at"`
}
