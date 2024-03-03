package dto

import "time"

type Tenant struct {
	ID         int32      `json:"id"`
	Name       string     `json:"name"`
	Mode       string     `json:"mode"`
	Properties Properties `json:"properties"`
	Status     string     `json:"status"`
	UpdatedBy  string     `json:"updated_by"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type Properties struct {
	Brokers string `json:"brokers"`
	Topic   string `json:"topic"`
}
