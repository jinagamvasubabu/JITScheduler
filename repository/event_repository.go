package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Buddy-Git/JITScheduler-svc/adapters/logger"
	"github.com/Buddy-Git/JITScheduler-svc/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type eventRepository struct {
	db *gorm.DB
}

type EventRepository interface {
	AddEvent(ctx context.Context, event *model.Event) error
	DeactivateEvent(ctx context.Context, id string, TenantID int32, user string) error
	UpdateStatus(ctx context.Context, id string, TenantID int32, status string) error
	UpdateCurrentQueue(ctx context.Context, id string, TenantID int32, currentQueue string) error
	FetchEvent(ctx context.Context, eventId string, TenantID int32) (*model.Event, error)
	FetchAllEventsUnderTenant(ctx context.Context, TenantID int32) ([]*model.Event, error)
	GetUnProcessedEventsCount() ([]*model.UnProcessedEventsCount, error)
	GetUnProcessedEvents() ([]*model.Event, error)
	FetchEventsWithInNextXmins(x int32, status string) ([]*model.Event, error)
	FetchFailedEventsWithInXminutes(x int32, status string) ([]*model.Event, error)
	GetQueuedEventsCount() ([]*model.QueuedEventsInfo, error)
}

func NewEventRepository(ctx context.Context, db *gorm.DB) *eventRepository {
	return &eventRepository{
		db: db,
	}
}

func (e *eventRepository) AddEvent(ctx context.Context, event *model.Event) error {
	tenant := &model.Tenant{}
	if err := e.db.Where("id = ?", event.TenantID).First(tenant).Error; err != nil {
		logger.Error("Error while checking the tenant", zap.Error(err), zap.Any("tenantId", event.TenantID))
		return err
	}
	if tenant.Status == "DISABLED" {
		return errors.New("can't add event to a disabled tenant")
	}

	return e.db.Where(model.Event{ID: event.ID, TenantID: event.TenantID}).Assign(*event).FirstOrCreate(event).Error
}

func (e *eventRepository) DeactivateEvent(ctx context.Context, id string, TenantID int32, user string) error {
	existingEvent := &model.Event{}
	if db := e.db.Where("id = ? AND tenant_id = ?", id, TenantID).First(existingEvent).
		Updates(model.Event{Status: "CANCELLED", UpdatedBy: user, UpdatedAt: time.Now()}); db.Error != nil {
		return db.Error
	}
	return nil
}

func (e *eventRepository) UpdateStatus(ctx context.Context, id string, TenantID int32, status string) error {
	if db := e.db.Model(&model.Event{ID: id, TenantID: TenantID}).
		Updates(model.Event{Status: status, UpdatedAt: time.Now().UTC()}); db.Error != nil {
		return db.Error
	}
	return nil
}

func (e *eventRepository) UpdateCurrentQueue(ctx context.Context, id string, TenantID int32, currentQueue string) error {
	if db := e.db.Model(&model.Event{ID: id, TenantID: TenantID}).
		Updates(model.Event{CurrentQueue: currentQueue}); db.Error != nil {
		return db.Error
	}
	return nil
}

func (e *eventRepository) FetchEvent(ctx context.Context, eventId string, TenantID int32) (*model.Event, error) {
	event := &model.Event{}
	if db := e.db.Where("id = ? and tenant_id = ?", eventId, TenantID).First(event); db.Error != nil {
		return nil, db.Error
	}
	return event, nil
}

func (e *eventRepository) FetchAllEventsUnderTenant(ctx context.Context, TenantID int32) ([]*model.Event, error) {
	tenant := &model.Tenant{}
	if err := e.db.Where("id = ?", TenantID).First(tenant).Error; err != nil {
		return nil, err
	}
	var events []*model.Event
	if err := e.db.Where("tenant_id = ?", TenantID).Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (e *eventRepository) FetchEventsWithInNextXmins(x int32, status string) ([]*model.Event, error) {
	var events []*model.Event
	query := fmt.Sprintf("SELECT * FROM events WHERE status = 'REQUESTED' and process_at <= (NOW()::timestamp + interval '%d minutes') ORDER BY process_at asc", x)
	if err := e.db.
		Raw(query).Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (e *eventRepository) FetchFailedEventsWithInXminutes(x int32, status string) ([]*model.Event, error) {
	var events []*model.Event
	query := fmt.Sprintf("SELECT * FROM events WHERE status = 'FAILED' and (process_at <= (NOW()::timestamp + interval '%d minutes') and process_at >= (NOW()::timestamp - interval '%d minutes')) ORDER BY process_at asc", x, x)
	if err := e.db.
		Raw(query).Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (e *eventRepository) GetUnProcessedEventsCount() ([]*model.UnProcessedEventsCount, error) {
	var unProcessedEventsCount []*model.UnProcessedEventsCount
	err := e.db.
		Raw("SELECT status, count(*) from events WHERE status in (?, ?) and process_at < now() GROUP BY status", model.Status.REQUESTED, model.Status.SCHEDULED).
		Scan(&unProcessedEventsCount).Error
	return unProcessedEventsCount, err
}

func (e *eventRepository) GetUnProcessedEvents() ([]*model.Event, error) {
	var unProcessedEvents []*model.Event
	err := e.db.
		Raw("SELECT * from events where status in (?, ?) and process_at < now()", model.Status.REQUESTED, model.Status.SCHEDULED).
		Scan(&unProcessedEvents).Error
	return unProcessedEvents, err
}

func (e *eventRepository) GetQueuedEventsCount() ([]*model.QueuedEventsInfo, error) {
	var QueuedEventsInfo []*model.QueuedEventsInfo
	err := e.db.
		Raw("SELECT e.tenant_id, t.mode, count(e.id) FROM events as e INNER JOIN tenants as t ON e.tenant_id = t.id WHERE e.status = 'SCHEDULED' and e.process_at >= NOW() GROUP BY e.tenant_id, t.mode").
		Scan(&QueuedEventsInfo).Error
	return QueuedEventsInfo, err
}
