package repository

import (
	"context"
	"encoding/json"

	"github.com/Buddy-Git/JITScheduler-svc/model"
	"github.com/Buddy-Git/JITScheduler-svc/service/kafka"
)

type Manager struct {
	TenantRepository TenantRepository
	EventRepository  EventRepository
}

func NewManager(tenantRepository TenantRepository, eventRepository EventRepository) *Manager {
	return &Manager{
		TenantRepository: tenantRepository,
		EventRepository:  eventRepository,
	}
}

func (m *Manager) EventHandOff(ctx context.Context, event *model.Event) (err error) {
	var tenant *model.Tenant
	if tenant, err = m.TenantRepository.FetchTenantById(ctx, event.TenantID); err != nil {
		return err
	}
	var props map[string]string
	json.Unmarshal([]byte(tenant.Properties), &props)
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	producer, err := kafka.NewSyncProducer()
	if err != nil {
		return err
	}
	err = producer.PublishMessage(ctx, props["topic"], event.ID, payload)
	return err
}
